// Package clienthints provides Client Hints spoofing for modern browser fingerprinting
// Client Hints are HTTP headers that provide information about the client device and preferences
package clienthints

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// ClientHints represents all Sec-CH-UA-* headers
type ClientHints struct {
	// Core hints
	SecChUa                  string // Brand and version list
	SecChUaMobile            string // ?0 or ?1
	SecChUaPlatform          string // Platform name
	SecChUaPlatformVersion   string // Platform version
	SecChUaArch              string // CPU architecture
	SecChUaBitness           string // 32 or 64
	SecChUaModel             string // Device model (mobile only)
	SecChUaFullVersion       string // Full browser version
	SecChUaFullVersionList   string // Full version list
	SecChUaWoW64             string // Windows on Windows 64
	SecChUaFormFactor        string // Form factor (desktop, mobile, tablet, etc.)

	// Preference hints
	SecChPrefersColorScheme  string // dark or light
	SecChPrefersReducedMotion string // reduce or no-preference
	SecChPrefersReducedTransparency string

	// Device hints
	SecChDeviceMemory        string // Device memory in GB
	SecChDpr                 string // Device pixel ratio
	SecChViewportWidth       string // Viewport width
	SecChViewportHeight      string // Viewport height

	// Network hints
	SecChDownlink            string // Downlink speed
	SecChEct                 string // Effective connection type
	SecChRtt                 string // Round trip time
	SecChSaveData            string // Save data preference
}

// DeviceProfile represents a complete device profile for Client Hints
type DeviceProfile struct {
	Name        string
	Type        string // desktop, mobile, tablet
	Platform    string
	Hints       ClientHints
	UserAgent   string
}

// ChromeWindowsDesktop returns Chrome on Windows desktop profile
func ChromeWindowsDesktop(version int) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	return DeviceProfile{
		Name:     fmt.Sprintf("Chrome %d Windows Desktop", version),
		Type:     "desktop",
		Platform: "Windows",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`, version, version),
			SecChUaMobile:          "?0",
			SecChUaPlatform:        `"Windows"`,
			SecChUaPlatformVersion: `"15.0.0"`,
			SecChUaArch:            `"x86"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           `""`,
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Google Chrome";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Desktop"`,
			SecChPrefersColorScheme: "light",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "8",
			SecChDpr:               "1",
			SecChViewportWidth:     "1920",
			SecChViewportHeight:    "1080",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", versionStr),
	}
}

// ChromeMacDesktop returns Chrome on macOS desktop profile
func ChromeMacDesktop(version int) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	return DeviceProfile{
		Name:     fmt.Sprintf("Chrome %d macOS Desktop", version),
		Type:     "desktop",
		Platform: "macOS",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`, version, version),
			SecChUaMobile:          "?0",
			SecChUaPlatform:        `"macOS"`,
			SecChUaPlatformVersion: `"14.2.0"`,
			SecChUaArch:            `"arm"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           `""`,
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Google Chrome";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Desktop"`,
			SecChPrefersColorScheme: "dark",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "16",
			SecChDpr:               "2",
			SecChViewportWidth:     "1440",
			SecChViewportHeight:    "900",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", versionStr),
	}
}

// ChromeLinuxDesktop returns Chrome on Linux desktop profile
func ChromeLinuxDesktop(version int) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	return DeviceProfile{
		Name:     fmt.Sprintf("Chrome %d Linux Desktop", version),
		Type:     "desktop",
		Platform: "Linux",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`, version, version),
			SecChUaMobile:          "?0",
			SecChUaPlatform:        `"Linux"`,
			SecChUaPlatformVersion: `"6.5.0"`,
			SecChUaArch:            `"x86"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           `""`,
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Google Chrome";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Desktop"`,
			SecChPrefersColorScheme: "light",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "8",
			SecChDpr:               "1",
			SecChViewportWidth:     "1920",
			SecChViewportHeight:    "1080",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", versionStr),
	}
}

// ChromeAndroidMobile returns Chrome on Android mobile profile
func ChromeAndroidMobile(version int, deviceModel string) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	if deviceModel == "" {
		deviceModel = "Pixel 7"
	}
	return DeviceProfile{
		Name:     fmt.Sprintf("Chrome %d Android Mobile (%s)", version, deviceModel),
		Type:     "mobile",
		Platform: "Android",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`, version, version),
			SecChUaMobile:          "?1",
			SecChUaPlatform:        `"Android"`,
			SecChUaPlatformVersion: `"14.0.0"`,
			SecChUaArch:            `"arm"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           fmt.Sprintf(`"%s"`, deviceModel),
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Google Chrome";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Mobile"`,
			SecChPrefersColorScheme: "light",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "8",
			SecChDpr:               "2.75",
			SecChViewportWidth:     "412",
			SecChViewportHeight:    "915",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (Linux; Android 14; %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Mobile Safari/537.36", deviceModel, versionStr),
	}
}

