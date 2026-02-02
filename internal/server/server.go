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
	"eroshit/internal/reporter"
	"eroshit/internal/simulator"
	"eroshit/pkg/useragent"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	mu           sync.Mutex
	cfg          *config.Config
	sim          *simulator.Simulator
	cancel       context.CancelFunc
	agentLoader  *useragent.Loader
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

	return &Server{
		cfg:          cfg,
		agentLoader:  agentLoader,
	}, nil
}

type configFile struct {
	PROXY_HOST           string `json:"PROXY_HOST"`
	PROXY_PORT           int    `json:"PROXY_PORT"`
	PROXY_USER           string `json:"PROXY_USER"`
	PROXY_PASS           string `json:"PROXY_PASS"`
	TargetDomain         string `json:"targetDomain"`
	FallbackGAID         string `json:"fallbackGAID"`
	MaxPages             int    `json:"maxPages"`
	DurationMinutes      int    `json:"durationMinutes"`
	HitsPerMinute        int    `json:"hitsPerMinute"`
	MaxConcurrentVisits  int    `json:"maxConcurrentVisits"`
	OutputDir            string `json:"outputDir"`
	ExportFormat         string `json:"exportFormat"`
	CanvasFingerprint       bool     `json:"canvasFingerprint"`
	ScrollStrategy          string   `json:"scrollStrategy"`
	SendScrollEvent         bool     `json:"sendScrollEvent"`
	UseSitemap              bool     `json:"useSitemap"`
	SitemapHomepageWeight   int      `json:"sitemapHomepageWeight"`
	Keywords                []string `json:"keywords"`
}

func saveConfigToFile(cfg *config.Config) {
	wd, _ := os.Getwd()
	paths := []string{filepath.Join(wd, "config.json"), filepath.Join(wd, "..", "config.json"), "config.json"}
	for _, p := range paths {
		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0755); err == nil {
			data, _ := json.MarshalIndent(configFile{
				PROXY_HOST:          cfg.ProxyHost,
				PROXY_PORT:          cfg.ProxyPort,
				PROXY_USER:          cfg.ProxyUser,
				PROXY_PASS:          cfg.ProxyPass,
				TargetDomain:        cfg.TargetDomain,
				FallbackGAID:        cfg.GtagID,
				MaxPages:            cfg.MaxPages,
				DurationMinutes:     cfg.DurationMinutes,
				HitsPerMinute:       cfg.HitsPerMinute,
				MaxConcurrentVisits: cfg.MaxConcurrentVisits,
				OutputDir:           cfg.OutputDir,
				ExportFormat:        cfg.ExportFormat,
				CanvasFingerprint:   cfg.CanvasFingerprint,
				ScrollStrategy:          cfg.ScrollStrategy,
				SendScrollEvent:         cfg.SendScrollEvent,
				UseSitemap:              cfg.UseSitemap,
				SitemapHomepageWeight:   cfg.SitemapHomepageWeight,
				Keywords:                cfg.Keywords,
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

	return mux
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		s.mu.Lock()
		cfg := s.cfg
		s.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"target_domain":         cfg.TargetDomain,
			"max_pages":             cfg.MaxPages,
			"duration_minutes":      cfg.DurationMinutes,
			"hits_per_minute":       cfg.HitsPerMinute,
			"max_concurrent_visits": cfg.MaxConcurrentVisits,
			"output_dir":            cfg.OutputDir,
			"export_format":         cfg.ExportFormat,
			"canvas_fingerprint":    cfg.CanvasFingerprint,
			"scroll_strategy":       cfg.ScrollStrategy,
			"send_scroll_event":       cfg.SendScrollEvent,
			"use_sitemap":             cfg.UseSitemap,
			"sitemap_homepage_weight": cfg.SitemapHomepageWeight,
			"keywords":                cfg.Keywords,
			"proxy_host":              cfg.ProxyHost,
			"proxy_port":            cfg.ProxyPort,
			"proxy_user":            cfg.ProxyUser,
			"proxy_pass":            cfg.ProxyPass,
			"gtag_id":               cfg.GtagID,
		})
		return
	}
	if r.Method == http.MethodPost {
		var body struct {
			TargetDomain         string `json:"target_domain"`
			MaxPages             int    `json:"max_pages"`
			DurationMinutes      int    `json:"duration_minutes"`
			HitsPerMinute        int    `json:"hits_per_minute"`
			MaxConcurrentVisits  int    `json:"max_concurrent_visits"`
			OutputDir            string `json:"output_dir"`
			ExportFormat         string `json:"export_format"`
			CanvasFingerprint    bool   `json:"canvas_fingerprint"`
			ScrollStrategy         string   `json:"scroll_strategy"`
			SendScrollEvent        bool     `json:"send_scroll_event"`
			UseSitemap             bool     `json:"use_sitemap"`
			SitemapHomepageWeight  int      `json:"sitemap_homepage_weight"`
			Keywords               []string `json:"keywords"`
			ProxyHost              string   `json:"proxy_host"`
			ProxyPort            int    `json:"proxy_port"`
			ProxyUser            string `json:"proxy_user"`
			ProxyPass            string `json:"proxy_pass"`
			GtagID               string `json:"gtag_id"`
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
	sim, err := simulator.New(s.cfg, s.agentLoader, rep)
	if err != nil {
		s.mu.Unlock()
		http.Error(w, err.Error(), 500)
		return
	}
	s.sim = sim
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

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

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	running := s.cancel != nil
	var metrics reporter.Metrics
	if s.sim != nil {
		metrics = s.sim.Reporter().GetMetrics()
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running":       running,
		"total_hits":    metrics.TotalHits,
		"success_hits":  metrics.SuccessHits,
		"failed_hits":   metrics.FailedHits,
		"avg_response_ms": metrics.AvgResponseTime,
		"min_response_ms": metrics.MinResponseTime,
		"max_response_ms": metrics.MaxResponseTime,
	})
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

	logChan := sim.Reporter().LogChan()
	for {
		select {
		case msg, ok := <-logChan:
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

func escapeSSE(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

