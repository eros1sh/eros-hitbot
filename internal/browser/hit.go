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

	"eroshit/internal/reporter"
	"eroshit/pkg/analytics"
	"eroshit/pkg/behavior"
	"eroshit/pkg/canvas"
	"eroshit/pkg/engagement"
	"eroshit/pkg/fingerprint"
	"eroshit/pkg/mobile"
	"eroshit/pkg/referrer"
	"eroshit/pkg/stealth"
	"eroshit/pkg/useragent"
)

// Varsayılan ziyaret süresi; yavaş sayfalar ve yüksek paralellik için yeterli süre.
const defaultVisitTimeout = 90 * time.Second

type HitVisitorConfig struct {
	ProxyURL          string
	ProxyUser         string
	ProxyPass         string
	GtagID            string
	CanvasFingerprint bool   // canvas/webgl/audio noise
	ScrollStrategy    string   // "gradual","fast","reader"
	SendScrollEvent   bool     // GA4 scroll %75 event
	AnalyticsManager  *analytics.Manager
	Keywords          []string // Arama referrer için anahtar kelimeler
	VisitTimeout      time.Duration // 0 ise defaultVisitTimeout kullanılır
	// Cihaz emülasyonu
	DeviceType        string   // "desktop", "mobile", "tablet", "mixed"
	DeviceBrands      []string // ["apple", "samsung", "google", "windows", "linux"]
	// Referrer ayarları
	ReferrerKeyword   string   // Google arama referrer için kelime
	ReferrerEnabled   bool     // Referrer simülasyonu aktif mi
}

// HitVisitor JS çalıştıran, her ziyarette farklı fingerprint, proxy destekli
type HitVisitor struct {
	agentProvider interface {
		RandomWithHeaders() (ua string, headers map[string]string)
	}
	reporter *reporter.Reporter
	config   HitVisitorConfig
	allocCtx context.Context
	allocCan context.CancelFunc
	mu       sync.Mutex
}

func NewHitVisitor(agentProvider interface {
	RandomWithHeaders() (ua string, headers map[string]string)
}, rep *reporter.Reporter, cfg HitVisitorConfig) (*HitVisitor, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		// Headless bypass - KRITIK
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process,TranslateUI"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-sync", true),
	)

	// SECURITY FIX: Proxy URL'den auth bilgisini ayır
	// Chrome --proxy-server flag'i auth bilgisi kabul etmiyor
	// Auth bilgisi fetch.EventAuthRequired ile ayrı olarak işlenmeli
	if cfg.ProxyURL != "" {
		proxyServerURL := cfg.ProxyURL
		
		// URL'yi parse et ve auth bilgisini çıkar
		if parsedURL, err := url.Parse(cfg.ProxyURL); err == nil {
			// Auth bilgisi varsa, URL'den çıkar
			if parsedURL.User != nil {
				// Kullanıcı adı ve şifreyi config'e kaydet (zaten varsa override etme)
				if cfg.ProxyUser == "" {
					cfg.ProxyUser = parsedURL.User.Username()
				}
				if cfg.ProxyPass == "" {
					if pass, ok := parsedURL.User.Password(); ok {
						cfg.ProxyPass = pass
					}
				}
				// Auth olmadan proxy URL oluştur
				proxyServerURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
			}
		}
		
		opts = append(opts,
			chromedp.ProxyServer(proxyServerURL),
			chromedp.Flag("proxy-bypass-list", "<-loopback>"),
		)
	}

	allocCtx, allocCan := chromedp.NewExecAllocator(context.Background(), opts...)

	return &HitVisitor{
		agentProvider: agentProvider,
		reporter:      rep,
		config:        cfg,
		allocCtx:      allocCtx,
		allocCan:      allocCan,
	}, nil
}

func (h *HitVisitor) Close() {
	h.allocCan()
}

var blockTypes = map[network.ResourceType]bool{
	network.ResourceTypeImage:      true,
	network.ResourceTypeStylesheet: true,
	network.ResourceTypeFont:       true,
	network.ResourceTypeMedia:      true,
}

