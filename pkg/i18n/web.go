package i18n

// Web UI çeviri anahtarları
const (
	// Genel
	WebLangTitle    = "web.lang.title"
	WebLangSubtitle = "web.lang.subtitle"

	// Header
	WebAppTitle       = "web.app.title"
	WebAppSubtitle    = "web.app.subtitle"
	WebStatusRunning  = "web.status.running"
	WebStatusStopped  = "web.status.stopped"
	WebStatusError    = "web.status.error"
	WebStatusComplete = "web.status.complete"

	// Tabs
	WebTabBasic     = "web.tab.basic"
	WebTabTraffic   = "web.tab.traffic"
	WebTabSeo       = "web.tab.seo"
	WebTabAdvanced  = "web.tab.advanced"
	WebTabNetwork   = "web.tab.network"
	WebTabSystem    = "web.tab.system"
	WebTabVM        = "web.tab.vm"
	WebTabProxy     = "web.tab.proxy"
	WebTabDashboard = "web.tab.dashboard"
	WebTabLogs      = "web.tab.logs"

	// Sections
	WebSectionTarget               = "web.section.target"
	WebSectionPerformance          = "web.section.performance"
	WebSectionQuickToggles         = "web.section.quick_toggles"
	WebSectionDevice               = "web.section.device"
	WebSectionBehavior             = "web.section.behavior"
	WebSectionBounceSettings       = "web.section.bounce_settings"
	WebSectionReferrer             = "web.section.referrer"
	WebSectionKeywordsGeo          = "web.section.keywords_geo"
	WebSectionAnalytics            = "web.section.analytics"
	WebSectionGSC                  = "web.section.gsc"
	WebSectionBrowserProfile       = "web.section.browser_profile"
	WebSectionReturning            = "web.section.returning"
	WebSectionNetworkOptimizations = "web.section.network_optimizations"
	WebSectionConnPoolSettings     = "web.section.conn_pool_settings"
	WebSectionSystemOptimizations  = "web.section.system_optimizations"
	WebSectionSystemInfo           = "web.section.system_info"
	WebSectionVMSpoofing           = "web.section.vm_spoofing"
	WebSectionVMScore              = "web.section.vm_score"
	WebSectionProxy                = "web.section.proxy"
	WebSectionTrafficChart         = "web.section.traffic_chart"
	WebSectionLiveLogs             = "web.section.live_logs"
	WebSectionStructuredLogs       = "web.section.structured_logs"

	// Labels
	WebLabelDomain            = "web.label.domain"
	WebLabelGtag              = "web.label.gtag"
	WebLabelOutputDir         = "web.label.output_dir"
	WebLabelMaxPages          = "web.label.max_pages"
	WebLabelDuration          = "web.label.duration"
	WebLabelHpm               = "web.label.hpm"
	WebLabelMaxConcurrent     = "web.label.max_concurrent"
	WebLabelDeviceType        = "web.label.device_type"
	WebLabelMinPageDuration   = "web.label.min_page_duration"
	WebLabelMaxPageDuration   = "web.label.max_page_duration"
	WebLabelTargetBounceRate  = "web.label.target_bounce_rate"
	WebLabelSessionMinPages   = "web.label.session_min_pages"
	WebLabelSessionMaxPages   = "web.label.session_max_pages"
	WebLabelReferrerSource    = "web.label.referrer_source"
	WebLabelReferrerKeyword   = "web.label.referrer_keyword"
	WebLabelKeywords          = "web.label.keywords"
	WebLabelGeoCountry        = "web.label.geo_country"
	WebLabelGeoLanguage       = "web.label.geo_language"
	WebLabelGeoTimezone       = "web.label.geo_timezone"
	WebLabelGscProperty       = "web.label.gsc_property"
	WebLabelGscApiKey         = "web.label.gsc_api_key"
	WebLabelProfilePath       = "web.label.profile_path"
	WebLabelMaxProfiles       = "web.label.max_profiles"
	WebLabelReturningRate     = "web.label.returning_rate"
	WebLabelReturningDays     = "web.label.returning_days"
	WebLabelMaxIdleConns      = "web.label.max_idle_conns"
	WebLabelMaxConnsPerHost   = "web.label.max_conns_per_host"
	WebLabelCpuCount          = "web.label.cpu_count"
	WebLabelNumaNodes         = "web.label.numa_nodes"
	WebLabelAffinityStatus    = "web.label.affinity_status"
	WebLabelNumaStatus        = "web.label.numa_status"
	WebLabelVMType            = "web.label.vm_type"
	WebLabelVMDetectionScore  = "web.label.vm_detection_score"
	WebLabelProxyList         = "web.label.proxy_list"

	// Toggle Titles
	WebToggleAntiDetectTitle     = "web.toggle.antidetect.title"
	WebToggleCanvasFpTitle       = "web.toggle.canvas_fp.title"
	WebToggleSitemapTitle        = "web.toggle.sitemap.title"
	WebToggleScrollTitle         = "web.toggle.scroll.title"
	WebToggleMouseTitle          = "web.toggle.mouse.title"
	WebToggleClicksTitle         = "web.toggle.clicks.title"
	WebToggleKeyboardTitle       = "web.toggle.keyboard.title"
	WebToggleFocusTitle          = "web.toggle.focus.title"
	WebToggleSessionDepthTitle   = "web.toggle.session_depth.title"
	WebToggleBounceTitle         = "web.toggle.bounce.title"
	WebToggleReferrerTitle       = "web.toggle.referrer.title"
	WebToggleGscTitle            = "web.toggle.gsc.title"
	WebToggleGscQueriesTitle     = "web.toggle.gsc_queries.title"
	WebToggleProfileTitle        = "web.toggle.profile.title"
	WebToggleCookiesTitle        = "web.toggle.cookies.title"
	WebToggleReturningTitle      = "web.toggle.returning.title"
	WebToggleHTTP3Title          = "web.toggle.http3.title"
	WebToggleConnPoolTitle       = "web.toggle.conn_pool.title"
	WebToggleTCPTitle            = "web.toggle.tcp.title"
	WebToggleAffinityTitle       = "web.toggle.affinity.title"
	WebToggleNUMATitle           = "web.toggle.numa.title"
	WebToggleVMSpoofTitle        = "web.toggle.vm_spoof.title"
	WebToggleHideVMTitle         = "web.toggle.hide_vm.title"
	WebToggleHardwareTitle       = "web.toggle.hardware.title"
	WebToggleRandomizeTitle      = "web.toggle.randomize.title"
	WebToggleProxyTitle          = "web.toggle.proxy.title"

	// Toggle Descriptions
	WebToggleAntiDetectDesc     = "web.toggle.antidetect.desc"
	WebToggleCanvasFpDesc       = "web.toggle.canvas_fp.desc"
	WebToggleSitemapDesc        = "web.toggle.sitemap.desc"
	WebToggleScrollDesc         = "web.toggle.scroll.desc"
	WebToggleMouseDesc          = "web.toggle.mouse.desc"
	WebToggleClicksDesc         = "web.toggle.clicks.desc"
	WebToggleKeyboardDesc       = "web.toggle.keyboard.desc"
	WebToggleFocusDesc          = "web.toggle.focus.desc"
	WebToggleSessionDepthDesc   = "web.toggle.session_depth.desc"
	WebToggleBounceDesc         = "web.toggle.bounce.desc"
	WebToggleReferrerDesc       = "web.toggle.referrer.desc"
	WebTogglePageViewDesc       = "web.toggle.pageview.desc"
	WebToggleSessionStartDesc   = "web.toggle.session_start.desc"
	WebToggleEngagementDesc     = "web.toggle.engagement.desc"
	WebToggleFirstVisitDesc     = "web.toggle.first_visit.desc"
	WebToggleGscDesc            = "web.toggle.gsc.desc"
	WebToggleGscQueriesDesc     = "web.toggle.gsc_queries.desc"
	WebToggleProfileDesc        = "web.toggle.profile.desc"
	WebToggleCookiesDesc        = "web.toggle.cookies.desc"
	WebToggleReturningDesc      = "web.toggle.returning.desc"
	WebToggleHTTP3Desc          = "web.toggle.http3.desc"
	WebToggleConnPoolDesc       = "web.toggle.conn_pool.desc"
	WebToggleTCPDesc            = "web.toggle.tcp.desc"
	WebToggleAffinityDesc       = "web.toggle.affinity.desc"
	WebToggleNUMADesc           = "web.toggle.numa.desc"
	WebToggleVMSpoofDesc        = "web.toggle.vm_spoof.desc"
	WebToggleHideVMDesc         = "web.toggle.hide_vm.desc"
	WebToggleHardwareDesc       = "web.toggle.hardware.desc"
	WebToggleRandomizeDesc      = "web.toggle.randomize.desc"
	WebToggleProxyDesc          = "web.toggle.proxy.desc"

	// Options
	WebOptDeviceMixed   = "web.opt.device.mixed"
	WebOptDeviceDesktop = "web.opt.device.desktop"
	WebOptDeviceMobile  = "web.opt.device.mobile"
	WebOptDeviceTablet  = "web.opt.device.tablet"
	WebOptMixed         = "web.opt.mixed"
	WebOptDirect        = "web.opt.direct"
	WebOptRandom        = "web.opt.random"
	WebOptAuto          = "web.opt.auto"
	WebOptNone          = "web.opt.none"

	// Hints
	WebHintBounceRate       = "web.hint.bounce_rate"
	WebHintReferrerKeyword  = "web.hint.referrer_keyword"
	WebHintGscApiKey        = "web.hint.gsc_api_key"
	WebHintMaxIdleConns     = "web.hint.max_idle_conns"
	WebHintMaxConnsPerHost  = "web.hint.max_conns_per_host"
	WebHintVMType           = "web.hint.vm_type"
	WebHintVMScore          = "web.hint.vm_score"
	WebHintProxyFormat      = "web.hint.proxy_format"

	// Buttons
	WebBtnDefaults = "web.btn.defaults"
	WebBtnReset    = "web.btn.reset"
	WebBtnSave     = "web.btn.save"
	WebBtnStart    = "web.btn.start"
	WebBtnStop     = "web.btn.stop"

	// Stats
	WebStatTotalHits    = "web.stat.total_hits"
	WebStatSuccessRate  = "web.stat.success_rate"
	WebStatActiveProxies = "web.stat.active_proxies"
	WebStatAvgResponse  = "web.stat.avg_response"

	// Filters
	WebFilterAll = "web.filter.all"

	// Messages
	WebMsgNoLogs = "web.msg.no_logs"

	// Toast
	WebToastSaved   = "web.toast.saved"
	WebToastStarted = "web.toast.started"
	WebToastStopped = "web.toast.stopped"
	WebToastError   = "web.toast.error"
)

