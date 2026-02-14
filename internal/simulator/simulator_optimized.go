// Package simulator provides optimized simulation engine with browser pooling.
package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"eroshit/internal/config"
	"eroshit/internal/crawler"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/pkg/analytics"
	"eroshit/pkg/behavior"
	"eroshit/pkg/browser"
	"eroshit/pkg/canvas"
	"eroshit/pkg/delay"
	"eroshit/pkg/engagement"
	"eroshit/pkg/fingerprint"
	"eroshit/pkg/i18n"
	"eroshit/pkg/mobile"
	"eroshit/pkg/referrer"
	"eroshit/pkg/sitemap"
	"eroshit/pkg/stealth"
	"eroshit/pkg/useragent"
)

// OptimizedSimulator uses browser pool for better performance.
type OptimizedSimulator struct {
	cfg           *config.Config
	crawler       *crawler.Crawler
	agentProvider crawler.AgentProvider
	browserPool   *browser.BrowserPool
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

	poolConfig := browser.PoolConfig{
		MaxInstances:        cfg.MaxConcurrentVisits,
		MinInstances:        minInt(cfg.MaxConcurrentVisits/5, 5),
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
	for i := 0; i < workers; i++ {
		slotFreed <- struct{}{}
	}

	// Main event loop - BUG FIX #2 pattern: tek slot tüketimi
	ticker := time.NewTicker(20 * time.Millisecond)
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
				wg.Add(1)
				visitPage := s.pickPage()
				go func(targetURL string) {
					defer wg.Done()
					defer func() { slotFreed <- struct{}{} }()

					if err := tb.Take(ctx); err != nil {
						return
					}
					if time.Now().After(deadline) {
						return
					}

					instance, err := s.browserPool.Acquire(ctx)
					if err != nil {
						s.visitErrAgg.add(s.reporter, targetURL, err)
						return
					}
					defer s.browserPool.Release(instance)

					visitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
					defer cancel()

					if err := s.visitWithPooledBrowser(visitCtx, instance, targetURL); err != nil {
						s.visitErrAgg.add(s.reporter, targetURL, err)
					} else {
						n := atomic.AddInt64(&hitCount, 1)
						if n%10 == 0 {
							m := s.reporter.GetMetrics()
							s.reporter.LogT(i18n.MsgProgress,
								n, m.TotalHits, m.SuccessHits, m.FailedHits, m.AvgResponseTime)
						}
					}
				}(visitPage)
			default:
			}
		}
	}
}

