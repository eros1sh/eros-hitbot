package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// LiveProxy çalışan proxy + hız ve ülke bilgisi
type LiveProxy struct {
	*ProxyConfig
	SpeedMs   int64  `json:"speed_ms"`
	Country   string `json:"country"`
	CheckedAt time.Time `json:"checked_at"`
}

// ip-api.com yanıtı (minimal)
type ipAPIResponse struct {
	Country string `json:"country"`
	Query   string `json:"query"`
	Status  string `json:"status"`
}

// Checker proxy'leri test eder; çalışanları LiveProxy olarak döner
type Checker struct {
	Client       *http.Client
	TestURL      string // Boşsa ip-api kullanılır
	Workers      int
	TimeoutPerProxy time.Duration
}

// NewChecker varsayılan ayarlarla checker oluşturur
func NewChecker(workers int) *Checker {
	if workers <= 0 {
		workers = 10 // SECURITY FIX: Default 25'ten 10'a düşürüldü, connection exhausting riski
	}
	if workers > 50 {
		workers = 50 // SECURITY FIX: Hard limit, sistem kaynaklarını korumak için
	}
	return &Checker{
		Client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				// PERFORMANCE: Daha iyi connection pooling
				MaxIdleConns:        50,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     20,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  true,
				// PERFORMANCE: Keep-alive ayarları
				ForceAttemptHTTP2:   false, // HTTP/1.1 daha stabil
			},
		},
		TestURL:         "", // ip-api kullan
		Workers:         workers,
		TimeoutPerProxy: 10 * time.Second, // 12'den 10'a düşürüldü
	}
}

// CheckOne proxy'yi test eder; çalışıyorsa LiveProxy döner
func (c *Checker) CheckOne(ctx context.Context, proxy *ProxyConfig) (*LiveProxy, error) {
	testURL := c.TestURL
	if testURL == "" {
		testURL = "http://ip-api.com/json/?fields=status,country,query"
	}
	proxyURL := proxy.ToURL()
	transport := &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		IdleConnTimeout:       10 * time.Second,
		DisableCompression:    true,
		MaxIdleConnsPerHost:   2,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   c.TimeoutPerProxy,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Milliseconds()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var apiResp ipAPIResponse
	if json.NewDecoder(resp.Body).Decode(&apiResp) != nil {
		return &LiveProxy{
			ProxyConfig: proxy,
			SpeedMs:     elapsed,
			Country:     "",
			CheckedAt:   time.Now(),
		}, nil
	}
	country := apiResp.Country
	if country == "" {
		country = "?"
	}
	return &LiveProxy{
		ProxyConfig: proxy,
		SpeedMs:     elapsed,
		Country:     country,
		CheckedAt:   time.Now(),
	}, nil
}

// Run queue'dan proxy alır, test eder; çalışanları onLive gönderir
func (c *Checker) Run(ctx context.Context, queue <-chan *ProxyConfig, onLive func(*LiveProxy)) {
	sem := make(chan struct{}, c.Workers)
	for {
		select {
		case <-ctx.Done():
			return
		case proxy, ok := <-queue:
			if !ok {
				return
			}
			sem <- struct{}{}
			go func(p *ProxyConfig) {
				defer func() { <-sem }()
				live, err := c.CheckOne(ctx, p)
				if err == nil && live != nil {
					onLive(live)
				}
			}(proxy)
		}
	}
}

// RunSlice queue slice'ı ile çalıştırır; çalışanları liveChan'a yazar; bittiğinde liveChan'ı kapatır
func (c *Checker) RunSlice(ctx context.Context, queue []*ProxyConfig, liveChan chan<- *LiveProxy) {
	if liveChan == nil {
		return
	}
	
	// PERFORMANCE: Batch processing ve adaptive rate limiting
	sem := make(chan struct{}, c.Workers)
	var wg sync.WaitGroup
	
	// PERFORMANCE: Her batch arasında kısa pause, sistem yükünü azaltmak için
	batchSize := c.Workers * 2
	for i, p := range queue {
		// Batch boundary'de kısa pause
		if i > 0 && i%batchSize == 0 {
			select {
			case <-ctx.Done():
				wg.Wait()
				close(liveChan)
				return
			case <-time.After(100 * time.Millisecond):
				// Kısa nefes alma süresi
			}
		}
		
		select {
		case <-ctx.Done():
			wg.Wait()
			close(liveChan)
			return
		default:
		}
		
		sem <- struct{}{}
		wg.Add(1)
		go func(proxy *ProxyConfig) {
			defer wg.Done()
			defer func() { <-sem }()
			live, err := c.CheckOne(ctx, proxy)
			if err == nil && live != nil {
				select {
				case liveChan <- live:
				case <-ctx.Done():
				}
			}
		}(p)
	}
	wg.Wait()
	close(liveChan)
}
