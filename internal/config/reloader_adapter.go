package config

import (
	"time"

	configpkg "eroshit/pkg/config"
)

// Reloader wraps pkg/config.Reloader for internal use
// This adapter allows seamless integration with the existing internal/config package
type Reloader struct {
	inner     *configpkg.Reloader
	callbacks []func(*Config)
}

// NewReloader creates a new config reloader that works with internal/config.Config
func NewReloader(configPath string) *Reloader {
	return &Reloader{
		inner:     configpkg.NewReloader(configPath),
		callbacks: make([]func(*Config), 0),
	}
}

// SetLogger sets a custom logger
func (r *Reloader) SetLogger(logger configpkg.Logger) {
	r.inner.SetLogger(logger)
}

// SetDebounceDelay sets the debounce delay
func (r *Reloader) SetDebounceDelay(delay time.Duration) {
	r.inner.SetDebounceDelay(delay)
}

// OnChange registers a callback for config changes
// The callback receives internal/config.Config type
func (r *Reloader) OnChange(callback func(*Config)) {
	r.callbacks = append(r.callbacks, callback)
	
	// Wrap the callback to convert config types
	r.inner.OnChange(func(pkgCfg *configpkg.Config) {
		cfg := convertFromPkgConfig(pkgCfg)
		callback(cfg)
	})
}

// Load loads the config from file
func (r *Reloader) Load() error {
	return r.inner.Load()
}

// Start starts watching the config file
func (r *Reloader) Start() error {
	return r.inner.Start()
}

// Stop stops watching the config file
func (r *Reloader) Stop() error {
	return r.inner.Stop()
}

// GetConfig returns the current config as internal/config.Config
func (r *Reloader) GetConfig() *Config {
	pkgCfg := r.inner.GetConfig()
	if pkgCfg == nil {
		return nil
	}
	return convertFromPkgConfig(pkgCfg)
}

