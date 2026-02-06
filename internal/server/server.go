package server

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"eroshit/internal/config"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/internal/simulator"
	"eroshit/pkg/metrics"
	"eroshit/pkg/useragent"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// SECURITY: Server start time for health endpoint
var serverStartTime = time.Now()

// SECURITY: Rate limiter - 100 requests per second with burst of 200
var apiLimiter = rate.NewLimiter(rate.Limit(100), 200)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	mu              sync.Mutex
	cfg             *config.Config
	sim             *simulator.Simulator
	cancel          context.CancelFunc
	agentLoader     *useragent.Loader
	proxyService    *proxy.Service
	hub             *Hub
	metrics         *metrics.MetricsCollector
	metricsWS       *MetricsWebSocket
}

// Hub WebSocket ve SSE abonelerine broadcast (status + log)
type Hub struct {
	mu       sync.RWMutex
	conns    map[*websocket.Conn]chan []byte
	logSubs  []chan string
}

func NewHub() *Hub {
	return &Hub{conns: make(map[*websocket.Conn]chan []byte)}
}

func (h *Hub) Register(conn *websocket.Conn) {
	ch := make(chan []byte, 128)
	h.mu.Lock()
	h.conns[conn] = ch
	h.mu.Unlock()
	go func() {
		for msg := range ch {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	if ch, ok := h.conns[conn]; ok {
		close(ch)
		delete(h.conns, conn)
	}
	h.mu.Unlock()
}

func (h *Hub) SubscribeLog() chan string {
	ch := make(chan string, 64)
	h.mu.Lock()
	h.logSubs = append(h.logSubs, ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) UnsubscribeLog(ch chan string) {
	h.mu.Lock()
	for i, c := range h.logSubs {
		if c == ch {
			h.logSubs = append(h.logSubs[:i], h.logSubs[i+1:]...)
			close(ch)
			break
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(typ string, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{"type": typ, "data": data})
	if err != nil {
		return
	}
	h.mu.RLock()
	for _, ch := range h.conns {
		select {
		case ch <- payload:
		default:
		}
	}
	if typ == "log" {
		if s, ok := data.(string); ok {
			for _, sub := range h.logSubs {
				select {
				case sub <- s:
				default:
				}
			}
		}
	}
	h.mu.RUnlock()
}

func New() (*Server, error) {
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	baseDirs := []string{wd, ".", "..", exeDir, filepath.Join(exeDir, ".."), filepath.Join(wd, "..")}
	agentLoader := useragent.LoadFromDirs(baseDirs)

	cfg, err := loadConfig(baseDirs)
	if err != nil {
		cfg = &config.Config{
			TargetDomain:    "",
			MaxPages:        5,
			DurationMinutes: 60,
			HitsPerMinute:   35,
			OutputDir:       "./reports",
			ExportFormat:    "both",
		}
		cfg.ApplyDefaults()
		cfg.ComputeDerived()
	}

	// Initialize metrics collector
	metricsCollector := metrics.GetGlobalCollector()

	s := &Server{
		cfg:          cfg,
		agentLoader:  agentLoader,
		proxyService: proxy.NewService(),
		hub:          NewHub(),
		metrics:      metricsCollector,
		metricsWS:    NewMetricsWebSocket(metricsCollector),
	}
	go s.broadcastStatusLoop()
	go s.metricsUpdateLoop()
	return s, nil
}

func (s *Server) broadcastStatusLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.hub.Broadcast("status", s.buildStatusMap())
	}
}

// metricsUpdateLoop periodically updates metrics from simulator state
func (s *Server) metricsUpdateLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.updateMetricsFromState()
	}
}

// updateMetricsFromState updates high-level metrics based on current simulator/proxy state.
// NOTE: Ayrıntılı hit/success istatistikleri doğrudan metrics collector tarafından tutulur.
func (s *Server) updateMetricsFromState() {
	s.mu.Lock()
	ps := s.proxyService
	sim := s.sim
	s.mu.Unlock()

	// Proxy metrikleri
	if ps != nil {
		st := ps.Status()
		s.metrics.SetActiveProxies(int64(st.LiveCount))
		s.metrics.SetQueueSize(int64(st.QueueCount))
	}

	// Oturum metrikleri (reporter'dan sadece aktiflik için faydalanıyoruz)
	if sim != nil {
		repMetrics := sim.Reporter().GetMetrics()
		if repMetrics.TotalHits > 0 {
			// TotalHits'i kaba bir aktif oturum proxysi olarak kullan
			s.metrics.SetActiveSessions(int64(repMetrics.TotalHits))
		}
	}
}

// RecordHit records a hit in metrics (called from simulator)
func (s *Server) RecordHit(url string, proxy string, duration time.Duration, success bool) {
	s.metrics.RecordHit()
	s.metrics.RecordResponseTime(duration)
	if proxy != "" {
		s.metrics.RecordProxyLatency(proxy, duration)
	}
	if success {
		s.metrics.RecordSuccess(proxy)
	} else {
		s.metrics.RecordFailure(proxy)
	}

	// Broadcast via WebSocket
	if s.metricsWS != nil {
		s.metricsWS.BroadcastHit(HitEvent{
			URL:          url,
			Proxy:        proxy,
			ResponseTime: duration,
			Success:      success,
		})
	}
}

// RecordSessionEvent records a session event
func (s *Server) RecordSessionEvent(sessionID string, action string, duration time.Duration, pages int) {
	if s.metricsWS != nil {
		s.metricsWS.BroadcastSession(SessionEvent{
			SessionID:    sessionID,
			Action:       action,
			Duration:     duration,
			PagesVisited: pages,
		})
	}

	switch action {
	case "started":
		s.metrics.SetActiveSessions(s.metrics.GetSnapshot().ActiveSessions + 1)
	case "ended":
		// CODE FIX: Use renamed maxInt64 function
		s.metrics.SetActiveSessions(maxInt64(0, s.metrics.GetSnapshot().ActiveSessions-1))
	case "bounce":
		s.metrics.RecordBounce()
	}
}

// RecordProxyStatus records proxy status change
func (s *Server) RecordProxyStatus(proxy string, status string, latency time.Duration, failCount int) {
	if s.metricsWS != nil {
		s.metricsWS.BroadcastProxyStatus(ProxyStatusEvent{
			Proxy:     proxy,
			Status:    status,
			Latency:   float64(latency.Milliseconds()),
			FailCount: failCount,
		})
	}
}

// maxInt64 returns the larger of two int64 values
// CODE FIX: Renamed from 'max' to avoid conflict with Go 1.21+ built-in max function
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

type configFile struct {
	PROXY_HOST             string   `json:"PROXY_HOST"`
	PROXY_PORT             int      `json:"PROXY_PORT"`
	PROXY_USER             string   `json:"PROXY_USER"`
	PROXY_PASS             string   `json:"PROXY_PASS"`
	TargetDomain           string   `json:"targetDomain"`
	FallbackGAID           string   `json:"fallbackGAID"`
	MaxPages               int      `json:"maxPages"`
	DurationMinutes        int      `json:"durationMinutes"`
	HitsPerMinute          int      `json:"hitsPerMinute"`
	MaxConcurrentVisits    int      `json:"maxConcurrentVisits"`
	OutputDir              string   `json:"outputDir"`
	ExportFormat           string   `json:"exportFormat"`
	CanvasFingerprint      bool     `json:"canvasFingerprint"`
	ScrollStrategy         string   `json:"scrollStrategy"`
	SendScrollEvent        bool     `json:"sendScrollEvent"`
	UseSitemap             bool     `json:"useSitemap"`
	SitemapHomepageWeight  int      `json:"sitemapHomepageWeight"`
	Keywords               []string `json:"keywords"`
	UsePublicProxy         bool     `json:"usePublicProxy"`
	ProxySourceURLs        []string `json:"proxySourceURLs"`
	GitHubRepos            []string `json:"githubRepos"`
	CheckerWorkers         int      `json:"checkerWorkers"`
	// Private proxy alanları
	PrivateProxies    []privateProxyFile `json:"privateProxies"`
	UsePrivateProxy   bool               `json:"usePrivateProxy"`
	// Yeni alanlar
	DeviceType        string   `json:"deviceType"`
	DeviceBrands      []string `json:"deviceBrands"`
	ReferrerKeyword   string   `json:"referrerKeyword"`
	ReferrerEnabled   bool     `json:"referrerEnabled"`
}

type privateProxyFile struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Protocol string `json:"protocol"`
}

