package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config uygulama konfigürasyonu
type Config struct {
	TargetDomain        string        `yaml:"target_domain"`
	MaxPages            int           `yaml:"max_pages"`
	DurationMinutes     int           `yaml:"duration_minutes"`
	HitsPerMinute       int           `yaml:"hits_per_minute"`
	DisableImages       bool          `yaml:"disable_images"`
	DisableJSExecution  bool          `yaml:"disable_js_execution"`
	ProxyEnabled        bool          `yaml:"proxy_enabled"`
	ProxyHost           string        `yaml:"proxy_host"`
	ProxyPort           int           `yaml:"proxy_port"`
	ProxyUser           string        `yaml:"proxy_user"`
	ProxyPass           string        `yaml:"proxy_pass"`
	ProxyURL            string        `yaml:"-"`
	ProxyBaseURL        string        `yaml:"-"` // auth olmadan host:port
	GtagID               string        `yaml:"gtag_id"`
	LogLevel             string        `yaml:"log_level"`
	ExportFormat         string        `yaml:"export_format"`
	OutputDir            string        `yaml:"output_dir"`
	MaxConcurrentVisits  int           `yaml:"max_concurrent_visits"`
	CanvasFingerprint    bool          `yaml:"canvas_fingerprint"`
	ScrollStrategy       string        `yaml:"scroll_strategy"`
	SendScrollEvent       bool          `yaml:"send_scroll_event"`
	UseSitemap            bool          `yaml:"use_sitemap"`
	SitemapHomepageWeight int           `yaml:"sitemap_homepage_weight"` // 0-100, anasayfa yüzdesi
	Keywords              []string      `yaml:"keywords"`
	// Public proxy: listelerden çek, checker ile test et, çalışanlarla vur
	UsePublicProxy   bool     `yaml:"use_public_proxy"`
	ProxySourceURLs  []string `yaml:"proxy_source_urls"`  // Boşsa varsayılan listeler
	GitHubRepos      []string `yaml:"github_repos"`      // GitHub repo URL'leri: tüm .txt indirilir, test yok
	CheckerWorkers   int     `yaml:"checker_workers"`   // Aynı anda test eden worker sayısı
	Duration              time.Duration `yaml:"-"`
	RequestInterval       time.Duration `yaml:"-"`
}

// LoadFromFile YAML dosyasından config yükler
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.ApplyDefaults()
	cfg.ComputeDerived()
	return &cfg, nil
}

// LoadFromEnv Ortam değişkenlerinden config yükler (override için)
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("EROSHIT_TARGET_DOMAIN"); v != "" {
		c.TargetDomain = v
	}
	if v := os.Getenv("EROSHIT_MAX_PAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxPages = n
		}
	}
	if v := os.Getenv("EROSHIT_DURATION_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.DurationMinutes = n
		}
	}
	if v := os.Getenv("EROSHIT_HITS_PER_MINUTE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.HitsPerMinute = n
		}
	}
}

// ApplyDefaults Varsayılan değerleri uygular
func (c *Config) ApplyDefaults() {
	if c.TargetDomain == "" {
		c.TargetDomain = "example.com"
	}
	if c.MaxPages <= 0 {
		c.MaxPages = 5
	}
	if c.MaxPages > 100 {
		c.MaxPages = 100
	}
	if c.DurationMinutes <= 0 {
		c.DurationMinutes = 60
	}
	if c.HitsPerMinute <= 0 {
		c.HitsPerMinute = 35
	}
	if c.HitsPerMinute > 120 {
		c.HitsPerMinute = 120
	}
	if c.OutputDir == "" {
		c.OutputDir = "./reports"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.ExportFormat == "" {
		c.ExportFormat = "both"
	}
	if c.MaxConcurrentVisits <= 0 {
		c.MaxConcurrentVisits = 10
	}
	if c.MaxConcurrentVisits > 50 {
		c.MaxConcurrentVisits = 50
	}
	if c.SitemapHomepageWeight <= 0 {
		c.SitemapHomepageWeight = 60
	}
	if c.SitemapHomepageWeight > 100 {
		c.SitemapHomepageWeight = 100
	}
	if c.CheckerWorkers <= 0 {
		c.CheckerWorkers = 25
	}
	if c.CheckerWorkers > 100 {
		c.CheckerWorkers = 100
	}
	c.TargetDomain = strings.TrimSpace(strings.TrimPrefix(c.TargetDomain, "https://"))
	c.TargetDomain = strings.TrimPrefix(c.TargetDomain, "http://")
	c.TargetDomain = strings.TrimSuffix(c.TargetDomain, "/")
}

// ComputeDerived Türetilmiş değerleri hesaplar
func (c *Config) ComputeDerived() {
	c.Duration = time.Duration(c.DurationMinutes) * time.Minute
	if c.HitsPerMinute > 0 {
		c.RequestInterval = time.Minute / time.Duration(c.HitsPerMinute)
	} else {
		c.RequestInterval = 2 * time.Second
	}
	if c.ProxyHost != "" && c.ProxyPort > 0 {
		c.ProxyBaseURL = fmt.Sprintf("http://%s:%d", c.ProxyHost, c.ProxyPort)
		c.ProxyURL = buildProxyURL(c.ProxyHost, c.ProxyPort, c.ProxyUser, c.ProxyPass)
		c.ProxyEnabled = true
	}
	if c.GtagID == "" {
		c.GtagID = "G-5WW6MDM5EN"
	}
}