// ChromeAndroidTablet returns Chrome on Android tablet profile
func ChromeAndroidTablet(version int, deviceModel string) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	if deviceModel == "" {
		deviceModel = "SM-X710"
	}
	return DeviceProfile{
		Name:     fmt.Sprintf("Chrome %d Android Tablet (%s)", version, deviceModel),
		Type:     "tablet",
		Platform: "Android",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Google Chrome";v="%d"`, version, version),
			SecChUaMobile:          "?0",
			SecChUaPlatform:        `"Android"`,
			SecChUaPlatformVersion: `"14.0.0"`,
			SecChUaArch:            `"arm"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           fmt.Sprintf(`"%s"`, deviceModel),
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Google Chrome";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Tablet"`,
			SecChPrefersColorScheme: "light",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "8",
			SecChDpr:               "2",
			SecChViewportWidth:     "800",
			SecChViewportHeight:    "1280",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (Linux; Android 14; %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", deviceModel, versionStr),
	}
}

// EdgeWindowsDesktop returns Edge on Windows desktop profile
func EdgeWindowsDesktop(version int) DeviceProfile {
	versionStr := fmt.Sprintf("%d", version)
	return DeviceProfile{
		Name:     fmt.Sprintf("Edge %d Windows Desktop", version),
		Type:     "desktop",
		Platform: "Windows",
		Hints: ClientHints{
			SecChUa:                fmt.Sprintf(`"Not_A Brand";v="8", "Chromium";v="%d", "Microsoft Edge";v="%d"`, version, version),
			SecChUaMobile:          "?0",
			SecChUaPlatform:        `"Windows"`,
			SecChUaPlatformVersion: `"15.0.0"`,
			SecChUaArch:            `"x86"`,
			SecChUaBitness:         `"64"`,
			SecChUaModel:           `""`,
			SecChUaFullVersion:     fmt.Sprintf(`"%d.0.0.0"`, version),
			SecChUaFullVersionList: fmt.Sprintf(`"Not_A Brand";v="8.0.0.0", "Chromium";v="%d.0.0.0", "Microsoft Edge";v="%d.0.0.0"`, version, version),
			SecChUaWoW64:           "?0",
			SecChUaFormFactor:      `"Desktop"`,
			SecChPrefersColorScheme: "light",
			SecChPrefersReducedMotion: "no-preference",
			SecChDeviceMemory:      "8",
			SecChDpr:               "1",
			SecChViewportWidth:     "1920",
			SecChViewportHeight:    "1080",
			SecChDownlink:          "10",
			SecChEct:               "4g",
			SecChRtt:               "50",
			SecChSaveData:          "?0",
		},
		UserAgent: fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 Edg/%s.0.0.0", versionStr, versionStr),
	}
}

// ClientHintsManager manages Client Hints generation and spoofing
type ClientHintsManager struct {
	currentProfile DeviceProfile
	randomize      bool
}

// NewClientHintsManager creates a new Client Hints manager
func NewClientHintsManager(randomize bool) *ClientHintsManager {
	return &ClientHintsManager{
		currentProfile: RandomDesktopProfile(),
		randomize:      randomize,
	}
}

// GetProfile returns the current device profile
func (m *ClientHintsManager) GetProfile() DeviceProfile {
	if m.randomize {
		m.currentProfile = RandomProfile()
	}
	return m.currentProfile
}

// SetProfile sets a specific device profile
func (m *ClientHintsManager) SetProfile(profile DeviceProfile) {
	m.currentProfile = profile
}

// GetHeaders returns all Client Hints as HTTP headers
func (m *ClientHintsManager) GetHeaders() map[string]string {
	hints := m.currentProfile.Hints
	headers := make(map[string]string)

	if hints.SecChUa != "" {
		headers["Sec-CH-UA"] = hints.SecChUa
	}
	if hints.SecChUaMobile != "" {
		headers["Sec-CH-UA-Mobile"] = hints.SecChUaMobile
	}
	if hints.SecChUaPlatform != "" {
		headers["Sec-CH-UA-Platform"] = hints.SecChUaPlatform
	}
	if hints.SecChUaPlatformVersion != "" {
		headers["Sec-CH-UA-Platform-Version"] = hints.SecChUaPlatformVersion
	}
	if hints.SecChUaArch != "" {
		headers["Sec-CH-UA-Arch"] = hints.SecChUaArch
	}
	if hints.SecChUaBitness != "" {
		headers["Sec-CH-UA-Bitness"] = hints.SecChUaBitness
	}
	if hints.SecChUaModel != "" {
		headers["Sec-CH-UA-Model"] = hints.SecChUaModel
	}
	if hints.SecChUaFullVersion != "" {
		headers["Sec-CH-UA-Full-Version"] = hints.SecChUaFullVersion
	}
	if hints.SecChUaFullVersionList != "" {
		headers["Sec-CH-UA-Full-Version-List"] = hints.SecChUaFullVersionList
	}
	if hints.SecChUaFormFactor != "" {
		headers["Sec-CH-UA-Form-Factor"] = hints.SecChUaFormFactor
	}

	return headers
}

