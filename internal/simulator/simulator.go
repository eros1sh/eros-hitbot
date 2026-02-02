package simulator

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"eroshit/internal/browser"
	"eroshit/internal/config"
	"eroshit/internal/crawler"
	"eroshit/internal/reporter"
	"eroshit/pkg/analytics"
	"eroshit/pkg/delay"
	"eroshit/pkg/i18n"
	"eroshit/pkg/sitemap"
)

// Simulator trafik simülasyonu orkestratörü
type Simulator struct {
	cfg          *config.Config
	crawler      *crawler.Crawler
	hitVisitor   *browser.HitVisitor
	reporter     *reporter.Reporter
	pages        []string
	homepageURL  string // anasayfa (ağırlıklı seçim için)
}

// New simulator oluşturur. agentProvider ve rep nil olabilir.
// Hit aşaması chromedp ile (JS çalışır, analytics tetiklenir). Resim/CSS bloklanır.
func New(cfg *config.Config, agentProvider crawler.AgentProvider, rep *reporter.Reporter) (*Simulator, error) {
	if rep == nil {
		rep = reporter.New(cfg.OutputDir, cfg.ExportFormat, cfg.TargetDomain)
	}
	rep.LogT(i18n.MsgStarting)

	proxyURL := ""
	if cfg.ProxyEnabled {
		proxyURL = cfg.ProxyBaseURL
		if proxyURL == "" {
			proxyURL = cfg.ProxyURL
		}
	}

	c, err := crawler.New(cfg.TargetDomain, cfg.MaxPages, rep, cfg.ProxyURL, agentProvider)
	if err != nil {
		return nil, err
	}

	analyticsMgr := &analytics.Manager{
		GA4Enabled:       cfg.GtagID != "",
		GA4MeasurementID: cfg.GtagID,
	}
	hitVisitor, err := browser.NewHitVisitor(agentProvider, rep, browser.HitVisitorConfig{
		ProxyURL:          proxyURL,
		ProxyUser:         cfg.ProxyUser,
		ProxyPass:         cfg.ProxyPass,
		GtagID:            cfg.GtagID,
		CanvasFingerprint: cfg.CanvasFingerprint,
		ScrollStrategy:    cfg.ScrollStrategy,
		SendScrollEvent:   cfg.SendScrollEvent,
		AnalyticsManager:  analyticsMgr,
		Keywords:          cfg.Keywords,
	})
	if err != nil {
		return nil, err
	}

	return &Simulator{
		cfg:        cfg,
		crawler:    c,
		hitVisitor: hitVisitor,
		reporter:   rep,
		pages:      nil,
	}, nil
}

