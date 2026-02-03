package antidetect

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	mrand "math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// AntiDetectConfig anti-detection yapılandırması
type AntiDetectConfig struct {
	WebRTCProtection     bool // WebRTC IP leak önleme
	CanvasNoise          bool // Canvas fingerprint noise
	AudioNoise           bool // AudioContext fingerprint noise
	WebGLNoise           bool // WebGL fingerprint noise
	FontNoise            bool // Font fingerprint noise
	PluginSpoof          bool // Plugin spoofing
	BatterySpoof         bool // Battery API spoofing
	SensorSpoof          bool // Sensor API spoofing (mobil)
	TimingRandomization  bool // Keyboard/mouse timing randomization
	NavigatorSpoof       bool // Navigator property spoofing
	ScreenSpoof          bool // Screen property spoofing
	HardwareSpoof        bool // Hardware concurrency spoofing
	MemorySpoof          bool // Device memory spoofing
	ConnectionSpoof      bool // Network connection spoofing
}

// AntiDetect anti-detection yöneticisi
type AntiDetect struct {
	config AntiDetectConfig
	mu     sync.Mutex
	rng    *mrand.Rand
}

// NewAntiDetect yeni anti-detect oluşturur
func NewAntiDetect(config AntiDetectConfig) *AntiDetect {
	return &AntiDetect{
		config: config,
		rng:    mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
}

// NewDefaultAntiDetect varsayılan ayarlarla anti-detect oluşturur
func NewDefaultAntiDetect() *AntiDetect {
	return NewAntiDetect(AntiDetectConfig{
		WebRTCProtection:    true,
		CanvasNoise:         true,
		AudioNoise:          true,
		WebGLNoise:          true,
		FontNoise:           true,
		PluginSpoof:         true,
		BatterySpoof:        true,
		SensorSpoof:         true,
		TimingRandomization: true,
		NavigatorSpoof:      true,
		ScreenSpoof:         true,
		HardwareSpoof:       true,
		MemorySpoof:         true,
		ConnectionSpoof:     true,
	})
}

// InjectAll tüm anti-detection scriptlerini enjekte eder
func (ad *AntiDetect) InjectAll(ctx context.Context) error {
	scripts := ad.GenerateAllScripts()
	
	for _, script := range scripts {
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			// Hata kritik değil, devam et
			_ = err
		}
	}
	
	return nil
}

// GenerateAllScripts tüm anti-detection scriptlerini oluşturur
func (ad *AntiDetect) GenerateAllScripts() []string {
	var scripts []string
	
	if ad.config.WebRTCProtection {
		scripts = append(scripts, ad.generateWebRTCScript())
	}
	if ad.config.CanvasNoise {
		scripts = append(scripts, ad.generateCanvasNoiseScript())
	}
	if ad.config.AudioNoise {
		scripts = append(scripts, ad.generateAudioNoiseScript())
	}
	if ad.config.WebGLNoise {
		scripts = append(scripts, ad.generateWebGLNoiseScript())
	}
	if ad.config.FontNoise {
		scripts = append(scripts, ad.generateFontNoiseScript())
	}
	if ad.config.PluginSpoof {
		scripts = append(scripts, ad.generatePluginSpoofScript())
	}
	if ad.config.BatterySpoof {
		scripts = append(scripts, ad.generateBatterySpoofScript())
	}
	if ad.config.SensorSpoof {
		scripts = append(scripts, ad.generateSensorSpoofScript())
	}
	if ad.config.TimingRandomization {
		scripts = append(scripts, ad.generateTimingRandomizationScript())
	}
	if ad.config.NavigatorSpoof {
		scripts = append(scripts, ad.generateNavigatorSpoofScript())
	}
	if ad.config.ConnectionSpoof {
		scripts = append(scripts, ad.generateConnectionSpoofScript())
	}
	
	return scripts
}

// generateWebRTCScript WebRTC IP leak önleme scripti
func (ad *AntiDetect) generateWebRTCScript() string {
	return `
(function() {
	// WebRTC IP leak önleme
	if (typeof RTCPeerConnection !== 'undefined') {
		const originalRTCPeerConnection = RTCPeerConnection;
		
		RTCPeerConnection = function(config) {
			if (config && config.iceServers) {
				config.iceServers = [];
			}
			return new originalRTCPeerConnection(config);
		};
		
		RTCPeerConnection.prototype = originalRTCPeerConnection.prototype;
	}
	
	// webkitRTCPeerConnection için de
	if (typeof webkitRTCPeerConnection !== 'undefined') {
		const originalWebkitRTC = webkitRTCPeerConnection;
		
		webkitRTCPeerConnection = function(config) {
			if (config && config.iceServers) {
				config.iceServers = [];
			}
			return new originalWebkitRTC(config);
		};
		
		webkitRTCPeerConnection.prototype = originalWebkitRTC.prototype;
	}
})();
`
}

// generateCanvasNoiseScript Canvas fingerprint noise scripti
func (ad *AntiDetect) generateCanvasNoiseScript() string {
	noise := ad.generateNoise()
	return fmt.Sprintf(`
(function() {
	const noise = %f;
	
	// Canvas 2D noise
	const originalGetImageData = CanvasRenderingContext2D.prototype.getImageData;
	CanvasRenderingContext2D.prototype.getImageData = function(x, y, w, h) {
		const imageData = originalGetImageData.call(this, x, y, w, h);
		const data = imageData.data;
		
		for (let i = 0; i < data.length; i += 4) {
			// RGB kanallarına küçük noise ekle
			data[i] = Math.max(0, Math.min(255, data[i] + Math.floor((Math.random() - 0.5) * noise * 2)));
			data[i + 1] = Math.max(0, Math.min(255, data[i + 1] + Math.floor((Math.random() - 0.5) * noise * 2)));
			data[i + 2] = Math.max(0, Math.min(255, data[i + 2] + Math.floor((Math.random() - 0.5) * noise * 2)));
		}
		
		return imageData;
	};
	
	// toDataURL noise
	const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
	HTMLCanvasElement.prototype.toDataURL = function(type, quality) {
		const ctx = this.getContext('2d');
		if (ctx) {
			const imageData = ctx.getImageData(0, 0, this.width, this.height);
			const data = imageData.data;
			
			for (let i = 0; i < data.length; i += 4) {
				data[i] = Math.max(0, Math.min(255, data[i] + Math.floor((Math.random() - 0.5) * noise)));
			}
			
			ctx.putImageData(imageData, 0, 0);
		}
		return originalToDataURL.call(this, type, quality);
	};
})();
`, noise)
}

// generateAudioNoiseScript AudioContext fingerprint noise scripti
func (ad *AntiDetect) generateAudioNoiseScript() string {
	noise := ad.generateNoise() / 1000
	return fmt.Sprintf(`
(function() {
	const noise = %f;
	
	// AudioContext noise
	if (typeof AudioContext !== 'undefined') {
		const originalCreateOscillator = AudioContext.prototype.createOscillator;
		AudioContext.prototype.createOscillator = function() {
			const oscillator = originalCreateOscillator.call(this);
			const originalFrequency = oscillator.frequency.value;
			oscillator.frequency.value = originalFrequency + (Math.random() - 0.5) * noise;
			return oscillator;
		};
		
		const originalCreateAnalyser = AudioContext.prototype.createAnalyser;
		AudioContext.prototype.createAnalyser = function() {
			const analyser = originalCreateAnalyser.call(this);
			const originalGetFloatFrequencyData = analyser.getFloatFrequencyData.bind(analyser);
			
			analyser.getFloatFrequencyData = function(array) {
				originalGetFloatFrequencyData(array);
				for (let i = 0; i < array.length; i++) {
					array[i] += (Math.random() - 0.5) * noise * 100;
				}
			};
			
			return analyser;
		};
	}
	
	// OfflineAudioContext için de
	if (typeof OfflineAudioContext !== 'undefined') {
		const originalRender = OfflineAudioContext.prototype.startRendering;
		OfflineAudioContext.prototype.startRendering = function() {
			return originalRender.call(this).then(function(buffer) {
				const channelData = buffer.getChannelData(0);
				for (let i = 0; i < channelData.length; i++) {
					channelData[i] += (Math.random() - 0.5) * noise;
				}
				return buffer;
			});
		};
	}
})();
`, noise)
}

