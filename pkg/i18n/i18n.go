package i18n

import "fmt"

// Msg log mesajı anahtarları
const (
	MsgStarting     = "starting"
	MsgTarget       = "target"
	MsgDiscovery    = "discovery"
	MsgDiscoveryErr = "discovery_err"
	MsgPagesFound   = "pages_found"
	MsgCancel       = "cancel"
	MsgDeadline     = "deadline"
	MsgVisitErr     = "visit_err"
	MsgProgress     = "progress"
	MsgSummary      = "summary"
	MsgSummaryLine  = "summary_line"
	MsgSummaryRT    = "summary_rt"
	MsgExportErr    = "export_err"
	MsgReportCSV     = "report_csv"
	MsgReportJSON    = "report_json"
	MsgReportHTML    = "report_html"
	MsgSitemapFound  = "sitemap_found"
	MsgSitemapNone   = "sitemap_none"
)

var tr = map[string]string{
	MsgStarting:     "Eros - Hit Botu başlatılıyor...",
	MsgTarget:       "Hedef: %s | Max sayfa: %d | Süre: %d dk | HPM: %d | Paralel: %d",
	MsgDiscovery:    "Sayfa keşfi başlıyor...",
	MsgDiscoveryErr: "Keşif hatası: %s",
	MsgPagesFound:   "%d sayfa bulundu: %v",
	MsgCancel:       "İptal sinyali alındı, kapatılıyor...",
	MsgDeadline:     "Test süresi doldu.",
	MsgVisitErr:     "Ziyaret hatası [%s]: %v",
	MsgProgress:     "[%d] Toplam: %d | OK: %d | Hata: %d | Ort. RT: %.0f ms",
	MsgSummary:      "--- Özet ---",
	MsgSummaryLine:  "Toplam istek: %d | Başarılı: %d | Hatalı: %d",
	MsgSummaryRT:    "Ort. yanıt: %.0f ms | Min: %d ms | Max: %d ms",
	MsgExportErr:    "Rapor export hatası: %s",
	MsgReportCSV:     "Rapor kaydedildi: %s",
	MsgReportJSON:    "Rapor kaydedildi: %s",
	MsgReportHTML:   "HTML rapor: %s",
	MsgSitemapFound: "Sitemap bulundu: %d URL (anasayfa ağırlığı %%%d)",
	MsgSitemapNone:  "Sitemap bulunamadı, sayfa keşfi kullanılıyor.",
}

var en = map[string]string{
	MsgStarting:     "Eros - Hit Bot starting...",
	MsgTarget:       "Target: %s | Max pages: %d | Duration: %d min | HPM: %d | Parallel: %d",
	MsgDiscovery:    "Page discovery starting...",
	MsgDiscoveryErr: "Discovery error: %s",
	MsgPagesFound:   "%d pages found: %v",
	MsgCancel:       "Cancel signal received, shutting down...",
	MsgDeadline:     "Test duration expired.",
	MsgVisitErr:     "Visit error [%s]: %v",
	MsgProgress:     "[%d] Total: %d | OK: %d | Fail: %d | Avg RT: %.0f ms",
	MsgSummary:      "--- Summary ---",
	MsgSummaryLine:  "Total requests: %d | Success: %d | Failed: %d",
	MsgSummaryRT:    "Avg response: %.0f ms | Min: %d ms | Max: %d ms",
	MsgExportErr:    "Report export error: %s",
	MsgReportCSV:     "Report saved: %s",
	MsgReportJSON:    "Report saved: %s",
	MsgReportHTML:   "HTML report: %s",
	MsgSitemapFound: "Sitemap found: %d URLs (homepage weight %d%%)",
	MsgSitemapNone:  "Sitemap not found, using page discovery.",
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