// Run simülasyonu başlatır
func (s *Simulator) Run(ctx context.Context) error {
	workers := s.cfg.MaxConcurrentVisits
	if workers <= 0 {
		workers = 10
	}
	hpm := s.cfg.HitsPerMinute
	if hpm <= 0 {
		hpm = 35
	}
	s.reporter.LogT(i18n.MsgTarget,
		s.cfg.TargetDomain, s.cfg.MaxPages, s.cfg.DurationMinutes, hpm, workers)

	// 1. Sayfa keşfi (ve isteğe bağlı sitemap)
	baseURL := s.cfg.TargetDomain
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + strings.TrimPrefix(baseURL, "//")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	s.homepageURL = baseURL

	s.reporter.LogT(i18n.MsgDiscovery)
	var pages []string
	if s.cfg.UseSitemap {
		sitemapURLs, errSitemap := sitemap.Fetch(baseURL, nil)
		if errSitemap == nil && len(sitemapURLs) > 0 {
			pages = sitemapURLs
			weight := s.cfg.SitemapHomepageWeight
			if weight <= 0 {
				weight = 60
			}
			s.reporter.LogT(i18n.MsgSitemapFound, len(pages), weight)
		} else {
			s.reporter.LogT(i18n.MsgSitemapNone)
		}
	}
	if len(pages) == 0 {
		var errDiscover error
		pages, errDiscover = s.crawler.Discover()
		if errDiscover != nil {
			s.reporter.LogT(i18n.MsgDiscoveryErr, errDiscover.Error())
			pages = []string{s.homepageURL}
		}
		s.reporter.LogT(i18n.MsgPagesFound, len(pages), pages)
	}
	s.pages = pages
	if len(s.pages) == 0 {
		s.pages = []string{s.homepageURL}
	}

	// 2. HPM sınırı: token bucket (başta workers kadar burst, sonra dakikada hpm refill)
	tb := delay.NewTokenBucket(ctx, hpm, workers)
	defer tb.Stop()

	deadline := time.Now().Add(s.cfg.Duration)
	sem := make(chan struct{}, workers)
	slotFreed := make(chan struct{}, workers)
	var hitCount int64
	var wg sync.WaitGroup

	// Slot boşalınca hemen yeni ziyaret başlat (token varsa); HPM token bucket ile sınırlı
	startVisit := func() {
		if err := tb.Take(ctx); err != nil {
			return
		}
		if time.Now().After(deadline) {
			return
		}
		select {
		case sem <- struct{}{}:
			wg.Add(1)
			page := s.pickPage()
			go func(url string) {
				defer wg.Done()
				defer func() { <-sem; slotFreed <- struct{}{} }()
				if err := s.hitVisitor.VisitURL(ctx, url); err != nil {
					s.reporter.LogT(i18n.MsgVisitErr, url, err)
				} else {
					n := atomic.AddInt64(&hitCount, 1)
					if n%10 == 0 {
						m := s.reporter.GetMetrics()
						s.reporter.LogT(i18n.MsgProgress,
							n, m.TotalHits, m.SuccessHits, m.FailedHits, m.AvgResponseTime)
					}
				}
			}(page)
		default:
			// Slot yok (teoride olmamalı; token aldık, slot bekliyoruz)
		}
	}

	// Başta tüm slotları hemen doldur (workers kadar burst)
	for i := 0; i < workers; i++ {
		slotFreed <- struct{}{}
	}

	deadlineTimer := time.NewTimer(time.Until(deadline))
	defer func() {
		if !deadlineTimer.Stop() {
			select { case <-deadlineTimer.C: default: }
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.reporter.LogT(i18n.MsgCancel)
			tb.Stop()
			wg.Wait()
			s.finish()
			return ctx.Err()

		case <-deadlineTimer.C:
			s.reporter.LogT(i18n.MsgDeadline)
			tb.Stop()
			wg.Wait()
			s.finish()
			return nil

		case <-slotFreed:
			if time.Now().After(deadline) {
				continue
			}
			go startVisit() // Take() bloklayabilir; select bloklanmasın
		}
	}
}

func (s *Simulator) pickPage() string {
	if len(s.pages) == 0 {
		return s.homepageURL
	}
	weight := s.cfg.SitemapHomepageWeight
	if weight <= 0 {
		weight = 60
	}
	// Anasayfa yoğunluğu: weight% anasayfa, (100-weight)% sitemap/diğer sayfalar
	if s.homepageURL != "" && rand.Intn(100) < weight {
		return s.homepageURL
	}
	return s.pages[rand.Intn(len(s.pages))]
}

// Reporter reporter instance döner (log kanalı için)
func (s *Simulator) Reporter() *reporter.Reporter {
	return s.reporter
}

func (s *Simulator) finish() {
	if s.hitVisitor != nil {
		s.hitVisitor.Close()
	}
	s.reporter.Finalize()
	m := s.reporter.GetMetrics()
	s.reporter.LogT(i18n.MsgSummary)
	s.reporter.LogT(i18n.MsgSummaryLine, m.TotalHits, m.SuccessHits, m.FailedHits)
	s.reporter.LogT(i18n.MsgSummaryRT, m.AvgResponseTime, m.MinResponseTime, m.MaxResponseTime)

	if err := s.reporter.Export(); err != nil {
		s.reporter.LogT(i18n.MsgExportErr, err.Error())
	}
}