// generateWebGLNoiseScript WebGL fingerprint noise scripti
func (ad *AntiDetect) generateWebGLNoiseScript() string {
	vendor := ad.getRandomWebGLVendor()
	renderer := ad.getRandomWebGLRenderer()
	
	return fmt.Sprintf(`
(function() {
	const vendor = '%s';
	const renderer = '%s';
	
	// WebGL vendor/renderer spoofing
	const getParameterOriginal = WebGLRenderingContext.prototype.getParameter;
	WebGLRenderingContext.prototype.getParameter = function(parameter) {
		if (parameter === 37445) { // UNMASKED_VENDOR_WEBGL
			return vendor;
		}
		if (parameter === 37446) { // UNMASKED_RENDERER_WEBGL
			return renderer;
		}
		return getParameterOriginal.call(this, parameter);
	};
	
	// WebGL2 için de
	if (typeof WebGL2RenderingContext !== 'undefined') {
		const getParameter2Original = WebGL2RenderingContext.prototype.getParameter;
		WebGL2RenderingContext.prototype.getParameter = function(parameter) {
			if (parameter === 37445) {
				return vendor;
			}
			if (parameter === 37446) {
				return renderer;
			}
			return getParameter2Original.call(this, parameter);
		};
	}
})();
`, vendor, renderer)
}

// generateFontNoiseScript Font fingerprint noise scripti
func (ad *AntiDetect) generateFontNoiseScript() string {
	return `
(function() {
	// Font detection noise
	const originalOffsetWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');
	const originalOffsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight');
	
	Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
		get: function() {
			const width = originalOffsetWidth.get.call(this);
			if (this.style.fontFamily) {
				return width + Math.floor((Math.random() - 0.5) * 2);
			}
			return width;
		}
	});
	
	Object.defineProperty(HTMLElement.prototype, 'offsetHeight', {
		get: function() {
			const height = originalOffsetHeight.get.call(this);
			if (this.style.fontFamily) {
				return height + Math.floor((Math.random() - 0.5) * 2);
			}
			return height;
		}
	});
})();
`
}

// generatePluginSpoofScript Plugin spoofing scripti
func (ad *AntiDetect) generatePluginSpoofScript() string {
	return `
(function() {
	// Plugin spoofing
	const plugins = [
		{name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format'},
		{name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: ''},
		{name: 'Native Client', filename: 'internal-nacl-plugin', description: ''}
	];
	
	Object.defineProperty(navigator, 'plugins', {
		get: function() {
			const pluginArray = {
				length: plugins.length,
				item: function(i) { return plugins[i]; },
				namedItem: function(name) { return plugins.find(p => p.name === name); },
				refresh: function() {}
			};
			
			plugins.forEach((p, i) => {
				pluginArray[i] = p;
			});
			
			return pluginArray;
		}
	});
	
	Object.defineProperty(navigator, 'mimeTypes', {
		get: function() {
			return {
				length: 2,
				item: function(i) { return this[i]; },
				namedItem: function(name) { return null; },
				0: {type: 'application/pdf', suffixes: 'pdf', description: 'Portable Document Format'},
				1: {type: 'text/pdf', suffixes: 'pdf', description: 'Portable Document Format'}
			};
		}
	});
})();
`
}

// generateBatterySpoofScript Battery API spoofing scripti
func (ad *AntiDetect) generateBatterySpoofScript() string {
	level := 0.5 + ad.rng.Float64()*0.5 // 0.5-1.0 arası
	return fmt.Sprintf(`
(function() {
	// Battery API spoofing
	if (navigator.getBattery) {
		const originalGetBattery = navigator.getBattery;
		navigator.getBattery = function() {
			return Promise.resolve({
				charging: true,
				chargingTime: Infinity,
				dischargingTime: Infinity,
				level: %f,
				addEventListener: function() {},
				removeEventListener: function() {}
			});
		};
	}
})();
`, level)
}

// generateSensorSpoofScript Sensor API spoofing scripti
func (ad *AntiDetect) generateSensorSpoofScript() string {
	return `
(function() {
	// DeviceMotionEvent spoofing
	if (typeof DeviceMotionEvent !== 'undefined') {
		const originalAddEventListener = EventTarget.prototype.addEventListener;
		EventTarget.prototype.addEventListener = function(type, listener, options) {
			if (type === 'devicemotion' || type === 'deviceorientation') {
				// Sensör eventlerini engelle veya sahte veri gönder
				return;
			}
			return originalAddEventListener.call(this, type, listener, options);
		};
	}
	
	// Accelerometer, Gyroscope, Magnetometer spoofing
	['Accelerometer', 'Gyroscope', 'Magnetometer', 'AbsoluteOrientationSensor', 'RelativeOrientationSensor'].forEach(function(sensor) {
		if (typeof window[sensor] !== 'undefined') {
			window[sensor] = function() {
				return {
					start: function() {},
					stop: function() {},
					addEventListener: function() {},
					removeEventListener: function() {}
				};
			};
		}
	});
})();
`
}

// generateTimingRandomizationScript Timing randomization scripti
func (ad *AntiDetect) generateTimingRandomizationScript() string {
	return `
(function() {
	// Performance.now() noise
	const originalNow = Performance.prototype.now;
	Performance.prototype.now = function() {
		return originalNow.call(this) + (Math.random() * 0.1);
	};
	
	// Date.now() noise
	const originalDateNow = Date.now;
	Date.now = function() {
		return originalDateNow() + Math.floor(Math.random() * 2);
	};
	
	// requestAnimationFrame timing noise
	const originalRAF = window.requestAnimationFrame;
	window.requestAnimationFrame = function(callback) {
		return originalRAF.call(window, function(timestamp) {
			callback(timestamp + Math.random() * 0.1);
		});
	};
})();
`
}

// generateNavigatorSpoofScript Navigator property spoofing scripti
func (ad *AntiDetect) generateNavigatorSpoofScript() string {
	return `
(function() {
	// webdriver property gizle
	Object.defineProperty(navigator, 'webdriver', {
		get: function() { return undefined; },
		configurable: true
	});
	
	// Automation flags gizle
	delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
	delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
	delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;
	
	// Chrome automation extension gizle
	if (window.chrome) {
		window.chrome.runtime = undefined;
	}
	
	// Permissions API spoof
	if (navigator.permissions) {
		const originalQuery = navigator.permissions.query;
		navigator.permissions.query = function(parameters) {
			if (parameters.name === 'notifications') {
				return Promise.resolve({state: 'prompt', onchange: null});
			}
			return originalQuery.call(this, parameters);
		};
	}
})();
`
}

// generateConnectionSpoofScript Network connection spoofing scripti
func (ad *AntiDetect) generateConnectionSpoofScript() string {
	connections := []string{"4g", "3g", "wifi"}
	connection := connections[ad.rng.Intn(len(connections))]
	
	return fmt.Sprintf(`
(function() {
	// Network Information API spoofing
	if (navigator.connection) {
		Object.defineProperty(navigator.connection, 'effectiveType', {
			get: function() { return '%s'; }
		});
		Object.defineProperty(navigator.connection, 'downlink', {
			get: function() { return 10 + Math.random() * 20; }
		});
		Object.defineProperty(navigator.connection, 'rtt', {
			get: function() { return 50 + Math.floor(Math.random() * 100); }
		});
	}
})();
`, connection)
}

// generateNoise rastgele noise değeri üretir
func (ad *AntiDetect) generateNoise() float64 {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	return 1.0 + ad.rng.Float64()*4.0 // 1-5 arası
}

// getRandomWebGLVendor rastgele WebGL vendor döner
func (ad *AntiDetect) getRandomWebGLVendor() string {
	vendors := []string{
		"Google Inc. (NVIDIA)",
		"Google Inc. (Intel)",
		"Google Inc. (AMD)",
		"Google Inc. (Microsoft)",
		"Intel Inc.",
		"NVIDIA Corporation",
	}
	ad.mu.Lock()
	defer ad.mu.Unlock()
	return vendors[ad.rng.Intn(len(vendors))]
}

