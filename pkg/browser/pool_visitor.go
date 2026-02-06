// Package browser provides pool-based visitor for high-performance hit generation
// This is the integration layer between the pool and the existing HitVisitor
package browser



import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"eroshit/pkg/behavior"
	"eroshit/pkg/canvas"
	"eroshit/pkg/engagement"
	"eroshit/pkg/fingerprint"
	"eroshit/pkg/mobile"
	"eroshit/pkg/referrer"
	"eroshit/pkg/stealth"
	"eroshit/pkg/useragent"
)

// PooledHitVisitor is a high-performance visitor that uses browser pool
// instead of creating new Chrome instances for each visit
type PooledHitVisitor struct {
	pool             *BrowserPool
	agentProvider    interface{ RandomWithHeaders() (ua string, headers map[string]string) }
	defaultUserAgent string
	profilePool      *behavior.ProfilePool
	mixedModeOpts    *MixedModeOptions
	mu               sync.RWMutex
}

// PooledHitVisitorConfig configuration for pooled visitor
type PooledHitVisitorConfig struct {
	// Pool configuration
	PoolConfig PoolConfig
	
	// Agent provider for User-Agent selection
	AgentProvider interface{ RandomWithHeaders() (ua string, headers map[string]string) }
	
	// Default User-Agent if agent provider fails
	DefaultUserAgent string
}

// NewPooledHitVisitor creates a new pooled visitor with the given configuration
func NewPooledHitVisitor(config PooledHitVisitorConfig) (*PooledHitVisitor, error) {
	// Create pool
	pool, err := NewBrowserPool(config.PoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser pool: %w", err)
	}

	visitor := &PooledHitVisitor{
		pool:             pool,
		agentProvider:    config.AgentProvider,
		defaultUserAgent: config.DefaultUserAgent,
		profilePool:      nil,
		mixedModeOpts:    nil,
	}

	return visitor, nil
}

// SetMixedMode mixed mode'u etkinleştirir
func (v *PooledHitVisitor) SetMixedMode(opts MixedModeOptions) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.mixedModeOpts = &opts
	
	if opts.EnableMixedMode && opts.ProfilePoolSize > 0 {
		v.profilePool = behavior.NewProfilePool(opts.ProfilePoolSize)
	}
}

// GenerateRandomProfile rastgele davranış profili oluşturur
func (v *PooledHitVisitor) GenerateRandomProfile() *behavior.BehavioralProfile {
	return behavior.GenerateRandomProfile()
}

// GetNextMixedProfile mixed mode için sıradaki profili döner
func (v *PooledHitVisitor) GetNextMixedProfile() *behavior.BehavioralProfile {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	if v.profilePool != nil {
		return v.profilePool.GetRandom()
	}
	
	// Profil havuzu yoksa yeni profil oluştur
	return behavior.GenerateRandomProfile()
}

// Close shuts down the visitor and its pool
func (v *PooledHitVisitor) Close() error {
	if v.pool != nil {
		return v.pool.Close()
	}
	return nil
}

// VisitOptions options for a single visit
type VisitOptions struct {
	URL               string
	ProxyURL          string
	ProxyUser         string
	ProxyPass         string
	UserAgent         string
	DeviceType        string
	DeviceBrands      []string
	CanvasFingerprint bool
	ScrollStrategy    string
	ReferrerKeyword   string
	ReferrerEnabled   bool
	Keywords          []string
	VisitTimeout      time.Duration
	GtagID            string
	BehavioralProfile *behavior.BehavioralProfile // Davranış profili (opsiyonel)
}

// MixedModeOptions mixed mode için ziyaret seçenekleri
type MixedModeOptions struct {
	EnableMixedMode   bool                          // Mixed mode aktif mi
	ProfilePoolSize   int                           // Profil havuzu boyutu
	BehavioralMix     map[behavior.ProfileType]float64 // Profil tipi dağılımı
	BehavioralProfile *behavior.BehavioralProfile   // Mevcut ziyaret için profil
}

