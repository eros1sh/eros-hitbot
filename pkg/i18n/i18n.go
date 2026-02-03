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
)

var tr = map[string]string{
	MsgStarting:        "Eros - Hit Botu v2.3.0 başlatılıyor...",
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
}

var en = map[string]string{
	MsgStarting:        "Eros - Hit Bot v2.3.0 starting...",
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