// getRandomWebGLRenderer rastgele WebGL renderer döner
func (ad *AntiDetect) getRandomWebGLRenderer() string {
	renderers := []string{
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1660 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (Intel, Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (AMD, AMD Radeon RX 580 Series Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (Intel, Intel(R) Iris(R) Xe Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)",
	}
	ad.mu.Lock()
	defer ad.mu.Unlock()
	return renderers[ad.rng.Intn(len(renderers))]
}

// GenerateUniqueID benzersiz ID üretir
func GenerateUniqueID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateClientID benzersiz client ID üretir (GA4 formatında)
func GenerateClientID() string {
	timestamp := time.Now().Unix()
	random := mrand.Int63n(int64(math.Pow(10, 9)))
	return fmt.Sprintf("%d.%d", random, timestamp)
}

// GetOnNewDocumentScript sayfa yüklenmeden önce çalışacak script
func (ad *AntiDetect) GetOnNewDocumentScript() string {
	scripts := ad.GenerateAllScripts()
	combined := ""
	for _, script := range scripts {
		combined += script + "\n"
	}
	return combined
}

// ============================================================================
// KEYWORD CLUSTERING & ROTATION SYSTEM
// ============================================================================

// KeywordCluster semantik olarak ilişkili keyword gruplarını temsil eder
type KeywordCluster struct {
	ID          string            // Cluster benzersiz ID
	Name        string            // Cluster adı (örn: "e-commerce", "tech")
	PrimaryKW   string            // Ana keyword
	Variations  []string          // Varyasyonlar
	LongTails   []string          // Long-tail keywordler
	Synonyms    []string          // Eş anlamlılar
	Related     []string          // İlişkili keywordler
	Modifiers   []string          // Modifier'lar (best, top, cheap, vb.)
	Locations   []string          // Lokasyon modifierleri
	Intent      SearchIntent      // Arama niyeti
	Weight      float64           // Cluster ağırlığı (0-1)
	UsageCount  int               // Kullanım sayısı
	LastUsed    time.Time         // Son kullanım zamanı
	Metadata    map[string]string // Ek metadata
}

// SearchIntent arama niyeti türleri
type SearchIntent string

const (
	IntentInformational SearchIntent = "informational" // Bilgi arama
	IntentNavigational  SearchIntent = "navigational"  // Site/sayfa arama
	IntentTransactional SearchIntent = "transactional" // Satın alma niyeti
	IntentCommercial    SearchIntent = "commercial"    // Araştırma + satın alma
	IntentLocal         SearchIntent = "local"         // Yerel arama
)

// KeywordRotationStrategy rotation stratejisi
type KeywordRotationStrategy string

const (
	RotationRoundRobin   KeywordRotationStrategy = "round_robin"   // Sıralı döngü
	RotationWeighted     KeywordRotationStrategy = "weighted"      // Ağırlıklı seçim
	RotationRandom       KeywordRotationStrategy = "random"        // Rastgele
	RotationTimeBased    KeywordRotationStrategy = "time_based"    // Zaman bazlı
	RotationAdaptive     KeywordRotationStrategy = "adaptive"      // Performans bazlı adaptif
	RotationAntiPattern  KeywordRotationStrategy = "anti_pattern"  // Pattern detection önleme
)

// KeywordClusterManager keyword cluster yöneticisi
type KeywordClusterManager struct {
	mu                sync.RWMutex
	clusters          map[string]*KeywordCluster
	rotationStrategy  KeywordRotationStrategy
	rotationIndex     map[string]int           // Cluster başına rotation index
	usageHistory      []KeywordUsageRecord     // Kullanım geçmişi
	patternDetector   *PatternDetector         // Pattern detection
	rng               *mrand.Rand
	maxHistorySize    int
	cooldownPeriod    time.Duration            // Aynı keyword için bekleme süresi
	clusterCooldown   time.Duration            // Aynı cluster için bekleme süresi
}

// KeywordUsageRecord keyword kullanım kaydı
type KeywordUsageRecord struct {
	Keyword     string
	ClusterID   string
	Timestamp   time.Time
	SessionID   string
	Success     bool
	ResponseMS  int64
}

// PatternDetector pattern detection sistemi
type PatternDetector struct {
	mu              sync.RWMutex
	sequenceWindow  int                      // Analiz penceresi
	sequences       []string                 // Son keyword dizisi
	patternScores   map[string]float64       // Pattern skorları
	alertThreshold  float64                  // Uyarı eşiği
}

// NewKeywordClusterManager yeni cluster manager oluşturur
func NewKeywordClusterManager(strategy KeywordRotationStrategy) *KeywordClusterManager {
	return &KeywordClusterManager{
		clusters:         make(map[string]*KeywordCluster),
		rotationStrategy: strategy,
		rotationIndex:    make(map[string]int),
		usageHistory:     make([]KeywordUsageRecord, 0, 10000),
		patternDetector:  NewPatternDetector(50, 0.7),
		rng:              mrand.New(mrand.NewSource(time.Now().UnixNano())),
		maxHistorySize:   10000,
		cooldownPeriod:   5 * time.Minute,
		clusterCooldown:  2 * time.Minute,
	}
}

// NewPatternDetector yeni pattern detector oluşturur
func NewPatternDetector(windowSize int, threshold float64) *PatternDetector {
	return &PatternDetector{
		sequenceWindow: windowSize,
		sequences:      make([]string, 0, windowSize),
		patternScores:  make(map[string]float64),
		alertThreshold: threshold,
	}
}

// AddCluster yeni cluster ekler
func (kcm *KeywordClusterManager) AddCluster(cluster *KeywordCluster) {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()
	
	if cluster.ID == "" {
		cluster.ID = generateClusterID(cluster.PrimaryKW)
	}
	if cluster.Metadata == nil {
		cluster.Metadata = make(map[string]string)
	}
	
	kcm.clusters[cluster.ID] = cluster
	kcm.rotationIndex[cluster.ID] = 0
}

// CreateClusterFromKeyword tek keyword'den cluster oluşturur
func (kcm *KeywordClusterManager) CreateClusterFromKeyword(keyword string, intent SearchIntent) *KeywordCluster {
	cluster := &KeywordCluster{
		ID:         generateClusterID(keyword),
		Name:       keyword,
		PrimaryKW:  keyword,
		Variations: generateVariations(keyword),
		LongTails:  generateLongTails(keyword, intent),
		Synonyms:   []string{},
		Related:    []string{},
		Modifiers:  getDefaultModifiers(intent),
		Locations:  []string{},
		Intent:     intent,
		Weight:     1.0,
		UsageCount: 0,
		LastUsed:   time.Time{},
		Metadata:   make(map[string]string),
	}
	
	kcm.AddCluster(cluster)
	return cluster
}

// GetNextKeyword rotation stratejisine göre sonraki keyword'ü döner
func (kcm *KeywordClusterManager) GetNextKeyword(clusterID string) (string, error) {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()
	
	cluster, exists := kcm.clusters[clusterID]
	if !exists {
		return "", fmt.Errorf("cluster not found: %s", clusterID)
	}
	
	// Tüm keyword'leri birleştir
	allKeywords := kcm.getAllKeywordsFromCluster(cluster)
	if len(allKeywords) == 0 {
		return cluster.PrimaryKW, nil
	}
	
	var selectedKW string
	
	switch kcm.rotationStrategy {
	case RotationRoundRobin:
		selectedKW = kcm.roundRobinSelect(clusterID, allKeywords)
	case RotationWeighted:
		selectedKW = kcm.weightedSelect(allKeywords, cluster.Weight)
	case RotationRandom:
		selectedKW = kcm.randomSelect(allKeywords)
	case RotationTimeBased:
		selectedKW = kcm.timeBasedSelect(clusterID, allKeywords)
	case RotationAdaptive:
		selectedKW = kcm.adaptiveSelect(clusterID, allKeywords)
	case RotationAntiPattern:
		selectedKW = kcm.antiPatternSelect(clusterID, allKeywords)
	default:
		selectedKW = kcm.randomSelect(allKeywords)
	}
	
	// Kullanım kaydı
	cluster.UsageCount++
	cluster.LastUsed = time.Now()
	
	// Pattern detector'a ekle
	kcm.patternDetector.AddSequence(selectedKW)
	
	return selectedKW, nil
}

// GetNextKeywordWithModifier modifier ile keyword döner
func (kcm *KeywordClusterManager) GetNextKeywordWithModifier(clusterID string) (string, error) {
	kcm.mu.RLock()
	cluster, exists := kcm.clusters[clusterID]
	kcm.mu.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("cluster not found: %s", clusterID)
	}
	
	baseKW, err := kcm.GetNextKeyword(clusterID)
	if err != nil {
		return "", err
	}
	
	// %60 ihtimalle modifier ekle
	if kcm.rng.Float64() < 0.6 && len(cluster.Modifiers) > 0 {
		modifier := cluster.Modifiers[kcm.rng.Intn(len(cluster.Modifiers))]
		
		// Modifier pozisyonu: %70 önde, %30 arkada
		if kcm.rng.Float64() < 0.7 {
			return modifier + " " + baseKW, nil
		}
		return baseKW + " " + modifier, nil
	}
	
	// %20 ihtimalle lokasyon ekle
	if kcm.rng.Float64() < 0.2 && len(cluster.Locations) > 0 {
		location := cluster.Locations[kcm.rng.Intn(len(cluster.Locations))]
		return baseKW + " " + location, nil
	}
	
	return baseKW, nil
}

// GetKeywordBatch toplu keyword döner (anti-pattern için)
func (kcm *KeywordClusterManager) GetKeywordBatch(clusterID string, count int) ([]string, error) {
	kcm.mu.RLock()
	cluster, exists := kcm.clusters[clusterID]
	kcm.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", clusterID)
	}
	
	allKeywords := kcm.getAllKeywordsFromCluster(cluster)
	if len(allKeywords) == 0 {
		return []string{cluster.PrimaryKW}, nil
	}
	
	// Shuffle ve seç
	shuffled := make([]string, len(allKeywords))
	copy(shuffled, allKeywords)
	kcm.shuffleStrings(shuffled)
	
	if count > len(shuffled) {
		count = len(shuffled)
	}
	
	result := shuffled[:count]
	
	// Anti-pattern: Aynı pattern'i tekrarlamamak için karıştır
	for i := range result {
		if kcm.rng.Float64() < 0.3 {
			// %30 ihtimalle modifier ekle
			if len(cluster.Modifiers) > 0 {
				modifier := cluster.Modifiers[kcm.rng.Intn(len(cluster.Modifiers))]
				if kcm.rng.Float64() < 0.5 {
					result[i] = modifier + " " + result[i]
				} else {
					result[i] = result[i] + " " + modifier
				}
			}
		}
	}
	
	return result, nil
}

