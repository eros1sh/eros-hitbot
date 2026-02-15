package i18n

import "fmt"

// Msg log mesajı anahtarları
const (
	MsgStarting        = "starting"
	MsgTarget          = "target"
	MsgDiscovery       = "discovery"
	MsgDiscoveryErr    = "discovery_err"
	MsgPagesFound      = "pages_found"
	MsgCancel          = "cancel"
	MsgDeadline        = "deadline"
	MsgVisitErr        = "visit_err"
	MsgVisitErrSummary = "visit_err_summary"
	MsgProgress        = "progress"
	MsgSummary         = "summary"
	MsgSummaryLine     = "summary_line"
	MsgSummaryRT       = "summary_rt"
	MsgExportErr       = "export_err"
	MsgReportCSV       = "report_csv"
	MsgReportJSON      = "report_json"
	MsgReportHTML      = "report_html"
	MsgSitemapFound    = "sitemap_found"
	MsgSitemapNone     = "sitemap_none"
	// v2.2.0 - New messages
	MsgProxyFetch      = "proxy_fetch"
	MsgProxyFetchErr   = "proxy_fetch_err"
	MsgProxyAdded      = "proxy_added"
	MsgProxyLive       = "proxy_live"
	MsgDeviceType      = "device_type"
	MsgReferrerSet     = "referrer_set"
	MsgGeoLocation     = "geo_location"
	MsgStealthEnabled  = "stealth_enabled"
	MsgAnalyticsEvent  = "analytics_event"
	MsgQualityScore    = "quality_score"
	// v2.3.0 - New messages
	MsgGscIntegration     = "gsc_integration"
	MsgGscQueriesFetched  = "gsc_queries_fetched"
	MsgGscQueryError      = "gsc_query_error"
	MsgBounceRateControl  = "bounce_rate_control"
	MsgSessionDepth       = "session_depth"
	MsgReturningVisitor   = "returning_visitor"
	MsgExitPageControl    = "exit_page_control"
	MsgBrowserProfile     = "browser_profile"
	MsgTlsFingerprint     = "tls_fingerprint"
	MsgCustomDimensions   = "custom_dimensions"
	MsgProxyRotation      = "proxy_rotation"
	MsgMultiProxy         = "multi_proxy"
	MsgJa3Randomization   = "ja3_randomization"
	MsgJa4Randomization   = "ja4_randomization"
	MsgProfilePersistence = "profile_persistence"
	MsgCookiePersistence  = "cookie_persistence"
	// v2.4.0 - Startup flow messages
	MsgSelectLanguage         = "select_language"
	MsgLanguageTurkish        = "language_turkish"
	MsgLanguageEnglish        = "language_english"
	MsgSelection              = "selection"
	MsgDetectingSystem        = "detecting_system"
	MsgSystemInfo             = "system_info"
	MsgRecommendedSettings    = "recommended_settings"
	MsgManualSettings         = "manual_settings"
	MsgSettingsQuestion       = "settings_question"
	MsgApplyingOptimization   = "applying_optimization"
	MsgOptimizationApplied    = "optimization_applied"
	MsgOptimizationCancelled  = "optimization_cancelled"
	MsgOpeningBrowser         = "opening_browser"
	MsgServerShutdown         = "server_shutdown"
	MsgServerShutdownComplete = "server_shutdown_complete"
	MsgServerError            = "server_error"
	MsgShutdownError          = "shutdown_error"
	MsgInvalidURL             = "invalid_url"
	MsgSecurityHTTPOnly       = "security_http_only"
	MsgSecurityLocalhost      = "security_localhost"
	MsgError                  = "error"
	MsgWarning                = "warning"
	// v2.4.0 - CLI messages
	MsgCLIMode              = "cli_mode"
	MsgCLITarget            = "cli_target"
	MsgCLIDuration          = "cli_duration"
	MsgCLIStopHint          = "cli_stop_hint"
	MsgCLIAutoOptimize      = "cli_auto_optimize"
	MsgCLIApplySettings     = "cli_apply_settings"
	MsgCLIConfigRequired    = "cli_config_required"
	MsgCLIExample           = "cli_example"
	MsgCLIFlags             = "cli_flags"
	MsgCLIFlagCli           = "cli_flag_cli"
	MsgCLIFlagSysinfo       = "cli_flag_sysinfo"
	MsgCLIFlagOptimize      = "cli_flag_optimize"
	MsgCLIFlagDomain        = "cli_flag_domain"
	MsgCLIFlagPages         = "cli_flag_pages"
	MsgCLIFlagDurationFlag  = "cli_flag_duration"
	MsgCLIFlagHpm           = "cli_flag_hpm"
	MsgCLIFlagConcurrent    = "cli_flag_concurrent"
	MsgSimulationError      = "simulation_error"
	// v2.4.0 - System info messages
	MsgSysOS              = "sys_os"
	MsgSysKernel          = "sys_kernel"
	MsgSysUptime          = "sys_uptime"
	MsgSysShell           = "sys_shell"
	MsgSysResolution      = "sys_resolution"
	MsgSysTerminal        = "sys_terminal"
	MsgSysCPU             = "sys_cpu"
	MsgSysCPUCores        = "sys_cpu_cores"
	MsgSysGPU             = "sys_gpu"
	MsgSysMemory          = "sys_memory"
	MsgSysDisk            = "sys_disk"
	MsgSysGoVersion       = "sys_go_version"
	MsgSysDays            = "sys_days"
	MsgSysHours           = "sys_hours"
	MsgSysMinutes         = "sys_minutes"
	// v2.4.0 - Optimization profile messages
	MsgOptProfileTitle       = "opt_profile_title"
	MsgOptRecommendedMode    = "opt_recommended_mode"
	MsgOptModeLow            = "opt_mode_low"
	MsgOptModeMedium         = "opt_mode_medium"
	MsgOptModeHigh           = "opt_mode_high"
	MsgOptModeUltra          = "opt_mode_ultra"
	MsgOptMaxConcurrent      = "opt_max_concurrent"
	MsgOptHitsPerMinute      = "opt_hits_per_minute"
	MsgOptBrowserPool        = "opt_browser_pool"
	MsgOptWorkerQueue        = "opt_worker_queue"
	MsgOptWarnLowRAM         = "opt_warn_low_ram"
	MsgOptWarnLowCPU         = "opt_warn_low_cpu"
	MsgOptWarnHighMemory     = "opt_warn_high_memory"
	MsgOptWarnLowDisk        = "opt_warn_low_disk"
	MsgOptRecResourceBlock   = "opt_rec_resource_block"
	MsgOptRecMediumSystem    = "opt_rec_medium_system"
	MsgOptRecGoodSystem      = "opt_rec_good_system"
	MsgOptRecPowerfulSystem  = "opt_rec_powerful_system"
	MsgOptRecStrongCPU       = "opt_rec_strong_cpu"
	MsgOptRecWindows         = "opt_rec_windows"
	MsgOptRecLinux           = "opt_rec_linux"
	MsgOptRecMacOS           = "opt_rec_macos"
	MsgOptApplyHint          = "opt_apply_hint"
	// v2.4.0 - Private proxy messages
	MsgPrivateProxyActive = "private_proxy_active"
	// v2.4.0 - Web interface messages
	MsgWebInterface = "web_interface"
	MsgOpenBrowser  = "open_browser"
	MsgStopHint     = "stop_hint"
)

