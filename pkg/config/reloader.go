// Package config provides hot-reload functionality for configuration files.
// This package wraps internal/config and adds file watching capabilities.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// Duration type alias for YAML parsing
type Duration time.Duration

// Config mirrors internal/config.Config for the reloader
// This avoids import cycles while maintaining compatibility
type Config struct {
	TargetDomain         string        `yaml:"target_domain"`
	MaxPages             int           `yaml:"max_pages"`
	DurationMinutes      int           `yaml:"duration_minutes"`
	HitsPerMinute        int           `yaml:"hits_per_minute"`
	DisableImages        bool          `yaml:"disable_images"`
	DisableJSExecution   bool          `yaml:"disable_js_execution"`
	ProxyEnabled         bool          `yaml:"proxy_enabled"`
	ProxyHost            string        `yaml:"proxy_host"`
	ProxyPort            int           `yaml:"proxy_port"`
	ProxyUser            string        `yaml:"proxy_user"`
	ProxyPass            string        `yaml:"proxy_pass"`
	ProxyURL             string        `yaml:"-"`
	ProxyBaseURL         string        `yaml:"-"`
	GtagID               string        `yaml:"gtag_id"`
	LogLevel             string        `yaml:"log_level"`
	ExportFormat         string        `yaml:"export_format"`
	OutputDir            string        `yaml:"output_dir"`
	MaxConcurrentVisits  int           `yaml:"max_concurrent_visits"`
	CanvasFingerprint    bool          `yaml:"canvas_fingerprint"`
	ScrollStrategy       string        `yaml:"scroll_strategy"`
	SendScrollEvent      bool          `yaml:"send_scroll_event"`
	UseSitemap           bool          `yaml:"use_sitemap"`
	SitemapHomepageWeight int          `yaml:"sitemap_homepage_weight"`
	Keywords             []string      `yaml:"keywords"`
	UsePublicProxy       bool          `yaml:"use_public_proxy"`
	ProxySourceURLs      []string      `yaml:"proxy_source_urls"`
	GitHubRepos          []string      `yaml:"github_repos"`
	CheckerWorkers       int           `yaml:"checker_workers"`
	PrivateProxies       []PrivateProxy `yaml:"private_proxies"`
	UsePrivateProxy      bool           `yaml:"use_private_proxy"`
	DeviceType           string        `yaml:"device_type"`
	DeviceBrands         []string      `yaml:"device_brands"`
	ReferrerKeyword      string        `yaml:"referrer_keyword"`
	ReferrerEnabled      bool          `yaml:"referrer_enabled"`
	ReferrerSource       string        `yaml:"referrer_source"`
	MinPageDuration      int           `yaml:"min_page_duration"`
	MaxPageDuration      int           `yaml:"max_page_duration"`
	MinScrollPercent     int           `yaml:"min_scroll_percent"`
	MaxScrollPercent     int           `yaml:"max_scroll_percent"`
	ClickProbability     int           `yaml:"click_probability"`
	SessionMinPages      int           `yaml:"session_min_pages"`
	SessionMaxPages      int           `yaml:"session_max_pages"`
	EnableSessionDepth   bool          `yaml:"enable_session_depth"`
	TargetBounceRate     int           `yaml:"target_bounce_rate"`
	EnableBounceControl  bool          `yaml:"enable_bounce_control"`
	SimulateMouseMove    bool          `yaml:"simulate_mouse_move"`
	SimulateKeyboard     bool          `yaml:"simulate_keyboard"`
	SimulateClicks       bool          `yaml:"simulate_clicks"`
	SimulateFocus        bool          `yaml:"simulate_focus"`
	GeoCountry           string        `yaml:"geo_country"`
	GeoTimezone          string        `yaml:"geo_timezone"`
	GeoLanguage          string        `yaml:"geo_language"`
	SendPageView         bool          `yaml:"send_page_view"`
	SendSessionStart     bool          `yaml:"send_session_start"`
	SendUserEngagement   bool          `yaml:"send_user_engagement"`
	SendFirstVisit       bool          `yaml:"send_first_visit"`
	CustomDimensions     string        `yaml:"custom_dimensions"`
	CustomMetrics        string        `yaml:"custom_metrics"`
	EnableCustomDimensions bool        `yaml:"enable_custom_dimensions"`
	GscPropertyUrl       string        `yaml:"gsc_property_url"`
	GscApiKey            string        `yaml:"gsc_api_key"`
	EnableGscIntegration bool          `yaml:"enable_gsc_integration"`
	UseGscQueries        bool          `yaml:"use_gsc_queries"`
	ReturningVisitorRate int           `yaml:"returning_visitor_rate"`
	ReturningVisitorDays int           `yaml:"returning_visitor_days"`
	EnableReturningVisitor bool        `yaml:"enable_returning_visitor"`
	ExitPages            []string      `yaml:"exit_pages"`
	EnableExitPageControl bool         `yaml:"enable_exit_page_control"`
	BrowserProfilePath   string        `yaml:"browser_profile_path"`
	MaxBrowserProfiles   int           `yaml:"max_browser_profiles"`
	EnableBrowserProfile bool          `yaml:"enable_browser_profile"`
	PersistCookies       bool          `yaml:"persist_cookies"`
	PersistLocalStorage  bool          `yaml:"persist_local_storage"`
	TlsFingerprintMode   string        `yaml:"tls_fingerprint_mode"`
	EnableJa3Randomization bool        `yaml:"enable_ja3_randomization"`
	EnableJa4Randomization bool        `yaml:"enable_ja4_randomization"`
	ProxyRotationMode    string        `yaml:"proxy_rotation_mode"`
	ProxyRotationInterval int          `yaml:"proxy_rotation_interval"`
	EnableProxyRotation  bool          `yaml:"enable_proxy_rotation"`
	EnableHttp2Fingerprint bool        `yaml:"enable_http2_fingerprint"`
	EnableHttp3Fingerprint bool        `yaml:"enable_http3_fingerprint"`
	Http2FingerprintMode string        `yaml:"http2_fingerprint_mode"`
	EnableClientHints    bool          `yaml:"enable_client_hints"`
	SpoofSecChUa         bool          `yaml:"spoof_sec_ch_ua"`
	SpoofSecChUaPlatform bool          `yaml:"spoof_sec_ch_ua_platform"`
	SpoofSecChUaArch     bool          `yaml:"spoof_sec_ch_ua_arch"`
	BypassPuppeteer      bool          `yaml:"bypass_puppeteer"`
	BypassPlaywright     bool          `yaml:"bypass_playwright"`
	BypassSelenium       bool          `yaml:"bypass_selenium"`
	BypassCDP            bool          `yaml:"bypass_cdp"`
	BypassRenderingDetection bool      `yaml:"bypass_rendering_detection"`
	BypassWebdriver      bool          `yaml:"bypass_webdriver"`
	SerpSearchEngine     string        `yaml:"serp_search_engine"`
	EnableSerpSimulation bool          `yaml:"enable_serp_simulation"`
	SerpScrollBeforeClick bool         `yaml:"serp_scroll_before_click"`
	BrowserPoolMin       int           `yaml:"browser_pool_min"`
	BrowserPoolMax       int           `yaml:"browser_pool_max"`
	EnableAutoScaling    bool          `yaml:"enable_auto_scaling"`
	WorkerQueueSize      int           `yaml:"worker_queue_size"`
	EnablePriorityQueue  bool          `yaml:"enable_priority_queue"`
	EnableFailureRecovery bool         `yaml:"enable_failure_recovery"`
	IosDeviceModel       string        `yaml:"ios_device_model"`
	EnableIosSafari      bool          `yaml:"enable_ios_safari"`
	EnableIosHaptics     bool          `yaml:"enable_ios_haptics"`
	AndroidDeviceModel   string        `yaml:"android_device_model"`
	EnableAndroidChrome  bool          `yaml:"enable_android_chrome"`
	EnableAndroidVibration bool        `yaml:"enable_android_vibration"`
	EnableTouchEvents    bool          `yaml:"enable_touch_events"`
	EnableMultiTouch     bool          `yaml:"enable_multi_touch"`
	EnableGestures       bool          `yaml:"enable_gestures"`
	EnableMobileKeyboard bool          `yaml:"enable_mobile_keyboard"`
	EnableAccelerometer  bool          `yaml:"enable_accelerometer"`
	EnableGyroscope      bool          `yaml:"enable_gyroscope"`
	EnableDeviceOrientation bool       `yaml:"enable_device_orientation"`
	EnableMagnetometer   bool          `yaml:"enable_magnetometer"`
	EnableTabletMode     bool          `yaml:"enable_tablet_mode"`
	EnableLandscapeMode  bool          `yaml:"enable_landscape_mode"`
	EnablePenInput       bool          `yaml:"enable_pen_input"`
	EnableSplitView      bool          `yaml:"enable_split_view"`
	StealthWebdriver     bool          `yaml:"stealth_webdriver"`
	StealthChrome        bool          `yaml:"stealth_chrome"`
	StealthPlugins       bool          `yaml:"stealth_plugins"`
	StealthWebGL         bool          `yaml:"stealth_webgl"`
	StealthAudio         bool          `yaml:"stealth_audio"`
	StealthCanvas        bool          `yaml:"stealth_canvas"`
	StealthTimezone      bool          `yaml:"stealth_timezone"`
	StealthLanguage      bool          `yaml:"stealth_language"`
	VisitTimeout         int           `yaml:"visit_timeout"`
	PageLoadWait         int           `yaml:"page_load_wait"`
	RetryCount           int           `yaml:"retry_count"`
	BlockImages          bool          `yaml:"block_images"`
	BlockStyles          bool          `yaml:"block_styles"`
	BlockFonts           bool          `yaml:"block_fonts"`
	BlockMedia           bool          `yaml:"block_media"`
	AntiDetectMode       bool          `yaml:"anti_detect_mode"`
	BrowserPoolEnabled   bool          `yaml:"browser_pool_enabled"`
	BrowserPoolSize      int           `yaml:"browser_pool_size"`
	BrowserPoolMinSize   int           `yaml:"browser_pool_min_size"`
	BrowserMaxSessions   int           `yaml:"browser_max_sessions"`
	BrowserMaxAge        int           `yaml:"browser_max_age"`
	ConnectionTimeout    int           `yaml:"connection_timeout"`
	MaxIdleConnections   int           `yaml:"max_idle_connections"`
	EnableKeepAlive      bool          `yaml:"enable_keep_alive"`
	EnableBufferPool     bool          `yaml:"enable_buffer_pool"`
	MaxMemoryMB          int           `yaml:"max_memory_mb"`
	Duration             time.Duration `yaml:"-"`
	RequestInterval      time.Duration `yaml:"-"`
}