// getAllKeywordsFromCluster cluster'daki tüm keyword'leri döner
func (kcm *KeywordClusterManager) getAllKeywordsFromCluster(cluster *KeywordCluster) []string {
	var all []string
	
	// Primary keyword'ü 3x ağırlıkla ekle
	all = append(all, cluster.PrimaryKW, cluster.PrimaryKW, cluster.PrimaryKW)
	
	// Variations 2x ağırlık
	for _, v := range cluster.Variations {
		all = append(all, v, v)
	}
	
	// Diğerleri 1x
	all = append(all, cluster.LongTails...)
	all = append(all, cluster.Synonyms...)
	all = append(all, cluster.Related...)
	
	return all
}

// roundRobinSelect sıralı seçim
func (kcm *KeywordClusterManager) roundRobinSelect(clusterID string, keywords []string) string {
	idx := kcm.rotationIndex[clusterID]
	selected := keywords[idx%len(keywords)]
	kcm.rotationIndex[clusterID] = (idx + 1) % len(keywords)
	return selected
}

// weightedSelect ağırlıklı seçim
func (kcm *KeywordClusterManager) weightedSelect(keywords []string, weight float64) string {
	// Ağırlığa göre primary keyword'e bias
	if kcm.rng.Float64() < weight*0.5 {
		return keywords[0] // Primary keyword
	}
	return keywords[kcm.rng.Intn(len(keywords))]
}

// randomSelect rastgele seçim
func (kcm *KeywordClusterManager) randomSelect(keywords []string) string {
	return keywords[kcm.rng.Intn(len(keywords))]
}

// timeBasedSelect zaman bazlı seçim
func (kcm *KeywordClusterManager) timeBasedSelect(clusterID string, keywords []string) string {
	hour := time.Now().Hour()
	
	// Gündüz saatlerinde daha çeşitli, gece daha tutarlı
	if hour >= 9 && hour <= 18 {
		// İş saatleri: Daha çeşitli
		return keywords[kcm.rng.Intn(len(keywords))]
	}
	
	// Gece: Primary keyword'e bias
	if kcm.rng.Float64() < 0.6 {
		return keywords[0]
	}
	return keywords[kcm.rng.Intn(len(keywords))]
}

// adaptiveSelect performans bazlı adaptif seçim
func (kcm *KeywordClusterManager) adaptiveSelect(clusterID string, keywords []string) string {
	// Son başarılı keyword'lere bias
	successfulKWs := kcm.getSuccessfulKeywords(clusterID, 10)
	
	if len(successfulKWs) > 0 && kcm.rng.Float64() < 0.4 {
		return successfulKWs[kcm.rng.Intn(len(successfulKWs))]
	}
	
	return keywords[kcm.rng.Intn(len(keywords))]
}

// antiPatternSelect pattern detection önleyici seçim
func (kcm *KeywordClusterManager) antiPatternSelect(clusterID string, keywords []string) string {
	// Son kullanılan keyword'leri al
	recentKWs := kcm.getRecentKeywords(5)
	
	// Pattern skoru yüksek olanları filtrele
	filtered := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		// Son 5'te kullanılmamış ve pattern skoru düşük olanları seç
		if !contains(recentKWs, kw) && kcm.patternDetector.GetPatternScore(kw) < 0.5 {
			filtered = append(filtered, kw)
		}
	}
	
	if len(filtered) == 0 {
		// Fallback: Rastgele seç
		return keywords[kcm.rng.Intn(len(keywords))]
	}
	
	return filtered[kcm.rng.Intn(len(filtered))]
}

// getSuccessfulKeywords başarılı keyword'leri döner
func (kcm *KeywordClusterManager) getSuccessfulKeywords(clusterID string, limit int) []string {
	var successful []string
	
	for i := len(kcm.usageHistory) - 1; i >= 0 && len(successful) < limit; i-- {
		record := kcm.usageHistory[i]
		if record.ClusterID == clusterID && record.Success {
			successful = append(successful, record.Keyword)
		}
	}
	
	return successful
}

// getRecentKeywords son kullanılan keyword'leri döner
func (kcm *KeywordClusterManager) getRecentKeywords(limit int) []string {
	var recent []string
	
	start := len(kcm.usageHistory) - limit
	if start < 0 {
		start = 0
	}
	
	for i := start; i < len(kcm.usageHistory); i++ {
		recent = append(recent, kcm.usageHistory[i].Keyword)
	}
	
	return recent
}

// RecordUsage kullanım kaydı ekler
func (kcm *KeywordClusterManager) RecordUsage(keyword, clusterID, sessionID string, success bool, responseMS int64) {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()
	
	record := KeywordUsageRecord{
		Keyword:    keyword,
		ClusterID:  clusterID,
		Timestamp:  time.Now(),
		SessionID:  sessionID,
		Success:    success,
		ResponseMS: responseMS,
	}
	
	kcm.usageHistory = append(kcm.usageHistory, record)
	
	// History boyutunu kontrol et
	if len(kcm.usageHistory) > kcm.maxHistorySize {
		kcm.usageHistory = kcm.usageHistory[len(kcm.usageHistory)-kcm.maxHistorySize:]
	}
}

// shuffleStrings string slice'ı karıştırır
func (kcm *KeywordClusterManager) shuffleStrings(s []string) {
	for i := len(s) - 1; i > 0; i-- {
		j := kcm.rng.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

// AddSequence pattern detector'a sequence ekler
func (pd *PatternDetector) AddSequence(keyword string) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	
	pd.sequences = append(pd.sequences, keyword)
	
	// Window boyutunu aşarsa eski kayıtları sil
	if len(pd.sequences) > pd.sequenceWindow {
		pd.sequences = pd.sequences[1:]
	}
	
	// Pattern skorlarını güncelle
	pd.updatePatternScores()
}

// updatePatternScores pattern skorlarını günceller
func (pd *PatternDetector) updatePatternScores() {
	// Keyword frekanslarını hesapla
	freq := make(map[string]int)
	for _, kw := range pd.sequences {
		freq[kw]++
	}
	
	// Skorları güncelle (normalize edilmiş frekans)
	total := float64(len(pd.sequences))
	for kw, count := range freq {
		pd.patternScores[kw] = float64(count) / total
	}
}

// GetPatternScore keyword için pattern skoru döner
func (pd *PatternDetector) GetPatternScore(keyword string) float64 {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	
	if score, exists := pd.patternScores[keyword]; exists {
		return score
	}
	return 0.0
}

// IsPatternDetected pattern tespit edildi mi
func (pd *PatternDetector) IsPatternDetected() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	
	for _, score := range pd.patternScores {
		if score > pd.alertThreshold {
			return true
		}
	}
	return false
}

