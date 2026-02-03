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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"eroshit/internal/config"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/internal/simulator"
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
	mu           sync.Mutex
	cfg          *config.Config
	sim          *simulator.Simulator
	cancel       context.CancelFunc
	agentLoader  *useragent.Loader
	proxyService *proxy.Service
	hub          *Hub
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

	s := &Server{
		cfg:          cfg,
		agentLoader:  agentLoader,
		proxyService: proxy.NewService(),
		hub:          NewHub(),
	}
	go s.broadcastStatusLoop()
	return s, nil
}

func (s *Server) broadcastStatusLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.hub.Broadcast("status", s.buildStatusMap())
	}
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
	wd, _ := os.Getwd()
	paths := []string{filepath.Join(wd, "config.json"), filepath.Join(wd, "..", "config.json"), "config.json"}
	
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
	
	for _, p := range paths {
		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0755); err == nil {
			data, _ := json.MarshalIndent(configFile{
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
			if err := os.WriteFile(p, data, 0644); err == nil {
				return
			}
		}
	}
}

func loadConfig(dirs []string) (*config.Config, error) {
	for _, d := range dirs {
		p := filepath.Join(d, "config.json")
		if _, err := os.Stat(p); err == nil {
			return config.LoadFromJSON(p)
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
			// Yeni alanlar
			"device_type":            cfg.DeviceType,
			"device_brands":          cfg.DeviceBrands,
			"referrer_keyword":       cfg.ReferrerKeyword,
			"referrer_enabled":       cfg.ReferrerEnabled,
		})
		return
	}
	if r.Method == http.MethodPost {
		var body struct {
			TargetDomain          string   `json:"target_domain"`
			MaxPages              int      `json:"max_pages"`
			DurationMinutes       int      `json:"duration_minutes"`
			HitsPerMinute         int      `json:"hits_per_minute"`
			MaxConcurrentVisits   int      `json:"max_concurrent_visits"`
			OutputDir             string   `json:"output_dir"`
			ExportFormat          string   `json:"export_format"`
			CanvasFingerprint      bool     `json:"canvas_fingerprint"`
			ScrollStrategy        string   `json:"scroll_strategy"`
			SendScrollEvent       bool     `json:"send_scroll_event"`
			UseSitemap            bool     `json:"use_sitemap"`
			SitemapHomepageWeight int      `json:"sitemap_homepage_weight"`
			Keywords              []string `json:"keywords"`
			ProxyHost             string   `json:"proxy_host"`
			ProxyPort             int      `json:"proxy_port"`
			ProxyUser             string   `json:"proxy_user"`
			ProxyPass             string   `json:"proxy_pass"`
			GtagID                string   `json:"gtag_id"`
			UsePublicProxy        bool     `json:"use_public_proxy"`
			ProxySourceURLs       []string `json:"proxy_source_urls"`
			GitHubRepos           []string `json:"github_repos"`
			CheckerWorkers        int      `json:"checker_workers"`
			// Private proxy alanları
			UsePrivateProxy   bool     `json:"use_private_proxy"`
			// Yeni alanlar
			DeviceType        string   `json:"device_type"`
			DeviceBrands      []string `json:"device_brands"`
			ReferrerKeyword   string   `json:"referrer_keyword"`
			ReferrerEnabled   bool     `json:"referrer_enabled"`
			// Private proxy listesi
			PrivateProxies    []struct {
				Host     string `json:"host"`
				Port     int    `json:"port"`
				User     string `json:"user"`
				Pass     string `json:"pass"`
				Protocol string `json:"protocol"`
			} `json:"private_proxies"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		s.mu.Lock()
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
		s.cfg.ProxyHost = body.ProxyHost
		s.cfg.ProxyPort = body.ProxyPort
		s.cfg.ProxyUser = body.ProxyUser
		s.cfg.ProxyPass = body.ProxyPass
		s.cfg.GtagID = body.GtagID
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
		
		// Yeni alanlar
		s.cfg.DeviceType = body.DeviceType
		s.cfg.DeviceBrands = body.DeviceBrands
		s.cfg.ReferrerKeyword = body.ReferrerKeyword
		s.cfg.ReferrerEnabled = body.ReferrerEnabled
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
		rep.Log(fmt.Sprintf("🔐 Private proxy modu aktif: %d proxy yüklendi", livePool.Count()))
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
	var metrics reporter.Metrics
	if s.sim != nil {
		metrics = s.sim.Reporter().GetMetrics()
	}
	ps := s.proxyService
	s.mu.Unlock()

	out := map[string]interface{}{
		"running":          running,
		"total_hits":       metrics.TotalHits,
		"success_hits":     metrics.SuccessHits,
		"failed_hits":      metrics.FailedHits,
		"avg_response_ms":  metrics.AvgResponseTime,
		"min_response_ms":  metrics.MinResponseTime,
		"max_response_ms":  metrics.MaxResponseTime,
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
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


