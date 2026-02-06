// Package http2fingerprint provides HTTP/2 and HTTP/3 (QUIC) fingerprint spoofing
// to evade advanced bot detection systems that analyze TLS and HTTP/2 fingerprints.
package http2fingerprint

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// HTTP2Fingerprint represents an HTTP/2 fingerprint configuration
type HTTP2Fingerprint struct {
	// SETTINGS frame parameters
	HeaderTableSize      uint32
	EnablePush           bool
	MaxConcurrentStreams uint32
	InitialWindowSize    uint32
	MaxFrameSize         uint32
	MaxHeaderListSize    uint32

	// WINDOW_UPDATE frame
	WindowUpdateIncrement uint32

	// Priority frames (PRIORITY or HEADERS with priority)
	PriorityFrames []PriorityFrame

	// Pseudo-header order
	PseudoHeaderOrder []string

	// AKAMAI fingerprint string
	AkamaiFingerprint string
}

// PriorityFrame represents HTTP/2 priority information
type PriorityFrame struct {
	StreamID   uint32
	Exclusive  bool
	DependsOn  uint32
	// BUG FIX: Weight changed from uint8 to uint16
	// HTTP/2 spec allows weight values 1-256, but uint8 max is 255
	// This caused compile error: "256 overflows uint8"
	Weight     uint16
}

// HTTP3Fingerprint represents an HTTP/3 (QUIC) fingerprint configuration
type HTTP3Fingerprint struct {
	// QUIC transport parameters
	MaxIdleTimeout                 uint64
	MaxUDPPayloadSize              uint64
	InitialMaxData                 uint64
	InitialMaxStreamDataBidiLocal  uint64
	InitialMaxStreamDataBidiRemote uint64
	InitialMaxStreamDataUni        uint64
	InitialMaxStreamsBidi          uint64
	InitialMaxStreamsUni           uint64
	ActiveConnectionIDLimit        uint64

	// QPACK settings
	QPACKMaxTableCapacity uint64
	QPACKBlockedStreams   uint64
}

// BrowserProfile represents a complete browser fingerprint profile
type BrowserProfile struct {
	Name            string
	HTTP2           HTTP2Fingerprint
	HTTP3           HTTP3Fingerprint
	JA3             string
	JA4             string
	UserAgent       string
	SecChUa         string
	SecChUaPlatform string
	SecChUaMobile   string
}