// VisitURL performs a visit using a pooled browser instance
func (v *PooledHitVisitor) VisitURL(ctx context.Context, opts VisitOptions) error {
	// Set default timeout
	timeout := opts.VisitTimeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}

	visitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Acquire browser instance from pool
	instance, err := v.pool.Acquire(visitCtx)
	if err != nil {
		return fmt.Errorf("failed to acquire browser: %w", err)
	}
	defer v.pool.Release(instance)

	// Get User-Agent
	ua := opts.UserAgent
	if ua == "" && v.agentProvider != nil {
		ua, _ = v.agentProvider.RandomWithHeaders()
	}
	if ua == "" {
		ua = v.defaultUserAgent
	}
	if ua == "" {
		ua = useragent.Random()
	}

	// Get device profile if needed
	var deviceProfile *mobile.DeviceProfile
	var isMobile bool
	
	if opts.DeviceType != "" && opts.DeviceType != "mixed" {
		device := mobile.GetRandomDeviceFiltered(opts.DeviceType, opts.DeviceBrands)
		deviceProfile = &device
		ua = device.UserAgent
		isMobile = device.Mobile
	} else if len(opts.DeviceBrands) > 0 {
		device := mobile.GetRandomDeviceFiltered("mixed", opts.DeviceBrands)
		deviceProfile = &device
		ua = device.UserAgent
		isMobile = device.Mobile
	}

	// Build fingerprint
	advFP := buildFingerprint(deviceProfile, ua)
	
	// Build stealth config
	stealthCfg := buildStealthConfig(advFP, ua)

	// Get tab context
	tabCtx := instance.GetContext()

	// Setup resource blocking
	blockTypes := map[network.ResourceType]bool{
		network.ResourceTypeImage:      true,
		network.ResourceTypeStylesheet: true,
		network.ResourceTypeFont:       true,
		network.ResourceTypeMedia:      true,
	}

	// Setup fetch listener for proxy auth and resource blocking
	setupFetchListener(tabCtx, opts, blockTypes)

	// Build navigation actions
	actions := buildNavigationActions(opts, stealthCfg, advFP, ua, isMobile)

	// Execute navigation
	start := time.Now()
	if err := chromedp.Run(tabCtx, actions...); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Inject gtag if needed
	if opts.GtagID != "" {
		injectGtag(tabCtx, opts.GtagID)
	}

	// Apply post-load actions
	// Mixed mode: profil atanmamışsa havuzdan al
	if opts.BehavioralProfile == nil && v.mixedModeOpts != nil && v.mixedModeOpts.EnableMixedMode {
		opts.BehavioralProfile = v.GetNextMixedProfile()
	}
	
	if err := applyPostLoadActions(tabCtx, opts, advFP); err != nil {
		// Post-load errors are not critical
		_ = err
	}

	elapsed := time.Since(start)
	_ = elapsed

	return nil
}

// GetMetrics returns pool metrics
func (v *PooledHitVisitor) GetMetrics() PoolMetrics {
	if v.pool != nil {
		return v.pool.GetMetrics()
	}
	return PoolMetrics{}
}

// ForceResetPool forces deep reset of all idle instances
func (v *PooledHitVisitor) ForceResetPool() {
	// This is a no-op as reset happens automatically on Release
	// but we could add explicit pool reset if needed
}

// buildFingerprint creates fingerprint from device profile or generates new one
func buildFingerprint(deviceProfile *mobile.DeviceProfile, ua string) *fingerprint.AdvancedFingerprint {
	if deviceProfile != nil {
		return &fingerprint.AdvancedFingerprint{
			UserAgent:           deviceProfile.UserAgent,
			Platform:            deviceProfile.Platform,
			ScreenWidth:         deviceProfile.ScreenWidth,
			ScreenHeight:        deviceProfile.ScreenHeight,
			ScreenPixelRatio:    deviceProfile.PixelRatio,
			MaxTouchPoints:      deviceProfile.MaxTouchPoints,
			Language:            "tr-TR",
			Languages:           []string{"tr-TR", "tr", "en"},
			HardwareConcurrency: 8,
			DeviceMemory:        8,
			ScreenColorDepth:    24,
			AvailWidth:          deviceProfile.ScreenWidth,
			AvailHeight:         deviceProfile.ScreenHeight - 40,
			Timezone:            "Europe/Istanbul",
			WebGLVendor:         "Google Inc.",
			WebGLRenderer:       "ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0)",
		}
	}

	advFP := fingerprint.GenerateAdvancedFingerprint()
	advFP.UserAgent = ua
	return advFP
}

