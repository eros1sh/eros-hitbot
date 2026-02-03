package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"eroshit/internal/browser"
	"eroshit/internal/config"
	"eroshit/internal/crawler"
	"eroshit/internal/proxy"
	"eroshit/internal/reporter"
	"eroshit/pkg/analytics"
	"eroshit/pkg/delay"
	"eroshit/pkg/i18n"
	"eroshit/pkg/sitemap"
)

// visitorSlot public proxy modunda her slot: bir visitor + bir proxy; başarısız olunca visitor kapatılır, proxy havuzdan silinir

// Simulator trafik simülasyonu orkestratörü
type Simulator struct {
	cfg          *config.Config
	crawler      *crawler.Crawler
	agentProvider crawler.AgentProvider
	hitVisitor   *browser.HitVisitor
	livePool     *proxy.LivePool
	reporter     *reporter.Reporter
	pages        []string
	homepageURL  string
	visitErrAgg  *visitErrAgg
}

type visitorSlot struct {
	mu      sync.Mutex
	visitor *browser.HitVisitor
	proxy   *proxy.ProxyConfig
}

// visitErrAgg ziyaret hatalarını toplar; log spam'ı azaltmak için özet satır yazar
type visitErrAgg struct {
	mu        sync.Mutex
	counts    map[string]int
	total     int
	lastFlush time.Time
	flushDur  time.Duration
	flushCap  int
}

func newVisitErrAgg() *visitErrAgg {
	return &visitErrAgg{
		counts:    make(map[string]int),
		flushDur:  25 * time.Second,
		flushCap:  12,
	}
}

func errKey(err error) string {
	s := err.Error()
	if i := strings.Index(s, "net::"); i >= 0 {
		rest := s[i+5:]
		if end := strings.IndexAny(rest, " ):\t"); end > 0 {
			return rest[:end]
		}
		return rest
	}
	if i := strings.Index(s, "ERR_"); i >= 0 {
		rest := s[i:]
		if end := strings.IndexAny(rest, " ):,\t"); end > 0 {
			return rest[:end]
		}
		return rest
	}
	if len(s) > 45 {
		return s[:45] + "…"
	}
	return s
}

func (a *visitErrAgg) add(rep *reporter.Reporter, _ string, err error) {
	key := errKey(err)
	a.mu.Lock()
	a.counts[key]++
	a.total++
	shouldFlush := a.total >= a.flushCap || time.Since(a.lastFlush) >= a.flushDur
	if shouldFlush {
		a.flushLocked(rep)
	}
	a.mu.Unlock()
}

func (a *visitErrAgg) flush(rep *reporter.Reporter) {
	a.mu.Lock()
	a.flushLocked(rep)
	a.mu.Unlock()
}

func (a *visitErrAgg) flushLocked(rep *reporter.Reporter) {
	if a.total == 0 {
		return
	}
	parts := make([]string, 0, len(a.counts))
	for k, n := range a.counts {
		parts = append(parts, fmt.Sprintf("%s: %d", k, n))
	}
	sort.Strings(parts)
	rep.LogT(i18n.MsgVisitErrSummary, a.total, strings.Join(parts, ", "))
	a.counts = make(map[string]int)
	a.total = 0
	a.lastFlush = time.Now()
}