// PrivateProxy represents a proxy configuration
type PrivateProxy struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Protocol string `yaml:"protocol"`
}

// ChangeCallback is called when config changes
type ChangeCallback func(newCfg *Config)

// Reloader watches config file for changes and reloads it
type Reloader struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
	
	watcher    *fsnotify.Watcher
	callbacks  []ChangeCallback
	cbMu       sync.RWMutex
	
	debounceTimer *time.Timer
	debounceMu    sync.Mutex
	debounceDelay time.Duration
	
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	logger Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// defaultLogger is a no-op logger
type defaultLogger struct{}

func (d *defaultLogger) Info(msg string, fields ...interface{})  {}
func (d *defaultLogger) Error(msg string, fields ...interface{}) {}

// NewReloader creates a new config reloader
func NewReloader(configPath string) *Reloader {
	return &Reloader{
		configPath:    configPath,
		callbacks:     make([]ChangeCallback, 0),
		debounceDelay: 1 * time.Second,
		logger:        &defaultLogger{},
	}
}

// SetLogger sets a custom logger
func (r *Reloader) SetLogger(logger Logger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = logger
}

// SetDebounceDelay sets the debounce delay (default: 1 second)
func (r *Reloader) SetDebounceDelay(delay time.Duration) {
	r.debounceMu.Lock()
	defer r.debounceMu.Unlock()
	r.debounceDelay = delay
}