var tr = map[string]string{
	MsgStarting:        "Eros - Hit Botu v2.5.0 başlatılıyor...",
	MsgTarget:          "Hedef: %s | Max sayfa: %d | Süre: %d dk | HPM: %d | Paralel: %d",
	MsgDiscovery:       "Sayfa keşfi başlıyor...",
	MsgDiscoveryErr:    "Keşif hatası: %s",
	MsgPagesFound:      "%d sayfa bulundu: %v",
	MsgCancel:          "İptal sinyali alındı, kapatılıyor...",
	MsgDeadline:        "Test süresi doldu.",
	MsgVisitErr:        "Ziyaret hatası [%s]: %v",
	MsgVisitErrSummary: "%d ziyaret hatası (%s)",
	MsgProgress:        "[%d] Toplam: %d | OK: %d | Hata: %d | Ort. RT: %.0f ms",
	MsgSummary:         "--- Özet ---",
	MsgSummaryLine:     "Toplam istek: %d | Başarılı: %d | Hatalı: %d",
	MsgSummaryRT:       "Ort. yanıt: %.0f ms | Min: %d ms | Max: %d ms",
	MsgExportErr:       "Rapor export hatası: %s",
	MsgReportCSV:       "Rapor kaydedildi: %s",
	MsgReportJSON:      "Rapor kaydedildi: %s",
	MsgReportHTML:      "HTML rapor: %s",
	MsgSitemapFound:    "Sitemap bulundu: %d URL (anasayfa ağırlığı %%%d)",
	MsgSitemapNone:     "Sitemap bulunamadı, sayfa keşfi kullanılıyor.",
	// v2.2.0 - New messages
	MsgProxyFetch:      "Proxy listeleri çekiliyor...",
	MsgProxyFetchErr:   "Proxy çekme hatası: %s",
	MsgProxyAdded:      "Havuza %d proxy eklendi.",
	MsgProxyLive:       "Canlı proxy sayısı: %d",
	MsgDeviceType:      "Cihaz tipi: %s | Markalar: %v",
	MsgReferrerSet:     "Referrer ayarlandı: %s",
	MsgGeoLocation:     "Coğrafi konum: %s | Saat dilimi: %s",
	MsgStealthEnabled:  "Stealth modu aktif: %d özellik",
	MsgAnalyticsEvent:  "Analytics eventi gönderildi: %s",
	MsgQualityScore:    "Trafik kalite skoru: %s (%%%d başarı)",
	// v2.3.0 - New messages
	MsgGscIntegration:     "GSC entegrasyonu aktif: %s",
	MsgGscQueriesFetched:  "GSC'den %d sorgu çekildi",
	MsgGscQueryError:      "GSC sorgu hatası: %s",
	MsgBounceRateControl:  "Bounce rate kontrolü aktif: hedef %%%d",
	MsgSessionDepth:       "Session depth simülasyonu: %d-%d sayfa/oturum",
	MsgReturningVisitor:   "Returning visitor simülasyonu: %%%d oran, %d gün aralık",
	MsgExitPageControl:    "Exit page kontrolü aktif: %d sayfa tanımlı",
	MsgBrowserProfile:     "Browser profil persistence aktif: %s dizini, max %d profil",
	MsgTlsFingerprint:     "TLS fingerprint randomization: %s modu",
	MsgCustomDimensions:   "Custom dimensions/metrics gönderiliyor",
	MsgProxyRotation:      "Proxy rotasyonu: %s modu, %d istek aralığı",
	MsgMultiProxy:         "Multi-proxy aktif: %d özel proxy tanımlı",
	MsgJa3Randomization:   "JA3 fingerprint randomization aktif",
	MsgJa4Randomization:   "JA4 fingerprint randomization aktif",
	MsgProfilePersistence: "Profil persistence: cookie=%v, localStorage=%v",
	MsgCookiePersistence:  "Cookie persistence aktif",
	// v2.4.0 - Startup flow messages
	MsgSelectLanguage:         "Dil Seçin / Select Language:",
	MsgLanguageTurkish:        "Türkçe",
	MsgLanguageEnglish:        "English",
	MsgSelection:              "Seçim (1/2) [1]: ",
	MsgDetectingSystem:        "🔍 Sistem bilgileri algılanıyor...",
	MsgSystemInfo:             "Sistem Bilgileri",
	MsgRecommendedSettings:    "1 = Önerilen ayarları kullan",
	MsgManualSettings:         "2 = Kendim ayar yapmak istiyorum",
	MsgSettingsQuestion:       "Seçiminiz (1/2) [1]: ",
	MsgApplyingOptimization:   "🔧 Önerilen ayarlar uygulanıyor...",
	MsgOptimizationApplied:    "✅ Optimizasyon uygulandı:",
	MsgOptimizationCancelled:  "❌ Optimizasyon iptal edildi. Varsayılan ayarlar kullanılacak.",
	MsgOpeningBrowser:         "🌐 Tarayıcı açılıyor...",
	MsgServerShutdown:         "🛑 Sunucu kapatılıyor...",
	MsgServerShutdownComplete: "✅ Sunucu başarıyla kapatıldı.",
	MsgServerError:            "Sunucu hatası: %v",
	MsgShutdownError:          "Shutdown hatası: %v",
	MsgInvalidURL:             "Geçersiz URL: %v",
	MsgSecurityHTTPOnly:       "Güvenlik: Sadece http/https URL'leri desteklenir",
	MsgSecurityLocalhost:      "Güvenlik: Sadece localhost URL'leri açılabilir",
	MsgError:                  "Hata: %v",
	MsgWarning:                "UYARI: %s",
	// v2.4.0 - CLI messages
	MsgCLIMode:              "ErosHit - CLI Modu",
	MsgCLITarget:            "Hedef: %s",
	MsgCLIDuration:          "Süre: %d dakika | HPM: %d | Paralel: %d",
	MsgCLIStopHint:          "Durdurmak için Ctrl+C",
	MsgCLIAutoOptimize:      "🔧 Otomatik optimizasyon modu aktif...",
	MsgCLIApplySettings:     "Bu ayarları uygulamak istiyor musunuz? (E/h): ",
	MsgCLIConfigRequired:    "target_domain gerekli. -domain example.com veya config.json'da belirtin.",
	MsgCLIExample:           "Örnek: eroshit -cli -domain mysite.com -pages 5 -duration 60 -hpm 35",
	MsgCLIFlags:             "Kullanılabilir flagler:",
	MsgCLIFlagCli:           "-cli           : CLI modunda çalıştır",
	MsgCLIFlagSysinfo:       "-sysinfo       : Sistem bilgilerini göster (neofetch benzeri)",
	MsgCLIFlagOptimize:      "-optimize      : Otomatik optimizasyon profili uygula",
	MsgCLIFlagDomain:        "-domain        : Hedef domain",
	MsgCLIFlagPages:         "-pages         : Max sayfa sayısı",
	MsgCLIFlagDurationFlag:  "-duration      : Süre (dakika)",
	MsgCLIFlagHpm:           "-hpm           : İstek/dakika",
	MsgCLIFlagConcurrent:    "-concurrent    : Paralel tarayıcı sayısı",
	MsgSimulationError:      "Simülasyon hatası: %v",
	// v2.4.0 - System info messages
	MsgSysOS:         "OS:",
	MsgSysKernel:     "Kernel:",
	MsgSysUptime:     "Uptime:",
	MsgSysShell:      "Shell:",
	MsgSysResolution: "Çözünürlük:",
	MsgSysTerminal:   "Terminal:",
	MsgSysCPU:        "CPU:",
	MsgSysCPUCores:   "%d çekirdek / %d thread",
	MsgSysGPU:        "GPU:",
	MsgSysMemory:     "Bellek:",
	MsgSysDisk:       "Disk:",
	MsgSysGoVersion:  "Go Sürümü:",
	MsgSysDays:       "%d gün, %d saat, %d dakika",
	MsgSysHours:      "%d saat, %d dakika",
	MsgSysMinutes:    "%d dakika",
	// v2.4.0 - Optimization profile messages
	MsgOptProfileTitle:       "🔧 OTOMATİK OPTİMİZASYON PROFİLİ",
	MsgOptRecommendedMode:    "Önerilen Mod:",
	MsgOptModeLow:            "DÜŞÜK",
	MsgOptModeMedium:         "ORTA",
	MsgOptModeHigh:           "YÜKSEK",
	MsgOptModeUltra:          "ULTRA",
	MsgOptMaxConcurrent:      "Max Paralel Tarayıcı:",
	MsgOptHitsPerMinute:      "İstek/Dakika:",
	MsgOptBrowserPool:        "Browser Pool:",
	MsgOptWorkerQueue:        "Worker Queue:",
	MsgOptWarnLowRAM:         "⚠️  Düşük RAM: %.1f GB - Performans sınırlı olabilir",
	MsgOptWarnLowCPU:         "⚠️  Düşük CPU thread sayısı: %d",
	MsgOptWarnHighMemory:     "⚠️  Yüksek bellek kullanımı: %.1f%% - Diğer uygulamaları kapatın",
	MsgOptWarnLowDisk:        "⚠️  Düşük disk alanı: %.1f GB - Raporlar için yer açın",
	MsgOptRecResourceBlock:   "💡 Kaynak engelleme (resim, CSS, font) aktif edilmeli",
	MsgOptRecMediumSystem:    "💡 Orta seviye sistem - Dengeli ayarlar önerilir",
	MsgOptRecGoodSystem:      "💡 İyi sistem - Yüksek performans ayarları kullanılabilir",
	MsgOptRecPowerfulSystem:  "🚀 Güçlü sistem - Maksimum performans ayarları kullanılabilir",
	MsgOptRecStrongCPU:       "💪 Güçlü CPU: %d thread - Paralel işlem kapasitesi yüksek",
	MsgOptRecWindows:         "💡 Windows Defender gerçek zamanlı korumayı geçici olarak devre dışı bırakabilirsiniz",
	MsgOptRecLinux:           "💡 ulimit -n değerini artırarak daha fazla dosya tanıtıcısı kullanabilirsiniz",
	MsgOptRecMacOS:           "💡 macOS'ta 'launchctl limit maxfiles' ile dosya limiti artırılabilir",
	MsgOptApplyHint:          "💡 Bu ayarları uygulamak için: eroshit -cli -optimize",
	// v2.4.0 - Private proxy messages
	MsgPrivateProxyActive: "🔐 Private proxy modu aktif: %d proxy yüklendi",
	// v2.4.0 - Web interface messages
	MsgWebInterface: "ErosHit - Web Arayüzü",
	MsgOpenBrowser:  "Tarayıcınızda açın: %s",
	MsgStopHint:     "Durdurmak için Ctrl+C",
}

