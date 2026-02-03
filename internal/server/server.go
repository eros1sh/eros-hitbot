package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"eroshit/internal/config"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/internal/simulator"
	"eroshit/pkg/useragent"

	"github.com/gorilla/websocket"
)

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
}

func saveConfigToFile(cfg *config.Config) {
	wd, _ := os.Getwd()
	paths := []string{filepath.Join(wd, "config.json"), filepath.Join(wd, "..", "config.json"), "config.json"}
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

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("/", http.FileServer(http.FS(sub)))

	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/api/ws", s.handleWebSocket)
	mux.HandleFunc("/api/proxy/fetch", s.handleProxyFetch)
	mux.HandleFunc("/api/proxy/status", s.handleProxyStatus)
	mux.HandleFunc("/api/proxy/live", s.handleProxyLive)
	mux.HandleFunc("/api/proxy/export", s.handleProxyExport)

	return mux
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		s.mu.Lock()
		cfg := s.cfg
		s.mu.Unlock()
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
	if s.cfg.UsePublicProxy && s.proxyService != nil {
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

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
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

func escapeSSE(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