// New simulator oluşturur. agentProvider ve rep nil olabilir.
// livePool verilirse public proxy modu: çalışan proxy'lerle ziyaret; başarısız proxy havuzdan silinir.
func New(cfg *config.Config, agentProvider crawler.AgentProvider, rep *reporter.Reporter, livePool *proxy.LivePool) (*Simulator, error) {
	if rep == nil {
		rep = reporter.New(cfg.OutputDir, cfg.ExportFormat, cfg.TargetDomain)
	}
	rep.LogT(i18n.MsgStarting)

	// SECURITY FIX: Proxy URL'yi doğru şekilde oluştur - auth bilgisi dahil
	proxyURL := ""
	if livePool == nil && cfg.ProxyEnabled {
		// Öncelik: ProxyURL (auth dahil), sonra ProxyBaseURL (auth yok)
		if cfg.ProxyURL != "" {
			proxyURL = cfg.ProxyURL
		} else if cfg.ProxyBaseURL != "" {
			proxyURL = cfg.ProxyBaseURL
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

	var hitVisitor *browser.HitVisitor
	if livePool == nil {
		var errHv error
		hitVisitor, errHv = browser.NewHitVisitor(agentProvider, rep, browser.HitVisitorConfig{
			ProxyURL:          proxyURL,
			ProxyUser:         cfg.ProxyUser,
			ProxyPass:         cfg.ProxyPass,
			GtagID:            cfg.GtagID,
			CanvasFingerprint: cfg.CanvasFingerprint,
			ScrollStrategy:    cfg.ScrollStrategy,
			SendScrollEvent:   cfg.SendScrollEvent,
			AnalyticsManager:  analyticsMgr,
			Keywords:          cfg.Keywords,
			// Yeni alanlar
			DeviceType:        cfg.DeviceType,
			DeviceBrands:      cfg.DeviceBrands,
			ReferrerKeyword:   cfg.ReferrerKeyword,
			ReferrerEnabled:   cfg.ReferrerEnabled,
		})
		if errHv != nil {
			return nil, errHv
		}
	}

	return &Simulator{
		cfg:           cfg,
		crawler:       c,
		agentProvider: agentProvider,
		hitVisitor:    hitVisitor,
		livePool:      livePool,
		reporter:      rep,
		pages:         nil,
		visitErrAgg:   newVisitErrAgg(),
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
	var hitCount int64
	var wg sync.WaitGroup

	if s.livePool == nil {
		// Tek proxy / proxy yok: tek HitVisitor, sem ile eşzamanlı ziyaret
		sem := make(chan struct{}, workers)
		slotFreed := make(chan struct{}, workers)
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
				go startVisit()
			}
		}
	}

	// Public proxy: slot başına bir visitor + proxy; başarısız ziyarette proxy havuzdan silinir, visitor yenilenir
	slots := make([]*visitorSlot, workers)
	for i := 0; i < workers; i++ {
		slots[i] = &visitorSlot{}
	}
	slotPool := make(chan int, workers)
	slotFreed := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		slotPool <- i
		slotFreed <- struct{}{}
	}
	analyticsMgr := &analytics.Manager{
		GA4Enabled:       s.cfg.GtagID != "",
		GA4MeasurementID: s.cfg.GtagID,
	}

	startVisitPublic := func() {
		if err := tb.Take(ctx); err != nil {
			return
		}
		if time.Now().After(deadline) {
			return
		}
		select {
		case idx := <-slotPool:
			slot := slots[idx]
			slot.mu.Lock()
			hv := slot.visitor
			pc := slot.proxy
			if hv == nil {
				pc = s.livePool.GetNext()
				if pc == nil {
					slot.mu.Unlock()
					slotPool <- idx
					slotFreed <- struct{}{}
					return
				}
				var errHv error
				hv, errHv = browser.NewHitVisitor(s.agentProvider, s.reporter, browser.HitVisitorConfig{
					ProxyURL:          pc.ToURLString(),
					ProxyUser:         s.cfg.ProxyUser,
					ProxyPass:         s.cfg.ProxyPass,
					GtagID:            s.cfg.GtagID,
					CanvasFingerprint: s.cfg.CanvasFingerprint,
					ScrollStrategy:    s.cfg.ScrollStrategy,
					SendScrollEvent:   s.cfg.SendScrollEvent,
					AnalyticsManager:  analyticsMgr,
					Keywords:          s.cfg.Keywords,
					// Yeni alanlar
					DeviceType:        s.cfg.DeviceType,
					DeviceBrands:      s.cfg.DeviceBrands,
					ReferrerKeyword:   s.cfg.ReferrerKeyword,
					ReferrerEnabled:   s.cfg.ReferrerEnabled,
				})
				if errHv != nil {
					slot.mu.Unlock()
					slotPool <- idx
					slotFreed <- struct{}{}
					return
				}
				slot.visitor = hv
				slot.proxy = pc
			}
			slot.mu.Unlock()
			
			// Nil pointer kontrolü - visitor nil ise devam etme
			if hv == nil {
				slotPool <- idx
				slotFreed <- struct{}{}
				return
			}
			
			page := s.pickPage()
			wg.Add(1)
			go func(url string, slotIdx int, visitor *browser.HitVisitor, proxyCfg *proxy.ProxyConfig) {
				defer wg.Done()
				defer func() { slotPool <- slotIdx; slotFreed <- struct{}{} }()
				
				// Ek nil kontrolü - goroutine içinde
				if visitor == nil {
					return
				}
				
				err := visitor.VisitURL(ctx, url)
				if err != nil {
					s.visitErrAgg.add(s.reporter, url, err)
					s.livePool.Remove(proxyCfg)
					visitor.Close()
					slots[slotIdx].mu.Lock()
					slots[slotIdx].visitor = nil
					slots[slotIdx].proxy = nil
					slots[slotIdx].mu.Unlock()
				} else {
					n := atomic.AddInt64(&hitCount, 1)
					if n%10 == 0 {
						m := s.reporter.GetMetrics()
						s.reporter.LogT(i18n.MsgProgress,
							n, m.TotalHits, m.SuccessHits, m.FailedHits, m.AvgResponseTime)
					}
				}
			}(page, idx, hv, pc)
		default:
			// Tüm slotlar meşgul
		}
	}

	deadlineTimer := time.NewTimer(time.Until(deadline))
	defer func() {
		if !deadlineTimer.Stop() {
			select { case <-deadlineTimer.C: default: }
		}
		for _, slot := range slots {
			slot.mu.Lock()
			if slot.visitor != nil {
				slot.visitor.Close()
				slot.visitor = nil
			}
			slot.proxy = nil
			slot.mu.Unlock()
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
			go startVisitPublic()
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
	
	// Log kanalını kapat - memory leak önleme
	s.reporter.Close()
}
