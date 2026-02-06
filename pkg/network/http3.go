//go:build http3
// +build http3

// HTTP/3 QUIC Support for ErosHit
// Requires: go get github.com/quic-go/quic-go
package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// HTTP3Config HTTP/3 configuration
type HTTP3Config struct {
	Enabled            bool
	MaxHeaderBytes     int
	MaxIncomingStreams int64
	Allow0RTT          bool // 0-RTT support for faster reconnections
}

// DefaultHTTP3Config returns default HTTP/3 configuration
func DefaultHTTP3Config() HTTP3Config {
	return HTTP3Config{
		Enabled:            false,
		MaxHeaderBytes:     1 << 20, // 1MB
		MaxIncomingStreams: 100,
		Allow0RTT:          true,
	}
}

// HTTP3Transport wraps http3.RoundTripper
type HTTP3Transport struct {
	config    HTTP3Config
	transport *http3.RoundTripper
	mu        sync.RWMutex
}

// NewHTTP3Transport creates HTTP/3 transport
func NewHTTP3Transport(tlsConfig *tls.Config) (*HTTP3Transport, error) {
	// Clone TLS config for HTTP/3
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
		}
	}
	
	transport := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig: &quic.Config{
			MaxIncomingStreams: 100,
			Allow0RTT:          true,
		},
	}
	
	return &HTTP3Transport{
		config:    DefaultHTTP3Config(),
		transport: transport,
	}, nil
}

// RoundTrip implements http.RoundTripper
func (h *HTTP3Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if !h.config.Enabled {
		return nil, fmt.Errorf("HTTP/3 is disabled")
	}
	
	return h.transport.RoundTrip(req)
}

// Close closes the transport
func (h *HTTP3Transport) Close() error {
	return h.transport.Close()
}

// Enable enables HTTP/3
func (h *HTTP3Transport) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.Enabled = true
}

// Disable disables HTTP/3
func (h *HTTP3Transport) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.Enabled = false
}

// IsEnabled returns HTTP/3 status
func (h *HTTP3Transport) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config.Enabled
}

// HTTP3FallbackTransport tries HTTP/3 first, falls back to HTTP/2/1
type HTTP3FallbackTransport struct {
	http3Transport *HTTP3Transport
	httpTransport  http.RoundTripper
	mu             sync.RWMutex
}

// NewHTTP3FallbackTransport creates fallback transport
func NewHTTP3FallbackTransport(http3Transport *HTTP3Transport, httpTransport http.RoundTripper) *HTTP3FallbackTransport {
	return &HTTP3FallbackTransport{
		http3Transport: http3Transport,
		httpTransport:  httpTransport,
	}
}

// RoundTrip implements http.RoundTripper with fallback
func (f *HTTP3FallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.mu.RLock()
	http3Enabled := f.http3Transport.IsEnabled()
	f.mu.RUnlock()
	
	if !http3Enabled {
		return f.httpTransport.RoundTrip(req)
	}
	
	// Try HTTP/3 first
	resp, err := f.http3Transport.RoundTrip(req)
	if err != nil {
		// Fallback to HTTP/2 or HTTP/1.1
		return f.httpTransport.RoundTrip(req)
	}
	
	return resp, nil
}

// EnableHTTP3 enables HTTP/3
func (f *HTTP3FallbackTransport) EnableHTTP3() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.http3Transport.Enable()
}

// DisableHTTP3 disables HTTP/3
func (f *HTTP3FallbackTransport) DisableHTTP3() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.http3Transport.Disable()
}

// IsHTTP3Enabled returns HTTP/3 status
func (f *HTTP3FallbackTransport) IsHTTP3Enabled() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.http3Transport.IsEnabled()
}
