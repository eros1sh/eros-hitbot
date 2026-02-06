//go:build windows
// +build windows

package network

import (
	"syscall"
)

// setTCPOpts sets TCP options on Windows systems
func setTCPOpts(fd int) error {
	handle := syscall.Handle(fd)
	
	// TCP_NODELAY
	if err := syscall.SetsockoptInt(handle, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1); err != nil {
		return err
	}
	
	// SO_KEEPALIVE
	if err := syscall.SetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	
	return nil
}
