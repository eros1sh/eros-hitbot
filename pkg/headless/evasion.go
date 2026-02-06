// Package headless provides comprehensive headless browser detection evasion
// including Puppeteer, Playwright, Selenium detection bypass and rendering-based detection
package headless

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// EvasionSuite provides comprehensive headless detection evasion
type EvasionSuite struct {
	EnableWebdriver          bool
	EnableChrome             bool
	EnablePlugins            bool
	EnableWebGL              bool
	EnableAudio              bool
	EnableCanvas             bool
	EnableTimezone           bool
	EnableLanguage           bool
	EnablePuppeteerBypass    bool
	EnablePlaywrightBypass   bool
	EnableSeleniumBypass     bool
	EnableCDPMasking         bool
	EnableRenderingDetection bool
}

// DefaultEvasionSuite returns a fully enabled evasion suite
func DefaultEvasionSuite() *EvasionSuite {
	return &EvasionSuite{
		EnableWebdriver:          true,
		EnableChrome:             true,
		EnablePlugins:            true,
		EnableWebGL:              true,
		EnableAudio:              true,
		EnableCanvas:             true,
		EnableTimezone:           true,
		EnableLanguage:           true,
		EnablePuppeteerBypass:    true,
		EnablePlaywrightBypass:   true,
		EnableSeleniumBypass:     true,
		EnableCDPMasking:         true,
		EnableRenderingDetection: true,
	}
}

// GenerateEvasionScript generates comprehensive JavaScript for headless detection evasion
func (e *EvasionSuite) GenerateEvasionScript() string {
	var scripts []string

	if e.EnableWebdriver {
		scripts = append(scripts, e.webdriverEvasion())
	}
	if e.EnableChrome {
		scripts = append(scripts, e.chromeEvasion())
	}
	if e.EnablePlugins {
		scripts = append(scripts, e.pluginsEvasion())
	}
	if e.EnablePuppeteerBypass {
		scripts = append(scripts, e.puppeteerBypass())
	}
	if e.EnablePlaywrightBypass {
		scripts = append(scripts, e.playwrightBypass())
	}
	if e.EnableSeleniumBypass {
		scripts = append(scripts, e.seleniumBypass())
	}
	if e.EnableCDPMasking {
		scripts = append(scripts, e.cdpMasking())
	}
	if e.EnableRenderingDetection {
		scripts = append(scripts, e.renderingDetectionBypass())
	}

	return fmt.Sprintf(`
(function() {
	'use strict';
	%s
	console.log('[EvasionSuite] All evasion techniques applied');
})();
`, strings.Join(scripts, "\n\n"))
}

func (e *EvasionSuite) webdriverEvasion() string {
	return `
	// Webdriver Detection Evasion
	Object.defineProperty(navigator, 'webdriver', {
		get: () => undefined,
		configurable: true
	});
	
	// Remove webdriver from prototype
	delete Navigator.prototype.webdriver;
	
	// Override Object.getOwnPropertyDescriptor for webdriver
	const originalGetOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
	Object.getOwnPropertyDescriptor = function(obj, prop) {
		if (prop === 'webdriver' && obj === navigator) {
			return undefined;
		}
		return originalGetOwnPropertyDescriptor.apply(this, arguments);
	};
	
	// Remove automation-related properties
	const automationProps = [
		'__webdriver_evaluate',
		'__selenium_evaluate',
		'__webdriver_script_function',
		'__webdriver_script_func',
		'__webdriver_script_fn',
		'__fxdriver_evaluate',
		'__driver_unwrapped',
		'__webdriver_unwrapped',
		'__driver_evaluate',
		'__selenium_unwrapped',
		'__fxdriver_unwrapped',
		'_Selenium_IDE_Recorder',
		'_selenium',
		'calledSelenium',
		'$cdc_asdjflasutopfhvcZLmcfl_',
		'$chrome_asyncScriptInfo',
		'__$webdriverAsyncExecutor',
		'webdriver',
		'__webdriverFunc',
		'domAutomation',
		'domAutomationController'
	];
	
	automationProps.forEach(prop => {
		try {
			if (window[prop]) delete window[prop];
			if (document[prop]) delete document[prop];
		} catch (e) {}
	});
`
}

