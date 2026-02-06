// Package metrics provides Prometheus-compatible metrics collection
// for real-time monitoring and dashboard integration.
package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector holds all application metrics with Prometheus compatibility
type MetricsCollector struct {
	// Hit metrikleri
	HitCounter   prometheus.Counter
	HitRate      prometheus.Gauge // Hits per minute
	hitsPerMin   *RateCalculator

	// Performans metrikleri
	ResponseTime prometheus.Histogram
	ProxyLatency *prometheus.HistogramVec // Proxy bazlı

	// Aktif durum
	ActiveSessions prometheus.Gauge
	ActiveProxies  prometheus.Gauge
	QueueSize      prometheus.Gauge

	// Başarı/Kalite
	SuccessRate prometheus.Gauge
	BounceRate  prometheus.Gauge
	ErrorRate   prometheus.Gauge

	// Proxy performansı
	ProxySuccess *prometheus.CounterVec
	ProxyFailure *prometheus.CounterVec

	// Internal tracking
	mu           sync.RWMutex
	startTime    time.Time
	sessionCount int64
	proxyCount   int64
	queueCount   int64
	successCount int64
	bounceCount  int64
	errorCount   int64
	totalHits    int64
}

// RateCalculator calculates hits per minute using a sliding window
type RateCalculator struct {
	mu       sync.Mutex
	hits     []time.Time
	window   time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
	current  float64
}

// NewRateCalculator creates a new rate calculator with specified window
func NewRateCalculator(window time.Duration) *RateCalculator {
	rc := &RateCalculator{
		hits:   make([]time.Time, 0, 1000),
		window: window,
		stopCh: make(chan struct{}),
	}
	go rc.cleanupLoop()
	return rc
}

// Record records a hit
func (rc *RateCalculator) Record() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.hits = append(rc.hits, time.Now())
}

// GetRate returns current hits per minute
func (rc *RateCalculator) GetRate() float64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cleanup(time.Now())
	return float64(len(rc.hits)) * (60.0 / rc.window.Seconds())
}

// cleanup removes old hits outside the window
func (rc *RateCalculator) cleanup(now time.Time) {
	cutoff := now.Add(-rc.window)
	idx := 0
	for i, t := range rc.hits {
		if t.After(cutoff) {
			idx = i
			break
		}
	}
	rc.hits = rc.hits[idx:]
}

// cleanupLoop periodically cleans up old hits
func (rc *RateCalculator) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rc.mu.Lock()
			rc.cleanup(time.Now())
			rc.current = float64(len(rc.hits)) * (60.0 / rc.window.Seconds())
			rc.mu.Unlock()
		case <-rc.stopCh:
			return
		}
	}
}

// Stop stops the rate calculator
func (rc *RateCalculator) Stop() {
	close(rc.stopCh)
}

// Namespace for all metrics
const namespace = "eroshit"

// NewMetricsCollector creates and initializes a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		startTime:  time.Now(),
		hitsPerMin: NewRateCalculator(time.Minute),
	}

	// Hit Counter
	mc.HitCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "hits_total",
		Help:      "Total number of hits",
	})

	// Hit Rate (hits per minute)
	mc.HitRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "hit_rate_per_minute",
		Help:      "Current hit rate per minute",
	})

	// Response Time Histogram
	mc.ResponseTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "response_time_seconds",
		Help:      "Response time distribution",
		Buckets:   prometheus.DefBuckets,
	})

	// Proxy Latency Histogram (per proxy)
	mc.ProxyLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "proxy_latency_seconds",
		Help:      "Proxy latency distribution by proxy",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"proxy"})

	// Active Sessions Gauge
	mc.ActiveSessions = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_sessions",
		Help:      "Number of active sessions",
	})

	// Active Proxies Gauge
	mc.ActiveProxies = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_proxies",
		Help:      "Number of active proxies",
	})

	// Queue Size Gauge
	mc.QueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "queue_size",
		Help:      "Current queue size",
	})

	// Success Rate Gauge
	mc.SuccessRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "success_rate",
		Help:      "Success rate (0-1)",
	})

	// Bounce Rate Gauge
	mc.BounceRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "bounce_rate",
		Help:      "Bounce rate (0-1)",
	})

	// Error Rate Gauge
	mc.ErrorRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "error_rate",
		Help:      "Error rate (0-1)",
	})

	// Proxy Success Counter (per proxy)
	mc.ProxySuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "proxy_success_total",
		Help:      "Total successful requests per proxy",
	}, []string{"proxy"})

	// Proxy Failure Counter (per proxy)
	mc.ProxyFailure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "proxy_failure_total",
		Help:      "Total failed requests per proxy",
	}, []string{"proxy"})

	// Register all metrics
	mc.register()

	// Start background updater
	go mc.updateLoop()

	return mc
}

// register registers all metrics with Prometheus
func (mc *MetricsCollector) register() {
	prometheus.MustRegister(
		mc.HitCounter,
		mc.HitRate,
		mc.ResponseTime,
		mc.ProxyLatency,
		mc.ActiveSessions,
		mc.ActiveProxies,
		mc.QueueSize,
		mc.SuccessRate,
		mc.BounceRate,
		mc.ErrorRate,
		mc.ProxySuccess,
		mc.ProxyFailure,
	)
}

// updateLoop periodically updates calculated metrics
func (mc *MetricsCollector) updateLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		mc.updateCalculatedMetrics()
	}
}