// buildStealthConfig creates stealth config from fingerprint
func buildStealthConfig(advFP *fingerprint.AdvancedFingerprint, ua string) stealth.StealthConfig {
	cfg := stealth.StealthConfig{
		UserAgent:           ua,
		Platform:            advFP.Platform,
		Vendor:              advFP.WebGLVendor,
		WebGLVendor:         advFP.WebGLVendor,
		WebGLRenderer:       advFP.WebGLRenderer,
		Languages:           advFP.Languages,
		Plugins:             stealth.GetDefaultStealthConfig().Plugins,
		ScreenWidth:         advFP.ScreenWidth,
		ScreenHeight:        advFP.ScreenHeight,
		AvailWidth:          advFP.AvailWidth,
		AvailHeight:         advFP.AvailHeight,
		ColorDepth:          advFP.ScreenColorDepth,
		PixelDepth:          advFP.ScreenColorDepth,
		HardwareConcurrency: advFP.HardwareConcurrency,
		DeviceMemory:        int(advFP.DeviceMemory),
	}

	// Validate dimensions
	if cfg.ScreenWidth <= 0 {
		cfg.ScreenWidth = 1920
	}
	if cfg.ScreenHeight <= 0 {
		cfg.ScreenHeight = 1080
	}
	if cfg.AvailWidth <= 0 {
		cfg.AvailWidth = cfg.ScreenWidth
	}
	if cfg.AvailHeight <= 0 {
		cfg.AvailHeight = cfg.ScreenHeight - 40
	}

	return cfg
}

// setupFetchListener sets up CDP fetch listener for resource blocking and proxy auth
func setupFetchListener(tabCtx context.Context, opts VisitOptions, blockTypes map[network.ResourceType]bool) {
	// Proxy auth handling
	if opts.ProxyUser != "" || opts.ProxyPass != "" {
		chromedp.ListenTarget(tabCtx, func(ev interface{}) {
			if ev, ok := ev.(*fetch.EventAuthRequired); ok && ev.AuthChallenge.Source == fetch.AuthChallengeSourceProxy {
				go func() {
					_ = chromedp.Run(tabCtx,
						fetch.ContinueWithAuth(ev.RequestID, &fetch.AuthChallengeResponse{
							Response: fetch.AuthChallengeResponseResponseProvideCredentials,
							Username: opts.ProxyUser,
							Password: opts.ProxyPass,
						}),
					)
				}()
			}
		})
	}

	// Resource blocking
	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if ev, ok := ev.(*fetch.EventRequestPaused); ok {
			go func() {
				rt := ev.ResourceType
				if rt == network.ResourceTypeDocument || rt == network.ResourceTypeScript || rt == "" {
					_ = chromedp.Run(tabCtx, fetch.ContinueRequest(ev.RequestID))
					return
				}
				if blockTypes[rt] {
					_ = chromedp.Run(tabCtx, fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient))
				} else {
					_ = chromedp.Run(tabCtx, fetch.ContinueRequest(ev.RequestID))
				}
			}()
		}
	})
}