func (e *EvasionSuite) chromeEvasion() string {
	return `
	// Chrome Runtime Evasion
	if (!window.chrome) {
		window.chrome = {};
	}
	
	window.chrome.runtime = {
		connect: function() { return { onMessage: { addListener: function() {} }, postMessage: function() {} }; },
		sendMessage: function() {},
		onMessage: { addListener: function() {} },
		onConnect: { addListener: function() {} },
		id: undefined
	};
	
	window.chrome.loadTimes = function() {
		return {
			commitLoadTime: Date.now() / 1000 - Math.random() * 2,
			connectionInfo: 'h2',
			finishDocumentLoadTime: Date.now() / 1000 - Math.random(),
			finishLoadTime: Date.now() / 1000 - Math.random() * 0.5,
			firstPaintAfterLoadTime: 0,
			firstPaintTime: Date.now() / 1000 - Math.random() * 1.5,
			navigationType: 'Other',
			npnNegotiatedProtocol: 'h2',
			requestTime: Date.now() / 1000 - Math.random() * 3,
			startLoadTime: Date.now() / 1000 - Math.random() * 2.5,
			wasAlternateProtocolAvailable: false,
			wasFetchedViaSpdy: true,
			wasNpnNegotiated: true
		};
	};
	
	window.chrome.csi = function() {
		return {
			onloadT: Date.now(),
			pageT: Math.random() * 1000 + 500,
			startE: Date.now() - Math.random() * 3000,
			tran: 15
		};
	};
	
	window.chrome.app = {
		isInstalled: false,
		InstallState: { DISABLED: 'disabled', INSTALLED: 'installed', NOT_INSTALLED: 'not_installed' },
		RunningState: { CANNOT_RUN: 'cannot_run', READY_TO_RUN: 'ready_to_run', RUNNING: 'running' },
		getDetails: function() { return null; },
		getIsInstalled: function() { return false; },
		runningState: function() { return 'cannot_run'; }
	};
`
}

func (e *EvasionSuite) pluginsEvasion() string {
	return `
	// Plugins Evasion - Create realistic plugin array
	const mockPlugins = [
		{
			name: 'Chrome PDF Plugin',
			description: 'Portable Document Format',
			filename: 'internal-pdf-viewer',
			length: 1,
			item: function(i) { return this[i]; },
			namedItem: function(name) { return this[name]; },
			0: { type: 'application/x-google-chrome-pdf', suffixes: 'pdf', description: 'Portable Document Format', enabledPlugin: null }
		},
		{
			name: 'Chrome PDF Viewer',
			description: '',
			filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai',
			length: 1,
			item: function(i) { return this[i]; },
			namedItem: function(name) { return this[name]; },
			0: { type: 'application/pdf', suffixes: 'pdf', description: '', enabledPlugin: null }
		},
		{
			name: 'Native Client',
			description: '',
			filename: 'internal-nacl-plugin',
			length: 2,
			item: function(i) { return this[i]; },
			namedItem: function(name) { return this[name]; },
			0: { type: 'application/x-nacl', suffixes: '', description: 'Native Client Executable', enabledPlugin: null },
			1: { type: 'application/x-pnacl', suffixes: '', description: 'Portable Native Client Executable', enabledPlugin: null }
		}
	];
	
	const pluginArray = {
		length: mockPlugins.length,
		item: function(i) { return mockPlugins[i]; },
		namedItem: function(name) { return mockPlugins.find(p => p.name === name); },
		refresh: function() {}
	};
	
	mockPlugins.forEach((plugin, i) => {
		pluginArray[i] = plugin;
		pluginArray[plugin.name] = plugin;
	});
	
	Object.defineProperty(navigator, 'plugins', {
		get: () => pluginArray,
		configurable: true
	});
	
	// MimeTypes
	const mimeTypes = [
		{ type: 'application/pdf', suffixes: 'pdf', description: '', enabledPlugin: mockPlugins[1] },
		{ type: 'application/x-google-chrome-pdf', suffixes: 'pdf', description: 'Portable Document Format', enabledPlugin: mockPlugins[0] },
		{ type: 'application/x-nacl', suffixes: '', description: 'Native Client Executable', enabledPlugin: mockPlugins[2] },
		{ type: 'application/x-pnacl', suffixes: '', description: 'Portable Native Client Executable', enabledPlugin: mockPlugins[2] }
	];
	
	const mimeTypeArray = {
		length: mimeTypes.length,
		item: function(i) { return mimeTypes[i]; },
		namedItem: function(name) { return mimeTypes.find(m => m.type === name); }
	};
	
	mimeTypes.forEach((mime, i) => {
		mimeTypeArray[i] = mime;
		mimeTypeArray[mime.type] = mime;
	});
	
	Object.defineProperty(navigator, 'mimeTypes', {
		get: () => mimeTypeArray,
		configurable: true
	});
`
}