// Turkish translations
var trWeb = map[string]string{
	WebLangTitle:    "Dil Seçin / Select Language",
	WebLangSubtitle: "Uygulama dilini seçin / Choose application language",

	WebAppTitle:       "ErosHit",
	WebAppSubtitle:    "Advanced SEO Traffic Bot",
	WebStatusRunning:  "Çalışıyor",
	WebStatusStopped:  "Durduruldu",
	WebStatusError:    "Hata",
	WebStatusComplete: "Tamamlandı",

	WebTabBasic:     "Temel",
	WebTabTraffic:   "Trafik",
	WebTabSeo:       "SEO",
	WebTabAdvanced:  "Gelişmiş",
	WebTabNetwork:   "Network",
	WebTabSystem:    "Sistem",
	WebTabVM:        "VM Spoof",
	WebTabProxy:     "Proxy",
	WebTabDashboard: "Dashboard",
	WebTabLogs:      "Loglar",

	WebSectionTarget:               "Hedef Ayarları",
	WebSectionPerformance:          "Performans Ayarları",
	WebSectionQuickToggles:         "Hızlı Ayarlar",
	WebSectionDevice:               "Cihaz Ayarları",
	WebSectionBehavior:             "Davranış Simülasyonu",
	WebSectionBounceSettings:       "Bounce Rate Ayarları",
	WebSectionReferrer:             "Referrer Ayarları",
	WebSectionKeywordsGeo:          "Anahtar Kelimeler & Konum",
	WebSectionAnalytics:            "Analytics Eventleri",
	WebSectionGSC:                  "Google Search Console",
	WebSectionBrowserProfile:       "Browser Profilleri",
	WebSectionReturning:            "Returning Visitor",
	WebSectionNetworkOptimizations: "Network Optimizasyonları",
	WebSectionConnPoolSettings:     "Connection Pool Ayarları",
	WebSectionSystemOptimizations:  "Sistem Optimizasyonları",
	WebSectionSystemInfo:           "Sistem Bilgisi",
	WebSectionVMSpoofing:           "VM Fingerprint Spoofing",
	WebSectionVMScore:              "VM Tespit Skoru",
	WebSectionProxy:                "Proxy Ayarları",
	WebSectionTrafficChart:         "Trafik Grafiği",
	WebSectionLiveLogs:             "Canlı Loglar",
	WebSectionStructuredLogs:       "Yapılandırılmış Loglar",

	WebLabelDomain:           "Domain",
	WebLabelGtag:             "GA4 Tracking ID",
	WebLabelOutputDir:        "Çıktı Dizini",
	WebLabelMaxPages:         "Max Sayfa",
	WebLabelDuration:         "Süre (dk)",
	WebLabelHpm:              "Hit/Dakika",
	WebLabelMaxConcurrent:    "Max Eşzamanlı",
	WebLabelDeviceType:       "Cihaz Tipi",
	WebLabelMinPageDuration:  "Min Sayfa Süresi (sn)",
	WebLabelMaxPageDuration:  "Max Sayfa Süresi (sn)",
	WebLabelTargetBounceRate: "Hedef Bounce Rate (%)",
	WebLabelSessionMinPages:  "Min Sayfa/Oturum",
	WebLabelSessionMaxPages:  "Max Sayfa/Oturum",
	WebLabelReferrerSource:   "Referrer Kaynağı",
	WebLabelReferrerKeyword:  "Anahtar Kelime",
	WebLabelKeywords:         "Anahtar Kelimeler (virgülle ayırın)",
	WebLabelGeoCountry:       "Ülke",
	WebLabelGeoLanguage:      "Dil",
	WebLabelGeoTimezone:      "Saat Dilimi",
	WebLabelGscProperty:      "GSC Property URL",
	WebLabelGscApiKey:        "Service Account JSON",
	WebLabelProfilePath:      "Profil Dizini",
	WebLabelMaxProfiles:      "Max Profil Sayısı",
	WebLabelReturningRate:    "Returning Rate (%)",
	WebLabelReturningDays:    "Tekrar Ziyaret (gün)",
	WebLabelMaxIdleConns:     "Max Idle Connections",
	WebLabelMaxConnsPerHost:  "Max Conns Per Host",
	WebLabelCpuCount:         "CPU Çekirdek",
	WebLabelNumaNodes:        "NUMA Node",
	WebLabelAffinityStatus:   "Affinity",
	WebLabelNumaStatus:       "NUMA",
	WebLabelVMType:           "VM Tipi",
	WebLabelVMDetectionScore: "Tespit Olasılığı",
	WebLabelProxyList:        "Proxy Listesi (satır başına bir tane)",

	WebToggleAntiDetectTitle:   "Anti-Detect Modu",
	WebToggleCanvasFpTitle:     "Canvas Fingerprint",
	WebToggleSitemapTitle:      "Sitemap Kullan",
	WebToggleScrollTitle:       "Scroll Eventleri",
	WebToggleMouseTitle:        "Mouse Hareketi",
	WebToggleClicksTitle:       "Tıklama Simülasyonu",
	WebToggleKeyboardTitle:     "Klavye Simülasyonu",
	WebToggleFocusTitle:        "Focus/Blur Eventleri",
	WebToggleSessionDepthTitle: "Session Depth",
	WebToggleBounceTitle:       "Bounce Rate Kontrolü",
	WebToggleReferrerTitle:     "Referrer Simülasyonu",
	WebToggleGscTitle:          "GSC Entegrasyonu",
	WebToggleGscQueriesTitle:   "Sorguları Kullan",
	WebToggleProfileTitle:      "Profil Persistence",
	WebToggleCookiesTitle:      "Cookie'leri Kaydet",
	WebToggleReturningTitle:    "Returning Visitor",
	WebToggleHTTP3Title:        "HTTP/3 QUIC",
	WebToggleConnPoolTitle:     "Connection Pooling",
	WebToggleTCPTitle:          "TCP Fast Open",
	WebToggleAffinityTitle:     "CPU Affinity",
	WebToggleNUMATitle:         "NUMA Awareness",
	WebToggleVMSpoofTitle:      "VM Spoofing Aktif",
	WebToggleHideVMTitle:       "VM Göstergelerini Gizle",
	WebToggleHardwareTitle:     "Hardware ID Spoofing",
	WebToggleRandomizeTitle:    "Rastgele VM Parametreleri",
	WebToggleProxyTitle:        "Proxy Kullan",

	WebToggleAntiDetectDesc:   "Bot algılama sistemlerini atlatmak için gelişmiş gizleme",
	WebToggleCanvasFpDesc:     "Benzersiz canvas parmak izi ile gerçek tarayıcı taklidi",
	WebToggleSitemapDesc:      "Site haritasından URL'leri otomatik çek ve ziyaret et",
	WebToggleScrollDesc:       "Google Analytics'e scroll derinliği eventleri gönder",
	WebToggleMouseDesc:        "İnsan benzeri mouse hareketleri ve hover efektleri",
	WebToggleClicksDesc:       "Rastgele bağlantılara tıklayarak gezinti davranışı",
	WebToggleKeyboardDesc:     "Form alanlarına yazma ve klavye eventleri",
	WebToggleFocusDesc:        "Sekme değişimi ve pencere focus/blur simülasyonu",
	WebToggleSessionDepthDesc: "Oturum başına birden fazla sayfa ziyareti",
	WebToggleBounceDesc:       "Hedef bounce rate'e göre oturum davranışı ayarı",
	WebToggleReferrerDesc:     "Organik trafik gibi görünmesi için referrer bilgisi ekle",
	WebTogglePageViewDesc:     "Her sayfa yüklemesinde GA4 page_view eventi gönder",
	WebToggleSessionStartDesc: "Yeni oturum başlatıldığında session_start eventi",
	WebToggleEngagementDesc:   "Kullanıcı etkileşimini ölçen engagement eventleri",
	WebToggleFirstVisitDesc:   "İlk ziyaret için first_visit eventi",
	WebToggleGscDesc:          "Gerçek arama sorgularını çek ve simüle et",
	WebToggleGscQueriesDesc:   "GSC'den çekilen sorguları anahtar kelime olarak kullan",
	WebToggleProfileDesc:      "Tarayıcı profillerini kaydet ve returning visitor simüle et",
	WebToggleCookiesDesc:      "Oturumlar arası cookie persistence",
	WebToggleReturningDesc:    "Aynı client_id ile tekrar ziyaretler",
	WebToggleHTTP3Desc:        "Google tarafından geliştirilen yeni nesil protokol. Daha hızlı bağlantı kurulumu ve çoklu stream desteği ile performans artışı sağlar.",
	WebToggleConnPoolDesc:     "Bağlantıları yeniden kullanarak TCP el sıkışma süresinden tasarruf et. HTTP/2 ve Keep-Alive desteği ile %40+ performans artışı.",
	WebToggleTCPDesc:          "SYN paketinde data göndererek 1-RTT tasarrufu sağlar. Yüksek gecikmeli bağlantılarda bağlantı süresini %30-50 azaltır.",
	WebToggleAffinityDesc:     "Her thread'i belirli bir CPU çekirdeğine sabitleyerek cache verimliliğini artır ve context switch'i azalt. Yüksek yükte %15-25 performans kazancı.",
	WebToggleNUMADesc:         "Non-Uniform Memory Access optimizasyonu. Thread'leri local NUMA node'a yerleştirerek memory latency'sini azalt ve bant genişliğini artır.",
	WebToggleVMSpoofDesc:      "Sanal makine ortamının izlerini gizle ve gerçek donanım gibi görün. VM detection scriptlerini atlatır.",
	WebToggleHideVMDesc:       "window.VM, navigator.deviceMemory gibi VM belirteçlerini temizle veya sahte değerler ata.",
	WebToggleHardwareDesc:     "CPU ID, MAC adresi, BIOS seri numarası gibi hardware tanımlayıcıları rastgele veya belirli bir VM tipine göre değiştir.",
	WebToggleRandomizeDesc:    "Her oturumda farklı VM parametreleri üreterek fingerprint çeşitliliği sağla ve izlenmeyi zorlaştır.",
	WebToggleProxyDesc:        "Trafik için proxy rotasyonu aktif et",

	WebOptDeviceMixed:   "Karışık (Tümü)",
	WebOptDeviceDesktop: "Masaüstü",
	WebOptDeviceMobile:  "Mobil",
	WebOptDeviceTablet:  "Tablet",
	WebOptMixed:         "Karışık",
	WebOptDirect:        "Direkt",
	WebOptRandom:        "Rastgele",
	WebOptAuto:          "Otomatik",
	WebOptNone:          "Otomatik / Yok",

	WebHintBounceRate:      "Düşük bounce rate için ziyaretçiler birden fazla sayfa gezer",
	WebHintReferrerKeyword: "Bu kelime ile Google'dan gelmiş gibi görünür",
	WebHintGscApiKey:       "Google Cloud Console'dan alınan JSON key",
	WebHintMaxIdleConns:    "Beklemede tutulacak max bağlantı sayısı",
	WebHintMaxConnsPerHost: "Her host için max eşzamanlı bağlantı",
	WebHintVMType:          "Belirli bir VM tipini taklit et (boş bırakılırsa rastgele)",
	WebHintVMScore:         "Düşük skor = Daha az tespit. VM spoofing aktif edildiğinde skor düşer.",
	WebHintProxyFormat:     "Format: protocol://user:pass@host:port veya protocol://host:port",

	WebBtnDefaults: "Varsayılanlar",
	WebBtnReset:    "Sıfırla",
	WebBtnSave:     "Kaydet",
	WebBtnStart:    "Başlat",
	WebBtnStop:     "Durdur",

	WebStatTotalHits:     "Toplam Hit",
	WebStatSuccessRate:   "Başarı Oranı",
	WebStatActiveProxies: "Aktif Proxy",
	WebStatAvgResponse:   "Ort. Yanıt",

	WebFilterAll: "Tümü",

	WebMsgNoLogs: "Log bekleniyor...",

	WebToastSaved:   "Ayarlar kaydedildi",
	WebToastStarted: "Bot başlatıldı",
	WebToastStopped: "Bot durduruldu",
	WebToastError:   "Hata oluştu",
}