// Chrome120Profile returns Chrome 120 fingerprint profile
func Chrome120Profile() BrowserProfile {
	return BrowserProfile{
		Name: "Chrome 120",
		HTTP2: HTTP2Fingerprint{
			HeaderTableSize:       65536,
			EnablePush:            false,
			MaxConcurrentStreams:  1000,
			InitialWindowSize:     6291456,
			MaxFrameSize:          16384,
			MaxHeaderListSize:     262144,
			WindowUpdateIncrement: 15663105,
			PriorityFrames: []PriorityFrame{
				{StreamID: 1, Exclusive: true, DependsOn: 0, Weight: 256},
				{StreamID: 3, Exclusive: false, DependsOn: 0, Weight: 201},
				{StreamID: 5, Exclusive: false, DependsOn: 0, Weight: 101},
				{StreamID: 7, Exclusive: false, DependsOn: 0, Weight: 1},
				{StreamID: 9, Exclusive: false, DependsOn: 7, Weight: 1},
				{StreamID: 11, Exclusive: false, DependsOn: 3, Weight: 1},
			},
			PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
			AkamaiFingerprint: "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
		},
		HTTP3: HTTP3Fingerprint{
			MaxIdleTimeout:                 30000,
			MaxUDPPayloadSize:              65527,
			InitialMaxData:                 15728640,
			InitialMaxStreamDataBidiLocal:  6291456,
			InitialMaxStreamDataBidiRemote: 6291456,
			InitialMaxStreamDataUni:        6291456,
			InitialMaxStreamsBidi:          100,
			InitialMaxStreamsUni:           100,
			ActiveConnectionIDLimit:        8,
			QPACKMaxTableCapacity:          16384,
			QPACKBlockedStreams:            100,
		},
		JA3:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
		JA4:             "t13d1516h2_8daaf6152771_b0da82dd1658",
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		SecChUa:         `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		SecChUaPlatform: `"Windows"`,
		SecChUaMobile:   "?0",
	}
}

// Chrome121Profile returns Chrome 121 fingerprint profile
func Chrome121Profile() BrowserProfile {
	return BrowserProfile{
		Name: "Chrome 121",
		HTTP2: HTTP2Fingerprint{
			HeaderTableSize:       65536,
			EnablePush:            false,
			MaxConcurrentStreams:  1000,
			InitialWindowSize:     6291456,
			MaxFrameSize:          16384,
			MaxHeaderListSize:     262144,
			WindowUpdateIncrement: 15663105,
			PriorityFrames: []PriorityFrame{
				{StreamID: 1, Exclusive: true, DependsOn: 0, Weight: 256},
				{StreamID: 3, Exclusive: false, DependsOn: 0, Weight: 201},
				{StreamID: 5, Exclusive: false, DependsOn: 0, Weight: 101},
				{StreamID: 7, Exclusive: false, DependsOn: 0, Weight: 1},
				{StreamID: 9, Exclusive: false, DependsOn: 7, Weight: 1},
				{StreamID: 11, Exclusive: false, DependsOn: 3, Weight: 1},
			},
			PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
			AkamaiFingerprint: "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
		},
		HTTP3: HTTP3Fingerprint{
			MaxIdleTimeout:                 30000,
			MaxUDPPayloadSize:              65527,
			InitialMaxData:                 15728640,
			InitialMaxStreamDataBidiLocal:  6291456,
			InitialMaxStreamDataBidiRemote: 6291456,
			InitialMaxStreamDataUni:        6291456,
			InitialMaxStreamsBidi:          100,
			InitialMaxStreamsUni:           100,
			ActiveConnectionIDLimit:        8,
			QPACKMaxTableCapacity:          16384,
			QPACKBlockedStreams:            100,
		},
		JA3:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
		JA4:             "t13d1516h2_8daaf6152771_b0da82dd1658",
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		SecChUa:         `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`,
		SecChUaPlatform: `"Windows"`,
		SecChUaMobile:   "?0",
	}
}

// Firefox121Profile returns Firefox 121 fingerprint profile
func Firefox121Profile() BrowserProfile {
	return BrowserProfile{
		Name: "Firefox 121",
		HTTP2: HTTP2Fingerprint{
			HeaderTableSize:       65536,
			EnablePush:            true,
			MaxConcurrentStreams:  0, // Firefox doesn't set this initially
			InitialWindowSize:     131072,
			MaxFrameSize:          16384,
			MaxHeaderListSize:     0, // Firefox doesn't set this
			WindowUpdateIncrement: 12517377,
			PriorityFrames:        []PriorityFrame{}, // Firefox uses different priority scheme
			PseudoHeaderOrder:     []string{":method", ":path", ":authority", ":scheme"},
			AkamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|0|m,p,a,s",
		},
		HTTP3: HTTP3Fingerprint{
			MaxIdleTimeout:                 30000,
			MaxUDPPayloadSize:              65527,
			InitialMaxData:                 10485760,
			InitialMaxStreamDataBidiLocal:  1048576,
			InitialMaxStreamDataBidiRemote: 1048576,
			InitialMaxStreamDataUni:        1048576,
			InitialMaxStreamsBidi:          100,
			InitialMaxStreamsUni:           100,
			ActiveConnectionIDLimit:        8,
			QPACKMaxTableCapacity:          0,
			QPACKBlockedStreams:            0,
		},
		JA3:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
		JA4:             "t13d1715h2_5b57614c22b0_3d5424432f57",
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		SecChUa:         "",
		SecChUaPlatform: "",
		SecChUaMobile:   "",
	}
}

// Safari17Profile returns Safari 17 fingerprint profile
func Safari17Profile() BrowserProfile {
	return BrowserProfile{
		Name: "Safari 17",
		HTTP2: HTTP2Fingerprint{
			HeaderTableSize:       4096,
			EnablePush:            false,
			MaxConcurrentStreams:  100,
			InitialWindowSize:     2097152,
			MaxFrameSize:          16384,
			MaxHeaderListSize:     0,
			WindowUpdateIncrement: 10485760,
			PriorityFrames:        []PriorityFrame{},
			PseudoHeaderOrder:     []string{":method", ":scheme", ":path", ":authority"},
			AkamaiFingerprint:     "1:4096,3:100,4:2097152|10485760|0|m,s,p,a",
		},
		HTTP3: HTTP3Fingerprint{
			MaxIdleTimeout:                 30000,
			MaxUDPPayloadSize:              65527,
			InitialMaxData:                 8388608,
			InitialMaxStreamDataBidiLocal:  1048576,
			InitialMaxStreamDataBidiRemote: 1048576,
			InitialMaxStreamDataUni:        1048576,
			InitialMaxStreamsBidi:          100,
			InitialMaxStreamsUni:           100,
			ActiveConnectionIDLimit:        4,
			QPACKMaxTableCapacity:          0,
			QPACKBlockedStreams:            0,
		},
		JA3:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
		JA4:             "t13d1715h2_5b57614c22b0_06cda9e17597",
		UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		SecChUa:         "",
		SecChUaPlatform: "",
		SecChUaMobile:   "",
	}
}

// Edge120Profile returns Edge 120 fingerprint profile
func Edge120Profile() BrowserProfile {
	return BrowserProfile{
		Name: "Edge 120",
		HTTP2: HTTP2Fingerprint{
			HeaderTableSize:       65536,
			EnablePush:            false,
			MaxConcurrentStreams:  1000,
			InitialWindowSize:     6291456,
			MaxFrameSize:          16384,
			MaxHeaderListSize:     262144,
			WindowUpdateIncrement: 15663105,
			PriorityFrames: []PriorityFrame{
				{StreamID: 1, Exclusive: true, DependsOn: 0, Weight: 256},
				{StreamID: 3, Exclusive: false, DependsOn: 0, Weight: 201},
				{StreamID: 5, Exclusive: false, DependsOn: 0, Weight: 101},
				{StreamID: 7, Exclusive: false, DependsOn: 0, Weight: 1},
				{StreamID: 9, Exclusive: false, DependsOn: 7, Weight: 1},
				{StreamID: 11, Exclusive: false, DependsOn: 3, Weight: 1},
			},
			PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
			AkamaiFingerprint: "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
		},
		HTTP3: HTTP3Fingerprint{
			MaxIdleTimeout:                 30000,
			MaxUDPPayloadSize:              65527,
			InitialMaxData:                 15728640,
			InitialMaxStreamDataBidiLocal:  6291456,
			InitialMaxStreamDataBidiRemote: 6291456,
			InitialMaxStreamDataUni:        6291456,
			InitialMaxStreamsBidi:          100,
			InitialMaxStreamsUni:           100,
			ActiveConnectionIDLimit:        8,
			QPACKMaxTableCapacity:          16384,
			QPACKBlockedStreams:            100,
		},
		JA3:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
		JA4:             "t13d1516h2_8daaf6152771_b0da82dd1658",
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		SecChUa:         `"Not_A Brand";v="8", "Chromium";v="120", "Microsoft Edge";v="120"`,
		SecChUaPlatform: `"Windows"`,
		SecChUaMobile:   "?0",
	}
}

// AllProfiles returns all available browser profiles
func AllProfiles() []BrowserProfile {
	return []BrowserProfile{
		Chrome120Profile(),
		Chrome121Profile(),
		Firefox121Profile(),
		Safari17Profile(),
		Edge120Profile(),
	}
}

// RandomProfile returns a random browser profile
func RandomProfile() BrowserProfile {
	profiles := AllProfiles()
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(profiles))))
	return profiles[n.Int64()]
}

// RandomProfileByType returns a random profile of specified browser type
func RandomProfileByType(browserType string) BrowserProfile {
	browserType = strings.ToLower(browserType)
	switch browserType {
	case "chrome":
		if randomBool() {
			return Chrome120Profile()
		}
		return Chrome121Profile()
	case "firefox":
		return Firefox121Profile()
	case "safari":
		return Safari17Profile()
	case "edge":
		return Edge120Profile()
	default:
		return RandomProfile()
	}
}

// FingerprintManager manages HTTP/2 and HTTP/3 fingerprints
type FingerprintManager struct {
	currentProfile BrowserProfile
	randomize      bool
}

// NewFingerprintManager creates a new fingerprint manager
func NewFingerprintManager(randomize bool) *FingerprintManager {
	return &FingerprintManager{
		currentProfile: RandomProfile(),
		randomize:      randomize,
	}
}

// GetProfile returns the current browser profile
func (fm *FingerprintManager) GetProfile() BrowserProfile {
	if fm.randomize {
		fm.currentProfile = RandomProfile()
	}
	return fm.currentProfile
}

// SetProfile sets a specific browser profile
func (fm *FingerprintManager) SetProfile(profile BrowserProfile) {
	fm.currentProfile = profile
}

// GetHTTP2Settings returns HTTP/2 SETTINGS frame parameters as a map
func (fm *FingerprintManager) GetHTTP2Settings() map[uint16]uint32 {
	fp := fm.currentProfile.HTTP2
	settings := make(map[uint16]uint32)

	settings[1] = fp.HeaderTableSize
	if fp.EnablePush {
		settings[2] = 1
	} else {
		settings[2] = 0
	}
	if fp.MaxConcurrentStreams > 0 {
		settings[3] = fp.MaxConcurrentStreams
	}
	settings[4] = fp.InitialWindowSize
	settings[5] = fp.MaxFrameSize
	if fp.MaxHeaderListSize > 0 {
		settings[6] = fp.MaxHeaderListSize
	}

	return settings
}

// GetAkamaiFingerprint returns the AKAMAI fingerprint string
func (fm *FingerprintManager) GetAkamaiFingerprint() string {
	return fm.currentProfile.HTTP2.AkamaiFingerprint
}

// GetJA3 returns the JA3 fingerprint
func (fm *FingerprintManager) GetJA3() string {
	return fm.currentProfile.JA3
}

// GetJA4 returns the JA4 fingerprint
func (fm *FingerprintManager) GetJA4() string {
	return fm.currentProfile.JA4
}

// GenerateRandomJA3 generates a randomized JA3 fingerprint based on a template
func GenerateRandomJA3(template string) string {
	parts := strings.Split(template, ",")
	if len(parts) < 5 {
		return template
	}

	// Slightly randomize cipher order while keeping it valid
	ciphers := strings.Split(parts[1], "-")
	if len(ciphers) > 2 {
		// Swap two random ciphers occasionally
		if randomBool() && len(ciphers) > 3 {
			i := randomInt(0, len(ciphers)-1)
			j := randomInt(0, len(ciphers)-1)
			ciphers[i], ciphers[j] = ciphers[j], ciphers[i]
		}
		parts[1] = strings.Join(ciphers, "-")
	}

	return strings.Join(parts, ",")
}

// GenerateRandomJA4 generates a randomized JA4 fingerprint
func GenerateRandomJA4() string {
	// JA4 format: t13d1516h2_8daaf6152771_b0da82dd1658
	// Protocol_CipherHash_ExtensionHash

	protocols := []string{"t13", "t12", "q13"}
	protocol := protocols[randomInt(0, len(protocols)-1)]

	// Generate random hashes
	cipherHash := randomHex(12)
	extHash := randomHex(12)

	return fmt.Sprintf("%sd1516h2_%s_%s", protocol, cipherHash, extHash)
}

// ToChromedpScript generates JavaScript to inject fingerprint spoofing
func (fm *FingerprintManager) ToChromedpScript() string {
	profile := fm.currentProfile

	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// HTTP/2 fingerprint spoofing via Performance API manipulation
	const originalGetEntries = performance.getEntries;
	performance.getEntries = function() {
		const entries = originalGetEntries.apply(this, arguments);
		return entries.map(entry => {
			if (entry.nextHopProtocol) {
				Object.defineProperty(entry, 'nextHopProtocol', {
					value: 'h2',
					writable: false,
					configurable: true
				});
			}
			return entry;
		});
	};
	
	// Spoof connection info
	if (navigator.connection) {
		Object.defineProperty(navigator.connection, 'effectiveType', {
			get: function() { return '4g'; },
			configurable: true
		});
		Object.defineProperty(navigator.connection, 'downlink', {
			get: function() { return 10; },
			configurable: true
		});
		Object.defineProperty(navigator.connection, 'rtt', {
			get: function() { return 50; },
			configurable: true
		});
	}
	
	// Store profile info for debugging
	window.__http2Profile = {
		name: '%s',
		akamai: '%s',
		ja3: '%s',
		ja4: '%s'
	};
})();
`, profile.Name, profile.HTTP2.AkamaiFingerprint, profile.JA3, profile.JA4)
}

// Helper functions
func randomBool() bool {
	n, _ := rand.Int(rand.Reader, big.NewInt(2))
	return n.Int64() == 1
}

func randomInt(min, max int) int {
	if max <= min {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}

func randomHex(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