// OnChange registers a callback to be called when config changes
func (r *Reloader) OnChange(callback ChangeCallback) {
	r.cbMu.Lock()
	defer r.cbMu.Unlock()
	r.callbacks = append(r.callbacks, callback)
}

// GetConfig returns the current config (thread-safe)
func (r *Reloader) GetConfig() *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// Load loads the config from file (initial load)
func (r *Reloader) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	cfg, err := r.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	r.config = cfg
	r.logger.Info("config_loaded", "path", r.configPath)
	return nil
}

// Start starts watching the config file for changes
func (r *Reloader) Start() error {
	if r.ctx != nil {
		return fmt.Errorf("reloader already started")
	}
	
	// Load initial config
	if err := r.Load(); err != nil {
		return err
	}
	
	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	r.watcher = watcher
	
	// Get the directory and filename
	dir := filepath.Dir(r.configPath)
	
	// Watch the directory (to catch renames/atomic writes)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch directory: %w", err)
	}
	
	// Also try to watch the file directly if it exists
	if _, err := os.Stat(r.configPath); err == nil {
		if err := watcher.Add(r.configPath); err != nil {
			// Log but don't fail - directory watching might be enough
			r.logger.Error("failed_to_watch_file", "path", r.configPath, "error", err)
		}
	}
	
	// Setup context
	r.ctx, r.cancel = context.WithCancel(context.Background())
	
	// Start watching
	r.wg.Add(1)
	go r.watch()
	
	r.logger.Info("config_reloader_started", "path", r.configPath)
	return nil
}