// GetPatternReport pattern raporu döner
func (pd *PatternDetector) GetPatternReport() map[string]float64 {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	
	report := make(map[string]float64)
	for kw, score := range pd.patternScores {
		report[kw] = score
	}
	return report
}

// ============================================================================
// SESSION FINGERPRINT ROTATION
// ============================================================================

// SessionFingerprint oturum parmak izi
type SessionFingerprint struct {
	ID              string
	UserAgent       string
	AcceptLanguage  string
	AcceptEncoding  string
	Platform        string
	Vendor          string
	ScreenRes       string
	ColorDepth      int
	Timezone        string
	DoNotTrack      string
	CookiesEnabled  bool
	JavaEnabled     bool
	WebGLVendor     string
	WebGLRenderer   string
	CanvasHash      string
	AudioHash       string
	FontsHash       string
	PluginsHash     string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	UsageCount      int
}

// FingerprintRotator fingerprint rotation yöneticisi
type FingerprintRotator struct {
	mu              sync.RWMutex
	fingerprints    []*SessionFingerprint
	activeIndex     int
	rotationCount   int
	maxFingerprints int
	rotationInterval time.Duration
	lastRotation    time.Time
	rng             *mrand.Rand
}

// NewFingerprintRotator yeni fingerprint rotator oluşturur
func NewFingerprintRotator(maxFingerprints int, rotationInterval time.Duration) *FingerprintRotator {
	fr := &FingerprintRotator{
		fingerprints:     make([]*SessionFingerprint, 0, maxFingerprints),
		activeIndex:      0,
		rotationCount:    0,
		maxFingerprints:  maxFingerprints,
		rotationInterval: rotationInterval,
		lastRotation:     time.Now(),
		rng:              mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
	
	// Başlangıç fingerprint'leri oluştur
	for i := 0; i < maxFingerprints; i++ {
		fr.fingerprints = append(fr.fingerprints, fr.generateFingerprint())
	}
	
	return fr
}

// GetCurrentFingerprint aktif fingerprint'i döner
func (fr *FingerprintRotator) GetCurrentFingerprint() *SessionFingerprint {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	
	if len(fr.fingerprints) == 0 {
		return nil
	}
	
	return fr.fingerprints[fr.activeIndex]
}

// Rotate fingerprint'i döndürür
func (fr *FingerprintRotator) Rotate() *SessionFingerprint {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	
	if len(fr.fingerprints) == 0 {
		fp := fr.generateFingerprint()
		fr.fingerprints = append(fr.fingerprints, fp)
		return fp
	}
	
	// Sonraki fingerprint'e geç
	fr.activeIndex = (fr.activeIndex + 1) % len(fr.fingerprints)
	fr.rotationCount++
	fr.lastRotation = time.Now()
	
	// Kullanım sayısını artır
	fr.fingerprints[fr.activeIndex].UsageCount++
	
	// Çok kullanılmış fingerprint'leri yenile
	if fr.fingerprints[fr.activeIndex].UsageCount > 100 {
		fr.fingerprints[fr.activeIndex] = fr.generateFingerprint()
	}
	
	return fr.fingerprints[fr.activeIndex]
}

// ShouldRotate rotation gerekli mi
func (fr *FingerprintRotator) ShouldRotate() bool {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	
	return time.Since(fr.lastRotation) > fr.rotationInterval
}

// AutoRotate otomatik rotation (gerekirse)
func (fr *FingerprintRotator) AutoRotate() *SessionFingerprint {
	if fr.ShouldRotate() {
		return fr.Rotate()
	}
	return fr.GetCurrentFingerprint()
}

// generateFingerprint yeni fingerprint oluşturur
func (fr *FingerprintRotator) generateFingerprint() *SessionFingerprint {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	}
	
	languages := []string{
		"en-US,en;q=0.9",
		"en-GB,en;q=0.9",
		"en-US,en;q=0.9,de;q=0.8",
		"en-US,en;q=0.9,fr;q=0.8",
		"en-US,en;q=0.9,es;q=0.8",
	}
	
	platforms := []string{"Win32", "MacIntel", "Linux x86_64"}
	
	screenResolutions := []string{
		"1920x1080", "2560x1440", "1366x768", "1536x864",
		"1440x900", "1680x1050", "2560x1600", "3840x2160",
	}
	
	timezones := []string{
		"America/New_York", "America/Los_Angeles", "America/Chicago",
		"Europe/London", "Europe/Paris", "Europe/Berlin",
		"Asia/Tokyo", "Asia/Shanghai", "Australia/Sydney",
	}
	
	colorDepths := []int{24, 32}
	
	webglVendors := []string{
		"Google Inc. (NVIDIA)",
		"Google Inc. (Intel)",
		"Google Inc. (AMD)",
		"Intel Inc.",
		"NVIDIA Corporation",
	}
	
	webglRenderers := []string{
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1660 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (Intel, Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (AMD, AMD Radeon RX 580 Series Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (Intel, Intel(R) Iris(R) Xe Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)",
	}
	
	return &SessionFingerprint{
		ID:             GenerateUniqueID(),
		UserAgent:      userAgents[fr.rng.Intn(len(userAgents))],
		AcceptLanguage: languages[fr.rng.Intn(len(languages))],
		AcceptEncoding: "gzip, deflate, br",
		Platform:       platforms[fr.rng.Intn(len(platforms))],
		Vendor:         "Google Inc.",
		ScreenRes:      screenResolutions[fr.rng.Intn(len(screenResolutions))],
		ColorDepth:     colorDepths[fr.rng.Intn(len(colorDepths))],
		Timezone:       timezones[fr.rng.Intn(len(timezones))],
		DoNotTrack:     "1",
		CookiesEnabled: true,
		JavaEnabled:    false,
		WebGLVendor:    webglVendors[fr.rng.Intn(len(webglVendors))],
		WebGLRenderer:  webglRenderers[fr.rng.Intn(len(webglRenderers))],
		CanvasHash:     generateRandomHash(),
		AudioHash:      generateRandomHash(),
		FontsHash:      generateRandomHash(),
		PluginsHash:    generateRandomHash(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		UsageCount:     0,
	}
}

// ============================================================================
// REQUEST PATTERN RANDOMIZATION
// ============================================================================

// RequestPattern istek pattern'i
type RequestPattern struct {
	MinDelay        time.Duration
	MaxDelay        time.Duration
	BurstSize       int           // Burst modunda kaç istek
	BurstProbability float64      // Burst olasılığı
	JitterPercent   float64       // Delay jitter yüzdesi
	TimeOfDayFactor bool          // Gün saatine göre ayarlama
}

// RequestPatternRandomizer istek pattern randomizer
type RequestPatternRandomizer struct {
	mu              sync.RWMutex
	patterns        []RequestPattern
	activePattern   int
	requestCount    int
	lastRequest     time.Time
	rng             *mrand.Rand
	humanSimulation bool
}

// NewRequestPatternRandomizer yeni pattern randomizer oluşturur
func NewRequestPatternRandomizer(humanSimulation bool) *RequestPatternRandomizer {
	rpr := &RequestPatternRandomizer{
		patterns: []RequestPattern{
			// Normal browsing pattern
			{MinDelay: 2 * time.Second, MaxDelay: 8 * time.Second, BurstSize: 1, BurstProbability: 0.0, JitterPercent: 0.3, TimeOfDayFactor: true},
			// Fast reader pattern
			{MinDelay: 1 * time.Second, MaxDelay: 4 * time.Second, BurstSize: 2, BurstProbability: 0.2, JitterPercent: 0.2, TimeOfDayFactor: true},
			// Slow reader pattern
			{MinDelay: 5 * time.Second, MaxDelay: 15 * time.Second, BurstSize: 1, BurstProbability: 0.0, JitterPercent: 0.4, TimeOfDayFactor: true},
			// Research pattern (çok sayfa açma)
			{MinDelay: 500 * time.Millisecond, MaxDelay: 2 * time.Second, BurstSize: 5, BurstProbability: 0.4, JitterPercent: 0.25, TimeOfDayFactor: false},
			// Mobile pattern
			{MinDelay: 3 * time.Second, MaxDelay: 10 * time.Second, BurstSize: 1, BurstProbability: 0.1, JitterPercent: 0.35, TimeOfDayFactor: true},
		},
		activePattern:   0,
		requestCount:    0,
		lastRequest:     time.Now(),
		rng:             mrand.New(mrand.NewSource(time.Now().UnixNano())),
		humanSimulation: humanSimulation,
	}
	
	return rpr
}

// GetNextDelay sonraki istek için delay döner
func (rpr *RequestPatternRandomizer) GetNextDelay() time.Duration {
	rpr.mu.Lock()
	defer rpr.mu.Unlock()
	
	pattern := rpr.patterns[rpr.activePattern]
	
	// Base delay hesapla
	delayRange := pattern.MaxDelay - pattern.MinDelay
	baseDelay := pattern.MinDelay + time.Duration(rpr.rng.Float64()*float64(delayRange))
	
	// Jitter ekle
	jitter := time.Duration(float64(baseDelay) * pattern.JitterPercent * (rpr.rng.Float64()*2 - 1))
	delay := baseDelay + jitter
	
	// Time of day factor
	if pattern.TimeOfDayFactor {
		delay = rpr.applyTimeOfDayFactor(delay)
	}
	
	// Human simulation
	if rpr.humanSimulation {
		delay = rpr.applyHumanFactors(delay)
	}
	
	rpr.requestCount++
	rpr.lastRequest = time.Now()
	
	// Her 50 istekte pattern değiştir
	if rpr.requestCount%50 == 0 {
		rpr.activePattern = rpr.rng.Intn(len(rpr.patterns))
	}
	
	return delay
}

// applyTimeOfDayFactor gün saatine göre delay ayarlar
func (rpr *RequestPatternRandomizer) applyTimeOfDayFactor(delay time.Duration) time.Duration {
	hour := time.Now().Hour()
	
	var factor float64
	switch {
	case hour >= 0 && hour < 6:
		// Gece: Daha yavaş
		factor = 1.5
	case hour >= 6 && hour < 9:
		// Sabah: Normal
		factor = 1.0
	case hour >= 9 && hour < 12:
		// Sabah iş saatleri: Hızlı
		factor = 0.8
	case hour >= 12 && hour < 14:
		// Öğle: Biraz yavaş
		factor = 1.1
	case hour >= 14 && hour < 18:
		// Öğleden sonra: Hızlı
		factor = 0.85
	case hour >= 18 && hour < 22:
		// Akşam: Normal
		factor = 1.0
	default:
		// Gece geç: Yavaş
		factor = 1.3
	}
	
	return time.Duration(float64(delay) * factor)
}

// applyHumanFactors insan davranışı simülasyonu
func (rpr *RequestPatternRandomizer) applyHumanFactors(delay time.Duration) time.Duration {
	// %10 ihtimalle "dikkat dağılması" - uzun pause
	if rpr.rng.Float64() < 0.1 {
		delay += time.Duration(rpr.rng.Intn(10)+5) * time.Second
	}
	
	// %5 ihtimalle "hızlı tıklama" - kısa pause
	if rpr.rng.Float64() < 0.05 {
		delay = time.Duration(float64(delay) * 0.3)
	}
	
	// %3 ihtimalle "kahve molası" - çok uzun pause
	if rpr.rng.Float64() < 0.03 {
		delay += time.Duration(rpr.rng.Intn(30)+30) * time.Second
	}
	
	return delay
}

// ShouldBurst burst modu aktif mi
func (rpr *RequestPatternRandomizer) ShouldBurst() (bool, int) {
	rpr.mu.RLock()
	defer rpr.mu.RUnlock()
	
	pattern := rpr.patterns[rpr.activePattern]
	
	if rpr.rng.Float64() < pattern.BurstProbability {
		return true, pattern.BurstSize
	}
	
	return false, 0
}

// SetPattern belirli pattern'i aktif eder
func (rpr *RequestPatternRandomizer) SetPattern(index int) {
	rpr.mu.Lock()
	defer rpr.mu.Unlock()
	
	if index >= 0 && index < len(rpr.patterns) {
		rpr.activePattern = index
	}
}

// AddCustomPattern özel pattern ekler
func (rpr *RequestPatternRandomizer) AddCustomPattern(pattern RequestPattern) {
	rpr.mu.Lock()
	defer rpr.mu.Unlock()
	
	rpr.patterns = append(rpr.patterns, pattern)
}

// ============================================================================
// BEHAVIORAL CLUSTERING
// ============================================================================

// BehaviorCluster davranış kümesi
type BehaviorCluster struct {
	ID              string
	Name            string
	Description     string
	ScrollPattern   ScrollBehavior
	ClickPattern    ClickBehavior
	ReadingPattern  ReadingBehavior
	NavigationPattern NavigationBehavior
	MousePattern    MouseBehavior
	Weight          float64
}

// ScrollBehavior scroll davranışı
type ScrollBehavior struct {
	Speed           string  // slow, medium, fast
	Pattern         string  // smooth, stepped, erratic
	PauseFrequency  float64 // 0-1 arası
	ScrollBackProb  float64 // Geri scroll olasılığı
	BottomReachProb float64 // Sayfa sonuna ulaşma olasılığı
}

// ClickBehavior tıklama davranışı
type ClickBehavior struct {
	Precision       string  // precise, sloppy
	DoubleClickProb float64
	RightClickProb  float64
	MissClickProb   float64 // Yanlış tıklama olasılığı
	HoverBeforeClick bool
	HoverDuration   time.Duration
}

// ReadingBehavior okuma davranışı
type ReadingBehavior struct {
	Speed           string  // fast, medium, slow
	SkimProb        float64 // Göz gezdirme olasılığı
	HighlightProb   float64 // Metin seçme olasılığı
	CopyProb        float64 // Kopyalama olasılığı
	ImageViewProb   float64 // Resme bakma olasılığı
}

// NavigationBehavior navigasyon davranışı
type NavigationBehavior struct {
	BackButtonProb  float64 // Geri butonu kullanma
	NewTabProb      float64 // Yeni sekme açma
	BookmarkProb    float64 // Bookmark kullanma
	SearchProb      float64 // Site içi arama
	MenuUsageProb   float64 // Menü kullanma
}

// MouseBehavior mouse davranışı
type MouseBehavior struct {
	MovementStyle   string  // linear, curved, erratic
	Speed           string  // slow, medium, fast
	IdleMovement    bool    // Boşta hareket
	TrailNoise      float64 // Hareket gürültüsü
}

// BehaviorClusterManager davranış cluster yöneticisi
type BehaviorClusterManager struct {
	mu              sync.RWMutex
	clusters        map[string]*BehaviorCluster
	activeCluster   string
	sessionBehavior *BehaviorCluster
	rng             *mrand.Rand
}

// NewBehaviorClusterManager yeni behavior cluster manager oluşturur
func NewBehaviorClusterManager() *BehaviorClusterManager {
	bcm := &BehaviorClusterManager{
		clusters: make(map[string]*BehaviorCluster),
		rng:      mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
	
	// Varsayılan davranış kümeleri
	bcm.initDefaultClusters()
	
	return bcm
}

// initDefaultClusters varsayılan kümeleri oluşturur
func (bcm *BehaviorClusterManager) initDefaultClusters() {
	// Casual Browser - Rahat gezinen kullanıcı
	bcm.clusters["casual"] = &BehaviorCluster{
		ID:          "casual",
		Name:        "Casual Browser",
		Description: "Rahat, yavaş gezinen kullanıcı",
		ScrollPattern: ScrollBehavior{
			Speed:           "slow",
			Pattern:         "smooth",
			PauseFrequency:  0.4,
			ScrollBackProb:  0.2,
			BottomReachProb: 0.3,
		},
		ClickPattern: ClickBehavior{
			Precision:        "sloppy",
			DoubleClickProb:  0.05,
			RightClickProb:   0.02,
			MissClickProb:    0.08,
			HoverBeforeClick: true,
			HoverDuration:    500 * time.Millisecond,
		},
		ReadingPattern: ReadingBehavior{
			Speed:         "slow",
			SkimProb:      0.2,
			HighlightProb: 0.1,
			CopyProb:      0.05,
			ImageViewProb: 0.4,
		},
		NavigationPattern: NavigationBehavior{
			BackButtonProb: 0.3,
			NewTabProb:     0.1,
			BookmarkProb:   0.02,
			SearchProb:     0.1,
			MenuUsageProb:  0.2,
		},
		MousePattern: MouseBehavior{
			MovementStyle: "curved",
			Speed:         "slow",
			IdleMovement:  true,
			TrailNoise:    0.3,
		},
		Weight: 0.35,
	}
	
	// Power User - Deneyimli kullanıcı
	bcm.clusters["power_user"] = &BehaviorCluster{
		ID:          "power_user",
		Name:        "Power User",
		Description: "Hızlı, deneyimli kullanıcı",
		ScrollPattern: ScrollBehavior{
			Speed:           "fast",
			Pattern:         "stepped",
			PauseFrequency:  0.1,
			ScrollBackProb:  0.1,
			BottomReachProb: 0.6,
		},
		ClickPattern: ClickBehavior{
			Precision:        "precise",
			DoubleClickProb:  0.1,
			RightClickProb:   0.05,
			MissClickProb:    0.02,
			HoverBeforeClick: false,
			HoverDuration:    100 * time.Millisecond,
		},
		ReadingPattern: ReadingBehavior{
			Speed:         "fast",
			SkimProb:      0.5,
			HighlightProb: 0.05,
			CopyProb:      0.1,
			ImageViewProb: 0.2,
		},
		NavigationPattern: NavigationBehavior{
			BackButtonProb: 0.1,
			NewTabProb:     0.4,
			BookmarkProb:   0.05,
			SearchProb:     0.3,
			MenuUsageProb:  0.1,
		},
		MousePattern: MouseBehavior{
			MovementStyle: "linear",
			Speed:         "fast",
			IdleMovement:  false,
			TrailNoise:    0.1,
		},
		Weight: 0.25,
	}
	
	// Researcher - Araştırmacı
	bcm.clusters["researcher"] = &BehaviorCluster{
		ID:          "researcher",
		Name:        "Researcher",
		Description: "Detaylı araştırma yapan kullanıcı",
		ScrollPattern: ScrollBehavior{
			Speed:           "medium",
			Pattern:         "smooth",
			PauseFrequency:  0.5,
			ScrollBackProb:  0.4,
			BottomReachProb: 0.8,
		},
		ClickPattern: ClickBehavior{
			Precision:        "precise",
			DoubleClickProb:  0.15,
			RightClickProb:   0.1,
			MissClickProb:    0.03,
			HoverBeforeClick: true,
			HoverDuration:    300 * time.Millisecond,
		},
		ReadingPattern: ReadingBehavior{
			Speed:         "medium",
			SkimProb:      0.3,
			HighlightProb: 0.3,
			CopyProb:      0.2,
			ImageViewProb: 0.5,
		},
		NavigationPattern: NavigationBehavior{
			BackButtonProb: 0.2,
			NewTabProb:     0.5,
			BookmarkProb:   0.1,
			SearchProb:     0.4,
			MenuUsageProb:  0.3,
		},
		MousePattern: MouseBehavior{
			MovementStyle: "curved",
			Speed:         "medium",
			IdleMovement:  true,
			TrailNoise:    0.2,
		},
		Weight: 0.2,
	}
	
	// Mobile User - Mobil kullanıcı
	bcm.clusters["mobile"] = &BehaviorCluster{
		ID:          "mobile",
		Name:        "Mobile User",
		Description: "Mobil cihaz kullanıcısı",
		ScrollPattern: ScrollBehavior{
			Speed:           "medium",
			Pattern:         "erratic",
			PauseFrequency:  0.3,
			ScrollBackProb:  0.15,
			BottomReachProb: 0.4,
		},
		ClickPattern: ClickBehavior{
			Precision:        "sloppy",
			DoubleClickProb:  0.02,
			RightClickProb:   0.0,
			MissClickProb:    0.12,
			HoverBeforeClick: false,
			HoverDuration:    0,
		},
		ReadingPattern: ReadingBehavior{
			Speed:         "medium",
			SkimProb:      0.4,
			HighlightProb: 0.02,
			CopyProb:      0.02,
			ImageViewProb: 0.3,
		},
		NavigationPattern: NavigationBehavior{
			BackButtonProb: 0.4,
			NewTabProb:     0.05,
			BookmarkProb:   0.01,
			SearchProb:     0.2,
			MenuUsageProb:  0.4,
		},
		MousePattern: MouseBehavior{
			MovementStyle: "erratic",
			Speed:         "medium",
			IdleMovement:  false,
			TrailNoise:    0.4,
		},
		Weight: 0.2,
	}
}

// SelectRandomCluster ağırlıklı rastgele cluster seçer
func (bcm *BehaviorClusterManager) SelectRandomCluster() *BehaviorCluster {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	
	// Toplam ağırlık
	var totalWeight float64
	for _, cluster := range bcm.clusters {
		totalWeight += cluster.Weight
	}
	
	// Rastgele seçim
	r := bcm.rng.Float64() * totalWeight
	var cumWeight float64
	
	for id, cluster := range bcm.clusters {
		cumWeight += cluster.Weight
		if r <= cumWeight {
			bcm.activeCluster = id
			bcm.sessionBehavior = cluster
			return cluster
		}
	}
	
	// Fallback
	for id, cluster := range bcm.clusters {
		bcm.activeCluster = id
		bcm.sessionBehavior = cluster
		return cluster
	}
	
	return nil
}

// GetActiveCluster aktif cluster'ı döner
func (bcm *BehaviorClusterManager) GetActiveCluster() *BehaviorCluster {
	bcm.mu.RLock()
	defer bcm.mu.RUnlock()
	
	return bcm.sessionBehavior
}

// GetCluster ID ile cluster döner
func (bcm *BehaviorClusterManager) GetCluster(id string) *BehaviorCluster {
	bcm.mu.RLock()
	defer bcm.mu.RUnlock()
	
	return bcm.clusters[id]
}

// AddCluster yeni cluster ekler
func (bcm *BehaviorClusterManager) AddCluster(cluster *BehaviorCluster) {
	bcm.mu.Lock()
	defer bcm.mu.Unlock()
	
	bcm.clusters[cluster.ID] = cluster
}

// GenerateScrollDelay scroll delay üretir
func (bcm *BehaviorClusterManager) GenerateScrollDelay() time.Duration {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return 100 * time.Millisecond
	}
	
	var baseDelay time.Duration
	switch cluster.ScrollPattern.Speed {
	case "slow":
		baseDelay = 150 * time.Millisecond
	case "medium":
		baseDelay = 80 * time.Millisecond
	case "fast":
		baseDelay = 30 * time.Millisecond
	default:
		baseDelay = 80 * time.Millisecond
	}
	
	// Jitter ekle
	jitter := time.Duration(bcm.rng.Float64() * float64(baseDelay) * 0.5)
	return baseDelay + jitter
}

// ShouldPauseScroll scroll sırasında durmalı mı
func (bcm *BehaviorClusterManager) ShouldPauseScroll() bool {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return false
	}
	
	return bcm.rng.Float64() < cluster.ScrollPattern.PauseFrequency
}

// ShouldScrollBack geri scroll yapmalı mı
func (bcm *BehaviorClusterManager) ShouldScrollBack() bool {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return false
	}
	
	return bcm.rng.Float64() < cluster.ScrollPattern.ScrollBackProb
}

// GenerateClickDelay tıklama öncesi delay
func (bcm *BehaviorClusterManager) GenerateClickDelay() time.Duration {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return 200 * time.Millisecond
	}
	
	if cluster.ClickPattern.HoverBeforeClick {
		return cluster.ClickPattern.HoverDuration + time.Duration(bcm.rng.Intn(200))*time.Millisecond
	}
	
	return time.Duration(50+bcm.rng.Intn(150)) * time.Millisecond
}