// BUG FIX #1: Gerçek ziyaret implementasyonu - stub kaldırıldı
func (s *OptimizedSimulator) visitWithPooledBrowser(ctx context.Context, instance *browser.BrowserInstance, urlStr string) error {
	if instance == nil {
		return fmt.Errorf("no browser instance available")
	}

	// Yeni tab oluştur (pooled browser'da)
	tabCtx, tabCancel, err := instance.CreateNewTab(ctx)
	if err != nil {
		return fmt.Errorf("failed to create tab: %w", err)
	}
	defer tabCancel()

	// Timeout uygula
	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, 90*time.Second)
	defer timeoutCancel()

	start := time.Now()

	// UA ve fingerprint oluştur
	var ua string
	var isMobile bool

	if s.cfg.DeviceType != "" && s.cfg.DeviceType != "mixed" {
		device := mobile.GetRandomDeviceFiltered(s.cfg.DeviceType, s.cfg.DeviceBrands)
		ua = device.UserAgent
		isMobile = device.Mobile
	} else if len(s.cfg.DeviceBrands) > 0 {
		device := mobile.GetRandomDeviceFiltered("mixed", s.cfg.DeviceBrands)
		ua = device.UserAgent
		isMobile = device.Mobile
	} else if s.agentProvider != nil {
		ua, _ = s.agentProvider.RandomWithHeaders()
	}
	if ua == "" {
		ua = useragent.Random()
	}

	advFP := fingerprint.GenerateAdvancedFingerprint()
	advFP.UserAgent = ua
	fp := fingerprint.FP{
		Platform:     advFP.Platform,
		Language:     advFP.Language,
		Languages:    strings.Join(advFP.Languages, ", "),
		InnerW:       advFP.ScreenWidth - 22,
		InnerH:       advFP.ScreenHeight - 100,
		DevicePixel:  advFP.ScreenPixelRatio,
		Timezone:     advFP.Timezone,
		HardwareConc: advFP.HardwareConcurrency,
		DeviceMem:    int64(advFP.DeviceMemory),
		Vendor:       advFP.WebGLVendor,
	}
	if fp.InnerW <= 0 {
		fp.InnerW = 1366
	}
	if fp.InnerH <= 0 {
		fp.InnerH = 768
	}

	// Stealth config
	stealthCfg := stealth.StealthConfig{
		UserAgent:           ua,
		Platform:            advFP.Platform,
		Vendor:              advFP.WebGLVendor,
		WebGLVendor:         advFP.WebGLVendor,
		WebGLRenderer:       advFP.WebGLRenderer,
		Languages:           advFP.Languages,
		Plugins:             stealth.GetDefaultStealthConfig().Plugins,
		ScreenWidth:         advFP.ScreenWidth,
		ScreenHeight:        advFP.ScreenHeight,
		AvailWidth:          advFP.AvailWidth,
		AvailHeight:         advFP.AvailHeight,
		ColorDepth:          advFP.ScreenColorDepth,
		PixelDepth:          advFP.ScreenColorDepth,
		HardwareConcurrency: advFP.HardwareConcurrency,
		DeviceMemory:        int(advFP.DeviceMemory),
	}
	if stealthCfg.ScreenWidth <= 0 {
		stealthCfg.ScreenWidth = 1920
	}
	if stealthCfg.ScreenHeight <= 0 {
		stealthCfg.ScreenHeight = 1080
	}
	if stealthCfg.AvailWidth <= 0 {
		stealthCfg.AvailWidth = stealthCfg.ScreenWidth
	}
	if stealthCfg.AvailHeight <= 0 {
		stealthCfg.AvailHeight = stealthCfg.ScreenHeight - 40
	}

	stealthScript := stealth.GetOnNewDocumentScript(stealthCfg)

	// Gerçek HTTP status kodu yakalama
	var realStatusCode int
	var statusMu sync.Mutex
	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if resp, ok := ev.(*network.EventResponseReceived); ok {
			if resp.Type == network.ResourceTypeDocument {
				statusMu.Lock()
				realStatusCode = int(resp.Response.Status)
				statusMu.Unlock()
			}
		}
	})

	// Referrer oluştur
	targetDomain := urlStr
	if idx := strings.Index(urlStr, "://"); idx >= 0 {
		targetDomain = urlStr[idx+3:]
	}
	if idx := strings.Index(targetDomain, "/"); idx >= 0 {
		targetDomain = targetDomain[:idx]
	}

	var referrerURL string
	if s.cfg.ReferrerEnabled && s.cfg.ReferrerKeyword != "" {
		encodedKeyword := url.QueryEscape(s.cfg.ReferrerKeyword)
		referrerURL = fmt.Sprintf("https://www.google.com/search?q=%s", encodedKeyword)
	} else if len(s.cfg.Keywords) > 0 {
		refCfg := &referrer.ReferrerConfig{
			GooglePercent: 50, BingPercent: 20, DirectPercent: 30,
			Keywords: s.cfg.Keywords,
		}
		refChain := referrer.NewReferrerChain(targetDomain, refCfg)
		src := refChain.Generate()
		if src != nil && src.URL != "" && (src.Type == "search" || src.Type == "social") {
			referrerURL = src.URL
		}
	}

	// Navigation actions
	navActions := []chromedp.Action{
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(stealthScript).Do(ctx)
			return err
		}),
		emulation.SetUserAgentOverride(ua),
		emulation.SetDeviceMetricsOverride(int64(fp.InnerW), int64(fp.InnerH), fp.DevicePixel, isMobile),
		emulation.SetTimezoneOverride(fp.Timezone),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.ClearBrowserCookies().Do(ctx)
		}),
	}

	// Touch emülasyonu (mobil)
	if isMobile {
		navActions = append(navActions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetTouchEmulationEnabled(true).WithMaxTouchPoints(5).Do(ctx)
		}))
	}

	// Navigate with or without referrer
	if referrerURL != "" {
		navActions = append(navActions, chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, _, err := page.Navigate(urlStr).WithReferrer(referrerURL).Do(ctx)
			return err
		}))
	} else {
		navActions = append(navActions, chromedp.Navigate(urlStr))
	}

	navActions = append(navActions,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)

	navErr := chromedp.Run(tabCtx, navActions...)

	// GA4 injection
	if navErr == nil && s.cfg.GtagID != "" {
		gtagID := s.cfg.GtagID
		isValid := false
		if len(gtagID) >= 10 && len(gtagID) <= 15 {
			if strings.HasPrefix(gtagID, "G-") || strings.HasPrefix(gtagID, "GT-") {
				suffix := gtagID[strings.Index(gtagID, "-")+1:]
				isValid = true
				for _, c := range suffix {
					if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
						isValid = false
						break
					}
				}
			}
		}
		if !isValid && len(gtagID) >= 10 && len(gtagID) <= 20 && strings.HasPrefix(gtagID, "UA-") {
			isValid = true
			for _, c := range gtagID[3:] {
				if !((c >= '0' && c <= '9') || c == '-') {
					isValid = false
					break
				}
			}
		}
		if isValid {
			gtagScript := `(function(){
				var s=document.createElement('script');s.async=true;
				s.src='https://www.googletagmanager.com/gtag/js?id=` + gtagID + `';
				document.head.appendChild(s);
				window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}
				gtag('js',new Date());
				gtag('config','` + gtagID + `',{send_page_view:true});
			})();`
			_ = chromedp.Run(tabCtx, chromedp.Evaluate(gtagScript, nil))
			_ = chromedp.Run(tabCtx, chromedp.Sleep(1000*time.Millisecond))
		}
	}

	// Stealth scripts already injected via AddScriptToEvaluateOnNewDocument (runs before page load).
	// Re-injection removed — redundant 14 CDP round-trips eliminated.

	if navErr == nil {
		// Canvas/WebGL/Audio fingerprint — single batched CDP call
		if s.cfg.CanvasFingerprint {
			cf := canvas.GenerateFingerprint()
			_ = cf.InjectAll(tabCtx)
		}

		// Scroll
		strategy := s.cfg.ScrollStrategy
		if strategy == "" {
			strategy = "gradual"
		}
		_ = engagement.HumanScroll(tabCtx, engagement.ScrollBehavior{
			Strategy:  strategy,
			ReadSpeed: 200,
		})

		// Scroll event (GA4)
		if s.cfg.SendScrollEvent && s.cfg.GtagID != "" {
			analyticsMgr := &analytics.Manager{
				GA4Enabled:       true,
				GA4MeasurementID: s.cfg.GtagID,
			}
			_ = analyticsMgr.SendEvent(tabCtx, analytics.Event{
				Type: analytics.EventScroll, Category: "engagement",
				Action: "scroll", Label: "75%", Value: 75,
			})
		}

		// Human behavior
		mouseMoveProb := 0.5
		if isMobile {
			mouseMoveProb = 0.0
		}
		hum := behavior.NewHumanBehavior(&behavior.BehaviorConfig{
			MinPageDuration:      1 * time.Second,
			MaxPageDuration:      3 * time.Second,
			ScrollProbability:    0.5,
			MouseMoveProbability: mouseMoveProb,
			ClickProbability:     0,
		})
		hum.SimulatePageVisit(tabCtx, 0)
	}

	elapsed := time.Since(start).Milliseconds()

	if navErr != nil {
		s.reporter.Record(reporter.HitRecord{
			Timestamp: time.Now(),
			URL:       urlStr,
			Error:     navErr.Error(),
			UserAgent: ua,
		})
		return navErr
	}

	statusMu.Lock()
	statusCode := realStatusCode
	statusMu.Unlock()
	if statusCode == 0 {
		statusCode = 200
	}

	s.reporter.Record(reporter.HitRecord{
		Timestamp:    time.Now(),
		URL:          urlStr,
		StatusCode:   statusCode,
		ResponseTime: elapsed,
		UserAgent:    ua,
	})
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

// BUG FIX #20: finish sıralaması düzeltildi - Close en son
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
	// Log ve export bittikten sonra Close
	s.reporter.Close()
}

// minInt returns the smaller of two int values
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
