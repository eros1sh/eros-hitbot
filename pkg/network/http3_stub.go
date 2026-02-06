//go:build !http3
// +build !http3

// HTTP/3 QUIC Stub - HTTP/3 disabled by default
package network

import (
	"fmt"
)

// HTTP3Config stub
type HTTP3Config struct {
	Enabled            bool
	MaxHeaderBytes     int
	MaxIncomingStreams int64
	Allow0RTT          bool
}

// DefaultHTTP3Config returns default config
func DefaultHTTP3Config() HTTP3Config {
	return HTTP3Config{
		Enabled:            false,
		MaxHeaderBytes:     1 << 20,
		MaxIncomingStreams: 100,
		Allow0RTT:          false,
	}
}

// HTTP3Transport stub
type HTTP3Transport struct {
	config HTTP3Config
}

// NewHTTP3Transport creates stub transport
func NewHTTP3Transport(tlsConfig interface{}) (*HTTP3Transport, error) {
	return &HTTP3Transport{
		config: DefaultHTTP3Config(),
	}, nil
}

// RoundTrip implements http.RoundTripper
func (h *HTTP3Transport) RoundTrip(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("HTTP/3 not enabled. Build with -tags http3 to enable")
}

// Close closes the transport
func (h *HTTP3Transport) Close() error {
	return nil
}

// Enable enables HTTP/3
func (h *HTTP3Transport) Enable() {
	h.config.Enabled = false
}

// Disable disables HTTP/3
func (h *HTTP3Transport) Disable() {
	h.config.Enabled = false
}

// IsEnabled returns HTTP/3 status
func (h *HTTP3Transport) IsEnabled() bool {
	return false
}

// HTTP3FallbackTransport stub
type HTTP3FallbackTransport struct {
	httpTransport interface{}
}

// NewHTTP3FallbackTransport creates stub
func NewHTTP3FallbackTransport(http3Transport *HTTP3Transport, httpTransport interface{}) *HTTP3FallbackTransport {
	return &HTTP3FallbackTransport{
		httpTransport: httpTransport,
	}
}

// EnableHTTP3 enables HTTP/3
func (f *HTTP3FallbackTransport) EnableHTTP3() {}

// DisableHTTP3 disables HTTP/3
func (f *HTTP3FallbackTransport) DisableHTTP3() {}

// IsHTTP3Enabled returns HTTP/3 status
func (f *HTTP3FallbackTransport) IsHTTP3Enabled() bool {
	return false
}