func (e *EvasionSuite) puppeteerBypass() string {
	return `
	// Puppeteer Detection Bypass
	
	// Remove Puppeteer-specific properties
	const puppeteerProps = [
		'__puppeteer_evaluation_script__',
		'__puppeteer_utility_world__'
	];
	
	puppeteerProps.forEach(prop => {
		try {
			if (window[prop]) delete window[prop];
		} catch (e) {}
	});
	
	// Override Function.prototype.toString to hide native code modifications
	const originalFunctionToString = Function.prototype.toString;
	const nativeCodeRegex = /\[native code\]/;
	
	Function.prototype.toString = function() {
		const result = originalFunctionToString.call(this);
		// If this is a modified native function, return fake native code
		if (this === navigator.permissions.query) {
			return 'function query() { [native code] }';
		}
		if (this === navigator.plugins.item) {
			return 'function item() { [native code] }';
		}
		return result;
	};
	
	// Permissions API spoofing
	const originalQuery = navigator.permissions.query;
	navigator.permissions.query = function(parameters) {
		if (parameters.name === 'notifications') {
			return Promise.resolve({ state: Notification.permission, onchange: null });
		}
		return originalQuery.apply(this, arguments);
	};
	
	// Override Error stack traces to remove puppeteer references
	const originalPrepareStackTrace = Error.prepareStackTrace;
	Error.prepareStackTrace = function(error, structuredStackTrace) {
		const filtered = structuredStackTrace.filter(frame => {
			const fileName = frame.getFileName() || '';
			return !fileName.includes('puppeteer') && !fileName.includes('pptr');
		});
		if (originalPrepareStackTrace) {
			return originalPrepareStackTrace(error, filtered);
		}
		return filtered.map(frame => '    at ' + frame.toString()).join('\n');
	};
`
}

func (e *EvasionSuite) playwrightBypass() string {
	return `
	// Playwright Detection Bypass
	
	// Remove Playwright-specific properties
	const playwrightProps = [
		'__playwright',
		'__pw_manual',
		'__PW_inspect'
	];
	
	playwrightProps.forEach(prop => {
		try {
			if (window[prop]) delete window[prop];
		} catch (e) {}
	});
	
	// Playwright uses specific binding names
	const bindingNames = Object.keys(window).filter(key => 
		key.startsWith('__playwright') || 
		key.startsWith('__pw') ||
		key.includes('playwright')
	);
	
	bindingNames.forEach(name => {
		try {
			delete window[name];
		} catch (e) {}
	});
	
	// Override window.open to prevent detection via popup behavior
	const originalWindowOpen = window.open;
	window.open = function() {
		const result = originalWindowOpen.apply(this, arguments);
		if (result) {
			// Remove any playwright markers from new window
			try {
				playwrightProps.forEach(prop => {
					if (result[prop]) delete result[prop];
				});
			} catch (e) {}
		}
		return result;
	};
`
}

