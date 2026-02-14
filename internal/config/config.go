package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// PrivateProxy kullanıcının kendi proxy'si
type PrivateProxy struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Pass     string `yaml:"pass" json:"pass"`
	Protocol string `yaml:"protocol" json:"protocol"` // http, socks5
}

// ToURL proxy URL'i oluşturur
func (p *PrivateProxy) ToURL() string {
	if p.Host == "" || p.Port <= 0 {
		return ""
	}
	protocol := p.Protocol
	if protocol == "" {
		protocol = "http"
	}
	hostPort := fmt.Sprintf("%s:%d", p.Host, p.Port)
	if p.User != "" || p.Pass != "" {
		userInfo := url.UserPassword(p.User, p.Pass)
		return fmt.Sprintf("%s://%s@%s", protocol, userInfo.String(), hostPort)
	}
	return fmt.Sprintf("%s://%s", protocol, hostPort)
}

// Key benzersiz proxy anahtarı
func (p *PrivateProxy) Key() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

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
	// Private proxy listesi (kullanıcının kendi proxy'leri)
	PrivateProxies   []PrivateProxy `yaml:"private_proxies"`
	UsePrivateProxy  bool           `yaml:"use_private_proxy"` // Private proxy modu aktif mi
	// Cihaz emülasyonu ayarları
	DeviceType         string   `yaml:"device_type"`          // "desktop", "mobile", "tablet", "mixed"
	DeviceBrands       []string `yaml:"device_brands"`        // ["apple", "samsung", "google", "windows", "linux"]
	// Referrer ayarları
	ReferrerKeyword    string   `yaml:"referrer_keyword"`     // Google arama referrer için kelime
	ReferrerEnabled    bool     `yaml:"referrer_enabled"`     // Referrer simülasyonu aktif mi
	ReferrerSource     string   `yaml:"referrer_source"`      // google, bing, yahoo, duckduckgo, mixed, direct
	
	// Traffic Simulation Settings
	MinPageDuration    int      `yaml:"min_page_duration"`    // Minimum sayfa süresi (saniye)
	MaxPageDuration    int      `yaml:"max_page_duration"`    // Maximum sayfa süresi (saniye)
	MinScrollPercent   int      `yaml:"min_scroll_percent"`   // Minimum scroll yüzdesi
	MaxScrollPercent   int      `yaml:"max_scroll_percent"`   // Maximum scroll yüzdesi
	ClickProbability   int      `yaml:"click_probability"`    // Tıklama olasılığı (%)
	
	// Session Depth Simulation
	SessionMinPages      int  `yaml:"session_min_pages"`      // Oturum başına minimum sayfa
	SessionMaxPages      int  `yaml:"session_max_pages"`      // Oturum başına maximum sayfa
	EnableSessionDepth   bool `yaml:"enable_session_depth"`   // Session depth aktif mi
	
	// Bounce Rate Control
	TargetBounceRate     int  `yaml:"target_bounce_rate"`     // Hedef bounce rate (%)
	EnableBounceControl  bool `yaml:"enable_bounce_control"`  // Bounce kontrolü aktif mi
	
	// Behavior Simulation
	SimulateMouseMove    bool `yaml:"simulate_mouse_move"`    // Mouse hareketi simülasyonu
	SimulateKeyboard     bool `yaml:"simulate_keyboard"`      // Klavye simülasyonu
	SimulateClicks       bool `yaml:"simulate_clicks"`        // Tıklama simülasyonu
	SimulateFocus        bool `yaml:"simulate_focus"`         // Focus/blur simülasyonu
	
	// Geo Location
	GeoCountry           string `yaml:"geo_country"`          // Ülke kodu (TR, US, GB, etc.)
	GeoTimezone          string `yaml:"geo_timezone"`         // Saat dilimi
	GeoLanguage          string `yaml:"geo_language"`         // Dil kodu (tr-TR, en-US, etc.)
	
	// Analytics Events
	SendPageView         bool `yaml:"send_page_view"`         // Page view eventi gönder
	SendSessionStart     bool `yaml:"send_session_start"`     // Session start eventi gönder
	SendUserEngagement   bool `yaml:"send_user_engagement"`   // User engagement eventi gönder
	SendFirstVisit       bool `yaml:"send_first_visit"`       // First visit eventi gönder
	
	// Custom Dimensions & Metrics (GA4)
	CustomDimensions       string `yaml:"custom_dimensions"`       // JSON formatında custom dimensions
	CustomMetrics          string `yaml:"custom_metrics"`          // JSON formatında custom metrics
	EnableCustomDimensions bool   `yaml:"enable_custom_dimensions"` // Custom dimensions aktif mi
	
	// GSC Integration
	GscPropertyUrl       string `yaml:"gsc_property_url"`       // GSC property URL
	GscApiKey            string `yaml:"gsc_api_key"`            // GSC API key (JSON)
	EnableGscIntegration bool   `yaml:"enable_gsc_integration"` // GSC entegrasyonu aktif mi
	UseGscQueries        bool   `yaml:"use_gsc_queries"`        // GSC sorgularını kullan
	
	// Returning Visitor Simulation
	ReturningVisitorRate   int  `yaml:"returning_visitor_rate"`   // Returning visitor oranı (%)
	ReturningVisitorDays   int  `yaml:"returning_visitor_days"`   // Tekrar ziyaret aralığı (gün)
	EnableReturningVisitor bool `yaml:"enable_returning_visitor"` // Returning visitor aktif mi
	
	// Exit Page Control
	ExitPages              []string `yaml:"exit_pages"`              // Çıkış sayfaları
	EnableExitPageControl  bool     `yaml:"enable_exit_page_control"` // Exit page kontrolü aktif mi
	
	// Browser Profile Persistence
	BrowserProfilePath     string `yaml:"browser_profile_path"`     // Profil kayıt dizini
	MaxBrowserProfiles     int    `yaml:"max_browser_profiles"`     // Max profil sayısı
	EnableBrowserProfile   bool   `yaml:"enable_browser_profile"`   // Profil persistence aktif mi
	PersistCookies         bool   `yaml:"persist_cookies"`          // Cookie'leri kaydet
	PersistLocalStorage    bool   `yaml:"persist_local_storage"`    // LocalStorage kaydet
	
	// Advanced Session Management
	SessionPersistence      bool   `yaml:"session_persistence"`       // Session persistence aktif mi
	SessionStoragePath      string `yaml:"session_storage_path"`      // Session kayıt dizini
	SessionEncryption       bool   `yaml:"session_encryption"`        // Session şifreleme aktif mi
	SessionEncryptionKey    string `yaml:"session_encryption_key"`    // Session şifreleme anahtarı
	SessionTTLHours         int    `yaml:"session_ttl_hours"`         // Session TTL (saat)
	SessionIndexedDBPersist bool   `yaml:"session_indexeddb_persist"` // IndexedDB persistence
	SessionCanvasFingerprint bool  `yaml:"session_canvas_fingerprint"` // Canvas fingerprint kullan
	
	// TLS Fingerprint Randomization
	TlsFingerprintMode     string `yaml:"tls_fingerprint_mode"`     // random, chrome, firefox, safari, edge
	EnableJa3Randomization bool   `yaml:"enable_ja3_randomization"` // JA3 randomization aktif mi
	EnableJa4Randomization bool   `yaml:"enable_ja4_randomization"` // JA4 randomization aktif mi
	
	// Proxy Rotation
	ProxyRotationMode      string `yaml:"proxy_rotation_mode"`      // round-robin, random, least-used, fastest
	ProxyRotationInterval  int    `yaml:"proxy_rotation_interval"`  // Rotasyon aralığı (istek)
	EnableProxyRotation    bool   `yaml:"enable_proxy_rotation"`    // Proxy rotasyonu aktif mi
	
	// Headless Detection Evasion
	EnableHttp2Fingerprint   bool   `yaml:"enable_http2_fingerprint"`   // HTTP/2 fingerprint aktif mi
	EnableHttp3Fingerprint   bool   `yaml:"enable_http3_fingerprint"`   // HTTP/3 (QUIC) aktif mi
	Http2FingerprintMode     string `yaml:"http2_fingerprint_mode"`     // random, chrome, firefox, safari, edge
	EnableClientHints        bool   `yaml:"enable_client_hints"`        // Client hints aktif mi
	SpoofSecChUa             bool   `yaml:"spoof_sec_ch_ua"`            // Sec-CH-UA spoof
	SpoofSecChUaPlatform     bool   `yaml:"spoof_sec_ch_ua_platform"`   // Platform spoof
	SpoofSecChUaArch         bool   `yaml:"spoof_sec_ch_ua_arch"`       // Architecture spoof
	
	// Headless Detection Bypass
	BypassPuppeteer          bool `yaml:"bypass_puppeteer"`           // Puppeteer bypass
	BypassPlaywright         bool `yaml:"bypass_playwright"`          // Playwright bypass
	BypassSelenium           bool `yaml:"bypass_selenium"`            // Selenium bypass
	BypassCDP                bool `yaml:"bypass_cdp"`                 // CDP masking
	BypassRenderingDetection bool `yaml:"bypass_rendering_detection"` // Rendering detection bypass
	BypassWebdriver          bool `yaml:"bypass_webdriver"`           // Webdriver gizle
	
	// SERP Click Simulation
	SerpSearchEngine       string `yaml:"serp_search_engine"`       // google, bing, yahoo, duckduckgo, yandex
	EnableSerpSimulation   bool   `yaml:"enable_serp_simulation"`   // SERP simülasyonu aktif mi
	SerpScrollBeforeClick  bool   `yaml:"serp_scroll_before_click"` // Tıklamadan önce scroll
	
	// Browser Pool & Worker Architecture
	BrowserPoolMin         int  `yaml:"browser_pool_min"`         // Min browser instance
	BrowserPoolMax         int  `yaml:"browser_pool_max"`         // Max browser instance
	EnableAutoScaling      bool `yaml:"enable_auto_scaling"`      // Auto-scaling aktif mi
	WorkerQueueSize        int  `yaml:"worker_queue_size"`        // Worker queue size
	EnablePriorityQueue    bool `yaml:"enable_priority_queue"`    // Priority queue aktif mi
	EnableFailureRecovery  bool `yaml:"enable_failure_recovery"`  // Failure recovery aktif mi
	
	// Mobile Emulation
	IosDeviceModel         string `yaml:"ios_device_model"`         // iOS cihaz modeli
	EnableIosSafari        bool   `yaml:"enable_ios_safari"`        // iOS Safari aktif mi
	EnableIosHaptics       bool   `yaml:"enable_ios_haptics"`       // Haptic feedback
	AndroidDeviceModel     string `yaml:"android_device_model"`     // Android cihaz modeli
	EnableAndroidChrome    bool   `yaml:"enable_android_chrome"`    // Android Chrome aktif mi
	EnableAndroidVibration bool   `yaml:"enable_android_vibration"` // Vibration API
	
	// Touch Event Simulation
	EnableTouchEvents      bool `yaml:"enable_touch_events"`      // Touch events aktif mi
	EnableMultiTouch       bool `yaml:"enable_multi_touch"`       // Multi-touch aktif mi
	EnableGestures         bool `yaml:"enable_gestures"`          // Gestures (swipe/pinch) aktif mi
	EnableMobileKeyboard   bool `yaml:"enable_mobile_keyboard"`   // Mobile keyboard aktif mi
	
	// Sensor Spoofing
	EnableAccelerometer    bool `yaml:"enable_accelerometer"`     // Accelerometer aktif mi
	EnableGyroscope        bool `yaml:"enable_gyroscope"`         // Gyroscope aktif mi
	EnableDeviceOrientation bool `yaml:"enable_device_orientation"` // Device orientation aktif mi
	EnableMagnetometer     bool `yaml:"enable_magnetometer"`      // Magnetometer aktif mi
	
	// Tablet-Specific Behavior
	EnableTabletMode       bool `yaml:"enable_tablet_mode"`       // Tablet mode aktif mi
	EnableLandscapeMode    bool `yaml:"enable_landscape_mode"`    // Landscape mode aktif mi
	EnablePenInput         bool `yaml:"enable_pen_input"`         // Pen/stylus input aktif mi
	EnableSplitView        bool `yaml:"enable_split_view"`        // Split view aktif mi
	
	// Stealth & Anti-Detect
	StealthWebdriver       bool `yaml:"stealth_webdriver"`        // Webdriver gizle
	StealthChrome          bool `yaml:"stealth_chrome"`           // Chrome gizle
	StealthPlugins         bool `yaml:"stealth_plugins"`          // Plugin spoof
	StealthWebGL           bool `yaml:"stealth_webgl"`            // WebGL noise
	StealthAudio           bool `yaml:"stealth_audio"`            // Audio noise
	StealthCanvas          bool `yaml:"stealth_canvas"`           // Canvas noise
	StealthTimezone        bool `yaml:"stealth_timezone"`         // Timezone spoof
	StealthLanguage        bool `yaml:"stealth_language"`         // Language spoof
	
	// Performance Settings
	VisitTimeout           int  `yaml:"visit_timeout"`            // Ziyaret timeout (saniye)
	PageLoadWait           int  `yaml:"page_load_wait"`           // Sayfa yükleme bekleme (ms)
	RetryCount             int  `yaml:"retry_count"`              // Hata tekrar sayısı
	
	// Resource Blocking
	BlockImages            bool `yaml:"block_images"`             // Görselleri engelle
	BlockStyles            bool `yaml:"block_styles"`             // CSS engelle
	BlockFonts             bool `yaml:"block_fonts"`              // Fontları engelle
	BlockMedia             bool `yaml:"block_media"`              // Medyayı engelle
	
	// Anti-Detect Mode (master switch)
	AntiDetectMode         bool `yaml:"anti_detect_mode"`         // Anti-detect modu aktif mi
	
	// PERFORMANCE SETTINGS
	BrowserPoolEnabled     bool `yaml:"browser_pool_enabled"`       // Browser pool kullan (performans)
	BrowserPoolSize        int  `yaml:"browser_pool_size"`          // Pool boyutu
	BrowserPoolMinSize     int  `yaml:"browser_pool_min_size"`      // Min pool boyutu
	BrowserMaxSessions     int  `yaml:"browser_max_sessions"`       // Instance basina max session
	BrowserMaxAge          int  `yaml:"browser_max_age"`            // Instance max yasi (dk)
	ConnectionTimeout      int  `yaml:"connection_timeout"`         // Connection timeout (sn)
	MaxIdleConnections     int  `yaml:"max_idle_connections"`       // Max idle connection
	EnableKeepAlive        bool `yaml:"enable_keep_alive"`          // HTTP keep-alive
	EnableBufferPool       bool `yaml:"enable_buffer_pool"`         // Buffer pool kullan
	MaxMemoryMB            int  `yaml:"max_memory_mb"`              // Max memory limit (MB)
	
	// NETWORK OPTIMIZATIONS
	EnableHTTP3            bool   `yaml:"enable_http3"`               // HTTP/3 QUIC support
	EnableConnectionPool   bool   `yaml:"enable_connection_pool"`     // Connection pooling
	EnableTCPFastOpen      bool   `yaml:"enable_tcp_fast_open"`       // TCP Fast Open
	ConnectionPoolSize     int    `yaml:"connection_pool_size"`       // Pool size
	MaxConnsPerHost        int    `yaml:"max_conns_per_host"`         // Max connections per host
	MaxIdleConns           int    `yaml:"max_idle_conns"`             // Max idle connections (alias for MaxIdleConnections)
	UseProxy               bool   `yaml:"use_proxy"`                  // Use proxy (alias for ProxyEnabled)
	
	// SYSTEM OPTIMIZATIONS  
	EnableCPUAffinity      bool   `yaml:"enable_cpu_affinity"`        // CPU affinity
	EnableNUMA             bool   `yaml:"enable_numa"`                // NUMA awareness
	CPUAffinityCores       []int  `yaml:"cpu_affinity_cores"`         // CPU cores to use
	NUMANodes              []int  `yaml:"numa_nodes"`                 // NUMA nodes to use
	
	// VM SPOOFING
	EnableVMSpoofing       bool   `yaml:"enable_vm_spoofing"`         // VM spoofing
	VMType                 string `yaml:"vm_type"`                    // VM type
	HideVMIndicators       bool   `yaml:"hide_vm_indicators"`         // Hide VM indicators
	SpoofHardwareIDs       bool   `yaml:"spoof_hardware_ids"`         // Spoof hardware IDs
	RandomizeVMParams      bool   `yaml:"randomize_vm_params"`        // Randomize VM params
	
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
	// Cihaz tipi varsayılanı
	if c.DeviceType == "" {
		c.DeviceType = "mixed"
	}
	// Geçerli cihaz tipleri kontrolü
	validDeviceTypes := map[string]bool{"desktop": true, "mobile": true, "tablet": true, "mixed": true}
	if !validDeviceTypes[c.DeviceType] {
		c.DeviceType = "mixed"
	}
	
	// Traffic Simulation Defaults
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
	
	// Session Depth Defaults
	if c.SessionMinPages <= 0 {
		c.SessionMinPages = 2
	}
	if c.SessionMaxPages <= 0 {
		c.SessionMaxPages = 5
	}
	
	// Bounce Rate Defaults
	if c.TargetBounceRate <= 0 {
		c.TargetBounceRate = 35
	}
	
	// Returning Visitor Defaults
	if c.ReturningVisitorRate <= 0 {
		c.ReturningVisitorRate = 30
	}
	if c.ReturningVisitorDays <= 0 {
		c.ReturningVisitorDays = 7
	}
	
	// Browser Profile Defaults
	if c.BrowserProfilePath == "" {
		c.BrowserProfilePath = "./browser_profiles"
	}
	if c.MaxBrowserProfiles <= 0 {
		c.MaxBrowserProfiles = 100
	}
	
	// Advanced Session Management Defaults
	if c.SessionStoragePath == "" {
		c.SessionStoragePath = "./sessions"
	}
	if c.SessionTTLHours <= 0 {
		c.SessionTTLHours = 168 // 7 days
	}
	
	// TLS Fingerprint Defaults
	if c.TlsFingerprintMode == "" {
		c.TlsFingerprintMode = "random"
	}
	
	// Proxy Rotation Defaults
	if c.ProxyRotationMode == "" {
		c.ProxyRotationMode = "round-robin"
	}
	if c.ProxyRotationInterval <= 0 {
		c.ProxyRotationInterval = 1
	}
	
	// HTTP/2 Fingerprint Defaults
	if c.Http2FingerprintMode == "" {
		c.Http2FingerprintMode = "random"
	}
	
	// SERP Defaults
	if c.SerpSearchEngine == "" {
		c.SerpSearchEngine = "google"
	}
	
	// Browser Pool Defaults
	if c.BrowserPoolMin <= 0 {
		c.BrowserPoolMin = 2
	}
	if c.BrowserPoolMax <= 0 {
		c.BrowserPoolMax = 10
	}
	if c.WorkerQueueSize <= 0 {
		c.WorkerQueueSize = 10000
	}
	
	// Mobile Emulation Defaults
	if c.IosDeviceModel == "" {
		c.IosDeviceModel = "random"
	}
	if c.AndroidDeviceModel == "" {
		c.AndroidDeviceModel = "random"
	}
	
	// Performance Defaults
	if c.VisitTimeout <= 0 {
		c.VisitTimeout = 90
	}
	if c.PageLoadWait <= 0 {
		c.PageLoadWait = 1500
	}
	if c.RetryCount <= 0 {
		c.RetryCount = 3
	}
	
	// Referrer Source Default
	if c.ReferrerSource == "" {
		c.ReferrerSource = "google"
	}
	
	// PERFORMANCE: Varsayilan performans ayarlari
	// CONFIG FIX: Boolean değerler için doğru varsayılan mantık
	// Not: Boolean'lar için "if !value" kullanmak yanlış - her zaman true yapar
	// Bunun yerine sadece ilk kez ayarlanmamışsa (zero value) varsayılan ata
	
	// BrowserPoolEnabled varsayılan olarak true olmalı - bu özel bir durum
	// Kullanıcı açıkça false yapmadıysa true olsun
	// NOT: YAML/JSON'dan false gelirse bu değer korunmalı, bu yüzden bu mantık kaldırıldı
	// Varsayılan true için config struct'ta default tag kullanılmalı
	
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
		c.BrowserMaxAge = 30 // dakika
	}
	if c.ConnectionTimeout <= 0 {
		c.ConnectionTimeout = 15 // saniye
	}
	if c.MaxIdleConnections <= 0 {
		c.MaxIdleConnections = 50
	}
	// CONFIG FIX: Boolean varsayılanlar kaldırıldı - kullanıcı tercihi korunmalı
	// EnableKeepAlive, EnableBufferPool gibi değerler kullanıcı tarafından ayarlanmalı
	if c.MaxMemoryMB <= 0 {
		c.MaxMemoryMB = 512
	}
	
	// NETWORK OPTIMIZATIONS defaults
	// CONFIG FIX: Boolean değerler için yanlış mantık düzeltildi
	// "if !c.EnableHTTP3 { c.EnableHTTP3 = false }" anlamsız - zaten false
	if c.ConnectionPoolSize <= 0 {
		c.ConnectionPoolSize = 100
	}
	if c.MaxConnsPerHost <= 0 {
		c.MaxConnsPerHost = 20
	}
	
	// SYSTEM OPTIMIZATIONS defaults
	// CONFIG FIX: Anlamsız boolean kontroller kaldırıldı
	// "if !c.EnableCPUAffinity { c.EnableCPUAffinity = false }" zaten false, anlamsız
	// Bu değerler varsayılan olarak false kalmalı, kullanıcı açıkça aktifleştirmeli
	
	// VM SPOOFING defaults
	// CONFIG FIX: Sadece string varsayılanları ayarla
	if c.VMType == "" {
		c.VMType = "none"
	}
	// CONFIG FIX: HideVMIndicators, SpoofHardwareIDs, RandomizeVMParams
	// Bu değerler güvenlik için varsayılan true olmalı
	// Ancak "if !value { value = true }" her zaman true yapar - bu yanlış
	// Çözüm: Bu değerler struct tanımında varsayılan true olarak işaretlenmeli
	// veya NewConfig() fonksiyonu ile başlatılmalı
	// Şimdilik bu mantığı kaldırıyoruz - kullanıcı config'den ayarlamalı
	
	// Domain'den scheme ve trailing slash temizle, path varsa koru
	c.TargetDomain = strings.TrimSpace(c.TargetDomain)
	c.TargetDomain = strings.TrimPrefix(c.TargetDomain, "https://")
	c.TargetDomain = strings.TrimPrefix(c.TargetDomain, "http://")
	c.TargetDomain = strings.TrimPrefix(c.TargetDomain, "//")
	c.TargetDomain = strings.TrimSuffix(c.TargetDomain, "/")
	// Path varsa sadece domain kısmını al
	if idx := strings.Index(c.TargetDomain, "/"); idx > 0 {
		c.TargetDomain = c.TargetDomain[:idx]
	}
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
	// SECURITY FIX: Removed hardcoded GA ID - user must provide their own
	// GtagID is now optional and will be empty if not configured
}