// convertFromPkgConfig converts pkg/config.Config to internal/config.Config
func convertFromPkgConfig(pkgCfg *configpkg.Config) *Config {
	if pkgCfg == nil {
		return nil
	}
	
	cfg := &Config{
		TargetDomain:         pkgCfg.TargetDomain,
		MaxPages:             pkgCfg.MaxPages,
		DurationMinutes:      pkgCfg.DurationMinutes,
		HitsPerMinute:        pkgCfg.HitsPerMinute,
		DisableImages:        pkgCfg.DisableImages,
		DisableJSExecution:   pkgCfg.DisableJSExecution,
		ProxyEnabled:         pkgCfg.ProxyEnabled,
		ProxyHost:            pkgCfg.ProxyHost,
		ProxyPort:            pkgCfg.ProxyPort,
		ProxyUser:            pkgCfg.ProxyUser,
		ProxyPass:            pkgCfg.ProxyPass,
		ProxyURL:             pkgCfg.ProxyURL,
		ProxyBaseURL:         pkgCfg.ProxyBaseURL,
		GtagID:               pkgCfg.GtagID,
		LogLevel:             pkgCfg.LogLevel,
		ExportFormat:         pkgCfg.ExportFormat,
		OutputDir:            pkgCfg.OutputDir,
		MaxConcurrentVisits:  pkgCfg.MaxConcurrentVisits,
		CanvasFingerprint:    pkgCfg.CanvasFingerprint,
		ScrollStrategy:       pkgCfg.ScrollStrategy,
		SendScrollEvent:      pkgCfg.SendScrollEvent,
		UseSitemap:           pkgCfg.UseSitemap,
		SitemapHomepageWeight: pkgCfg.SitemapHomepageWeight,
		Keywords:             pkgCfg.Keywords,
		UsePublicProxy:       pkgCfg.UsePublicProxy,
		ProxySourceURLs:      pkgCfg.ProxySourceURLs,
		GitHubRepos:          pkgCfg.GitHubRepos,
		CheckerWorkers:       pkgCfg.CheckerWorkers,
		UsePrivateProxy:      pkgCfg.UsePrivateProxy,
		DeviceType:           pkgCfg.DeviceType,
		DeviceBrands:         pkgCfg.DeviceBrands,
		ReferrerKeyword:      pkgCfg.ReferrerKeyword,
		ReferrerEnabled:      pkgCfg.ReferrerEnabled,
		ReferrerSource:       pkgCfg.ReferrerSource,
		MinPageDuration:      pkgCfg.MinPageDuration,
		MaxPageDuration:      pkgCfg.MaxPageDuration,
		MinScrollPercent:     pkgCfg.MinScrollPercent,
		MaxScrollPercent:     pkgCfg.MaxScrollPercent,
		ClickProbability:     pkgCfg.ClickProbability,
		SessionMinPages:      pkgCfg.SessionMinPages,
		SessionMaxPages:      pkgCfg.SessionMaxPages,
		EnableSessionDepth:   pkgCfg.EnableSessionDepth,
		TargetBounceRate:     pkgCfg.TargetBounceRate,
		EnableBounceControl:  pkgCfg.EnableBounceControl,
		SimulateMouseMove:    pkgCfg.SimulateMouseMove,
		SimulateKeyboard:     pkgCfg.SimulateKeyboard,
		SimulateClicks:       pkgCfg.SimulateClicks,
		SimulateFocus:        pkgCfg.SimulateFocus,
		GeoCountry:           pkgCfg.GeoCountry,
		GeoTimezone:          pkgCfg.GeoTimezone,
		GeoLanguage:          pkgCfg.GeoLanguage,
		SendPageView:         pkgCfg.SendPageView,
		SendSessionStart:     pkgCfg.SendSessionStart,
		SendUserEngagement:   pkgCfg.SendUserEngagement,
		SendFirstVisit:       pkgCfg.SendFirstVisit,
		CustomDimensions:     pkgCfg.CustomDimensions,
		CustomMetrics:        pkgCfg.CustomMetrics,
		EnableCustomDimensions: pkgCfg.EnableCustomDimensions,
		GscPropertyUrl:       pkgCfg.GscPropertyUrl,
		GscApiKey:            pkgCfg.GscApiKey,
		EnableGscIntegration: pkgCfg.EnableGscIntegration,
		UseGscQueries:        pkgCfg.UseGscQueries,
		ReturningVisitorRate: pkgCfg.ReturningVisitorRate,
		ReturningVisitorDays: pkgCfg.ReturningVisitorDays,
		EnableReturningVisitor: pkgCfg.EnableReturningVisitor,
		ExitPages:            pkgCfg.ExitPages,
		EnableExitPageControl: pkgCfg.EnableExitPageControl,
		BrowserProfilePath:   pkgCfg.BrowserProfilePath,
		MaxBrowserProfiles:   pkgCfg.MaxBrowserProfiles,
		EnableBrowserProfile: pkgCfg.EnableBrowserProfile,
		PersistCookies:       pkgCfg.PersistCookies,
		PersistLocalStorage:  pkgCfg.PersistLocalStorage,
		TlsFingerprintMode:   pkgCfg.TlsFingerprintMode,
		EnableJa3Randomization: pkgCfg.EnableJa3Randomization,
		EnableJa4Randomization: pkgCfg.EnableJa4Randomization,
		ProxyRotationMode:    pkgCfg.ProxyRotationMode,
		ProxyRotationInterval: pkgCfg.ProxyRotationInterval,
		EnableProxyRotation:  pkgCfg.EnableProxyRotation,
		EnableHttp2Fingerprint: pkgCfg.EnableHttp2Fingerprint,
		EnableHttp3Fingerprint: pkgCfg.EnableHttp3Fingerprint,
		Http2FingerprintMode: pkgCfg.Http2FingerprintMode,
		EnableClientHints:    pkgCfg.EnableClientHints,
		SpoofSecChUa:         pkgCfg.SpoofSecChUa,
		SpoofSecChUaPlatform: pkgCfg.SpoofSecChUaPlatform,
		SpoofSecChUaArch:     pkgCfg.SpoofSecChUaArch,
		BypassPuppeteer:      pkgCfg.BypassPuppeteer,
		BypassPlaywright:     pkgCfg.BypassPlaywright,
		BypassSelenium:       pkgCfg.BypassSelenium,
		BypassCDP:            pkgCfg.BypassCDP,
		BypassRenderingDetection: pkgCfg.BypassRenderingDetection,
		BypassWebdriver:      pkgCfg.BypassWebdriver,
		SerpSearchEngine:     pkgCfg.SerpSearchEngine,
		EnableSerpSimulation: pkgCfg.EnableSerpSimulation,
		SerpScrollBeforeClick: pkgCfg.SerpScrollBeforeClick,
		BrowserPoolMin:       pkgCfg.BrowserPoolMin,
		BrowserPoolMax:       pkgCfg.BrowserPoolMax,
		EnableAutoScaling:    pkgCfg.EnableAutoScaling,
		WorkerQueueSize:      pkgCfg.WorkerQueueSize,
		EnablePriorityQueue:  pkgCfg.EnablePriorityQueue,
		EnableFailureRecovery: pkgCfg.EnableFailureRecovery,
		IosDeviceModel:       pkgCfg.IosDeviceModel,
		EnableIosSafari:      pkgCfg.EnableIosSafari,
		EnableIosHaptics:     pkgCfg.EnableIosHaptics,
		AndroidDeviceModel:   pkgCfg.AndroidDeviceModel,
		EnableAndroidChrome:  pkgCfg.EnableAndroidChrome,
		EnableAndroidVibration: pkgCfg.EnableAndroidVibration,
		EnableTouchEvents:    pkgCfg.EnableTouchEvents,
		EnableMultiTouch:     pkgCfg.EnableMultiTouch,
		EnableGestures:       pkgCfg.EnableGestures,
		EnableMobileKeyboard: pkgCfg.EnableMobileKeyboard,
		EnableAccelerometer:  pkgCfg.EnableAccelerometer,
		EnableGyroscope:      pkgCfg.EnableGyroscope,
		EnableDeviceOrientation: pkgCfg.EnableDeviceOrientation,
		EnableMagnetometer:   pkgCfg.EnableMagnetometer,
		EnableTabletMode:     pkgCfg.EnableTabletMode,
		EnableLandscapeMode:  pkgCfg.EnableLandscapeMode,
		EnablePenInput:       pkgCfg.EnablePenInput,
		EnableSplitView:      pkgCfg.EnableSplitView,
		StealthWebdriver:     pkgCfg.StealthWebdriver,
		StealthChrome:        pkgCfg.StealthChrome,
		StealthPlugins:       pkgCfg.StealthPlugins,
		StealthWebGL:         pkgCfg.StealthWebGL,
		StealthAudio:         pkgCfg.StealthAudio,
		StealthCanvas:        pkgCfg.StealthCanvas,
		StealthTimezone:      pkgCfg.StealthTimezone,
		StealthLanguage:      pkgCfg.StealthLanguage,
		VisitTimeout:         pkgCfg.VisitTimeout,
		PageLoadWait:         pkgCfg.PageLoadWait,
		RetryCount:           pkgCfg.RetryCount,
		BlockImages:          pkgCfg.BlockImages,
		BlockStyles:          pkgCfg.BlockStyles,
		BlockFonts:           pkgCfg.BlockFonts,
		BlockMedia:           pkgCfg.BlockMedia,
		AntiDetectMode:       pkgCfg.AntiDetectMode,
		BrowserPoolEnabled:   pkgCfg.BrowserPoolEnabled,
		BrowserPoolSize:      pkgCfg.BrowserPoolSize,
		BrowserPoolMinSize:   pkgCfg.BrowserPoolMinSize,
		BrowserMaxSessions:   pkgCfg.BrowserMaxSessions,
		BrowserMaxAge:        pkgCfg.BrowserMaxAge,
		ConnectionTimeout:    pkgCfg.ConnectionTimeout,
		MaxIdleConnections:   pkgCfg.MaxIdleConnections,
		EnableKeepAlive:      pkgCfg.EnableKeepAlive,
		EnableBufferPool:     pkgCfg.EnableBufferPool,
		MaxMemoryMB:          pkgCfg.MaxMemoryMB,
		Duration:             pkgCfg.Duration,
		RequestInterval:      pkgCfg.RequestInterval,
	}
	
	// Convert PrivateProxies
	if len(pkgCfg.PrivateProxies) > 0 {
		cfg.PrivateProxies = make([]PrivateProxy, len(pkgCfg.PrivateProxies))
		for i, pp := range pkgCfg.PrivateProxies {
			cfg.PrivateProxies[i] = PrivateProxy{
				Host:     pp.Host,
				Port:     pp.Port,
				User:     pp.User,
				Pass:     pp.Pass,
				Protocol: pp.Protocol,
			}
		}
	}
	
	return cfg
}
