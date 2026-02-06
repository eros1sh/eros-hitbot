//go:build linux || darwin
// +build linux darwin

package network

import (
	"syscall"
)

// setTCPOpts sets TCP options on Unix systems
func setTCPOpts(fd int) error {
	// TCP_NODELAY
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1); err != nil {
		return err
	}
	
	// SO_KEEPALIVE
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	
	return nil
}
