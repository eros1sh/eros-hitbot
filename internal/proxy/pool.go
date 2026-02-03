package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ProxyConfig proxy yapılandırması
type ProxyConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Protocol string
}

// Key benzersiz proxy anahtarı
func (pc *ProxyConfig) Key() string {
	return fmt.Sprintf("%s:%d", pc.Host, pc.Port)
}

// ToURL proxy URL'i
func (pc *ProxyConfig) ToURL() *url.URL {
	hostPort := fmt.Sprintf("%s:%d", pc.Host, pc.Port)
	var userInfo *url.Userinfo
	if pc.Username != "" || pc.Password != "" {
		userInfo = url.UserPassword(pc.Username, pc.Password)
	}
	protocol := pc.Protocol
	if protocol == "" {
		protocol = "http"
	}
	return &url.URL{
		Scheme: protocol,
		Host:   hostPort,
		User:   userInfo,
	}
}

// ToURLString Colly/Chromedp için string URL
func (pc *ProxyConfig) ToURLString() string {
	u := pc.ToURL()
	return u.String()
}

// FailureInfo proxy başarısızlık bilgisi
type FailureInfo struct {
	FailCount   int
	LastFailure time.Time
	NextRetry   time.Time
}

// ProxyMetrics proxy performans metrikleri
type ProxyMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	AvgResponseTime time.Duration
	LastUsed        time.Time
}

// ProxyPool proxy havuzu
type ProxyPool struct {
	proxies        []*ProxyConfig
	current        int
	mu             sync.RWMutex
	healthCheck    bool
	healthInterval time.Duration
	failedProxies  map[string]*FailureInfo
	metrics        map[string]*ProxyMetrics
}

// NewProxyPool yeni proxy havuzu oluşturur
func NewProxyPool(proxies []*ProxyConfig, healthCheck bool) *ProxyPool {
	if len(proxies) == 0 {
		return nil
	}
	pool := &ProxyPool{
		proxies:        proxies,
		current:        0,
		healthCheck:    healthCheck,
		healthInterval: 5 * time.Minute,
		failedProxies:  make(map[string]*FailureInfo),
		metrics:        make(map[string]*ProxyMetrics),
	}
	for _, p := range proxies {
		pool.metrics[p.Key()] = &ProxyMetrics{}
	}
	return pool
}

// GetNext sıradaki proxy'yi döner (round-robin, başarısızları atlar)
func (p *ProxyPool) GetNext() *ProxyConfig {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.proxies) == 0 {
		return nil
	}

	maxAttempts := len(p.proxies)
	for attempts := 0; attempts < maxAttempts; attempts++ {
		proxy := p.proxies[p.current]
		p.current = (p.current + 1) % len(p.proxies)

		if failure, exists := p.failedProxies[proxy.Key()]; exists {
			if time.Now().Before(failure.NextRetry) {
				continue
			}
			delete(p.failedProxies, proxy.Key())
		}
		return proxy
	}
	return p.proxies[0]
}

// MarkSuccess başarılı istek kaydeder
func (p *ProxyPool) MarkSuccess(proxy *ProxyConfig, responseTime time.Duration) {
	if proxy == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	key := proxy.Key()
	m := p.metrics[key]
	if m == nil {
		m = &ProxyMetrics{}
		p.metrics[key] = m
	}

	m.TotalRequests++
	m.SuccessRequests++
	m.LastUsed = time.Now()

	if m.AvgResponseTime == 0 {
		m.AvgResponseTime = responseTime
	} else {
		alpha := 0.3
		m.AvgResponseTime = time.Duration(
			float64(m.AvgResponseTime)*(1-alpha) + float64(responseTime)*alpha,
		)
	}

	delete(p.failedProxies, key)
}

// MarkFailed başarısız istek kaydeder
func (p *ProxyPool) MarkFailed(proxy *ProxyConfig, err error) {
	if proxy == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	key := proxy.Key()
	m := p.metrics[key]
	if m != nil {
		m.TotalRequests++
		m.FailedRequests++
	}

	if failure, exists := p.failedProxies[key]; exists {
		failure.FailCount++
		failure.LastFailure = time.Now()
		exp := failure.FailCount - 1
		if exp > 6 {
			exp = 6
		}
		retryDelay := time.Minute * time.Duration(1<<exp)
		failure.NextRetry = time.Now().Add(retryDelay)
	} else {
		p.failedProxies[key] = &FailureInfo{
			FailCount:   1,
			LastFailure: time.Now(),
			NextRetry:   time.Now().Add(1 * time.Minute),
		}
	}
}

// StartHealthCheck arka planda sağlık kontrolü başlatır
func (p *ProxyPool) StartHealthCheck(ctx context.Context) {
	if !p.healthCheck {
		return
	}
	ticker := time.NewTicker(p.healthInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.runHealthChecks()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *ProxyPool) runHealthChecks() {
	p.mu.RLock()
	proxies := make([]*ProxyConfig, len(p.proxies))
	copy(proxies, p.proxies)
	p.mu.RUnlock()

	for _, px := range proxies {
		if p.checkProxyHealth(px) {
			p.MarkSuccess(px, 0)
		} else {
			p.MarkFailed(px, fmt.Errorf("health check failed"))
		}
	}
}

func (p *ProxyPool) checkProxyHealth(proxy *ProxyConfig) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy.ToURL()),
		},
	}
	// SECURITY FIX: Use HTTPS instead of HTTP for health checks
	resp, err := client.Get("https://www.google.com/robots.txt")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// GetMetrics tüm proxy metriklerini döner
func (p *ProxyPool) GetMetrics() map[string]*ProxyMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make(map[string]*ProxyMetrics)
	for k, v := range p.metrics {
		cp := *v
		out[k] = &cp
	}
	return out
}

// GetBestProxy en iyi performanslı proxy'yi döner
func (p *ProxyPool) GetBestProxy() *ProxyConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return nil
	}

	var best *ProxyConfig
	bestScore := -1.0

	for _, proxy := range p.proxies {
		key := proxy.Key()
		if _, failed := p.failedProxies[key]; failed {
			continue
		}
		m := p.metrics[key]
		if m == nil || m.TotalRequests == 0 {
			return proxy
		}
		successRate := float64(m.SuccessRequests) / float64(m.TotalRequests)
		speed := 1.0
		if m.AvgResponseTime.Milliseconds() > 0 {
			speed = 1000.0 / float64(m.AvgResponseTime.Milliseconds())
		}
		score := successRate * speed
		if score > bestScore {
			bestScore = score
			best = proxy
		}
	}

	if best == nil {
		return p.proxies[0]
	}
	return best
}