// Stop stops watching and cleans up resources
func (r *Reloader) Stop() error {
	if r.ctx == nil {
		return nil // Not started
	}
	
	// Cancel context
	r.cancel()
	
	// Close watcher
	if r.watcher != nil {
		r.watcher.Close()
	}
	
	// Stop debounce timer
	r.debounceMu.Lock()
	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
	}
	r.debounceMu.Unlock()
	
	// Wait for goroutine to finish
	r.wg.Wait()
	
	r.logger.Info("config_reloader_stopped")
	return nil
}

// watch is the main watch loop
func (r *Reloader) watch() {
	defer r.wg.Done()
	
	for {
		select {
		case <-r.ctx.Done():
			return
			
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}
			
			// Check if event is for our config file
			if filepath.Base(event.Name) != filepath.Base(r.configPath) {
				continue
			}
			
			// Handle write or create events
			if event.Op&fsnotify.Write == fsnotify.Write ||
			   event.Op&fsnotify.Create == fsnotify.Create ||
			   event.Op&fsnotify.Rename == fsnotify.Rename {
				r.logger.Info("config_file_changed", "op", event.Op.String())
				r.triggerReload()
			}
			
		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			r.logger.Error("watcher_error", "error", err)
		}
	}
}

// triggerReload triggers a debounced reload
func (r *Reloader) triggerReload() {
	r.debounceMu.Lock()
	defer r.debounceMu.Unlock()
	
	// Stop existing timer if any
	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
	}
	
	// Start new timer
	r.debounceTimer = time.AfterFunc(r.debounceDelay, func() {
		r.reload()
	})
}

// reload performs the actual config reload
func (r *Reloader) reload() {
	// Load new config
	newCfg, err := r.loadConfig()
	if err != nil {
		r.logger.Error("config_reload_failed", "error", err)
		return
	}
	
	// Get old config for diff
	r.mu.RLock()
	oldCfg := r.config
	r.mu.RUnlock()
	
	// Update config
	r.mu.Lock()
	r.config = newCfg
	r.mu.Unlock()
	
	r.logger.Info("config_reloaded", 
		"path", r.configPath,
		"target_domain", newCfg.TargetDomain,
		"hits_per_minute", newCfg.HitsPerMinute)
	
	// Notify callbacks
	r.notifyCallbacks(newCfg, oldCfg)
}

// loadConfig loads config from file
func (r *Reloader) loadConfig() (*Config, error) {
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	cfg.ApplyDefaults()
	cfg.ComputeDerived()
	
	return &cfg, nil
}

// notifyCallbacks calls all registered callbacks
func (r *Reloader) notifyCallbacks(newCfg, oldCfg *Config) {
	r.cbMu.RLock()
	callbacks := make([]ChangeCallback, len(r.callbacks))
	copy(callbacks, r.callbacks)
	r.cbMu.RUnlock()
	
	for _, cb := range callbacks {
		// Run callbacks in goroutine to prevent blocking
		go func(callback ChangeCallback) {
			defer func() {
				if rec := recover(); rec != nil {
					r.logger.Error("callback_panic", "recover", rec)
				}
			}()
			callback(newCfg)
		}(cb)
	}
}