func (h *HitVisitor) VisitURL(ctx context.Context, urlStr string) error {
	// Cihaz emülasyonu: DeviceType ve DeviceBrands'e göre cihaz seç
	var deviceProfile *mobile.DeviceProfile
	var ua string
	var isMobile bool
	
	if h.config.DeviceType != "" && h.config.DeviceType != "mixed" {
		// Belirli bir cihaz tipi seçilmiş
		device := mobile.GetRandomDeviceFiltered(h.config.DeviceType, h.config.DeviceBrands)
		deviceProfile = &device
		ua = device.UserAgent
		isMobile = device.Mobile
	} else if len(h.config.DeviceBrands) > 0 {
		// Sadece marka filtresi var
		device := mobile.GetRandomDeviceFiltered("mixed", h.config.DeviceBrands)
		deviceProfile = &device
		ua = device.UserAgent
		isMobile = device.Mobile
	} else {
		// Varsayılan davranış
		ua, _ = h.agentProvider.RandomWithHeaders()
		if ua == "" {
			ua = useragent.Random()
		}
	}
	
	var advFP *fingerprint.AdvancedFingerprint
	var fp fingerprint.FP
	
	if deviceProfile != nil {
		// Cihaz profilinden fingerprint oluştur
		advFP = &fingerprint.AdvancedFingerprint{
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
	} else {
		// Varsayılan fingerprint
		advFP = fingerprint.GenerateAdvancedFingerprint()
		advFP.UserAgent = ua
		fp = fingerprint.FP{
			Platform:     advFP.Platform,
			Language:     advFP.Language,
			Languages:    strings.Join(advFP.Languages, ", "),
			InnerW:       advFP.ScreenWidth - 22,
			InnerH:       advFP.ScreenHeight - 100,
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

	// Stealth config (headless bypass) - fingerprint değerleriyle
	stealthCfg := stealth.StealthConfig{
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
	if stealthCfg.ScreenWidth <= 0 {
		stealthCfg.ScreenWidth = 1920
	}
	if stealthCfg.ScreenHeight <= 0 {
		stealthCfg.ScreenHeight = 1080
	}
	if stealthCfg.AvailWidth <= 0 {
		stealthCfg.AvailWidth = stealthCfg.ScreenWidth
	}
	if stealthCfg.AvailHeight <= 0 {
		stealthCfg.AvailHeight = stealthCfg.ScreenHeight - 40
	}
	
	// Mobil cihaz için touch desteği
	_ = isMobile // Kullanılacak

	browserOpts := []chromedp.ContextOption{
		chromedp.WithLogf(func(string, ...interface{}) {}),
	}

	tabCtx, tabCancel := chromedp.NewContext(h.allocCtx, browserOpts...)
	defer tabCancel()

	visitTimeout := h.config.VisitTimeout
	if visitTimeout <= 0 {
		visitTimeout = defaultVisitTimeout
	}
	tabCtx, tabCancel2 := context.WithTimeout(tabCtx, visitTimeout)
	defer tabCancel2()

	start := time.Now()
	authDone := make(chan struct{})

	// Proxy auth (proxy kullanıcı/şifre varsa)
	if h.config.ProxyUser != "" || h.config.ProxyPass != "" {
		chromedp.ListenTarget(tabCtx, func(ev interface{}) {
			if ev, ok := ev.(*fetch.EventAuthRequired); ok && ev.AuthChallenge.Source == fetch.AuthChallengeSourceProxy {
				go func() {
					_ = chromedp.Run(tabCtx,
						fetch.ContinueWithAuth(ev.RequestID, &fetch.AuthChallengeResponse{
							Response: fetch.AuthChallengeResponseResponseProvideCredentials,
							Username: h.config.ProxyUser,
							Password: h.config.ProxyPass,
						}),
					)
					select {
					case authDone <- struct{}{}:
					default:
					}
				}()
			}
		})
	}

	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if ev, ok := ev.(*fetch.EventRequestPaused); ok {
			go func() {
				rt := ev.ResourceType
				// Ana belge ve script asla bloklama; yanlışlıkla sayfa yükünü kırma
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

	fetchOpt := fetch.Enable()
	if h.config.ProxyUser != "" || h.config.ProxyPass != "" {
		fetchOpt = fetch.Enable().WithHandleAuthRequests(true)
	}

	gtagScript := ""
	if h.config.GtagID != "" {
		gtagScript = `(function(){
			var s=document.createElement('script');s.async=true;
			s.src='https://www.googletagmanager.com/gtag/js?id=` + h.config.GtagID + `';
			document.head.appendChild(s);
			window.dataLayer=window.dataLayer||[];function gtag(){dataLayer.push(arguments);}
			gtag('js',new Date());
			gtag('config','` + h.config.GtagID + `',{send_page_view:true});
		})();`
	}

	// Stealth script - sayfa yüklenmeden ÖNCE (headless bypass)
	stealthScript := stealth.GetOnNewDocumentScript(stealthCfg)

	// Referrer (keyword tabanlı arama kaynağı) - hedef domain'den çıkar
	targetDomain := urlStr
	if idx := strings.Index(urlStr, "://"); idx >= 0 {
		targetDomain = urlStr[idx+3:]
	}
	if idx := strings.Index(targetDomain, "/"); idx >= 0 {
		targetDomain = targetDomain[:idx]
	}
	navActions := []chromedp.Action{
		fetchOpt,
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(stealthScript).Do(ctx)
			return err
		}),
		emulation.SetUserAgentOverride(ua),
		emulation.SetDeviceMetricsOverride(int64(fp.InnerW), int64(fp.InnerH), fp.DevicePixel, isMobile),
		emulation.SetTimezoneOverride(fp.Timezone),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.ClearBrowserCookies().Do(ctx)
		}),
	}
	
	// Mobil cihaz için touch emülasyonu
	if isMobile && deviceProfile != nil {
		navActions = append(navActions, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetTouchEmulationEnabled(true).WithMaxTouchPoints(int64(deviceProfile.MaxTouchPoints)).Do(ctx)
		}))
	}
	
	// Referrer ayarla - öncelik: ReferrerKeyword > Keywords
	var referrerURL string
	if h.config.ReferrerEnabled && h.config.ReferrerKeyword != "" {
		// Kullanıcının girdiği kelime ile Google arama referrer'ı oluştur
		encodedKeyword := url.QueryEscape(h.config.ReferrerKeyword)
		referrerURL = fmt.Sprintf("https://www.google.com/search?q=%s", encodedKeyword)
	} else if len(h.config.Keywords) > 0 {
		// Eski davranış: Keywords listesinden referrer oluştur
		refCfg := &referrer.ReferrerConfig{
			GooglePercent: 50, BingPercent: 20, DirectPercent: 30,
			Keywords: h.config.Keywords,
		}
		refChain := referrer.NewReferrerChain(targetDomain, refCfg)
		src := refChain.Generate()
		if src != nil && src.URL != "" && (src.Type == "search" || src.Type == "social") {
			referrerURL = src.URL
		}
	}
	
	// Referrer'ı page.Navigate ile birlikte ayarla (SetExtraHTTPHeaders yerine)
	// Bu şekilde sadece ana sayfa navigasyonuna referrer eklenir, alt kaynaklara değil
	if referrerURL != "" {
		navActions = append(navActions,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// page.Navigate ile referrer parametresi kullan
				_, _, _, err := page.Navigate(urlStr).WithReferrer(referrerURL).Do(ctx)
				return err
			}),
		)
	} else {
		navActions = append(navActions,
			chromedp.Navigate(urlStr),
		)
	}
	navActions = append(navActions,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
	)
	navErr := chromedp.Run(tabCtx, navActions...)

	if navErr == nil && gtagScript != "" {
		if err := chromedp.Run(tabCtx, chromedp.Evaluate(gtagScript, nil)); err != nil {
			// gtag script hatası kritik değil, devam et
			_ = err
		}
		if err := chromedp.Run(tabCtx, chromedp.Sleep(1500*time.Millisecond)); err != nil {
			_ = err
		}
	}

	// Stealth scriptleri sayfa yüklendikten sonra tekrar enjekte et (bazı siteler için gerekli)
	if navErr == nil {
		if err := stealth.InjectStealthScripts(tabCtx, stealthCfg); err != nil {
			// Stealth script hatası kritik değil, devam et
			_ = err
		}
	}

	if navErr == nil {
		// Canvas fingerprint (sayfa yüklendikten sonra)
		if h.config.CanvasFingerprint {
			cf := canvas.GenerateFingerprint()
			if err := cf.InjectCanvasNoise(tabCtx); err != nil {
				_ = err
			}
			if err := cf.InjectWebGLFingerprint(tabCtx); err != nil {
				_ = err
			}
			if err := cf.InjectAudioFingerprint(tabCtx); err != nil {
				_ = err
			}
		}

		// Scroll davranışı
		strategy := h.config.ScrollStrategy
		if strategy == "" {
			strategy = "gradual"
		}
		if err := engagement.HumanScroll(tabCtx, engagement.ScrollBehavior{
			Strategy:    strategy,
			ReadSpeed:   200,
		}); err != nil {
			_ = err
		}

		// Scroll event (GA4)
		if h.config.SendScrollEvent && h.config.AnalyticsManager != nil {
			if err := h.config.AnalyticsManager.SendEvent(tabCtx, analytics.Event{
				Type: analytics.EventScroll, Category: "engagement",
				Action: "scroll", Label: "75%", Value: 75,
			}); err != nil {
				_ = err
			}
		}

		// İnsan davranışı (kısa)
		hum := behavior.NewHumanBehavior(&behavior.BehaviorConfig{
			MinPageDuration: 1 * time.Second,
			MaxPageDuration: 3 * time.Second,
			ScrollProbability: 0.5, // Zaten scroll yaptık
			MouseMoveProbability: 0.5,
			ClickProbability: 0,
		})
		var pageLen int
		if err := chromedp.Evaluate(`document.body ? document.body.innerText.length : 0`, &pageLen).Do(tabCtx); err != nil {
			pageLen = 0
		}
		hum.SimulatePageVisit(tabCtx, pageLen)
	}

	elapsed := time.Since(start).Milliseconds()

	if navErr != nil {
		h.reporter.Record(reporter.HitRecord{
			Timestamp: time.Now(),
			URL:       urlStr,
			Error:     navErr.Error(),
			UserAgent: ua,
		})
		return navErr
	}

	h.reporter.Record(reporter.HitRecord{
		Timestamp:    time.Now(),
		URL:          urlStr,
		StatusCode:   200,
		ResponseTime: elapsed,
		UserAgent:    ua,
	})
	return nil
}
