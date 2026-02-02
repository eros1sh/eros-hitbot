package simulator

import (
	"context"
	"math/rand"
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
)

// Simulator trafik simülasyonu orkestratörü
type Simulator struct {
	cfg         *config.Config
	crawler     *crawler.Crawler
	hitVisitor  *browser.HitVisitor
	reporter    *reporter.Reporter
	pages       []string
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
	s.reporter.LogT(i18n.MsgTarget,
		s.cfg.TargetDomain, s.cfg.MaxPages, s.cfg.DurationMinutes, s.cfg.HitsPerMinute, workers)

	// 1. Sayfa keşfi
	s.reporter.LogT(i18n.MsgDiscovery)
	pages, err := s.crawler.Discover()
	if err != nil {
		s.reporter.LogT(i18n.MsgDiscoveryErr, err.Error())
		pages = []string{"https://" + s.cfg.TargetDomain}
	}
	s.pages = pages
	s.reporter.LogT(i18n.MsgPagesFound, len(pages), pages)

	// 2. Paralel işçi havuzu - en fazla workers kadar eşzamanlı ziyaret
	sem := make(chan struct{}, workers)
	var hitCount int64
	var wg sync.WaitGroup

	deadline := time.Now().Add(s.cfg.Duration)
	interval := delay.RequestInterval(s.cfg.HitsPerMinute)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.reporter.LogT(i18n.MsgCancel)
			wg.Wait()
			s.finish()
			return ctx.Err()

		case <-ticker.C:
			if time.Now().After(deadline) {
				s.reporter.LogT(i18n.MsgDeadline)
				wg.Wait()
				s.finish()
				return nil
			}

			// Yeni ziyaret başlat (paralel - semaphore ile sınırlı)
			select {
			case sem <- struct{}{}:
				// Slot aldık, goroutine'de ziyaret yap
				wg.Add(1)
				page := s.pickPage()
				go func(url string) {
					defer wg.Done()
					defer func() { <-sem }()
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
				// Tüm slotlar dolu, bu tick'i atla (bir sonraki tick'te tekrar dene)
			}

			jitter := delay.Jitter(100*time.Millisecond, 30)
			select {
			case <-time.After(jitter):
			case <-ctx.Done():
				wg.Wait()
				s.finish()
				return ctx.Err()
			}
		}
	}
}

func (s *Simulator) pickPage() string {
	if len(s.pages) == 0 {
		return "https://" + s.cfg.TargetDomain
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