func saveConfigToFile(cfg *config.Config) {
	// SECURITY FIX: Determine config file location - prioritize exe directory, then working directory
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	
	// Priority order: exe directory > working directory > current directory
	paths := []string{
		filepath.Join(exeDir, "config.json"),
		filepath.Join(wd, "config.json"),
		"config.json",
	}
	
	// Private proxy'leri dönüştür
	var privateProxies []privateProxyFile
	for _, pp := range cfg.PrivateProxies {
		privateProxies = append(privateProxies, privateProxyFile{
			Host:     pp.Host,
			Port:     pp.Port,
			User:     pp.User,
			Pass:     pp.Pass,
			Protocol: pp.Protocol,
		})
	}
	
	// SECURITY FIX: Save config to all possible locations, log success/failure
	var savedPath string
	var saveErr error
	
	for _, p := range paths {
		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		
		data, err := json.MarshalIndent(configFile{
			PROXY_HOST:            cfg.ProxyHost,
			PROXY_PORT:            cfg.ProxyPort,
			PROXY_USER:            cfg.ProxyUser,
			PROXY_PASS:            cfg.ProxyPass,
			TargetDomain:          cfg.TargetDomain,
			FallbackGAID:          cfg.GtagID,
			MaxPages:              cfg.MaxPages,
			DurationMinutes:       cfg.DurationMinutes,
			HitsPerMinute:         cfg.HitsPerMinute,
			MaxConcurrentVisits:   cfg.MaxConcurrentVisits,
			OutputDir:             cfg.OutputDir,
			ExportFormat:          cfg.ExportFormat,
			CanvasFingerprint:     cfg.CanvasFingerprint,
			ScrollStrategy:        cfg.ScrollStrategy,
			SendScrollEvent:       cfg.SendScrollEvent,
			UseSitemap:            cfg.UseSitemap,
			SitemapHomepageWeight: cfg.SitemapHomepageWeight,
			Keywords:              cfg.Keywords,
			UsePublicProxy:        cfg.UsePublicProxy,
			ProxySourceURLs:       cfg.ProxySourceURLs,
			GitHubRepos:           cfg.GitHubRepos,
			CheckerWorkers:        cfg.CheckerWorkers,
			// Private proxy alanları
			PrivateProxies:    privateProxies,
			UsePrivateProxy:   cfg.UsePrivateProxy,
			// Yeni alanlar
			DeviceType:        cfg.DeviceType,
			DeviceBrands:      cfg.DeviceBrands,
			ReferrerKeyword:   cfg.ReferrerKeyword,
			ReferrerEnabled:   cfg.ReferrerEnabled,
		}, "", "  ")
		if err != nil {
			saveErr = err
			continue
		}
		
		if err := os.WriteFile(p, data, 0644); err == nil {
			savedPath = p
			log.Printf("[INFO] Config kaydedildi: %s", p)
			// Save to first successful location, but also try others for redundancy
			break
		} else {
			saveErr = err
		}
	}
	
	if savedPath == "" {
		log.Printf("[ERROR] Config kaydedilemedi: %v", saveErr)
	}
}

func loadConfig(dirs []string) (*config.Config, error) {
	// SECURITY FIX: Priority order: exe directory > working directory > provided dirs
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	
	// Priority order: exe directory > working directory > provided dirs
	allDirs := make([]string, 0, len(dirs)+2)
	if exeDir != "" {
		allDirs = append(allDirs, exeDir)
	}
	if wd != "" && wd != exeDir {
		allDirs = append(allDirs, wd)
	}
	allDirs = append(allDirs, dirs...)
	
	for _, d := range allDirs {
		if d == "" {
			continue
		}
		p := filepath.Join(d, "config.json")
		if _, err := os.Stat(p); err == nil {
			log.Printf("[INFO] Config yüklendi: %s", p)
			cfg, err := config.LoadFromJSON(p)
			if err != nil {
				log.Printf("[ERROR] Config parse hatası (%s): %v", p, err)
				continue
			}
			return cfg, nil
		}
	}
	return nil, fmt.Errorf("config.json bulunamadı")
}

