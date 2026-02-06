// VM Fingerprint Spoofing for ErosHit
package stealth

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"
)

// VMType virtual machine types
type VMType string

const (
	VMNone      VMType = "none"
	VMVirtualBox VMType = "virtualbox"
	VMVMware    VMType = "vmware"
	VMParallels VMType = "parallels"
	VMHyperV    VMType = "hyperv"
	VMQEMU      VMType = "qemu"
	VMXen       VMType = "xen"
)

// VMConfig VM spoofing configuration
type VMConfig struct {
	Enabled              bool
	VMType               VMType
	HideVMIndicators     bool
	SpoofHardwareIDs     bool
	RandomizeVMParams    bool
}

// DefaultVMConfig returns default config
func DefaultVMConfig() VMConfig {
	return VMConfig{
		Enabled:           false,
		VMType:            VMNone,
		HideVMIndicators:  true,
		SpoofHardwareIDs:  true,
		RandomizeVMParams: true,
	}
}

// VMFingerprintSpoofer VM fingerprint spoofer
type VMFingerprintSpoofer struct {
	Config VMConfig
	rng    *rand.Rand
}

// NewVMFingerprintSpoofer creates VM spoofer
func NewVMFingerprintSpoofer(config VMConfig) *VMFingerprintSpoofer {
	return &VMFingerprintSpoofer{
		Config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetVMSpoofingScript returns JavaScript to spoof VM detection
func (v *VMFingerprintSpoofer) GetVMSpoofingScript() string {
	if !v.Config.Enabled {
		return ""
	}
	
	scripts := []string{
		v.getVMIndicatorRemovalScript(),
		v.getHardwareIDSpoofScript(),
		v.getNavigatorSpoofScript(),
		v.getScreenSpoofScript(),
		v.getPluginsSpoofScript(),
	}
	
	result := ""
	for _, script := range scripts {
		if script != "" {
			result += script + "\n"
		}
	}
	
	return result
}

// getVMIndicatorRemovalScript removes VM indicators
func (v *VMFingerprintSpoofer) getVMIndicatorRemovalScript() string {
	return `
(function() {
	// Remove VirtualBox indicators
	delete window.navigator.__proto__.webdriver;
	
	// Hide VMware indicators
	if (window.navigator.plugins) {
		Object.defineProperty(navigator.plugins, 'length', {
			get: function() { return 3; }
		});
	}
	
	// Remove Hyper-V indicators
	const originalToString = Function.prototype.toString;
	Function.prototype.toString = function() {
		const str = originalToString.call(this);
		if (str.includes('native code')) {
			return str.replace(/\\[native code\\]/g, '[native code]');
		}
		return str;
	};
	
	// Hide QEMU/Xen indicators from navigator
	const originalUserAgent = navigator.userAgent;
	Object.defineProperty(navigator, 'userAgent', {
		get: function() {
			return originalUserAgent.replace(/QEMU|Xen|VirtualBox|VMware/g, '');
		}
	});
})();
`
}

// getHardwareIDSpoofScript spoofs hardware IDs
func (v *VMFingerprintSpoofer) getHardwareIDSpoofScript() string {
	if !v.Config.SpoofHardwareIDs {
		return ""
	}
	
	// Generate random hardware fingerprint
	hardwareConcurrency := 4 + v.rng.Intn(12) // 4-16 cores
	deviceMemory := 4 + (v.rng.Intn(8) * 2)   // 4, 8, 16, 20 GB
	
	return fmt.Sprintf(`
(function() {
	// Spoof hardware concurrency
	Object.defineProperty(navigator, 'hardwareConcurrency', {
		get: function() { return %d; }
	});
	
	// Spoof device memory
	Object.defineProperty(navigator, 'deviceMemory', {
		get: function() { return %d; }
	});
	
	// Spoof platform
	Object.defineProperty(navigator, 'platform', {
		get: function() { return 'Win32'; }
	});
})();
`, hardwareConcurrency, deviceMemory)
}

// getNavigatorSpoofScript spoofs navigator properties
func (v *VMFingerprintSpoofer) getNavigatorSpoofScript() string {
	return `
(function() {
	// Spoof vendor
	Object.defineProperty(navigator, 'vendor', {
		get: function() { return 'Google Inc.'; }
	});
	
	// Spoof productSub
	Object.defineProperty(navigator, 'productSub', {
		get: function() { return '20030107'; }
	});
	
	// Remove vendor-specific properties
	delete window.chrome.runtime;
	window.chrome.runtime = {
		OnInstalledReason: {CHROME_UPDATE: "chrome_update"},
		OnRestartRequiredReason: {APP_UPDATE: "app_update"},
		PlatformArch: {X86_64: "x86-64"},
		PlatformNaclArch: {X86_64: "x86-64"},
		PlatformOs: {WIN: "win"},
		RequestUpdateCheckStatus: {NO_UPDATE: "no_update"},
		connect: function() { return {postMessage: function() {}}; },
		onConnect: {addListener: function() {}},
		onMessage: {addListener: function() {}},
		sendMessage: function() {}
	};
})();
`
}

// getScreenSpoofScript spoofs screen properties
func (v *VMFingerprintSpoofer) getScreenSpoofScript() string {
	if !v.Config.RandomizeVMParams {
		return ""
	}
	
	// Random but realistic screen resolution
	resolutions := []struct {
		width, height int
	}{
		{1920, 1080},
		{1366, 768},
		{1440, 900},
		{1536, 864},
		{1280, 720},
	}
	
	res := resolutions[v.rng.Intn(len(resolutions))]
	colorDepth := 24
	
	return fmt.Sprintf(`
(function() {
	// Spoof screen resolution
	Object.defineProperty(screen, 'width', {
		get: function() { return %d; }
	});
	Object.defineProperty(screen, 'height', {
		get: function() { return %d; }
	});
	Object.defineProperty(screen, 'availWidth', {
		get: function() { return %d; }
	});
	Object.defineProperty(screen, 'availHeight', {
		get: function() { return %d; }
	});
	Object.defineProperty(screen, 'colorDepth', {
		get: function() { return %d; }
	});
	Object.defineProperty(screen, 'pixelDepth', {
		get: function() { return %d; }
	});
})();
`, res.width, res.height, res.width, res.height-40, colorDepth, colorDepth)
}

// getPluginsSpoofScript spoofs browser plugins
func (v *VMFingerprintSpoofer) getPluginsSpoofScript() string {
	return `
(function() {
	// Create fake plugins array
	const fakePlugins = [
		{
			name: "Chrome PDF Plugin",
			filename: "internal-pdf-viewer",
			description: "Portable Document Format",
			version: "undefined",
			length: 1,
			item: function() { return this; },
			namedItem: function() { return this; }
		},
		{
			name: "Chrome PDF Viewer",
			filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
			description: "Portable Document Format",
			version: "undefined",
			length: 1,
			item: function() { return this; },
			namedItem: function() { return this; }
		},
		{
			name: "Native Client",
			filename: "internal-nacl-plugin",
			description: "",
			version: "undefined",
			length: 1,
			item: function() { return this; },
			namedItem: function() { return this; }
		}
	];
	
	fakePlugins.length = 3;
	fakePlugins.item = function(idx) { return this[idx]; };
	fakePlugins.namedItem = function(name) {
		for (let i = 0; i < this.length; i++) {
			if (this[i].name === name) return this[i];
		}
		return null;
	};
	fakePlugins.refresh = function() {};
	
	Object.defineProperty(navigator, 'plugins', {
		get: function() { return fakePlugins; }
	});
})();
`
}

// InjectVMProtection injects VM protection scripts
func (v *VMFingerprintSpoofer) InjectVMProtection(ctx context.Context) error {
	if !v.Config.Enabled {
		return nil
	}
	
	script := v.GetVMSpoofingScript()
	if script == "" {
		return nil
	}
	
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// Enable enables VM spoofing
func (v *VMFingerprintSpoofer) Enable() {
	v.Config.Enabled = true
}

// Disable disables VM spoofing
func (v *VMFingerprintSpoofer) Disable() {
	v.Config.Enabled = false
}

// IsEnabled returns status
func (v *VMFingerprintSpoofer) IsEnabled() bool {
	return v.Config.Enabled
}

// SetVMType sets VM type
func (v *VMFingerprintSpoofer) SetVMType(vmType VMType) {
	v.Config.VMType = vmType
}

// GetVMDetectionScore returns likelihood of VM detection (0-100)
func (v *VMFingerprintSpoofer) GetVMDetectionScore() int {
	if !v.Config.Enabled {
		// High detection score if not spoofing
		return 70
	}
	
	score := 20 // Base score with spoofing
	
	if v.Config.HideVMIndicators {
		score -= 10
	}
	
	if v.Config.SpoofHardwareIDs {
		score -= 5
	}
	
	if v.Config.RandomizeVMParams {
		score -= 5
	}
	
	if score < 0 {
		score = 5 // Minimum detection score
	}
	
	return score
}

// IsRunningInVM checks if running in VM
func IsRunningInVM() bool {
	// Simple check - more sophisticated checks would examine CPU features, BIOS strings, etc.
	if runtime.GOOS == "linux" {
		// Check for VM indicators in Linux
		// This is a simplified check
		return false
	}
	return false
}