func (e *EvasionSuite) seleniumBypass() string {
	return `
	// Selenium Detection Bypass
	
	// Remove Selenium-specific properties
	const seleniumProps = [
		'_Selenium_IDE_Recorder',
		'_selenium',
		'calledSelenium',
		'$cdc_asdjflasutopfhvcZLmcfl_',
		'$wdc_',
		'__selenium_evaluate',
		'__selenium_unwrapped',
		'__fxdriver_evaluate',
		'__fxdriver_unwrapped',
		'__webdriver_evaluate',
		'__webdriver_unwrapped',
		'__driver_evaluate',
		'__driver_unwrapped'
	];
	
	seleniumProps.forEach(prop => {
		try {
			if (window[prop]) delete window[prop];
			if (document[prop]) delete document[prop];
		} catch (e) {}
	});
	
	// Remove $cdc_ prefixed properties (ChromeDriver)
	Object.keys(window).forEach(key => {
		if (key.startsWith('$cdc_') || key.startsWith('$wdc_')) {
			try {
				delete window[key];
			} catch (e) {}
		}
	});
	
	// Override document.querySelector to hide selenium elements
	const originalQuerySelector = document.querySelector;
	document.querySelector = function(selector) {
		if (selector.includes('selenium') || selector.includes('webdriver')) {
			return null;
		}
		return originalQuerySelector.apply(this, arguments);
	};
	
	// Override document.querySelectorAll
	const originalQuerySelectorAll = document.querySelectorAll;
	document.querySelectorAll = function(selector) {
		if (selector.includes('selenium') || selector.includes('webdriver')) {
			return [];
		}
		return originalQuerySelectorAll.apply(this, arguments);
	};
	
	// Remove selenium from document.$
	if (document.$) {
		delete document.$;
	}
	if (document.$$) {
		delete document.$$;
	}
`
}

func (e *EvasionSuite) cdpMasking() string {
	return `
	// Chrome DevTools Protocol Fingerprint Masking
	
	// Override Runtime.evaluate detection
	const originalEval = window.eval;
	window.eval = function(code) {
		// Remove CDP markers from evaluated code
		if (typeof code === 'string') {
			code = code.replace(/__cdp_binding__/g, '');
			code = code.replace(/Runtime\.evaluate/g, '');
		}
		return originalEval.apply(this, arguments);
	};
	
	// Mask CDP-specific console methods
	const originalConsoleDebug = console.debug;
	console.debug = function() {
		const args = Array.from(arguments);
		const filtered = args.filter(arg => {
			if (typeof arg === 'string') {
				return !arg.includes('CDP') && !arg.includes('DevTools');
			}
			return true;
		});
		if (filtered.length > 0) {
			originalConsoleDebug.apply(this, filtered);
		}
	};
	
	// Override Performance.now() to add slight randomness
	const originalPerformanceNow = performance.now;
	performance.now = function() {
		return originalPerformanceNow.call(this) + (Math.random() * 0.1);
	};
	
	// Mask CDP Runtime domain
	if (window.Runtime) {
		delete window.Runtime;
	}
	
	// Override Object.keys to hide CDP properties
	const originalObjectKeys = Object.keys;
	Object.keys = function(obj) {
		const keys = originalObjectKeys.apply(this, arguments);
		if (obj === window) {
			return keys.filter(key => 
				!key.includes('cdp') && 
				!key.includes('CDP') && 
				!key.includes('__') ||
				key === '__proto__'
			);
		}
		return keys;
	};
`
}