// buildNavigationActions builds chromedp actions for page navigation
func buildNavigationActions(opts VisitOptions, stealthCfg stealth.StealthConfig, 
	advFP *fingerprint.AdvancedFingerprint, ua string, isMobile bool) []chromedp.Action {

	var fp fingerprint.FP
	if advFP != nil {
		fp = fingerprint.FP{
			Platform:     advFP.Platform,
			Language:     advFP.Language,
			Languages:    strings.Join(advFP.Languages, ", "),
			InnerW:       advFP.ScreenWidth,
			InnerH:       advFP.ScreenHeight,
			DevicePixel:  advFP.ScreenPixelRatio,
			Timezone:     advFP.Timezone,
			HardwareConc: advFP.HardwareConcurrency,
			DeviceMem:    int64(advFP.DeviceMemory),
			Vendor:       advFP.WebGLVendor,
		}
	}

	if fp.InnerW <= 0 {
		fp.InnerW = 1366
	}
	if fp.InnerH <= 0 {
		fp.InnerH = 768
	}

	// Build fetch options
	fetchOpt := fetch.Enable()
	if opts.ProxyUser != "" || opts.ProxyPass != "" {
		fetchOpt = fetch.Enable().WithHandleAuthRequests(true)
	}

	// Stealth script
	stealthScript := stealth.GetOnNewDocumentScript(stealthCfg)

	// Build actions
	actions := []chromedp.Action{
		fetchOpt,
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(stealthScript).Do(ctx)
			return err
		}),
		emulation.SetUserAgentOverride(ua),
		emulation.SetDeviceMetricsOverride(int64(fp.InnerW), int64(fp.InnerH), fp.DevicePixel, isMobile),
		emulation.SetTimezoneOverride(fp.Timezone),
	}

	// Mobile touch emulation
	if isMobile && advFP != nil {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetTouchEmulationEnabled(true).
				WithMaxTouchPoints(int64(advFP.MaxTouchPoints)).Do(ctx)
		}))
	}

	// Determine referrer
	var referrerURL string
	if opts.ReferrerEnabled && opts.ReferrerKeyword != "" {
		encodedKeyword := url.QueryEscape(opts.ReferrerKeyword)
		referrerURL = fmt.Sprintf("https://www.google.com/search?q=%s", encodedKeyword)
	} else if len(opts.Keywords) > 0 {
		targetDomain := opts.URL
		if idx := strings.Index(opts.URL, "://"); idx >= 0 {
			targetDomain = opts.URL[idx+3:]
		}
		if idx := strings.Index(targetDomain, "/"); idx >= 0 {
			targetDomain = targetDomain[:idx]
		}

		refCfg := &referrer.ReferrerConfig{
			GooglePercent: 50, BingPercent: 20, DirectPercent: 30,
			Keywords: opts.Keywords,
		}
		refChain := referrer.NewReferrerChain(targetDomain, refCfg)
		src := refChain.Generate()
		if src != nil && src.URL != "" && (src.Type == "search" || src.Type == "social") {
			referrerURL = src.URL
		}
	}

	// Navigate with or without referrer
	if referrerURL != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, _, err := page.Navigate(opts.URL).WithReferrer(referrerURL).Do(ctx)
			return err
		}))
	} else {
		actions = append(actions, chromedp.Navigate(opts.URL))
	}

	// Wait for page load
	actions = append(actions,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
	)

	return actions
}

// injectGtag injects Google Analytics gtag script
func injectGtag(tabCtx context.Context, gtagID string) {
	gtagScript := `(function(){
		var s=document.createElement('script');s.async=true;
		s.src='https://www.googletagmanager.com/gtag/js?id=` + gtagID + `';
		document.head.appendChild(s);
		window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}
		gtag('js',new Date());
		gtag('config','` + gtagID + `',{send_page_view:true});
	})();`

	chromedp.Run(tabCtx,
		chromedp.Evaluate(gtagScript, nil),
		chromedp.Sleep(1500*time.Millisecond),
	)
}

// applyPostLoadActions applies canvas fingerprint, scrolling and behavior simulation
func applyPostLoadActions(tabCtx context.Context, opts VisitOptions, advFP *fingerprint.AdvancedFingerprint) error {
	// Stealth scripts
	stealthCfg := buildStealthConfig(advFP, advFP.UserAgent)
	if err := stealth.InjectStealthScripts(tabCtx, stealthCfg); err != nil {
		// Non-critical
		_ = err
	}

	// Canvas fingerprint
	if opts.CanvasFingerprint {
		cf := canvas.GenerateFingerprint()
		cf.InjectCanvasNoise(tabCtx)
		cf.InjectWebGLFingerprint(tabCtx)
		cf.InjectAudioFingerprint(tabCtx)
	}

	// Scroll behavior
	strategy := opts.ScrollStrategy
	if strategy == "" {
		strategy = "gradual"
	}
	
	// Profil bazlı scroll hızı
	readSpeed := 200
	if opts.BehavioralProfile != nil {
		readSpeed = opts.BehavioralProfile.ReadingSpeed
	}
	
	engagement.HumanScroll(tabCtx, engagement.ScrollBehavior{
		Strategy:  strategy,
		ReadSpeed: readSpeed,
	})

	// Human behavior simulation with profile
	var hum *behavior.HumanBehavior
	
	if opts.BehavioralProfile != nil {
		// Belirtilen profili kullan
		hum = behavior.NewHumanBehaviorWithProfile(opts.BehavioralProfile)
	} else {
		// Varsayılan davranış
		hum = behavior.NewHumanBehavior(&behavior.BehaviorConfig{
			MinPageDuration:      1 * time.Second,
			MaxPageDuration:      3 * time.Second,
			ScrollProbability:    0.5,
			MouseMoveProbability: 0.5,
			ClickProbability:     0,
		})
	}

	var pageLen int
	chromedp.Evaluate(`document.body ? document.body.innerText.length : 0`, &pageLen).Do(tabCtx)
	hum.SimulatePageVisit(tabCtx, pageLen)

	return nil
}
