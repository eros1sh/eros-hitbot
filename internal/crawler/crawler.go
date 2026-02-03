package crawler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"

	"eroshit/internal/reporter"
	"eroshit/pkg/useragent"
)

// Crawler domain-scoped, etik crawler
type Crawler struct {
	collector     *colly.Collector
	baseURL       string
	domain        string
	maxPages      int
	visited       map[string]bool
	mu            sync.Mutex
	pages         []string
	reporter      *reporter.Reporter
	agentProvider AgentProvider
}

// AgentProvider UA ve opsiyonel headers sağlar
type AgentProvider interface {
	RandomWithHeaders() (ua string, headers map[string]string)
}

// New yeni crawler oluşturur. agentProvider nil ise varsayılan UA kullanılır.
func New(domain string, maxPages int, rep *reporter.Reporter, proxyURL string, agentProvider AgentProvider) (*Crawler, error) {
	domain = strings.TrimSpace(domain)
	if !strings.HasPrefix(domain, "http") {
		domain = "https://" + domain
	}

	u, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	baseDomain := u.Hostname()
	if baseDomain == "" {
		baseDomain = domain
	}

	c := colly.NewCollector(
		colly.AllowedDomains(baseDomain),
		colly.Async(true),
		colly.MaxDepth(2),
		colly.AllowURLRevisit(), // Aynı URL'e binlerce hit - trafik simülasyonu için
	)

	// Connection pool - trafik simülasyonu için yeterli paralellik
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 6,
		RandomDelay: 200 * time.Millisecond,
	})

	if proxyURL != "" {
		if err := c.SetProxy(proxyURL); err != nil {
			return nil, err
		}
	}

	// Keep-alive için transport
	c.WithTransport(&http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	})

	cr := &Crawler{
		collector:      c,
		baseURL:        strings.TrimSuffix(domain, "/"),
		domain:         baseDomain,
		maxPages:       maxPages,
		visited:        make(map[string]bool),
		pages:          make([]string, 0, maxPages),
		reporter:       rep,
		agentProvider:  agentProvider,
	}

	cr.setupHandlers()
	return cr, nil
}

func (cr *Crawler) setupHandlers() {
	var startTimeMu sync.Mutex
	startTime := make(map[string]time.Time)

	cr.collector.OnRequest(func(r *colly.Request) {
		var ua string
		var headers map[string]string
		if cr.agentProvider != nil {
			ua, headers = cr.agentProvider.RandomWithHeaders()
		} else {
			ua = useragent.Random()
		}
		r.Headers.Set("User-Agent", ua)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7")
		for k, v := range headers {
			r.Headers.Set(k, v)
		}
		startTimeMu.Lock()
		startTime[r.URL.String()] = time.Now()
		startTimeMu.Unlock()
	})

	cr.collector.OnResponse(func(r *colly.Response) {
		u := r.Request.URL.String()
		startTimeMu.Lock()
		start, ok := startTime[u]
		if ok {
			delete(startTime, u)
		}
		startTimeMu.Unlock()
		if !ok {
			start = time.Now()
		}
		elapsed := time.Since(start).Milliseconds()

		cr.reporter.Record(reporter.HitRecord{
			Timestamp:    time.Now(),
			URL:          u,
			StatusCode:   r.StatusCode,
			ResponseTime: elapsed,
			UserAgent:    r.Request.Headers.Get("User-Agent"),
		})
	})

	cr.collector.OnError(func(r *colly.Response, err error) {
		u := ""
		if r != nil && r.Request != nil {
			u = r.Request.URL.String()
		}
		cr.reporter.Record(reporter.HitRecord{
			Timestamp: time.Now(),
			URL:       u,
			Error:     err.Error(),
		})
	})

	cr.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Request.AbsoluteURL(e.Attr("href"))
		if href == "" {
			return
		}
		cr.maybeVisit(href)
	})

	cr.collector.OnScraped(func(r *colly.Response) {
		cr.mu.Lock()
		if !cr.visited[r.Request.URL.String()] {
			cr.visited[r.Request.URL.String()] = true
			cr.pages = append(cr.pages, r.Request.URL.String())
		}
		cr.mu.Unlock()
	})
}

func (cr *Crawler) maybeVisit(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if u.Hostname() != cr.domain {
		return
	}
	// Sadece HTTP(S)
	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}
	// Fragment ve query temizle (aynı sayfa tekrarı önle)
	u.Fragment = ""
	normalized := u.String()

	cr.mu.Lock()
	if cr.visited[normalized] || len(cr.pages) >= cr.maxPages {
		cr.mu.Unlock()
		return
	}
	cr.visited[normalized] = true
	cr.mu.Unlock()

	_ = cr.collector.Visit(normalized)
}

// Discover ana sayfadan başlayarak sayfaları keşfeder
func (cr *Crawler) Discover() ([]string, error) {
	if err := cr.collector.Visit(cr.baseURL); err != nil {
		return nil, err
	}
	cr.collector.Wait()

	cr.mu.Lock()
	pages := make([]string, len(cr.pages))
	copy(pages, cr.pages)
	cr.mu.Unlock()

	if len(pages) == 0 {
		pages = []string{cr.baseURL}
	}
	return pages, nil
}

// VisitURL tek bir URL'i ziyaret eder (simülasyon için)
func (cr *Crawler) VisitURL(urlStr string) error {
	return cr.collector.Visit(urlStr)
}