// English translations
var enWeb = map[string]string{
	WebLangTitle:    "Select Language / Dil Seçin",
	WebLangSubtitle: "Choose application language / Uygulama dilini seçin",

	WebAppTitle:       "ErosHit",
	WebAppSubtitle:    "Advanced SEO Traffic Bot",
	WebStatusRunning:  "Running",
	WebStatusStopped:  "Stopped",
	WebStatusError:    "Error",
	WebStatusComplete: "Complete",

	WebTabBasic:     "Basic",
	WebTabTraffic:   "Traffic",
	WebTabSeo:       "SEO",
	WebTabAdvanced:  "Advanced",
	WebTabNetwork:   "Network",
	WebTabSystem:    "System",
	WebTabVM:        "VM Spoof",
	WebTabProxy:     "Proxy",
	WebTabDashboard: "Dashboard",
	WebTabLogs:      "Logs",

	WebSectionTarget:               "Target Settings",
	WebSectionPerformance:          "Performance Settings",
	WebSectionQuickToggles:         "Quick Settings",
	WebSectionDevice:               "Device Settings",
	WebSectionBehavior:             "Behavior Simulation",
	WebSectionBounceSettings:       "Bounce Rate Settings",
	WebSectionReferrer:             "Referrer Settings",
	WebSectionKeywordsGeo:          "Keywords & Location",
	WebSectionAnalytics:            "Analytics Events",
	WebSectionGSC:                  "Google Search Console",
	WebSectionBrowserProfile:       "Browser Profiles",
	WebSectionReturning:            "Returning Visitor",
	WebSectionNetworkOptimizations: "Network Optimizations",
	WebSectionConnPoolSettings:     "Connection Pool Settings",
	WebSectionSystemOptimizations:  "System Optimizations",
	WebSectionSystemInfo:           "System Information",
	WebSectionVMSpoofing:           "VM Fingerprint Spoofing",
	WebSectionVMScore:              "VM Detection Score",
	WebSectionProxy:                "Proxy Settings",
	WebSectionTrafficChart:         "Traffic Chart",
	WebSectionLiveLogs:             "Live Logs",
	WebSectionStructuredLogs:       "Structured Logs",

	WebLabelDomain:           "Domain",
	WebLabelGtag:             "GA4 Tracking ID",
	WebLabelOutputDir:        "Output Directory",
	WebLabelMaxPages:         "Max Pages",
	WebLabelDuration:         "Duration (min)",
	WebLabelHpm:              "Hits/Minute",
	WebLabelMaxConcurrent:    "Max Concurrent",
	WebLabelDeviceType:       "Device Type",
	WebLabelMinPageDuration:  "Min Page Duration (sec)",
	WebLabelMaxPageDuration:  "Max Page Duration (sec)",
	WebLabelTargetBounceRate: "Target Bounce Rate (%)",
	WebLabelSessionMinPages:  "Min Pages/Session",
	WebLabelSessionMaxPages:  "Max Pages/Session",
	WebLabelReferrerSource:   "Referrer Source",
	WebLabelReferrerKeyword:  "Keyword",
	WebLabelKeywords:         "Keywords (comma separated)",
	WebLabelGeoCountry:       "Country",
	WebLabelGeoLanguage:      "Language",
	WebLabelGeoTimezone:      "Timezone",
	WebLabelGscProperty:      "GSC Property URL",
	WebLabelGscApiKey:        "Service Account JSON",
	WebLabelProfilePath:      "Profile Directory",
	WebLabelMaxProfiles:      "Max Profiles",
	WebLabelReturningRate:    "Returning Rate (%)",
	WebLabelReturningDays:    "Return Visit (days)",
	WebLabelMaxIdleConns:     "Max Idle Connections",
	WebLabelMaxConnsPerHost:  "Max Conns Per Host",
	WebLabelCpuCount:         "CPU Cores",
	WebLabelNumaNodes:        "NUMA Nodes",
	WebLabelAffinityStatus:   "Affinity",
	WebLabelNumaStatus:       "NUMA",
	WebLabelVMType:           "VM Type",
	WebLabelVMDetectionScore: "Detection Likelihood",
	WebLabelProxyList:        "Proxy List (one per line)",

	WebToggleAntiDetectTitle:   "Anti-Detect Mode",
	WebToggleCanvasFpTitle:     "Canvas Fingerprint",
	WebToggleSitemapTitle:      "Use Sitemap",
	WebToggleScrollTitle:       "Scroll Events",
	WebToggleMouseTitle:        "Mouse Movement",
	WebToggleClicksTitle:       "Click Simulation",
	WebToggleKeyboardTitle:     "Keyboard Simulation",
	WebToggleFocusTitle:        "Focus/Blur Events",
	WebToggleSessionDepthTitle: "Session Depth",
	WebToggleBounceTitle:       "Bounce Rate Control",
	WebToggleReferrerTitle:     "Referrer Simulation",
	WebToggleGscTitle:          "GSC Integration",
	WebToggleGscQueriesTitle:   "Use Queries",
	WebToggleProfileTitle:      "Profile Persistence",
	WebToggleCookiesTitle:      "Save Cookies",
	WebToggleReturningTitle:    "Returning Visitor",
	WebToggleHTTP3Title:        "HTTP/3 QUIC",
	WebToggleConnPoolTitle:     "Connection Pooling",
	WebToggleTCPTitle:          "TCP Fast Open",
	WebToggleAffinityTitle:     "CPU Affinity",
	WebToggleNUMATitle:         "NUMA Awareness",
	WebToggleVMSpoofTitle:      "VM Spoofing Active",
	WebToggleHideVMTitle:       "Hide VM Indicators",
	WebToggleHardwareTitle:     "Hardware ID Spoofing",
	WebToggleRandomizeTitle:    "Randomize VM Params",
	WebToggleProxyTitle:        "Use Proxy",

	WebToggleAntiDetectDesc:   "Advanced evasion to bypass bot detection systems",
	WebToggleCanvasFpDesc:     "Unique canvas fingerprint for realistic browser emulation",
	WebToggleSitemapDesc:      "Automatically fetch and visit URLs from sitemap",
	WebToggleScrollDesc:       "Send scroll depth events to Google Analytics",
	WebToggleMouseDesc:        "Human-like mouse movements and hover effects",
	WebToggleClicksDesc:       "Simulate navigation by clicking random links",
	WebToggleKeyboardDesc:     "Typing and keyboard events for form fields",
	WebToggleFocusDesc:        "Tab switching and window focus/blur simulation",
	WebToggleSessionDepthDesc: "Multiple page visits per session",
	WebToggleBounceDesc:       "Adjust session behavior based on target bounce rate",
	WebToggleReferrerDesc:     "Add referrer info to appear as organic traffic",
	WebTogglePageViewDesc:     "Send GA4 page_view event on every page load",
	WebToggleSessionStartDesc: "Send session_start event when new session starts",
	WebToggleEngagementDesc:   "Engagement events to measure user interaction",
	WebToggleFirstVisitDesc:   "Send first_visit event for new visitors",
	WebToggleGscDesc:          "Fetch and simulate real search queries",
	WebToggleGscQueriesDesc:   "Use GSC fetched queries as keywords",
	WebToggleProfileDesc:      "Save browser profiles and simulate returning visitors",
	WebToggleCookiesDesc:      "Cookie persistence between sessions",
	WebToggleReturningDesc:    "Revisit with same client_id",
	WebToggleHTTP3Desc:        "Next-gen protocol by Google. Faster connection setup and multi-stream support for performance boost.",
	WebToggleConnPoolDesc:     "Reuse connections to save TCP handshake time. HTTP/2 and Keep-Alive support for 40%+ performance gain.",
	WebToggleTCPDesc:          "Save 1-RTT by sending data in SYN packet. Reduces connection time by 30-50% on high latency connections.",
	WebToggleAffinityDesc:     "Pin each thread to specific CPU core to improve cache efficiency and reduce context switches. 15-25% performance gain under high load.",
	WebToggleNUMADesc:         "Non-Uniform Memory Access optimization. Place threads on local NUMA node to reduce memory latency and increase bandwidth.",
	WebToggleVMSpoofDesc:      "Hide virtual machine traces and appear as real hardware. Bypass VM detection scripts.",
	WebToggleHideVMDesc:       "Clean or fake VM indicators like window.VM, navigator.deviceMemory",
	WebToggleHardwareDesc:     "Change hardware identifiers like CPU ID, MAC address, BIOS serial based on VM type",
	WebToggleRandomizeDesc:    "Generate different VM params per session for fingerprint diversity",
	WebToggleProxyDesc:        "Enable proxy rotation for traffic",

	WebOptDeviceMixed:   "Mixed (All)",
	WebOptDeviceDesktop: "Desktop",
	WebOptDeviceMobile:  "Mobile",
	WebOptDeviceTablet:  "Tablet",
	WebOptMixed:         "Mixed",
	WebOptDirect:        "Direct",
	WebOptRandom:        "Random",
	WebOptAuto:          "Auto",
	WebOptNone:          "Auto / None",

	WebHintBounceRate:      "For low bounce rate, visitors browse multiple pages",
	WebHintReferrerKeyword: "Will appear as coming from Google with this keyword",
	WebHintGscApiKey:       "JSON key from Google Cloud Console",
	WebHintMaxIdleConns:    "Max connections to keep idle",
	WebHintMaxConnsPerHost: "Max concurrent connections per host",
	WebHintVMType:          "Emulate specific VM type (random if empty)",
	WebHintVMScore:         "Low score = Less detection. Score drops when VM spoofing is active.",
	WebHintProxyFormat:     "Format: protocol://user:pass@host:port or protocol://host:port",

	WebBtnDefaults: "Defaults",
	WebBtnReset:    "Reset",
	WebBtnSave:     "Save",
	WebBtnStart:    "Start",
	WebBtnStop:     "Stop",

	WebStatTotalHits:     "Total Hits",
	WebStatSuccessRate:   "Success Rate",
	WebStatActiveProxies: "Active Proxies",
	WebStatAvgResponse:   "Avg Response",

	WebFilterAll: "All",

	WebMsgNoLogs: "Waiting for logs...",

	WebToastSaved:   "Settings saved",
	WebToastStarted: "Bot started",
	WebToastStopped: "Bot stopped",
	WebToastError:   "An error occurred",
}

// WebT returns web UI translation
func WebT(locale, key string) string {
	var m map[string]string
	switch locale {
	case "tr":
		m = trWeb
	case "en":
		m = enWeb
	default:
		m = trWeb
	}
	if v, ok := m[key]; ok {
		return v
	}
	return key
}

// GetAllWebTranslations returns all web translations for a locale
func GetAllWebTranslations(locale string) map[string]string {
	switch locale {
	case "en":
		return enWeb
	case "tr":
		return trWeb
	default:
		return trWeb
	}
}
