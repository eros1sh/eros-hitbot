package browser

import (
	"context"
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
	"eroshit/pkg/referrer"
	"eroshit/pkg/stealth"
	"eroshit/pkg/useragent"
)

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

	if cfg.ProxyURL != "" {
		opts = append(opts,
			chromedp.ProxyServer(cfg.ProxyURL),
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
	ua, _ := h.agentProvider.RandomWithHeaders()
	if ua == "" {
		ua = useragent.Random()
	}
	advFP := fingerprint.GenerateAdvancedFingerprint()
	advFP.UserAgent = ua
	fp := fingerprint.FP{
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

	browserOpts := []chromedp.ContextOption{
		chromedp.WithLogf(func(string, ...interface{}) {}),
	}

	tabCtx, tabCancel := chromedp.NewContext(h.allocCtx, browserOpts...)
	defer tabCancel()

	tabCtx, tabCancel2 := context.WithTimeout(tabCtx, 30*time.Second)
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
				block := blockTypes[ev.ResourceType]
				if block {
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
		emulation.SetDeviceMetricsOverride(int64(fp.InnerW), int64(fp.InnerH), fp.DevicePixel, false),
		emulation.SetTimezoneOverride(fp.Timezone),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.ClearBrowserCookies().Do(ctx)
		}),
	}
	// Keyword'ler varsa referrer ayarla (arama motoru kaynağı simülasyonu)
	if len(h.config.Keywords) > 0 {
		refCfg := &referrer.ReferrerConfig{
			GooglePercent: 50, BingPercent: 20, DirectPercent: 30,
			Keywords: h.config.Keywords,
		}
		refChain := referrer.NewReferrerChain(targetDomain, refCfg)
		src := refChain.Generate()
		if src != nil && src.URL != "" && (src.Type == "search" || src.Type == "social") {
			navActions = append(navActions, chromedp.ActionFunc(func(ctx context.Context) error {
				return network.SetExtraHTTPHeaders(map[string]interface{}{"Referer": src.URL}).Do(ctx)
			}))
		}
	}
	navActions = append(navActions,
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
	)
	navErr := chromedp.Run(tabCtx, navActions...)

	if navErr == nil && gtagScript != "" {
		chromedp.Run(tabCtx, chromedp.Evaluate(gtagScript, nil))
		chromedp.Run(tabCtx, chromedp.Sleep(1500*time.Millisecond))
	}

	// Stealth scriptleri sayfa yüklendikten sonra tekrar enjekte et (bazı siteler için gerekli)
	if navErr == nil {
		_ = stealth.InjectStealthScripts(tabCtx, stealthCfg)
	}

	if navErr == nil {
		// Canvas fingerprint (sayfa yüklendikten sonra)
		if h.config.CanvasFingerprint {
			cf := canvas.GenerateFingerprint()
			_ = cf.InjectCanvasNoise(tabCtx)
			_ = cf.InjectWebGLFingerprint(tabCtx)
			_ = cf.InjectAudioFingerprint(tabCtx)
		}

		// Scroll davranışı
		strategy := h.config.ScrollStrategy
		if strategy == "" {
			strategy = "gradual"
		}
		_ = engagement.HumanScroll(tabCtx, engagement.ScrollBehavior{
			Strategy:    strategy,
			ReadSpeed:   200,
		})

		// Scroll event (GA4)
		if h.config.SendScrollEvent && h.config.AnalyticsManager != nil {
			_ = h.config.AnalyticsManager.SendEvent(tabCtx, analytics.Event{
				Type: analytics.EventScroll, Category: "engagement",
				Action: "scroll", Label: "75%", Value: 75,
			})
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
		_ = chromedp.Evaluate(`document.body ? document.body.innerText.length : 0`, &pageLen).Do(tabCtx)
		hum.SimulatePageVisit(tabCtx, pageLen)
	}

	elapsed := time.Since(start).Milliseconds()
	_ = authDone

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