// ToChromedpScript generates JavaScript to inject Client Hints spoofing
func (m *ClientHintsManager) ToChromedpScript() string {
	hints := m.currentProfile.Hints
	profile := m.currentProfile

	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// Spoof navigator.userAgentData (Client Hints API)
	const brands = [
		{ brand: "Not_A Brand", version: "8" },
		{ brand: "Chromium", version: "120" },
		{ brand: "Google Chrome", version: "120" }
	];
	
	const fullVersionList = [
		{ brand: "Not_A Brand", version: "8.0.0.0" },
		{ brand: "Chromium", version: "120.0.0.0" },
		{ brand: "Google Chrome", version: "120.0.0.0" }
	];
	
	const userAgentData = {
		brands: brands,
		mobile: %s === "?1",
		platform: %s,
		
		getHighEntropyValues: function(hints) {
			return Promise.resolve({
				brands: brands,
				fullVersionList: fullVersionList,
				mobile: this.mobile,
				platform: this.platform,
				platformVersion: %s,
				architecture: %s,
				bitness: %s,
				model: %s,
				uaFullVersion: %s,
				wow64: false,
				formFactor: %s
			});
		},
		
		toJSON: function() {
			return {
				brands: this.brands,
				mobile: this.mobile,
				platform: this.platform
			};
		}
	};
	
	// Override navigator.userAgentData
	Object.defineProperty(navigator, 'userAgentData', {
		get: function() { return userAgentData; },
		configurable: true
	});
	
	// Override navigator.userAgent
	Object.defineProperty(navigator, 'userAgent', {
		get: function() { return '%s'; },
		configurable: true
	});
	
	// Override navigator.platform
	const platformMap = {
		'Windows': 'Win32',
		'macOS': 'MacIntel',
		'Linux': 'Linux x86_64',
		'Android': 'Linux armv8l'
	};
	Object.defineProperty(navigator, 'platform', {
		get: function() { return platformMap['%s'] || 'Win32'; },
		configurable: true
	});
	
	// Spoof device memory
	Object.defineProperty(navigator, 'deviceMemory', {
		get: function() { return %s; },
		configurable: true
	});
	
	// Spoof hardware concurrency based on device type
	const concurrencyMap = {
		'desktop': 8,
		'mobile': 8,
		'tablet': 8
	};
	Object.defineProperty(navigator, 'hardwareConcurrency', {
		get: function() { return concurrencyMap['%s'] || 8; },
		configurable: true
	});
	
	// Spoof connection info
	if (navigator.connection) {
		Object.defineProperty(navigator.connection, 'effectiveType', {
			get: function() { return '%s'; },
			configurable: true
		});
		Object.defineProperty(navigator.connection, 'downlink', {
			get: function() { return %s; },
			configurable: true
		});
		Object.defineProperty(navigator.connection, 'rtt', {
			get: function() { return %s; },
			configurable: true
		});
		Object.defineProperty(navigator.connection, 'saveData', {
			get: function() { return %s === "?1"; },
			configurable: true
		});
	}
	
	// Spoof screen properties based on device
	const screenProps = {
		'desktop': { width: 1920, height: 1080, availWidth: 1920, availHeight: 1040, colorDepth: 24, pixelDepth: 24 },
		'mobile': { width: 412, height: 915, availWidth: 412, availHeight: 915, colorDepth: 24, pixelDepth: 24 },
		'tablet': { width: 800, height: 1280, availWidth: 800, availHeight: 1280, colorDepth: 24, pixelDepth: 24 }
	};
	const sp = screenProps['%s'] || screenProps['desktop'];
	
	Object.defineProperty(screen, 'width', { get: () => sp.width, configurable: true });
	Object.defineProperty(screen, 'height', { get: () => sp.height, configurable: true });
	Object.defineProperty(screen, 'availWidth', { get: () => sp.availWidth, configurable: true });
	Object.defineProperty(screen, 'availHeight', { get: () => sp.availHeight, configurable: true });
	Object.defineProperty(screen, 'colorDepth', { get: () => sp.colorDepth, configurable: true });
	Object.defineProperty(screen, 'pixelDepth', { get: () => sp.pixelDepth, configurable: true });
	
	// Spoof window dimensions
	Object.defineProperty(window, 'innerWidth', { get: () => sp.width, configurable: true });
	Object.defineProperty(window, 'innerHeight', { get: () => sp.height - 100, configurable: true });
	Object.defineProperty(window, 'outerWidth', { get: () => sp.width, configurable: true });
	Object.defineProperty(window, 'outerHeight', { get: () => sp.height, configurable: true });
	
	// Spoof devicePixelRatio
	const dprMap = { 'desktop': 1, 'mobile': 2.75, 'tablet': 2 };
	Object.defineProperty(window, 'devicePixelRatio', {
		get: function() { return dprMap['%s'] || 1; },
		configurable: true
	});
	
	// Spoof matchMedia for prefers-color-scheme
	const originalMatchMedia = window.matchMedia;
	window.matchMedia = function(query) {
		if (query === '(prefers-color-scheme: dark)') {
			return {
				matches: '%s' === 'dark',
				media: query,
				onchange: null,
				addListener: function() {},
				removeListener: function() {},
				addEventListener: function() {},
				removeEventListener: function() {},
				dispatchEvent: function() { return true; }
			};
		}
		if (query === '(prefers-reduced-motion: reduce)') {
			return {
				matches: '%s' === 'reduce',
				media: query,
				onchange: null,
				addListener: function() {},
				removeListener: function() {},
				addEventListener: function() {},
				removeEventListener: function() {},
				dispatchEvent: function() { return true; }
			};
		}
		return originalMatchMedia.call(this, query);
	};
	
	console.log('[ClientHints] Spoofing applied for: %s');
})();
`,
		hints.SecChUaMobile,
		hints.SecChUaPlatform,
		hints.SecChUaPlatformVersion,
		hints.SecChUaArch,
		hints.SecChUaBitness,
		hints.SecChUaModel,
		hints.SecChUaFullVersion,
		hints.SecChUaFormFactor,
		strings.ReplaceAll(profile.UserAgent, "'", "\\'"),
		profile.Platform,
		hints.SecChDeviceMemory,
		profile.Type,
		hints.SecChEct,
		hints.SecChDownlink,
		hints.SecChRtt,
		hints.SecChSaveData,
		profile.Type,
		profile.Type,
		hints.SecChPrefersColorScheme,
		hints.SecChPrefersReducedMotion,
		profile.Name,
	)
}

// RandomProfile returns a random device profile
func RandomProfile() DeviceProfile {
	profiles := AllProfiles()
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(profiles))))
	return profiles[n.Int64()]
}

// RandomDesktopProfile returns a random desktop profile
func RandomDesktopProfile() DeviceProfile {
	profiles := []DeviceProfile{
		ChromeWindowsDesktop(120),
		ChromeWindowsDesktop(121),
		ChromeMacDesktop(120),
		ChromeMacDesktop(121),
		ChromeLinuxDesktop(120),
		EdgeWindowsDesktop(120),
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(profiles))))
	return profiles[n.Int64()]
}

// RandomMobileProfile returns a random mobile profile
func RandomMobileProfile() DeviceProfile {
	models := []string{"Pixel 7", "Pixel 8", "SM-S918B", "SM-A546B", "22101316G"}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(models))))
	return ChromeAndroidMobile(120, models[n.Int64()])
}

// RandomTabletProfile returns a random tablet profile
func RandomTabletProfile() DeviceProfile {
	models := []string{"SM-X710", "SM-X810", "SM-T870"}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(models))))
	return ChromeAndroidTablet(120, models[n.Int64()])
}

// AllProfiles returns all available device profiles
func AllProfiles() []DeviceProfile {
	return []DeviceProfile{
		ChromeWindowsDesktop(120),
		ChromeWindowsDesktop(121),
		ChromeMacDesktop(120),
		ChromeMacDesktop(121),
		ChromeLinuxDesktop(120),
		ChromeAndroidMobile(120, "Pixel 7"),
		ChromeAndroidMobile(120, "SM-S918B"),
		ChromeAndroidTablet(120, "SM-X710"),
		EdgeWindowsDesktop(120),
	}
}

// GetProfileByType returns a profile matching the specified type
func GetProfileByType(deviceType string) DeviceProfile {
	switch strings.ToLower(deviceType) {
	case "mobile":
		return RandomMobileProfile()
	case "tablet":
		return RandomTabletProfile()
	case "desktop":
		return RandomDesktopProfile()
	default:
		return RandomProfile()
	}
}
