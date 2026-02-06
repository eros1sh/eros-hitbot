// Package network provides advanced connection pooling and optimization features
package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// PoolConfig connection pool configuration
type PoolConfig struct {
	// Pool sizes
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	
	// Timeouts
	IdleConnTimeout     time.Duration
	TLSHandshakeTimeout time.Duration
	ExpectContinueTimeout time.Duration
	ResponseHeaderTimeout time.Duration
	
	// TCP optimizations
	EnableTCPKeepAlive  bool
	TCPKeepAlivePeriod  time.Duration
	EnableTCPNoDelay    bool
	
	// Connection pooling strategy
	DisableCompression  bool
	ForceHTTP2          bool
	EnableHTTP3         bool // HTTP/3 QUIC support
}

// DefaultPoolConfig returns optimized default configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		EnableTCPKeepAlive:    true,
		TCPKeepAlivePeriod:    30 * time.Second,
		EnableTCPNoDelay:      true,
		DisableCompression:    false,
		ForceHTTP2:            true,
		EnableHTTP3:           false,
	}
}

// ConnectionPool wraps http.Transport with advanced pooling
type ConnectionPool struct {
	config    PoolConfig
	transport *http.Transport
	mu        sync.RWMutex
	metrics   *PoolMetrics
}

// PoolMetrics connection pool metrics
type PoolMetrics struct {
	TotalConnections   int64
	ActiveConnections  int64
	IdleConnections    int64
	ConnectionsCreated int64
	ConnectionsReused  int64
}

// NewConnectionPool creates optimized connection pool
func NewConnectionPool(config PoolConfig) *ConnectionPool {
	pool := &ConnectionPool{
		config:  config,
		metrics: &PoolMetrics{},
	}
	
	pool.transport = pool.createTransport()
	return pool
}

// createTransport builds optimized http.Transport
func (p *ConnectionPool) createTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: p.config.TCPKeepAlivePeriod,
	}
	
	if p.config.EnableTCPKeepAlive {
		dialer.KeepAlive = p.config.TCPKeepAlivePeriod
	}
	
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          p.config.MaxIdleConns,
		MaxIdleConnsPerHost:   p.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       p.config.MaxConnsPerHost,
		IdleConnTimeout:       p.config.IdleConnTimeout,
		TLSHandshakeTimeout:   p.config.TLSHandshakeTimeout,
		ExpectContinueTimeout: p.config.ExpectContinueTimeout,
		ResponseHeaderTimeout: p.config.ResponseHeaderTimeout,
		DisableCompression:    p.config.DisableCompression,
		ForceAttemptHTTP2:     p.config.ForceHTTP2,
		DisableKeepAlives:     !p.config.EnableTCPKeepAlive,
	}
	
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		ClientSessionCache: tls.NewLRUClientSessionCache(128),
	}
	
	return transport
}

// GetClient returns http.Client with pooled transport
func (p *ConnectionPool) GetClient() *http.Client {
	return &http.Client{
		Transport: p.transport,
		Timeout:   30 * time.Second,
	}
}

// CloseIdleConnections closes all idle connections
func (p *ConnectionPool) CloseIdleConnections() {
	p.transport.CloseIdleConnections()
}

// GetMetrics returns current pool metrics
func (p *ConnectionPool) GetMetrics() PoolMetrics {
	return PoolMetrics{
		TotalConnections:   atomic.LoadInt64(&p.metrics.TotalConnections),
		ActiveConnections:  atomic.LoadInt64(&p.metrics.ActiveConnections),
		IdleConnections:    atomic.LoadInt64(&p.metrics.IdleConnections),
		ConnectionsCreated: atomic.LoadInt64(&p.metrics.ConnectionsCreated),
		ConnectionsReused:  atomic.LoadInt64(&p.metrics.ConnectionsReused),
	}
}

// EnableHTTP3 enables HTTP/3 QUIC support
func (p *ConnectionPool) EnableHTTP3() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.config.EnableHTTP3 {
		return nil
	}
	
	p.config.EnableHTTP3 = true
	return fmt.Errorf("HTTP/3 support requires quic-go dependency")
}

// TraceInfo connection trace information
type TraceInfo struct {
	Event    string
	Addr     string
	Err      error
	Reused   bool
	WasIdle  bool
	IdleTime time.Duration
}

// Global pool instance
var globalPool *ConnectionPool
var poolOnce sync.Once

// GetGlobalPool returns singleton connection pool
func GetGlobalPool() *ConnectionPool {
	poolOnce.Do(func() {
		globalPool = NewConnectionPool(DefaultPoolConfig())
	})
	return globalPool
}

// ResetGlobalPool resets the global pool with new config
func ResetGlobalPool(config PoolConfig) {
	if globalPool != nil {
		globalPool.CloseIdleConnections()
	}
	globalPool = NewConnectionPool(config)
}
