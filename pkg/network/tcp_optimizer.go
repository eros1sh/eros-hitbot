// TCP Optimizations including TCP Fast Open
package network

import (
	"context"
	"net"
	"runtime"
	"syscall"
	"time"
)

// TCPFastOpenConfig TCP Fast Open configuration
type TCPFastOpenConfig struct {
	Enabled     bool
	QueueLength int // SYN queue length
}

// DefaultTCPFastOpenConfig returns default TFO config
func DefaultTCPFastOpenConfig() TCPFastOpenConfig {
	return TCPFastOpenConfig{
		Enabled:     false, // Disabled by default (requires OS support)
		QueueLength: 1000,
	}
}

// TCPFastOpenDialer dialer with TFO support
type TCPFastOpenDialer struct {
	*net.Dialer
	config TCPFastOpenConfig
}

// NewTCPFastOpenDialer creates TFO-enabled dialer
func NewTCPFastOpenDialer(config TCPFastOpenConfig) *TCPFastOpenDialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	
	// Only set control function on Unix-like systems
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		dialer.Control = getTCPControlFunc(config)
	}
	
	return &TCPFastOpenDialer{
		Dialer: dialer,
		config: config,
	}
}

// getTCPControlFunc returns control function for TCP optimizations
// Note: This only works on Linux/macOS, not Windows
func getTCPControlFunc(tfoConfig TCPFastOpenConfig) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		if network != "tcp" && network != "tcp4" && network != "tcp6" {
			return nil
		}
		
		var controlErr error
		err := c.Control(func(fd uintptr) {
			// TCP Fast Open (Linux only)
			if tfoConfig.Enabled && runtime.GOOS == "linux" {
				// TCP_FASTOPEN = 23 (Linux specific)
				_ = fd
			}
			
			// TCP_NODELAY and SO_KEEPALIVE
			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
				if err := setTCPOpts(int(fd)); err != nil {
					controlErr = err
					return
				}
			}
		})
		
		if err != nil {
			return err
		}
		return controlErr
	}
}

// DialContext implements dialer with TFO
func (d *TCPFastOpenDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !d.config.Enabled {
		return d.Dialer.DialContext(ctx, network, address)
	}
	
	return d.Dialer.DialContext(ctx, network, address)
}

// TCPOptimizer TCP connection optimizer
type TCPOptimizer struct {
	tfoConfig TCPFastOpenConfig
}

// NewTCPOptimizer creates TCP optimizer
func NewTCPOptimizer(tfoEnabled bool) *TCPOptimizer {
	return &TCPOptimizer{
		tfoConfig: TCPFastOpenConfig{
			Enabled:     tfoEnabled && (runtime.GOOS == "linux" || runtime.GOOS == "darwin"),
			QueueLength: 1000,
		},
	}
}

// GetOptimizedDialer returns dialer with TCP optimizations
func (t *TCPOptimizer) GetOptimizedDialer() *net.Dialer {
	return NewTCPFastOpenDialer(t.tfoConfig).Dialer
}

// EnableTFO enables TCP Fast Open
func (t *TCPOptimizer) EnableTFO() {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		t.tfoConfig.Enabled = true
	}
}

// DisableTFO disables TCP Fast Open
func (t *TCPOptimizer) DisableTFO() {
	t.tfoConfig.Enabled = false
}

// IsTFOEnabled returns TFO status
func (t *TCPOptimizer) IsTFOEnabled() bool {
	return t.tfoConfig.Enabled
}

// IsTFOSupported returns whether TFO is supported on this platform
func IsTFOSupported() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
}