var en = map[string]string{
	MsgStarting:        "Eros - Hit Bot v2.5.0 starting...",
	MsgTarget:          "Target: %s | Max pages: %d | Duration: %d min | HPM: %d | Parallel: %d",
	MsgDiscovery:       "Page discovery starting...",
	MsgDiscoveryErr:    "Discovery error: %s",
	MsgPagesFound:      "%d pages found: %v",
	MsgCancel:          "Cancel signal received, shutting down...",
	MsgDeadline:        "Test duration expired.",
	MsgVisitErr:        "Visit error [%s]: %v",
	MsgVisitErrSummary: "%d visit errors (%s)",
	MsgProgress:        "[%d] Total: %d | OK: %d | Fail: %d | Avg RT: %.0f ms",
	MsgSummary:         "--- Summary ---",
	MsgSummaryLine:     "Total requests: %d | Success: %d | Failed: %d",
	MsgSummaryRT:       "Avg response: %.0f ms | Min: %d ms | Max: %d ms",
	MsgExportErr:       "Report export error: %s",
	MsgReportCSV:       "Report saved: %s",
	MsgReportJSON:      "Report saved: %s",
	MsgReportHTML:      "HTML report: %s",
	MsgSitemapFound:    "Sitemap found: %d URLs (homepage weight %d%%)",
	MsgSitemapNone:     "Sitemap not found, using page discovery.",
	// v2.2.0 - New messages
	MsgProxyFetch:      "Fetching proxy lists...",
	MsgProxyFetchErr:   "Proxy fetch error: %s",
	MsgProxyAdded:      "Added %d proxies to pool.",
	MsgProxyLive:       "Live proxy count: %d",
	MsgDeviceType:      "Device type: %s | Brands: %v",
	MsgReferrerSet:     "Referrer set: %s",
	MsgGeoLocation:     "Geo location: %s | Timezone: %s",
	MsgStealthEnabled:  "Stealth mode active: %d features",
	MsgAnalyticsEvent:  "Analytics event sent: %s",
	MsgQualityScore:    "Traffic quality score: %s (%d%% success)",
	// v2.3.0 - New messages
	MsgGscIntegration:     "GSC integration active: %s",
	MsgGscQueriesFetched:  "Fetched %d queries from GSC",
	MsgGscQueryError:      "GSC query error: %s",
	MsgBounceRateControl:  "Bounce rate control active: target %d%%",
	MsgSessionDepth:       "Session depth simulation: %d-%d pages/session",
	MsgReturningVisitor:   "Returning visitor simulation: %d%% rate, %d day interval",
	MsgExitPageControl:    "Exit page control active: %d pages defined",
	MsgBrowserProfile:     "Browser profile persistence active: %s directory, max %d profiles",
	MsgTlsFingerprint:     "TLS fingerprint randomization: %s mode",
	MsgCustomDimensions:   "Sending custom dimensions/metrics",
	MsgProxyRotation:      "Proxy rotation: %s mode, %d request interval",
	MsgMultiProxy:         "Multi-proxy active: %d private proxies defined",
	MsgJa3Randomization:   "JA3 fingerprint randomization active",
	MsgJa4Randomization:   "JA4 fingerprint randomization active",
	MsgProfilePersistence: "Profile persistence: cookie=%v, localStorage=%v",
	MsgCookiePersistence:  "Cookie persistence active",
	// v2.4.0 - Startup flow messages
	MsgSelectLanguage:         "Select Language / Dil Seçin:",
	MsgLanguageTurkish:        "Türkçe",
	MsgLanguageEnglish:        "English",
	MsgSelection:              "Selection (1/2) [1]: ",
	MsgDetectingSystem:        "🔍 Detecting system information...",
	MsgSystemInfo:             "System Information",
	MsgRecommendedSettings:    "1 = Use recommended settings",
	MsgManualSettings:         "2 = I want to configure manually",
	MsgSettingsQuestion:       "Your choice (1/2) [1]: ",
	MsgApplyingOptimization:   "🔧 Applying recommended settings...",
	MsgOptimizationApplied:    "✅ Optimization applied:",
	MsgOptimizationCancelled:  "❌ Optimization cancelled. Default settings will be used.",
	MsgOpeningBrowser:         "🌐 Opening browser...",
	MsgServerShutdown:         "🛑 Server shutting down...",
	MsgServerShutdownComplete: "✅ Server shut down successfully.",
	MsgServerError:            "Server error: %v",
	MsgShutdownError:          "Shutdown error: %v",
	MsgInvalidURL:             "Invalid URL: %v",
	MsgSecurityHTTPOnly:       "Security: Only http/https URLs are supported",
	MsgSecurityLocalhost:      "Security: Only localhost URLs can be opened",
	MsgError:                  "Error: %v",
	MsgWarning:                "WARNING: %s",
	// v2.4.0 - CLI messages
	MsgCLIMode:              "ErosHit - CLI Mode",
	MsgCLITarget:            "Target: %s",
	MsgCLIDuration:          "Duration: %d min | HPM: %d | Parallel: %d",
	MsgCLIStopHint:          "Press Ctrl+C to stop",
	MsgCLIAutoOptimize:      "🔧 Auto-optimization mode active...",
	MsgCLIApplySettings:     "Do you want to apply these settings? (Y/n): ",
	MsgCLIConfigRequired:    "target_domain required. Specify with -domain example.com or in config.json.",
	MsgCLIExample:           "Example: eroshit -cli -domain mysite.com -pages 5 -duration 60 -hpm 35",
	MsgCLIFlags:             "Available flags:",
	MsgCLIFlagCli:           "-cli           : Run in CLI mode",
	MsgCLIFlagSysinfo:       "-sysinfo       : Show system info (neofetch-like)",
	MsgCLIFlagOptimize:      "-optimize      : Apply auto-optimization profile",
	MsgCLIFlagDomain:        "-domain        : Target domain",
	MsgCLIFlagPages:         "-pages         : Max page count",
	MsgCLIFlagDurationFlag:  "-duration      : Duration (minutes)",
	MsgCLIFlagHpm:           "-hpm           : Requests/minute",
	MsgCLIFlagConcurrent:    "-concurrent    : Parallel browser count",
	MsgSimulationError:      "Simulation error: %v",
	// v2.4.0 - System info messages
	MsgSysOS:         "OS:",
	MsgSysKernel:     "Kernel:",
	MsgSysUptime:     "Uptime:",
	MsgSysShell:      "Shell:",
	MsgSysResolution: "Resolution:",
	MsgSysTerminal:   "Terminal:",
	MsgSysCPU:        "CPU:",
	MsgSysCPUCores:   "%d cores / %d threads",
	MsgSysGPU:        "GPU:",
	MsgSysMemory:     "Memory:",
	MsgSysDisk:       "Disk:",
	MsgSysGoVersion:  "Go Version:",
	MsgSysDays:       "%d days, %d hours, %d minutes",
	MsgSysHours:      "%d hours, %d minutes",
	MsgSysMinutes:    "%d minutes",
	// v2.4.0 - Optimization profile messages
	MsgOptProfileTitle:       "🔧 AUTO-OPTIMIZATION PROFILE",
	MsgOptRecommendedMode:    "Recommended Mode:",
	MsgOptModeLow:            "LOW",
	MsgOptModeMedium:         "MEDIUM",
	MsgOptModeHigh:           "HIGH",
	MsgOptModeUltra:          "ULTRA",
	MsgOptMaxConcurrent:      "Max Concurrent Browsers:",
	MsgOptHitsPerMinute:      "Hits/Minute:",
	MsgOptBrowserPool:        "Browser Pool:",
	MsgOptWorkerQueue:        "Worker Queue:",
	MsgOptWarnLowRAM:         "⚠️  Low RAM: %.1f GB - Performance may be limited",
	MsgOptWarnLowCPU:         "⚠️  Low CPU thread count: %d",
	MsgOptWarnHighMemory:     "⚠️  High memory usage: %.1f%% - Close other applications",
	MsgOptWarnLowDisk:        "⚠️  Low disk space: %.1f GB - Free up space for reports",
	MsgOptRecResourceBlock:   "💡 Resource blocking (images, CSS, fonts) should be enabled",
	MsgOptRecMediumSystem:    "💡 Medium-tier system - Balanced settings recommended",
	MsgOptRecGoodSystem:      "💡 Good system - High performance settings can be used",
	MsgOptRecPowerfulSystem:  "🚀 Powerful system - Maximum performance settings can be used",
	MsgOptRecStrongCPU:       "💪 Strong CPU: %d threads - High parallel processing capacity",
	MsgOptRecWindows:         "💡 You can temporarily disable Windows Defender real-time protection",
	MsgOptRecLinux:           "💡 You can increase ulimit -n for more file descriptors",
	MsgOptRecMacOS:           "💡 On macOS, use 'launchctl limit maxfiles' to increase file limit",
	MsgOptApplyHint:          "💡 To apply these settings: eroshit -cli -optimize",
	// v2.4.0 - Private proxy messages
	MsgPrivateProxyActive: "🔐 Private proxy mode active: %d proxies loaded",
	// v2.4.0 - Web interface messages
	MsgWebInterface: "ErosHit - Web Interface",
	MsgOpenBrowser:  "Open in browser: %s",
	MsgStopHint:     "Press Ctrl+C to stop",
}

// T locale'e göre mesajı çevirir ve formatlar
func T(locale string, key string, args ...interface{}) string {
	m := tr
	if locale == "en" {
		m = en
	}
	tpl, ok := m[key]
	if !ok {
		tpl = tr[key]
	}
	if tpl == "" {
		return key
	}
	if len(args) == 0 {
		return tpl
	}
	return fmt.Sprintf(tpl, args...)
}

// GetModeNames returns mode names for the given locale
func GetModeNames(locale string) map[string]string {
	if locale == "en" {
		return map[string]string{
			"low":    T(locale, MsgOptModeLow),
			"medium": T(locale, MsgOptModeMedium),
			"high":   T(locale, MsgOptModeHigh),
			"ultra":  T(locale, MsgOptModeUltra),
		}
	}
	return map[string]string{
		"low":    T(locale, MsgOptModeLow),
		"medium": T(locale, MsgOptModeMedium),
		"high":   T(locale, MsgOptModeHigh),
		"ultra":  T(locale, MsgOptModeUltra),
	}
}