// SECURITY: Rate limiting middleware
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !apiLimiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// SECURITY: Health endpoint for monitoring
	mux.HandleFunc("/health", s.handleHealth)
	
	// API endpoints with rate limiting
	mux.HandleFunc("/api/config", rateLimitMiddleware(s.handleConfig))
	mux.HandleFunc("/api/start", rateLimitMiddleware(s.handleStart))
	mux.HandleFunc("/api/stop", rateLimitMiddleware(s.handleStop))
	mux.HandleFunc("/api/status", rateLimitMiddleware(s.handleStatus))
	mux.HandleFunc("/api/logs", rateLimitMiddleware(s.handleLogs))
	mux.HandleFunc("/api/ws", s.handleWebSocket) // WebSocket has its own handling
	mux.HandleFunc("/api/proxy/fetch", rateLimitMiddleware(s.handleProxyFetch))
	mux.HandleFunc("/api/proxy/status", rateLimitMiddleware(s.handleProxyStatus))
	mux.HandleFunc("/api/proxy/live", rateLimitMiddleware(s.handleProxyLive))
	mux.HandleFunc("/api/proxy/export", rateLimitMiddleware(s.handleProxyExport))
	mux.HandleFunc("/api/proxy/test", rateLimitMiddleware(s.handleProxyTest))
	mux.HandleFunc("/api/gsc/queries", rateLimitMiddleware(s.handleGSCQueries))

	// Metrics endpoints
	mux.HandleFunc("/api/metrics", MetricsHandler(s.metrics))               // Prometheus format
	mux.HandleFunc("/api/metrics/json", rateLimitMiddleware(MetricsJSONHandler(s.metrics))) // JSON format
	mux.HandleFunc("/api/metrics/stream", s.metricsWS.HandleWebSocket)      // Real-time WebSocket stream
	mux.HandleFunc("/api/metrics/dashboard", rateLimitMiddleware(DashboardHandler()))       // Grafana dashboard JSON
	
	// System Optimization endpoints
	mux.HandleFunc("/api/system/info", rateLimitMiddleware(s.handleSystemInfo))
	mux.HandleFunc("/api/system/optimize", rateLimitMiddleware(s.handleSystemOptimize))
	
	// Network Optimization endpoints
	mux.HandleFunc("/api/network/config", rateLimitMiddleware(s.handleNetworkConfig))
	
	// VM Spoofing endpoints
	mux.HandleFunc("/api/vm/status", rateLimitMiddleware(s.handleVMStatus))
	mux.HandleFunc("/api/vm/score", rateLimitMiddleware(s.handleVMScore))

	return mux
}