// ApplyDefaults applies default values to config
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
	if c.DeviceType == "" {
		c.DeviceType = "mixed"
	}
	validDeviceTypes := map[string]bool{"desktop": true, "mobile": true, "tablet": true, "mixed": true}
	if !validDeviceTypes[c.DeviceType] {
		c.DeviceType = "mixed"
	}
	if c.MinPageDuration <= 0 {
		c.MinPageDuration = 15
	}
	if c.MaxPageDuration <= 0 {
		c.MaxPageDuration = 120
	}
	if c.MinScrollPercent <= 0 {
		c.MinScrollPercent = 25
	}
	if c.MaxScrollPercent <= 0 {
		c.MaxScrollPercent = 100
	}
	if c.ClickProbability <= 0 {
		c.ClickProbability = 30
	}
	if c.SessionMinPages <= 0 {
		c.SessionMinPages = 2
	}
	if c.SessionMaxPages <= 0 {
		c.SessionMaxPages = 5
	}
	if c.TargetBounceRate <= 0 {
		c.TargetBounceRate = 35
	}
	if c.ReturningVisitorRate <= 0 {
		c.ReturningVisitorRate = 30
	}
	if c.ReturningVisitorDays <= 0 {
		c.ReturningVisitorDays = 7
	}
	if c.BrowserProfilePath == "" {
		c.BrowserProfilePath = "./browser_profiles"
	}
	if c.MaxBrowserProfiles <= 0 {
		c.MaxBrowserProfiles = 100
	}
	if c.TlsFingerprintMode == "" {
		c.TlsFingerprintMode = "random"
	}
	if c.ProxyRotationMode == "" {
		c.ProxyRotationMode = "round-robin"
	}
	if c.ProxyRotationInterval <= 0 {
		c.ProxyRotationInterval = 1
	}
	if c.Http2FingerprintMode == "" {
		c.Http2FingerprintMode = "random"
	}
	if c.SerpSearchEngine == "" {
		c.SerpSearchEngine = "google"
	}
	if c.BrowserPoolMin <= 0 {
		c.BrowserPoolMin = 2
	}
	if c.BrowserPoolMax <= 0 {
		c.BrowserPoolMax = 10
	}
	if c.WorkerQueueSize <= 0 {
		c.WorkerQueueSize = 10000
	}
	if c.IosDeviceModel == "" {
		c.IosDeviceModel = "random"
	}
	if c.AndroidDeviceModel == "" {
		c.AndroidDeviceModel = "random"
	}
	if c.VisitTimeout <= 0 {
		c.VisitTimeout = 90
	}
	if c.PageLoadWait <= 0 {
		c.PageLoadWait = 1500
	}
	if c.RetryCount <= 0 {
		c.RetryCount = 3
	}
	if c.ReferrerSource == "" {
		c.ReferrerSource = "google"
	}
	if !c.BrowserPoolEnabled {
		c.BrowserPoolEnabled = true
	}
	if c.BrowserPoolSize <= 0 {
		c.BrowserPoolSize = c.MaxConcurrentVisits
	}
	if c.BrowserPoolMinSize <= 0 {
		c.BrowserPoolMinSize = 2
	}
	if c.BrowserMaxSessions <= 0 {
		c.BrowserMaxSessions = 50
	}
	if c.BrowserMaxAge <= 0 {
		c.BrowserMaxAge = 30
	}
	if c.ConnectionTimeout <= 0 {
		c.ConnectionTimeout = 15
	}
	if c.MaxIdleConnections <= 0 {
		c.MaxIdleConnections = 50
	}
	if !c.EnableKeepAlive {
		c.EnableKeepAlive = true
	}
	if !c.EnableBufferPool {
		c.EnableBufferPool = true
	}
	if c.MaxMemoryMB <= 0 {
		c.MaxMemoryMB = 512
	}
	
	c.TargetDomain = strings.TrimSpace(strings.TrimPrefix(c.TargetDomain, "https://"))
	c.TargetDomain = strings.TrimPrefix(c.TargetDomain, "http://")
	c.TargetDomain = strings.TrimSuffix(c.TargetDomain, "/")
}

// ComputeDerived computes derived values
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
}

func buildProxyURL(host string, port int, user, pass string) string {
	if user != "" || pass != "" {
		return fmt.Sprintf("http://%s:%s@%s:%d", user, pass, host, port)
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}

// Diff returns the differences between two configs
func Diff(oldCfg, newCfg *Config) map[string]struct{ Old, New interface{} } {
	diff := make(map[string]struct{ Old, New interface{} })
	
	if oldCfg == nil || newCfg == nil {
		return diff
	}
	
	// Check important fields
	if oldCfg.TargetDomain != newCfg.TargetDomain {
		diff["target_domain"] = struct{ Old, New interface{} }{oldCfg.TargetDomain, newCfg.TargetDomain}
	}
	if oldCfg.HitsPerMinute != newCfg.HitsPerMinute {
		diff["hits_per_minute"] = struct{ Old, New interface{} }{oldCfg.HitsPerMinute, newCfg.HitsPerMinute}
	}
	if oldCfg.DurationMinutes != newCfg.DurationMinutes {
		diff["duration_minutes"] = struct{ Old, New interface{} }{oldCfg.DurationMinutes, newCfg.DurationMinutes}
	}
	if oldCfg.MaxPages != newCfg.MaxPages {
		diff["max_pages"] = struct{ Old, New interface{} }{oldCfg.MaxPages, newCfg.MaxPages}
	}
	if oldCfg.ProxyEnabled != newCfg.ProxyEnabled {
		diff["proxy_enabled"] = struct{ Old, New interface{} }{oldCfg.ProxyEnabled, newCfg.ProxyEnabled}
	}
	if oldCfg.LogLevel != newCfg.LogLevel {
		diff["log_level"] = struct{ Old, New interface{} }{oldCfg.LogLevel, newCfg.LogLevel}
	}
	
	return diff
}