// updateCalculatedMetrics updates derived metrics
func (mc *MetricsCollector) updateCalculatedMetrics() {
	mc.mu.RLock()
	total := mc.totalHits
	success := mc.successCount
	bounce := mc.bounceCount
	errors := mc.errorCount
	mc.mu.RUnlock()

	if total > 0 {
		mc.SuccessRate.Set(float64(success) / float64(total))
		mc.BounceRate.Set(float64(bounce) / float64(total))
		mc.ErrorRate.Set(float64(errors) / float64(total))
	}

	// Update hit rate from calculator
	mc.HitRate.Set(mc.hitsPerMin.GetRate())
}

// RecordHit records a hit
func (mc *MetricsCollector) RecordHit() {
	mc.HitCounter.Inc()
	mc.hitsPerMin.Record()
	mc.mu.Lock()
	mc.totalHits++
	mc.mu.Unlock()
}

// RecordResponseTime records response time
func (mc *MetricsCollector) RecordResponseTime(duration time.Duration) {
	mc.ResponseTime.Observe(duration.Seconds())
}

// RecordProxyLatency records proxy-specific latency
func (mc *MetricsCollector) RecordProxyLatency(proxy string, duration time.Duration) {
	mc.ProxyLatency.WithLabelValues(proxy).Observe(duration.Seconds())
}

// RecordSuccess records a successful hit
func (mc *MetricsCollector) RecordSuccess(proxy string) {
	mc.mu.Lock()
	mc.successCount++
	mc.mu.Unlock()
	if proxy != "" {
		mc.ProxySuccess.WithLabelValues(proxy).Inc()
	}
}

// RecordFailure records a failed hit
func (mc *MetricsCollector) RecordFailure(proxy string) {
	mc.mu.Lock()
	mc.errorCount++
	mc.mu.Unlock()
	if proxy != "" {
		mc.ProxyFailure.WithLabelValues(proxy).Inc()
	}
}

// RecordBounce records a bounce
func (mc *MetricsCollector) RecordBounce() {
	mc.mu.Lock()
	mc.bounceCount++
	mc.mu.Unlock()
}

// SetActiveSessions sets active sessions count
func (mc *MetricsCollector) SetActiveSessions(count int64) {
	mc.ActiveSessions.Set(float64(count))
	mc.mu.Lock()
	mc.sessionCount = count
	mc.mu.Unlock()
}

// SetActiveProxies sets active proxies count
func (mc *MetricsCollector) SetActiveProxies(count int64) {
	mc.ActiveProxies.Set(float64(count))
	mc.mu.Lock()
	mc.proxyCount = count
	mc.mu.Unlock()
}

// SetQueueSize sets queue size
func (mc *MetricsCollector) SetQueueSize(size int64) {
	mc.QueueSize.Set(float64(size))
	mc.mu.Lock()
	mc.queueCount = size
	mc.mu.Unlock()
}

// GetSnapshot returns current metrics snapshot
func (mc *MetricsCollector) GetSnapshot() Snapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return Snapshot{
		Timestamp:       time.Now(),
		TotalHits:       mc.totalHits,
		SuccessCount:    mc.successCount,
		ErrorCount:      mc.errorCount,
		BounceCount:     mc.bounceCount,
		ActiveSessions:  mc.sessionCount,
		ActiveProxies:   mc.proxyCount,
		QueueSize:       mc.queueCount,
		HitRatePerMin:   mc.hitsPerMin.GetRate(),
		SuccessRate:     calculateRate(mc.successCount, mc.totalHits),
		BounceRate:      calculateRate(mc.bounceCount, mc.totalHits),
		ErrorRate:       calculateRate(mc.errorCount, mc.totalHits),
		UptimeSeconds:   time.Since(mc.startTime).Seconds(),
	}
}

// Snapshot represents a point-in-time metrics snapshot
type Snapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	TotalHits      int64     `json:"total_hits"`
	SuccessCount   int64     `json:"success_count"`
	ErrorCount     int64     `json:"error_count"`
	BounceCount    int64     `json:"bounce_count"`
	ActiveSessions int64     `json:"active_sessions"`
	ActiveProxies  int64     `json:"active_proxies"`
	QueueSize      int64     `json:"queue_size"`
	HitRatePerMin  float64   `json:"hit_rate_per_min"`
	SuccessRate    float64   `json:"success_rate"`
	BounceRate     float64   `json:"bounce_rate"`
	ErrorRate      float64   `json:"error_rate"`
	UptimeSeconds  float64   `json:"uptime_seconds"`
}

func calculateRate(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

// MetricsHandler returns HTTP handler for Prometheus metrics
func (mc *MetricsCollector) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// JSONHandler returns metrics in JSON format
func (mc *MetricsCollector) JSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mc.GetSnapshot())
	}
}

// Close cleans up resources
func (mc *MetricsCollector) Close() {
	if mc.hitsPerMin != nil {
		mc.hitsPerMin.Stop()
	}
}

// Global instance for easy access
var globalCollector *MetricsCollector
var globalMu sync.Once

// GetGlobalCollector returns the global metrics collector instance
func GetGlobalCollector() *MetricsCollector {
	globalMu.Do(func() {
		globalCollector = NewMetricsCollector()
	})
	return globalCollector
}

// SetGlobalCollector sets the global metrics collector (for testing)
func SetGlobalCollector(mc *MetricsCollector) {
	globalCollector = mc
}