// ShouldMissClick yanlış tıklama yapmalı mı
func (bcm *BehaviorClusterManager) ShouldMissClick() bool {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return false
	}
	
	return bcm.rng.Float64() < cluster.ClickPattern.MissClickProb
}

// GenerateMousePath mouse hareket yolu üretir
func (bcm *BehaviorClusterManager) GenerateMousePath(startX, startY, endX, endY int) []struct{ X, Y int } {
	cluster := bcm.GetActiveCluster()
	if cluster == nil {
		return []struct{ X, Y int }{{endX, endY}}
	}
	
	var path []struct{ X, Y int }
	
	switch cluster.MousePattern.MovementStyle {
	case "linear":
		path = bcm.generateLinearPath(startX, startY, endX, endY)
	case "curved":
		path = bcm.generateCurvedPath(startX, startY, endX, endY)
	case "erratic":
		path = bcm.generateErraticPath(startX, startY, endX, endY)
	default:
		path = bcm.generateCurvedPath(startX, startY, endX, endY)
	}
	
	// Noise ekle
	if cluster.MousePattern.TrailNoise > 0 {
		path = bcm.addNoiseToPath(path, cluster.MousePattern.TrailNoise)
	}
	
	return path
}

// generateLinearPath doğrusal yol üretir
func (bcm *BehaviorClusterManager) generateLinearPath(startX, startY, endX, endY int) []struct{ X, Y int } {
	steps := 10 + bcm.rng.Intn(10)
	path := make([]struct{ X, Y int }, steps)
	
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		path[i] = struct{ X, Y int }{
			X: startX + int(float64(endX-startX)*t),
			Y: startY + int(float64(endY-startY)*t),
		}
	}
	
	return path
}