// SECURITY: Health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.mu.Lock()
	running := s.cancel != nil
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"uptime":     time.Since(serverStartTime).String(),
		"running":    running,
		"version":    "1.0.0",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		s.mu.Lock()
		cfg := s.cfg
		s.mu.Unlock()
		
		// Private proxy'leri API formatına dönüştür
		var privateProxiesAPI []map[string]interface{}
		for _, pp := range cfg.PrivateProxies {
			privateProxiesAPI = append(privateProxiesAPI, map[string]interface{}{
				"host":     pp.Host,
				"port":     pp.Port,
				"user":     pp.User,
				"pass":     pp.Pass,
				"protocol": pp.Protocol,
			})
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"target_domain":          cfg.TargetDomain,
			"max_pages":              cfg.MaxPages,
			"duration_minutes":       cfg.DurationMinutes,
			"hits_per_minute":        cfg.HitsPerMinute,
			"max_concurrent_visits":  cfg.MaxConcurrentVisits,
			"output_dir":             cfg.OutputDir,
			"export_format":          cfg.ExportFormat,
			"canvas_fingerprint":     cfg.CanvasFingerprint,
			"scroll_strategy":        cfg.ScrollStrategy,
			"send_scroll_event":      cfg.SendScrollEvent,
			"use_sitemap":            cfg.UseSitemap,
			"sitemap_homepage_weight": cfg.SitemapHomepageWeight,
			"keywords":               cfg.Keywords,
			"proxy_host":             cfg.ProxyHost,
			"proxy_port":             cfg.ProxyPort,
			"proxy_user":             cfg.ProxyUser,
			"proxy_pass":             cfg.ProxyPass,
			"gtag_id":                cfg.GtagID,
			"use_public_proxy":       cfg.UsePublicProxy,
			"proxy_source_urls":      cfg.ProxySourceURLs,
			"github_repos":           cfg.GitHubRepos,
			"checker_workers":        cfg.CheckerWorkers,
			// Private proxy alanları
			"private_proxies":        privateProxiesAPI,
			"use_private_proxy":      cfg.UsePrivateProxy,
			// Device & Referrer
			"device_type":            cfg.DeviceType,
			"device_brands":          cfg.DeviceBrands,
			"referrer_keyword":       cfg.ReferrerKeyword,
			"referrer_enabled":       cfg.ReferrerEnabled,
			"referrer_source":        cfg.ReferrerSource,
			// Traffic Simulation
			"min_page_duration":      cfg.MinPageDuration,
			"max_page_duration":      cfg.MaxPageDuration,
			"min_scroll_percent":     cfg.MinScrollPercent,
			"max_scroll_percent":     cfg.MaxScrollPercent,
			"click_probability":      cfg.ClickProbability,
			// Session Depth
			"session_min_pages":      cfg.SessionMinPages,
			"session_max_pages":      cfg.SessionMaxPages,
			"enable_session_depth":   cfg.EnableSessionDepth,
			// Bounce Rate
			"target_bounce_rate":     cfg.TargetBounceRate,
			"enable_bounce_control":  cfg.EnableBounceControl,
			// Behavior Simulation
			"simulate_mouse_move":    cfg.SimulateMouseMove,
			"simulate_keyboard":      cfg.SimulateKeyboard,
			"simulate_clicks":        cfg.SimulateClicks,
			"simulate_focus":         cfg.SimulateFocus,
			// Geo Location
			"geo_country":            cfg.GeoCountry,
			"geo_timezone":           cfg.GeoTimezone,
			"geo_language":           cfg.GeoLanguage,
			// Analytics Events
			"send_page_view":         cfg.SendPageView,
			"send_session_start":     cfg.SendSessionStart,
			"send_user_engagement":   cfg.SendUserEngagement,
			"send_first_visit":       cfg.SendFirstVisit,
			// Custom Dimensions
			"custom_dimensions":      cfg.CustomDimensions,
			"custom_metrics":         cfg.CustomMetrics,
			"enable_custom_dimensions": cfg.EnableCustomDimensions,
			// GSC Integration
			"gsc_property_url":       cfg.GscPropertyUrl,
			"enable_gsc_integration": cfg.EnableGscIntegration,
			"use_gsc_queries":        cfg.UseGscQueries,
			// Returning Visitor
			"returning_visitor_rate": cfg.ReturningVisitorRate,
			"returning_visitor_days": cfg.ReturningVisitorDays,
			"enable_returning_visitor": cfg.EnableReturningVisitor,
			// Exit Page
			"exit_pages":             cfg.ExitPages,
			"enable_exit_page_control": cfg.EnableExitPageControl,
			// Browser Profile
			"browser_profile_path":   cfg.BrowserProfilePath,
			"max_browser_profiles":   cfg.MaxBrowserProfiles,
			"enable_browser_profile": cfg.EnableBrowserProfile,
			"persist_cookies":        cfg.PersistCookies,
			"persist_local_storage":  cfg.PersistLocalStorage,
			// TLS Fingerprint
			"tls_fingerprint_mode":   cfg.TlsFingerprintMode,
			"enable_ja3_randomization": cfg.EnableJa3Randomization,
			"enable_ja4_randomization": cfg.EnableJa4Randomization,
			// Proxy Rotation
			"proxy_rotation_mode":    cfg.ProxyRotationMode,
			"proxy_rotation_interval": cfg.ProxyRotationInterval,
			"enable_proxy_rotation":  cfg.EnableProxyRotation,
			// HTTP/2 Fingerprint
			"http2_fingerprint_mode": cfg.Http2FingerprintMode,
			"enable_http2_fingerprint": cfg.EnableHttp2Fingerprint,
			"enable_http3_fingerprint": cfg.EnableHttp3Fingerprint,
			// Client Hints
			"enable_client_hints":    cfg.EnableClientHints,
			"spoof_sec_ch_ua":        cfg.SpoofSecChUa,
			"spoof_sec_ch_ua_platform": cfg.SpoofSecChUaPlatform,
			"spoof_sec_ch_ua_arch":   cfg.SpoofSecChUaArch,
			// Headless Bypass
			"bypass_puppeteer":       cfg.BypassPuppeteer,
			"bypass_playwright":      cfg.BypassPlaywright,
			"bypass_selenium":        cfg.BypassSelenium,
			"bypass_cdp":             cfg.BypassCDP,
			"bypass_rendering_detection": cfg.BypassRenderingDetection,
			"bypass_webdriver":       cfg.BypassWebdriver,
			// SERP
			"serp_search_engine":     cfg.SerpSearchEngine,
			"enable_serp_simulation": cfg.EnableSerpSimulation,
			"serp_scroll_before_click": cfg.SerpScrollBeforeClick,
			// Browser Pool
			"browser_pool_min":       cfg.BrowserPoolMin,
			"browser_pool_max":       cfg.BrowserPoolMax,
			"enable_auto_scaling":    cfg.EnableAutoScaling,
			"worker_queue_size":      cfg.WorkerQueueSize,
			"enable_priority_queue":  cfg.EnablePriorityQueue,
			"enable_failure_recovery": cfg.EnableFailureRecovery,
			// Mobile Emulation
			"ios_device_model":       cfg.IosDeviceModel,
			"enable_ios_safari":      cfg.EnableIosSafari,
			"enable_ios_haptics":     cfg.EnableIosHaptics,
			"android_device_model":   cfg.AndroidDeviceModel,
			"enable_android_chrome":  cfg.EnableAndroidChrome,
			"enable_android_vibration": cfg.EnableAndroidVibration,
			// Touch Events
			"enable_touch_events":    cfg.EnableTouchEvents,
			"enable_multi_touch":     cfg.EnableMultiTouch,
			"enable_gestures":        cfg.EnableGestures,
			"enable_mobile_keyboard": cfg.EnableMobileKeyboard,
			// Sensors
			"enable_accelerometer":   cfg.EnableAccelerometer,
			"enable_gyroscope":       cfg.EnableGyroscope,
			"enable_device_orientation": cfg.EnableDeviceOrientation,
			"enable_magnetometer":    cfg.EnableMagnetometer,
			// Tablet
			"enable_tablet_mode":     cfg.EnableTabletMode,
			"enable_landscape_mode":  cfg.EnableLandscapeMode,
			"enable_pen_input":       cfg.EnablePenInput,
			"enable_split_view":      cfg.EnableSplitView,
			// Stealth
			"stealth_webdriver":      cfg.StealthWebdriver,
			"stealth_chrome":         cfg.StealthChrome,
			"stealth_plugins":        cfg.StealthPlugins,
			"stealth_webgl":          cfg.StealthWebGL,
			"stealth_audio":          cfg.StealthAudio,
			"stealth_canvas":         cfg.StealthCanvas,
			"stealth_timezone":       cfg.StealthTimezone,
			"stealth_language":       cfg.StealthLanguage,
			// Performance
			"visit_timeout":          cfg.VisitTimeout,
			"page_load_wait":         cfg.PageLoadWait,
			"retry_count":            cfg.RetryCount,
			// Resource Blocking
			"block_images":           cfg.BlockImages,
			"block_styles":           cfg.BlockStyles,
			"block_fonts":            cfg.BlockFonts,
			"block_media":            cfg.BlockMedia,
			// Anti-Detect Mode
			"anti_detect_mode":       cfg.AntiDetectMode,
		})
		return
	}
	if r.Method == http.MethodPost {
		var body struct {
			// Basic Settings
			TargetDomain          string   `json:"target_domain"`
			MaxPages              int      `json:"max_pages"`
			DurationMinutes       int      `json:"duration_minutes"`
			HitsPerMinute         int      `json:"hits_per_minute"`
			MaxConcurrentVisits   int      `json:"max_concurrent_visits"`
			OutputDir             string   `json:"output_dir"`
			ExportFormat          string   `json:"export_format"`
			CanvasFingerprint     bool     `json:"canvas_fingerprint"`
			ScrollStrategy        string   `json:"scroll_strategy"`
			SendScrollEvent       bool     `json:"send_scroll_event"`
			UseSitemap            bool     `json:"use_sitemap"`
			SitemapHomepageWeight int      `json:"sitemap_homepage_weight"`
			Keywords              []string `json:"keywords"`
			GtagID                string   `json:"gtag_id"`
			AntiDetectMode        bool     `json:"anti_detect_mode"`
			
			// Device & Traffic
			DeviceType        string   `json:"device_type"`
			DeviceBrands      []string `json:"device_brands"`
			MinPageDuration   int      `json:"min_page_duration"`
			MaxPageDuration   int      `json:"max_page_duration"`
			
			// Session & Bounce
			EnableSessionDepth   bool `json:"enable_session_depth"`
			SessionMinPages      int  `json:"session_min_pages"`
			SessionMaxPages      int  `json:"session_max_pages"`
			EnableBounceControl  bool `json:"enable_bounce_control"`
			TargetBounceRate     int  `json:"target_bounce_rate"`
			
			// Behavior Simulation
			SimulateMouseMove   bool `json:"simulate_mouse_move"`
			SimulateKeyboard    bool `json:"simulate_keyboard"`
			SimulateClicks      bool `json:"simulate_clicks"`
			SimulateFocus       bool `json:"simulate_focus"`
			
			// Referrer
			ReferrerEnabled   bool   `json:"referrer_enabled"`
			ReferrerSource    string `json:"referrer_source"`
			ReferrerKeyword   string `json:"referrer_keyword"`
			
			// Geo
			GeoCountry   string `json:"geo_country"`
			GeoLanguage  string `json:"geo_language"`
			GeoTimezone  string `json:"geo_timezone"`
			
			// Analytics Events
			SendPageView        bool `json:"send_page_view"`
			SendSessionStart    bool `json:"send_session_start"`
			SendUserEngagement  bool `json:"send_user_engagement"`
			SendFirstVisit      bool `json:"send_first_visit"`
			
			// GSC
			EnableGscIntegration  bool   `json:"enable_gsc_integration"`
			UseGscQueries         bool   `json:"use_gsc_queries"`
			GscPropertyUrl        string `json:"gsc_property_url"`
			GscApiKey             string `json:"gsc_api_key"`
			
			// Browser Profile
			EnableBrowserProfile  bool   `json:"enable_browser_profile"`
			BrowserProfilePath    string `json:"browser_profile_path"`
			MaxBrowserProfiles    int    `json:"max_browser_profiles"`
			PersistCookies        bool   `json:"persist_cookies"`
			PersistLocalStorage   bool   `json:"persist_local_storage"`
			
			// Returning Visitor
			EnableReturningVisitor  bool `json:"enable_returning_visitor"`
			ReturningVisitorRate    int  `json:"returning_visitor_rate"`
			ReturningVisitorDays    int  `json:"returning_visitor_days"`
			
			// Network
			EnableHTTP3           bool `json:"enable_http3"`
			EnableConnectionPool  bool `json:"enable_connection_pool"`
			EnableTCPFastOpen     bool `json:"enable_tcp_fast_open"`
			MaxIdleConns          int  `json:"max_idle_conns"`
			MaxConnsPerHost       int  `json:"max_conns_per_host"`
			
			// System
			EnableCPUAffinity  bool `json:"enable_cpu_affinity"`
			EnableNUMA         bool `json:"enable_numa"`
			
			// VM Spoofing
			EnableVMSpoofing    bool   `json:"enable_vm_spoofing"`
			HideVMIndicators    bool   `json:"hide_vm_indicators"`
			SpoofHardwareIDs    bool   `json:"spoof_hardware_ids"`
			RandomizeVMParams   bool   `json:"randomize_vm_params"`
			VMType              string `json:"vm_type"`
			
			// Proxy
			UseProxy       bool   `json:"use_proxy"`
			ProxyHost      string `json:"proxy_host"`
			ProxyPort      int    `json:"proxy_port"`
			ProxyUser      string `json:"proxy_user"`
			ProxyPass      string `json:"proxy_pass"`
			UsePublicProxy bool   `json:"use_public_proxy"`
			ProxySourceURLs []string `json:"proxy_source_urls"`
			GitHubRepos     []string `json:"github_repos"`
			CheckerWorkers  int      `json:"checker_workers"`
			
			// Private proxy
			UsePrivateProxy   bool     `json:"use_private_proxy"`
			PrivateProxies    []struct {
				Host     string `json:"host"`
				Port     int    `json:"port"`
				User     string `json:"user"`
				Pass     string `json:"pass"`
				Protocol string `json:"protocol"`
			} `json:"private_proxies"`
			
			// Proxy List (textarea'dan gelen)
			ProxyList string `json:"proxy_list"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Printf("[ERROR] Config decode error: %v", err)
			http.Error(w, "Invalid JSON: "+err.Error(), 400)
			return
		}
		s.mu.Lock()
		// Basic Settings
		s.cfg.TargetDomain = body.TargetDomain
		s.cfg.MaxPages = body.MaxPages
		s.cfg.DurationMinutes = body.DurationMinutes
		s.cfg.HitsPerMinute = body.HitsPerMinute
		s.cfg.MaxConcurrentVisits = body.MaxConcurrentVisits
		s.cfg.OutputDir = body.OutputDir
		s.cfg.ExportFormat = body.ExportFormat
		s.cfg.CanvasFingerprint = body.CanvasFingerprint
		s.cfg.ScrollStrategy = body.ScrollStrategy
		s.cfg.SendScrollEvent = body.SendScrollEvent
		s.cfg.UseSitemap = body.UseSitemap
		s.cfg.SitemapHomepageWeight = body.SitemapHomepageWeight
		s.cfg.Keywords = body.Keywords
		s.cfg.GtagID = body.GtagID
		s.cfg.AntiDetectMode = body.AntiDetectMode
		
		// Device & Traffic
		s.cfg.DeviceType = body.DeviceType
		s.cfg.DeviceBrands = body.DeviceBrands
		s.cfg.MinPageDuration = body.MinPageDuration
		s.cfg.MaxPageDuration = body.MaxPageDuration
		
		// Session & Bounce
		s.cfg.EnableSessionDepth = body.EnableSessionDepth
		s.cfg.SessionMinPages = body.SessionMinPages
		s.cfg.SessionMaxPages = body.SessionMaxPages
		s.cfg.EnableBounceControl = body.EnableBounceControl
		s.cfg.TargetBounceRate = body.TargetBounceRate
		
		// Behavior Simulation
		s.cfg.SimulateMouseMove = body.SimulateMouseMove
		s.cfg.SimulateKeyboard = body.SimulateKeyboard
		s.cfg.SimulateClicks = body.SimulateClicks
		s.cfg.SimulateFocus = body.SimulateFocus
		
		// Referrer
		s.cfg.ReferrerEnabled = body.ReferrerEnabled
		s.cfg.ReferrerSource = body.ReferrerSource
		s.cfg.ReferrerKeyword = body.ReferrerKeyword
		
		// Geo
		s.cfg.GeoCountry = body.GeoCountry
		s.cfg.GeoLanguage = body.GeoLanguage
		s.cfg.GeoTimezone = body.GeoTimezone
		
		// Analytics Events
		s.cfg.SendPageView = body.SendPageView
		s.cfg.SendSessionStart = body.SendSessionStart
		s.cfg.SendUserEngagement = body.SendUserEngagement
		s.cfg.SendFirstVisit = body.SendFirstVisit
		
		// GSC
		s.cfg.EnableGscIntegration = body.EnableGscIntegration
		s.cfg.UseGscQueries = body.UseGscQueries
		s.cfg.GscPropertyUrl = body.GscPropertyUrl
		s.cfg.GscApiKey = body.GscApiKey
		
		// Browser Profile
		s.cfg.EnableBrowserProfile = body.EnableBrowserProfile
		s.cfg.BrowserProfilePath = body.BrowserProfilePath
		s.cfg.MaxBrowserProfiles = body.MaxBrowserProfiles
		s.cfg.PersistCookies = body.PersistCookies
		s.cfg.PersistLocalStorage = body.PersistLocalStorage
		
		// Returning Visitor
		s.cfg.EnableReturningVisitor = body.EnableReturningVisitor
		s.cfg.ReturningVisitorRate = body.ReturningVisitorRate
		s.cfg.ReturningVisitorDays = body.ReturningVisitorDays
		
		// Network
		s.cfg.EnableHTTP3 = body.EnableHTTP3
		s.cfg.EnableConnectionPool = body.EnableConnectionPool
		s.cfg.EnableTCPFastOpen = body.EnableTCPFastOpen
		s.cfg.MaxIdleConns = body.MaxIdleConns
		s.cfg.MaxConnsPerHost = body.MaxConnsPerHost
		
		// System
		s.cfg.EnableCPUAffinity = body.EnableCPUAffinity
		s.cfg.EnableNUMA = body.EnableNUMA
		
		// VM Spoofing
		s.cfg.EnableVMSpoofing = body.EnableVMSpoofing
		s.cfg.HideVMIndicators = body.HideVMIndicators
		s.cfg.SpoofHardwareIDs = body.SpoofHardwareIDs
		s.cfg.RandomizeVMParams = body.RandomizeVMParams
		s.cfg.VMType = body.VMType
		
		// Proxy
		s.cfg.UseProxy = body.UseProxy
		s.cfg.ProxyHost = body.ProxyHost
		s.cfg.ProxyPort = body.ProxyPort
		s.cfg.ProxyUser = body.ProxyUser
		s.cfg.ProxyPass = body.ProxyPass
		s.cfg.UsePublicProxy = body.UsePublicProxy
		s.cfg.ProxySourceURLs = body.ProxySourceURLs
		if body.GitHubRepos != nil {
			s.cfg.GitHubRepos = body.GitHubRepos
		}
		if body.CheckerWorkers > 0 {
			s.cfg.CheckerWorkers = body.CheckerWorkers
		}
		
		// Private proxy'leri config'e kaydet
		s.cfg.UsePrivateProxy = body.UsePrivateProxy
		s.cfg.PrivateProxies = nil // Önce temizle
		for _, pp := range body.PrivateProxies {
			if pp.Host != "" && pp.Port > 0 {
				protocol := pp.Protocol
				if protocol == "" {
					protocol = "http"
				}
				s.cfg.PrivateProxies = append(s.cfg.PrivateProxies, config.PrivateProxy{
					Host:     pp.Host,
					Port:     pp.Port,
					User:     pp.User,
					Pass:     pp.Pass,
					Protocol: protocol,
				})
			}
		}
		
		// Private proxy varsa UsePrivateProxy'yi otomatik aktifleştir
		if len(s.cfg.PrivateProxies) > 0 {
			s.cfg.UsePrivateProxy = true
		}
		
		// Proxy List parsing (textarea'dan gelen)
		if body.ProxyList != "" {
			// Satır satır parse et ve private_proxies'e ekle
			lines := strings.Split(body.ProxyList, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				
				// Parse proxy URL: protocol://[user:pass@]host:port
				proxyURL, err := url.Parse(line)
				if err != nil {
					log.Printf("[WARN] Invalid proxy URL skipped: %s - %v", line, err)
					continue
				}
				
				host := proxyURL.Hostname()
				portStr := proxyURL.Port()
				port := 0
				if portStr != "" {
					if p, err := strconv.Atoi(portStr); err == nil {
						port = p
					}
				}
				
				if host == "" || port == 0 {
					log.Printf("[WARN] Proxy missing host or port, skipped: %s", line)
					continue
				}
				
				protocol := proxyURL.Scheme
				if protocol == "" {
					protocol = "http"
				}
				
				user := ""
				pass := ""
				if proxyURL.User != nil {
					user = proxyURL.User.Username()
					if p, ok := proxyURL.User.Password(); ok {
						pass = p
					}
				}
				
				s.cfg.PrivateProxies = append(s.cfg.PrivateProxies, config.PrivateProxy{
					Host:     host,
					Port:     port,
					User:     user,
					Pass:     pass,
					Protocol: protocol,
				})
				
				log.Printf("[INFO] Added proxy: %s://%s:%d", protocol, host, port)
			}
			
			// Proxy varsa UsePrivateProxy ve UseProxy'yi aktifleştir
			if len(s.cfg.PrivateProxies) > 0 {
				s.cfg.UsePrivateProxy = true
				s.cfg.UseProxy = true
				s.cfg.ProxyEnabled = true
			}
		}
		
		s.cfg.ApplyDefaults()
		s.cfg.ComputeDerived()
		s.mu.Unlock()
		saveConfigToFile(s.cfg)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	http.Error(w, "Method not allowed", 405)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		http.Error(w, "Simülasyon zaten çalışıyor", 400)
		return
	}
	domain := s.cfg.TargetDomain
	if domain == "" {
		s.mu.Unlock()
		http.Error(w, "Lütfen hedef domain girin", 400)
		return
	}

	// İsteğe bağlı lang (client'tan gelen seçim)
	locale := "tr"
	if body, err := io.ReadAll(r.Body); err == nil && len(body) > 0 {
		var req struct {
			Lang string `json:"lang"`
		}
		if json.Unmarshal(body, &req) == nil && (req.Lang == "en" || req.Lang == "tr") {
			locale = req.Lang
		}
	}

	rep := reporter.NewWithLocale(s.cfg.OutputDir, s.cfg.ExportFormat, s.cfg.TargetDomain, locale)
	var livePool *proxy.LivePool
	
	// Private proxy modu: kullanıcının kendi proxy'lerini LivePool'a ekle
	if s.cfg.UsePrivateProxy && len(s.cfg.PrivateProxies) > 0 {
		// Yeni LivePool oluştur ve private proxy'leri ekle
		livePool = proxy.NewLivePool()
		for _, pp := range s.cfg.PrivateProxies {
			if pp.Host != "" && pp.Port > 0 {
				protocol := pp.Protocol
				if protocol == "" {
					protocol = "http"
				}
				livePool.AddUnchecked(&proxy.ProxyConfig{
					Host:     pp.Host,
					Port:     pp.Port,
					Username: pp.User,
					Password: pp.Pass,
					Protocol: protocol,
				})
			}
		}
		// Log: Private proxy sayısını bildir
		rep.Log(fmt.Sprintf("🔐 Private proxy mode active: %d proxies loaded", livePool.Count()))
	} else if s.cfg.UsePublicProxy && s.proxyService != nil {
		// Public proxy modu
		livePool = s.proxyService.LivePool
	}
	
	sim, err := simulator.New(s.cfg, s.agentLoader, rep, livePool)
	if err != nil {
		s.mu.Unlock()
		http.Error(w, err.Error(), 500)
		return
	}
	s.sim = sim
	
	// SECURITY FIX: Her hit için anlık server bildirimi - callback set et
	rep.SetHitCallback(func(url string, duration time.Duration, success bool, proxy string) {
		// Metrics collector'a kaydet
		s.RecordHit(url, proxy, duration, success)
		
		// Anlık WebSocket broadcast - status güncellemesi
		s.hub.Broadcast("status", s.buildStatusMap())
	})
	
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	logChan := sim.Reporter().LogChan()
	s.mu.Unlock()

	go func() {
		for msg := range logChan {
			s.hub.Broadcast("log", msg)
		}
	}()
	go func() {
		sim.Run(ctx)
		s.mu.Lock()
		s.cancel = nil
		s.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// buildStatusMap handleStatus ve WebSocket için ortak status verisi
func (s *Server) buildStatusMap() map[string]interface{} {
	s.mu.Lock()
	running := s.cancel != nil
	var repMetrics reporter.Metrics
	if s.sim != nil {
		repMetrics = s.sim.Reporter().GetMetrics()
	}
	ps := s.proxyService
	s.mu.Unlock()

	// Snapshot (prometheus collector'dan)
	metricsSnapshot := s.metrics.GetSnapshot()

	out := map[string]interface{}{
		"running":         running,
		"total_hits":      metricsSnapshot.TotalHits,
		"success_hits":    repMetrics.SuccessHits,
		"failed_hits":     repMetrics.FailedHits,
		"avg_response_ms": repMetrics.AvgResponseTime,
		"min_response_ms": repMetrics.MinResponseTime,
		"max_response_ms": repMetrics.MaxResponseTime,
		// Prometheus metrics - dashboard için ana kaynak
		"metrics": map[string]interface{}{
			"total_hits":      metricsSnapshot.TotalHits,
			"success_count":   metricsSnapshot.SuccessCount,
			"error_count":     metricsSnapshot.ErrorCount,
			"bounce_count":    metricsSnapshot.BounceCount,
			"hit_rate_per_min": metricsSnapshot.HitRatePerMin,
			"success_rate":     metricsSnapshot.SuccessRate,
			"bounce_rate":      metricsSnapshot.BounceRate,
			"error_rate":       metricsSnapshot.ErrorRate,
			"active_sessions":  metricsSnapshot.ActiveSessions,
			"active_proxies":   metricsSnapshot.ActiveProxies,
			"queue_size":       metricsSnapshot.QueueSize,
			"uptime_seconds":   metricsSnapshot.UptimeSeconds,
		},
		// Frontend'in doğrudan okuduğu kısayol alanlar
		"success_rate":   metricsSnapshot.SuccessRate,
		"active_proxies": metricsSnapshot.ActiveProxies,
	}
	if ps != nil {
		st := ps.Status()
		out["proxy_status"] = map[string]interface{}{
			"queue_count":    st.QueueCount,
			"live_count":     st.LiveCount,
			"checking":       st.Checking,
			"checked_done":   st.CheckedDone,
			"added_total":    st.AddedTotal,
			"removed_total":  st.RemovedTotal,
		}
		out["proxy_live"] = ps.LivePool.SnapshotForAPI()
	}
	return out
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.buildStatusMap())
}

// SECURITY FIX: WebSocket origin validation to prevent CSWSH attacks
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow same-origin requests (no Origin header)
		if origin == "" {
			return true
		}
		// Allow localhost origins for local development
		allowedOrigins := []string{
			"http://127.0.0.1",
			"http://localhost",
			"https://127.0.0.1",
			"https://localhost",
		}
		for _, allowed := range allowedOrigins {
			if strings.HasPrefix(origin, allowed) {
				return true
			}
		}
		return false
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.hub.Register(conn)
	defer s.hub.Unregister(conn)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	<-done
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	sim := s.sim
	s.mu.Unlock()
	if sim == nil {
		http.Error(w, "Simülasyon çalışmıyor", 400)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// SECURITY FIX: Restrict CORS to localhost only instead of wildcard
	origin := r.Header.Get("Origin")
	if origin == "" || strings.HasPrefix(origin, "http://127.0.0.1") ||
	   strings.HasPrefix(origin, "http://localhost") ||
	   strings.HasPrefix(origin, "https://127.0.0.1") ||
	   strings.HasPrefix(origin, "https://localhost") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	sub := s.hub.SubscribeLog()
	defer s.hub.UnsubscribeLog(sub)
	for {
		select {
		case msg, ok := <-sub:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", escapeSSE(msg))
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleProxyFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	ps := s.proxyService
	cfg := s.cfg
	s.mu.Unlock()
	if ps == nil {
		http.Error(w, "Proxy servisi yok", 500)
		return
	}
	sources := cfg.ProxySourceURLs
	checkerWorkers := cfg.CheckerWorkers
	if checkerWorkers <= 0 {
		checkerWorkers = 25
	}
	var githubRepos []string
	var bodySources []string
	if r.Body != nil {
		var body struct {
			Sources        []string `json:"sources"`
			GitHubRepos    []string `json:"github_repos"`
			CheckerWorkers int      `json:"checker_workers"`
		}
		if json.NewDecoder(r.Body).Decode(&body) == nil {
			bodySources = body.Sources
			if len(body.GitHubRepos) > 0 {
				githubRepos = body.GitHubRepos
			}
			if body.CheckerWorkers > 0 {
				checkerWorkers = body.CheckerWorkers
			}
		}
	}
	if len(githubRepos) == 0 && len(cfg.GitHubRepos) > 0 {
		githubRepos = cfg.GitHubRepos
	}
	// Kullanıcı ne GitHub ne kaynak URL girmemişse: varsayılan GitHub repolarından çek (test yok)
	if len(githubRepos) == 0 && len(bodySources) == 0 && len(cfg.ProxySourceURLs) == 0 {
		githubRepos = proxy.DefaultGitHubRepos
	}
	if len(bodySources) > 0 {
		sources = bodySources
	}
	if len(sources) == 0 && len(githubRepos) == 0 {
		sources = proxy.DefaultProxySourceURLs
	}
	// GitHub repo'ları verilmişse: tüm .txt indir, test yok, havuza ekle; başarısızlar kullanımda silinir
	if len(githubRepos) > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		added, err := ps.FetchFromGitHubNoCheck(ctx, githubRepos, nil)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error(), "added": added})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "added": added})
		return
	}
	ps.FetchAndCheckBackground(sources, checkerWorkers, nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	ps := s.proxyService
	s.mu.Unlock()
	if ps == nil {
		json.NewEncoder(w).Encode(proxy.Status{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps.Status())
}

func (s *Server) handleProxyLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	ps := s.proxyService
	s.mu.Unlock()
	if ps == nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps.LivePool.SnapshotForAPI())
}

func (s *Server) handleProxyExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	ps := s.proxyService
	s.mu.Unlock()
	if ps == nil {
		http.Error(w, "Proxy servisi yok", 500)
		return
	}
	data := ps.LivePool.ExportTxt()
	if len(data) == 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Canlı proxy yok\n"))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=live_proxies.txt")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
func (s *Server) handleProxyTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	
	var body struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		User string `json:"user"`
		Pass string `json:"pass"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	
	if body.Host == "" || body.Port == 0 {
		http.Error(w, "Host and port required", 400)
		return
	}
	
	// Proxy test - basit HTTP bağlantı testi
	proxyURL := fmt.Sprintf("http://%s:%d", body.Host, body.Port)
	if body.User != "" {
		proxyURL = fmt.Sprintf("http://%s:%s@%s:%d", body.User, body.Pass, body.Host, body.Port)
	}
	
	// Test için httpbin.org'a istek at
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
	}
	
	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     resp.StatusCode == 200,
		"status_code": resp.StatusCode,
	})
}

