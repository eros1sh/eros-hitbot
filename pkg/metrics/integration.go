// Package metrics provides integration utilities for connecting
// the metrics system with other components.
package metrics

import (
	"context"
	"time"
)

// SimulatorHooks provides hooks for simulator integration
type SimulatorHooks struct {
	collector *MetricsCollector
}

// NewSimulatorHooks creates new simulator hooks
func NewSimulatorHooks(collector *MetricsCollector) *SimulatorHooks {
	return &SimulatorHooks{collector: collector}
}

// OnHitStart records the start of a hit
func (h *SimulatorHooks) OnHitStart() {
	h.collector.RecordHit()
}

// OnHitComplete records a completed hit
func (h *SimulatorHooks) OnHitComplete(proxy string, duration time.Duration, success bool) {
	h.collector.RecordResponseTime(duration)
	if proxy != "" {
		h.collector.RecordProxyLatency(proxy, duration)
	}
	if success {
		h.collector.RecordSuccess(proxy)
	} else {
		h.collector.RecordFailure(proxy)
	}
}

// OnSessionStart records a new session
func (h *SimulatorHooks) OnSessionStart() {
	// Active sessions is tracked externally
}

// OnSessionEnd records session end
func (h *SimulatorHooks) OnSessionEnd() {
	// Active sessions is tracked externally
}

// OnBounce records a bounce
func (h *SimulatorHooks) OnBounce() {
	h.collector.RecordBounce()
}

// ProxyHooks provides hooks for proxy integration
type ProxyHooks struct {
	collector *MetricsCollector
}

// NewProxyHooks creates new proxy hooks
func NewProxyHooks(collector *MetricsCollector) *ProxyHooks {
	return &ProxyHooks{collector: collector}
}

// OnProxyAdd records proxy addition
func (h *ProxyHooks) OnProxyAdd(count int) {
	h.collector.SetActiveProxies(int64(count))
}

// OnProxyRemove records proxy removal
func (h *ProxyHooks) OnProxyRemove(count int) {
	h.collector.SetActiveProxies(int64(count))
}

// OnProxySuccess records proxy success
func (h *ProxyHooks) OnProxySuccess(proxy string) {
	h.collector.RecordSuccess(proxy)
}

// OnProxyFailure records proxy failure
func (h *ProxyHooks) OnProxyFailure(proxy string) {
	h.collector.RecordFailure(proxy)
}

// QueueHooks provides hooks for queue integration
type QueueHooks struct {
	collector *MetricsCollector
}

// NewQueueHooks creates new queue hooks
func NewQueueHooks(collector *MetricsCollector) *QueueHooks {
	return &QueueHooks{collector: collector}
}

// OnQueueSizeChange records queue size change
func (h *QueueHooks) OnQueueSizeChange(size int) {
	h.collector.SetQueueSize(int64(size))
}

// MetricsContext carries metrics through context
type ctxKey string

const metricsKey ctxKey = "metrics"

// WithContext adds metrics collector to context
func WithContext(ctx context.Context, collector *MetricsCollector) context.Context {
	return context.WithValue(ctx, metricsKey, collector)
}

// FromContext extracts metrics collector from context
func FromContext(ctx context.Context) *MetricsCollector {
	if v := ctx.Value(metricsKey); v != nil {
		if mc, ok := v.(*MetricsCollector); ok {
			return mc
		}
	}
	return nil
}

// RecordHitFromContext records a hit using collector from context
func RecordHitFromContext(ctx context.Context) {
	if mc := FromContext(ctx); mc != nil {
		mc.RecordHit()
	}
}

// Timer helps measure operation durations
type Timer struct {
	start     time.Time
	collector *MetricsCollector
	proxy     string
	histogram bool
}

// StartTimer starts a new timer
func (h *SimulatorHooks) StartTimer(proxy string) *Timer {
	return &Timer{
		start:     time.Now(),
		collector: h.collector,
		proxy:     proxy,
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop(success bool) time.Duration {
	duration := time.Since(t.start)
	t.collector.RecordResponseTime(duration)
	if t.proxy != "" {
		t.collector.RecordProxyLatency(t.proxy, duration)
	}
	if success {
		t.collector.RecordSuccess(t.proxy)
	} else {
		t.collector.RecordFailure(t.proxy)
	}
	return duration
}
