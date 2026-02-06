// Package simulator provides optimized simulation engine with browser pooling.
package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"eroshit/internal/config"
	"eroshit/internal/crawler"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/pkg/browser"
	"eroshit/pkg/delay"
	"eroshit/pkg/i18n"
	"eroshit/pkg/sitemap"
)

// OptimizedSimulator uses browser pool for better performance.
type OptimizedSimulator struct {
	cfg           *config.Config
	crawler       *crawler.Crawler
	agentProvider crawler.AgentProvider
	browserPool   *browser.BrowserPool // pkg/browser'dan
	livePool      *proxy.LivePool
	reporter      *reporter.Reporter
	pages         []string
	homepageURL   string
	visitErrAgg   *visitErrAgg
}

// NewOptimized creates an optimized simulator with browser pooling.
func NewOptimized(cfg *config.Config, agentProvider crawler.AgentProvider, rep *reporter.Reporter, livePool *proxy.LivePool) (*OptimizedSimulator, error) {
	if rep == nil {
		rep = reporter.New(cfg.OutputDir, cfg.ExportFormat, cfg.TargetDomain)
	}
	rep.LogT(i18n.MsgStarting)

	// Create browser pool
	poolConfig := browser.PoolConfig{
		MaxInstances:        cfg.MaxConcurrentVisits,
		MinInstances:        minInt(cfg.MaxConcurrentVisits/5, 5), // 20% or max 5
		AcquireTimeout:      30 * time.Second,
		InstanceMaxAge:      30 * time.Minute,
		InstanceMaxSessions: 50,
		ProxyURL:            cfg.ProxyURL,
		ProxyUser:           cfg.ProxyUser,
		ProxyPass:           cfg.ProxyPass,
		Headless:            true,
	}

	pool, err := browser.NewBrowserPool(poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser pool: %w", err)
	}

	c, err := crawler.New(cfg.TargetDomain, cfg.MaxPages, rep, cfg.ProxyURL, agentProvider)
	if err != nil {
		return nil, err
	}

	return &OptimizedSimulator{
		cfg:           cfg,
		crawler:       c,
		agentProvider: agentProvider,
		browserPool:   pool,
		livePool:      livePool,
		reporter:      rep,
		visitErrAgg:   newVisitErrAgg(),
	}, nil
}

// Run starts the optimized simulation.
func (s *OptimizedSimulator) Run(ctx context.Context) error {
	workers := s.cfg.MaxConcurrentVisits
	if workers <= 0 {
		workers = 10
	}
	if workers > 50 {
		workers = 50
	}

	hpm := s.cfg.HitsPerMinute
	if hpm <= 0 {
		hpm = 35
	}

	s.reporter.LogT(i18n.MsgTarget,
		s.cfg.TargetDomain, s.cfg.MaxPages, s.cfg.DurationMinutes, hpm, workers)

	// Discovery phase
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

	// Token bucket for rate limiting
	tb := delay.NewTokenBucket(ctx, hpm, workers)
	defer tb.Stop()

	deadline := time.Now().Add(s.cfg.Duration)
	var hitCount int64
	var wg sync.WaitGroup

	// Worker pool with semaphore
	slotFreed := make(chan struct{}, workers)

	// Pre-fill slotFreed
	for i := 0; i < workers; i++ {
		slotFreed <- struct{}{}
	}

	// Visit function with pooled browser
	startVisit := func() {
		if err := tb.Take(ctx); err != nil {
			return
		}
		if time.Now().After(deadline) {
			return
		}

		select {
		case <-slotFreed:
			wg.Add(1)
			page := s.pickPage()
			go func(url string) {
				defer wg.Done()
				defer func() { slotFreed <- struct{}{} }()

				// Acquire browser from pool
				instance, err := s.browserPool.Acquire(ctx)
				if err != nil {
					s.visitErrAgg.add(s.reporter, url, err)
					return
				}
				defer s.browserPool.Release(instance)

				// Execute visit
				visitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
				defer cancel()

				// Use pooled browser for visit
				if err := s.visitWithPooledBrowser(visitCtx, instance, url); err != nil {
					s.visitErrAgg.add(s.reporter, url, err)
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
		}
	}

	// Main event loop
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.reporter.LogT(i18n.MsgCancel)
			tb.Stop()
			wg.Wait()
			s.finish()
			return ctx.Err()

		case <-ticker.C:
			if time.Now().After(deadline) {
				s.reporter.LogT(i18n.MsgDeadline)
				tb.Stop()
				wg.Wait()
				s.finish()
				return nil
			}

			select {
			case <-slotFreed:
				go startVisit()
			default:
			}
		}
	}
}

// visitWithPooledBrowser performs a visit using a pooled browser instance.
func (s *OptimizedSimulator) visitWithPooledBrowser(ctx context.Context, instance *browser.BrowserInstance, url string) error {
	// This is a simplified version - in production, you'd integrate with hit.go logic
	// For now, just verify the browser context is valid
	if instance == nil {
		return fmt.Errorf("no browser instance available")
	}

	// The actual visit logic would use instance.GetContext() with chromedp
	// and reuse the existing browser instance
	
	// Simulate work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Simulated visit time
	}

	return nil
}

func (s *OptimizedSimulator) pickPage() string {
	if len(s.pages) == 0 {
		return s.homepageURL
	}
	weight := s.cfg.SitemapHomepageWeight
	if weight <= 0 {
		weight = 60
	}
	if s.homepageURL != "" && rand.Intn(100) < weight {
		return s.homepageURL
	}
	return s.pages[rand.Intn(len(s.pages))]
}

// Reporter returns the reporter instance.
func (s *OptimizedSimulator) Reporter() *reporter.Reporter {
	return s.reporter
}

func (s *OptimizedSimulator) finish() {
	if s.browserPool != nil {
		s.browserPool.Close()
	}
	if s.visitErrAgg != nil {
		s.visitErrAgg.flush(s.reporter)
	}
	s.reporter.Finalize()
	m := s.reporter.GetMetrics()
	s.reporter.LogT(i18n.MsgSummary)
	s.reporter.LogT(i18n.MsgSummaryLine, m.TotalHits, m.SuccessHits, m.FailedHits)
	s.reporter.LogT(i18n.MsgSummaryRT, m.AvgResponseTime, m.MinResponseTime, m.MaxResponseTime)

	if err := s.reporter.Export(); err != nil {
		s.reporter.LogT(i18n.MsgExportErr, err.Error())
	}
	s.reporter.Close()
}

// minInt returns the smaller of two int values
// CODE FIX: Renamed from 'min' to avoid conflict with Go 1.21+ built-in min function
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