func (s *Server) handleGSCQueries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	
	var body struct {
		PropertyURL string `json:"property_url"`
		APIKey      string `json:"api_key"`
		Days        int    `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	
	if body.PropertyURL == "" {
		http.Error(w, "Property URL required", 400)
		return
	}
	
	// Property URL'yi normalize et
	// Kullanıcı sadece domain girerse (örn: eros.sh), otomatik olarak sc-domain: formatına çevir
	propertyURL := strings.TrimSpace(body.PropertyURL)
	propertyURL = strings.TrimSuffix(propertyURL, "/")
	
	// Eğer http:// veya https:// ile başlamıyorsa ve sc-domain: değilse
	if !strings.HasPrefix(propertyURL, "http://") &&
	   !strings.HasPrefix(propertyURL, "https://") &&
	   !strings.HasPrefix(propertyURL, "sc-domain:") {
		// Domain property olarak ayarla (en yaygın format)
		propertyURL = "sc-domain:" + propertyURL
	}
	
	if body.APIKey == "" {
		http.Error(w, "API Key (Service Account JSON) required", 400)
		return
	}
	
	// Service Account JSON'ı parse et
	var serviceAccount struct {
		Type                    string `json:"type"`
		ProjectID               string `json:"project_id"`
		PrivateKeyID            string `json:"private_key_id"`
		PrivateKey              string `json:"private_key"`
		ClientEmail             string `json:"client_email"`
		ClientID                string `json:"client_id"`
		AuthURI                 string `json:"auth_uri"`
		TokenURI                string `json:"token_uri"`
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
		ClientX509CertURL       string `json:"client_x509_cert_url"`
	}
	
	if err := json.Unmarshal([]byte(body.APIKey), &serviceAccount); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid Service Account JSON format: " + err.Error(),
		})
		return
	}
	
	if serviceAccount.Type != "service_account" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid credential type. Expected 'service_account', got '" + serviceAccount.Type + "'",
		})
		return
	}
	
	if serviceAccount.PrivateKey == "" || serviceAccount.ClientEmail == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Service Account JSON missing required fields (private_key or client_email)",
		})
		return
	}
	
	// GSC API çağrısı yap
	days := body.Days
	if days <= 0 {
		days = 28 // Varsayılan 28 gün
	}
	
	queries, err := fetchGSCQueries(propertyURL, serviceAccount.ClientEmail, serviceAccount.PrivateKey, days)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "GSC API error: " + err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"queries": queries,
	})
}

// fetchGSCQueries Google Search Console API'den sorguları çeker
func fetchGSCQueries(propertyURL, clientEmail, privateKey string, days int) ([]map[string]interface{}, error) {
	// JWT token oluştur
	token, err := createGSCJWT(clientEmail, privateKey)
	if err != nil {
		return nil, fmt.Errorf("JWT oluşturma hatası: %w", err)
	}
	
	// Access token al
	accessToken, err := exchangeJWTForAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("Access token alma hatası: %w", err)
	}
	
	// GSC API'ye istek at
	endDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	
	requestBody := map[string]interface{}{
		"startDate":  startDate,
		"endDate":    endDate,
		"dimensions": []string{"query"},
		"rowLimit":   100,
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	
	// Property URL'yi encode et
	encodedProperty := url.QueryEscape(propertyURL)
	apiURL := fmt.Sprintf("https://www.googleapis.com/webmasters/v3/sites/%s/searchAnalytics/query", encodedProperty)
	
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GSC API hatası (%d): %s", resp.StatusCode, string(bodyBytes))
	}
	
	var gscResponse struct {
		Rows []struct {
			Keys        []string `json:"keys"`
			Clicks      float64  `json:"clicks"`
			Impressions float64  `json:"impressions"`
			CTR         float64  `json:"ctr"`
			Position    float64  `json:"position"`
		} `json:"rows"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&gscResponse); err != nil {
		return nil, fmt.Errorf("GSC yanıt parse hatası: %w", err)
	}
	
	queries := make([]map[string]interface{}, 0, len(gscResponse.Rows))
	for _, row := range gscResponse.Rows {
		if len(row.Keys) > 0 {
			queries = append(queries, map[string]interface{}{
				"query":       row.Keys[0],
				"clicks":      int(row.Clicks),
				"impressions": int(row.Impressions),
				"ctr":         row.CTR,
				"position":    row.Position,
			})
		}
	}
	
	return queries, nil
}