// generateCurvedPath eğri yol üretir (Bezier)
func (bcm *BehaviorClusterManager) generateCurvedPath(startX, startY, endX, endY int) []struct{ X, Y int } {
	steps := 15 + bcm.rng.Intn(10)
	path := make([]struct{ X, Y int }, steps)
	
	// Kontrol noktası
	ctrlX := (startX + endX) / 2 + bcm.rng.Intn(100) - 50
	ctrlY := (startY + endY) / 2 + bcm.rng.Intn(100) - 50
	
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		
		// Quadratic Bezier
		x := math.Pow(1-t, 2)*float64(startX) + 2*(1-t)*t*float64(ctrlX) + math.Pow(t, 2)*float64(endX)
		y := math.Pow(1-t, 2)*float64(startY) + 2*(1-t)*t*float64(ctrlY) + math.Pow(t, 2)*float64(endY)
		
		path[i] = struct{ X, Y int }{X: int(x), Y: int(y)}
	}
	
	return path
}

// generateErraticPath düzensiz yol üretir
func (bcm *BehaviorClusterManager) generateErraticPath(startX, startY, endX, endY int) []struct{ X, Y int } {
	steps := 20 + bcm.rng.Intn(15)
	path := make([]struct{ X, Y int }, steps)
	
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		
		// Base position
		baseX := startX + int(float64(endX-startX)*t)
		baseY := startY + int(float64(endY-startY)*t)
		
		// Random deviation
		devX := bcm.rng.Intn(40) - 20
		devY := bcm.rng.Intn(40) - 20
		
		// Deviation azalt sona doğru
		devX = int(float64(devX) * (1 - t))
		devY = int(float64(devY) * (1 - t))
		
		path[i] = struct{ X, Y int }{X: baseX + devX, Y: baseY + devY}
	}
	
	// Son nokta kesin hedef
	path[steps-1] = struct{ X, Y int }{X: endX, Y: endY}
	
	return path
}