func (e *EvasionSuite) renderingDetectionBypass() string {
	return `
	// Headless Mode Detection via Rendering Bypass
	
	// Spoof WebGL renderer info
	const getParameterProxyHandler = {
		apply: function(target, thisArg, args) {
			const param = args[0];
			const gl = thisArg;
			
			// UNMASKED_VENDOR_WEBGL
			if (param === 37445) {
				return 'Google Inc. (NVIDIA)';
			}
			// UNMASKED_RENDERER_WEBGL
			if (param === 37446) {
				return 'ANGLE (NVIDIA, NVIDIA GeForce GTX 1080 Direct3D11 vs_5_0 ps_5_0, D3D11)';
			}
			
			return target.apply(thisArg, args);
		}
	};
	
	// Apply to WebGL contexts
	const getContext = HTMLCanvasElement.prototype.getContext;
	HTMLCanvasElement.prototype.getContext = function(type, attributes) {
		const context = getContext.apply(this, arguments);
		if (context && (type === 'webgl' || type === 'webgl2' || type === 'experimental-webgl')) {
			const originalGetParameter = context.getParameter.bind(context);
			context.getParameter = new Proxy(originalGetParameter, getParameterProxyHandler);
		}
		return context;
	};
	
	// Spoof screen properties to avoid headless detection
	const screenProps = {
		availTop: 0,
		availLeft: 0,
		availHeight: screen.height,
		availWidth: screen.width
	};
	
	Object.keys(screenProps).forEach(prop => {
		Object.defineProperty(screen, prop, {
			get: () => screenProps[prop],
			configurable: true
		});
	});
	
	// Spoof window.outerWidth/outerHeight (headless often has 0)
	if (window.outerWidth === 0) {
		Object.defineProperty(window, 'outerWidth', {
			get: () => window.innerWidth + 16,
			configurable: true
		});
	}
	if (window.outerHeight === 0) {
		Object.defineProperty(window, 'outerHeight', {
			get: () => window.innerHeight + 88,
			configurable: true
		});
	}
	
	// Spoof Notification permission (headless often has 'denied')
	if (Notification.permission === 'denied') {
		Object.defineProperty(Notification, 'permission', {
			get: () => 'default',
			configurable: true
		});
	}
	
	// Add realistic touch support detection
	Object.defineProperty(navigator, 'maxTouchPoints', {
		get: () => 0, // Desktop typically has 0
		configurable: true
	});
	
	// Spoof connection info
	if (navigator.connection) {
		Object.defineProperty(navigator.connection, 'rtt', {
			get: () => 50 + Math.floor(Math.random() * 50),
			configurable: true
		});
	}
	
	// Override Image loading to simulate real rendering
	const originalImage = window.Image;
	window.Image = function(width, height) {
		const img = new originalImage(width, height);
		// Add slight delay to simulate real image loading
		const originalOnload = img.onload;
		Object.defineProperty(img, 'onload', {
			set: function(fn) {
				originalOnload = function(e) {
					setTimeout(() => fn && fn.call(this, e), Math.random() * 10);
				};
			},
			get: function() { return originalOnload; }
		});
		return img;
	};
	window.Image.prototype = originalImage.prototype;
`
}

// GenerateRandomNoise generates random noise values for fingerprint spoofing
func GenerateRandomNoise() float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return float64(n.Int64()) / 10000.0
}

// GenerateCanvasNoise generates canvas fingerprint noise script
func GenerateCanvasNoise() string {
	noise := GenerateRandomNoise()
	return fmt.Sprintf(`
(function() {
	const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
	HTMLCanvasElement.prototype.toDataURL = function(type) {
		if (type === 'image/png' || type === undefined) {
			const context = this.getContext('2d');
			if (context) {
				const imageData = context.getImageData(0, 0, this.width, this.height);
				for (let i = 0; i < imageData.data.length; i += 4) {
					imageData.data[i] = imageData.data[i] + (Math.random() * %f - %f);
				}
				context.putImageData(imageData, 0, 0);
			}
		}
		return originalToDataURL.apply(this, arguments);
	};
})();
`, noise*2, noise)
}

// GenerateAudioNoise generates audio fingerprint noise script
func GenerateAudioNoise() string {
	return `
(function() {
	const context = window.AudioContext || window.webkitAudioContext;
	if (context) {
		const originalCreateAnalyser = context.prototype.createAnalyser;
		context.prototype.createAnalyser = function() {
			const analyser = originalCreateAnalyser.apply(this, arguments);
			const originalGetFloatFrequencyData = analyser.getFloatFrequencyData;
			analyser.getFloatFrequencyData = function(array) {
				originalGetFloatFrequencyData.apply(this, arguments);
				for (let i = 0; i < array.length; i++) {
					array[i] = array[i] + (Math.random() * 0.0001 - 0.00005);
				}
			};
			return analyser;
		};
	}
})();
`
}