// createGSCJWT Service Account için JWT oluşturur
func createGSCJWT(clientEmail, privateKey string) (string, error) {
	// JWT Header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64URLEncode(headerJSON)
	
	// JWT Claims
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"iss":   clientEmail,
		"scope": "https://www.googleapis.com/auth/webmasters.readonly",
		"aud":   "https://oauth2.googleapis.com/token",
		"iat":   now,
		"exp":   now + 3600,
	}
	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64URLEncode(claimsJSON)
	
	// Signature
	signatureInput := headerB64 + "." + claimsB64
	signature, err := signRS256(signatureInput, privateKey)
	if err != nil {
		return "", err
	}
	
	return signatureInput + "." + signature, nil
}

// exchangeJWTForAccessToken JWT'yi access token ile değiştirir
func exchangeJWTForAccessToken(jwt string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", jwt)
	
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange hatası (%d): %s", resp.StatusCode, string(bodyBytes))
	}
	
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}
	
	return tokenResponse.AccessToken, nil
}

// base64URLEncode base64 URL encoding
func base64URLEncode(data []byte) string {
	encoded := encodeBase64(data)
	// URL-safe: + -> -, / -> _, padding kaldır
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.TrimRight(encoded, "=")
	return encoded
}

// encodeBase64 standart base64 encoding
func encodeBase64(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, ((len(data)+2)/3)*4)
	
	for i, j := 0, 0; i < len(data); i, j = i+3, j+4 {
		var n uint32
		remaining := len(data) - i
		
		n = uint32(data[i]) << 16
		if remaining > 1 {
			n |= uint32(data[i+1]) << 8
		}
		if remaining > 2 {
			n |= uint32(data[i+2])
		}
		
		result[j] = base64Chars[(n>>18)&0x3F]
		result[j+1] = base64Chars[(n>>12)&0x3F]
		
		if remaining > 1 {
			result[j+2] = base64Chars[(n>>6)&0x3F]
		} else {
			result[j+2] = '='
		}
		
		if remaining > 2 {
			result[j+3] = base64Chars[n&0x3F]
		} else {
			result[j+3] = '='
		}
	}
	
	return string(result)
}

// signRS256 RS256 imzalama
func signRS256(input, privateKeyPEM string) (string, error) {
	// PEM formatındaki private key'i parse et
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("PEM block bulunamadı")
	}
	
	var privateKey interface{}
	var err error
	
	// PKCS#8 veya PKCS#1 formatını dene
	privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("private key parse hatası: %w", err)
		}
	}
	
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("RSA private key değil")
	}
	
	// SHA256 hash
	h := sha256.New()
	h.Write([]byte(input))
	hashed := h.Sum(nil)
	
	// RSA imzala
	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("imzalama hatası: %w", err)
	}
	
	// Base64 URL encode
	return strings.TrimRight(strings.ReplaceAll(strings.ReplaceAll(
		encodeBase64(signature), "+", "-"), "/", "_"), "="), nil
}

func escapeSSE(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}