// addNoiseToPath yola noise ekler
func (bcm *BehaviorClusterManager) addNoiseToPath(path []struct{ X, Y int }, noise float64) []struct{ X, Y int } {
	for i := 1; i < len(path)-1; i++ { // İlk ve son nokta hariç
		noiseX := int(float64(bcm.rng.Intn(10)-5) * noise)
		noiseY := int(float64(bcm.rng.Intn(10)-5) * noise)
		path[i].X += noiseX
		path[i].Y += noiseY
	}
	return path
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateClusterID cluster ID üretir
func generateClusterID(keyword string) string {
	hash := sha256.Sum256([]byte(keyword + time.Now().String()))
	return hex.EncodeToString(hash[:8])
}

// generateRandomHash rastgele hash üretir
func generateRandomHash() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateVariations keyword varyasyonları üretir
func generateVariations(keyword string) []string {
	variations := []string{}
	words := strings.Fields(keyword)
	
	if len(words) > 1 {
		// Kelime sırası değişimi
		for i := 0; i < len(words)-1; i++ {
			swapped := make([]string, len(words))
			copy(swapped, words)
			swapped[i], swapped[i+1] = swapped[i+1], swapped[i]
			variations = append(variations, strings.Join(swapped, " "))
		}
	}
	
	// Plural/singular
	if strings.HasSuffix(keyword, "s") {
		variations = append(variations, strings.TrimSuffix(keyword, "s"))
	} else {
		variations = append(variations, keyword+"s")
	}
	
	return variations
}

// generateLongTails long-tail keywordler üretir
func generateLongTails(keyword string, intent SearchIntent) []string {
	longTails := []string{}
	
	prefixes := map[SearchIntent][]string{
		IntentInformational: {"what is", "how to", "why", "when", "where", "guide to", "tutorial"},
		IntentNavigational:  {"official", "login", "website", "homepage"},
		IntentTransactional: {"buy", "order", "purchase", "get", "download", "subscribe"},
		IntentCommercial:    {"best", "top", "review", "compare", "vs", "alternative"},
		IntentLocal:         {"near me", "in my area", "local", "nearby"},
	}
	
	suffixes := map[SearchIntent][]string{
		IntentInformational: {"explained", "guide", "tutorial", "tips", "examples"},
		IntentNavigational:  {"site", "page", "portal"},
		IntentTransactional: {"online", "now", "today", "discount", "deal"},
		IntentCommercial:    {"2024", "review", "comparison", "pros and cons"},
		IntentLocal:         {"open now", "hours", "directions"},
	}
	
	// Intent'e göre prefix ekle
	if prefs, ok := prefixes[intent]; ok {
		for _, pref := range prefs[:min(3, len(prefs))] {
			longTails = append(longTails, pref+" "+keyword)
		}
	}
	
	// Intent'e göre suffix ekle
	if suffs, ok := suffixes[intent]; ok {
		for _, suff := range suffs[:min(3, len(suffs))] {
			longTails = append(longTails, keyword+" "+suff)
		}
	}
	
	return longTails
}

// getDefaultModifiers varsayılan modifier'ları döner
func getDefaultModifiers(intent SearchIntent) []string {
	modifiers := map[SearchIntent][]string{
		IntentInformational: {"free", "easy", "simple", "complete", "detailed"},
		IntentNavigational:  {"official", "main", "new"},
		IntentTransactional: {"cheap", "affordable", "discount", "sale", "free shipping"},
		IntentCommercial:    {"best", "top", "recommended", "popular", "rated"},
		IntentLocal:         {"nearby", "closest", "best rated"},
	}
	
	if mods, ok := modifiers[intent]; ok {
		return mods
	}
	
	return []string{"best", "top", "free"}
}

// contains string slice'da eleman var mı
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// min iki int'in minimumunu döner
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetClusterStats cluster istatistiklerini döner
func (kcm *KeywordClusterManager) GetClusterStats() map[string]interface{} {
	kcm.mu.RLock()
	defer kcm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	
	clusterStats := make([]map[string]interface{}, 0)
	for id, cluster := range kcm.clusters {
		clusterStats = append(clusterStats, map[string]interface{}{
			"id":          id,
			"name":        cluster.Name,
			"primary_kw":  cluster.PrimaryKW,
			"usage_count": cluster.UsageCount,
			"last_used":   cluster.LastUsed,
			"weight":      cluster.Weight,
			"intent":      cluster.Intent,
		})
	}
	
	stats["clusters"] = clusterStats
	stats["total_clusters"] = len(kcm.clusters)
	stats["total_usage"] = len(kcm.usageHistory)
	stats["rotation_strategy"] = kcm.rotationStrategy
	stats["pattern_detected"] = kcm.patternDetector.IsPatternDetected()
	stats["pattern_scores"] = kcm.patternDetector.GetPatternReport()
	
	return stats
}

// ExportClusters cluster'ları export eder
func (kcm *KeywordClusterManager) ExportClusters() []*KeywordCluster {
	kcm.mu.RLock()
	defer kcm.mu.RUnlock()
	
	clusters := make([]*KeywordCluster, 0, len(kcm.clusters))
	for _, cluster := range kcm.clusters {
		clusters = append(clusters, cluster)
	}
	
	// ID'ye göre sırala
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].ID < clusters[j].ID
	})
	
	return clusters
}

// ImportClusters cluster'ları import eder
func (kcm *KeywordClusterManager) ImportClusters(clusters []*KeywordCluster) {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()
	
	for _, cluster := range clusters {
		kcm.clusters[cluster.ID] = cluster
		kcm.rotationIndex[cluster.ID] = 0
	}
}

// ClearHistory kullanım geçmişini temizler
func (kcm *KeywordClusterManager) ClearHistory() {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()
	
	kcm.usageHistory = make([]KeywordUsageRecord, 0, kcm.maxHistorySize)
}

// ResetPatternDetector pattern detector'ı sıfırlar
func (kcm *KeywordClusterManager) ResetPatternDetector() {
	kcm.patternDetector = NewPatternDetector(50, 0.7)
}
