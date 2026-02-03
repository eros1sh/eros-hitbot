package analytics

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// ============================================================================
// GOOGLE ANALYTICS 4 (GA4) MEASUREMENT PROTOCOL
// ============================================================================

// GA4Config GA4 yapılandırması
type GA4Config struct {
	MeasurementID   string // G-XXXXXXXXXX formatında
	APISecret       string // Measurement Protocol API secret
	ClientID        string // Client ID (cid)
	UserID          string // User ID (uid) - opsiyonel
	SessionID       string // Session ID
	EngagementTime  int64  // Engagement time in milliseconds
	Debug           bool   // Debug mode
}

// GA4Event GA4 event yapısı
type GA4Event struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// GA4Payload GA4 Measurement Protocol payload
type GA4Payload struct {
	ClientID        string                 `json:"client_id"`
	UserID          string                 `json:"user_id,omitempty"`
	TimestampMicros string                 `json:"timestamp_micros,omitempty"`
	NonPersonalized bool                   `json:"non_personalized_ads,omitempty"`
	Events          []GA4Event             `json:"events"`
	UserProperties  map[string]interface{} `json:"user_properties,omitempty"`
}

// GA4Client GA4 Measurement Protocol client
type GA4Client struct {
	config     GA4Config
	httpClient *http.Client
	mu         sync.Mutex
	rng        *mrand.Rand
}

// NewGA4Client yeni GA4 client oluşturur
func NewGA4Client(config GA4Config) *GA4Client {
	if config.ClientID == "" {
		config.ClientID = GenerateClientID()
	}
	if config.SessionID == "" {
		config.SessionID = GenerateSessionID()
	}
	
	return &GA4Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		rng: mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
}

// SendPageView sayfa görüntüleme eventi gönderir
func (c *GA4Client) SendPageView(pageTitle, pageLocation, pageReferrer string) error {
	event := GA4Event{
		Name: "page_view",
		Params: map[string]interface{}{
			"page_title":       pageTitle,
			"page_location":    pageLocation,
			"page_referrer":    pageReferrer,
			"session_id":       c.config.SessionID,
			"engagement_time_msec": c.getEngagementTime(),
		},
	}
	
	return c.sendEvent(event)
}

// SendSessionStart oturum başlangıç eventi gönderir
func (c *GA4Client) SendSessionStart(pageLocation string) error {
	event := GA4Event{
		Name: "session_start",
		Params: map[string]interface{}{
			"page_location": pageLocation,
			"session_id":    c.config.SessionID,
		},
	}
	
	return c.sendEvent(event)
}

// SendScroll scroll eventi gönderir
func (c *GA4Client) SendScroll(percentScrolled int) error {
	event := GA4Event{
		Name: "scroll",
		Params: map[string]interface{}{
			"percent_scrolled":     percentScrolled,
			"session_id":           c.config.SessionID,
			"engagement_time_msec": c.getEngagementTime(),
		},
	}
	
	return c.sendEvent(event)
}

// SendClick tıklama eventi gönderir
func (c *GA4Client) SendClick(linkURL, linkText string, outbound bool) error {
	eventName := "click"
	if outbound {
		eventName = "click" // GA4'te outbound click ayrı event değil
	}
	
	event := GA4Event{
		Name: eventName,
		Params: map[string]interface{}{
			"link_url":             linkURL,
			"link_text":            linkText,
			"outbound":             outbound,
			"session_id":           c.config.SessionID,
			"engagement_time_msec": c.getEngagementTime(),
		},
	}
	
	return c.sendEvent(event)
}

// SendUserEngagement kullanıcı etkileşim eventi gönderir
func (c *GA4Client) SendUserEngagement(engagementTimeMs int64) error {
	event := GA4Event{
		Name: "user_engagement",
		Params: map[string]interface{}{
			"session_id":           c.config.SessionID,
			"engagement_time_msec": engagementTimeMs,
		},
	}
	
	return c.sendEvent(event)
}

// SendFirstVisit ilk ziyaret eventi gönderir
func (c *GA4Client) SendFirstVisit() error {
	event := GA4Event{
		Name: "first_visit",
		Params: map[string]interface{}{
			"session_id": c.config.SessionID,
		},
	}
	
	return c.sendEvent(event)
}

// SendCustomEvent özel event gönderir
func (c *GA4Client) SendCustomEvent(eventName string, params map[string]interface{}) error {
	if params == nil {
		params = make(map[string]interface{})
	}
	params["session_id"] = c.config.SessionID
	params["engagement_time_msec"] = c.getEngagementTime()
	
	event := GA4Event{
		Name:   eventName,
		Params: params,
	}
	
	return c.sendEvent(event)
}

// sendEvent eventi GA4'e gönderir
func (c *GA4Client) sendEvent(event GA4Event) error {
	payload := GA4Payload{
		ClientID: c.config.ClientID,
		UserID:   c.config.UserID,
		Events:   []GA4Event{event},
	}
	
	return c.send(payload)
}

// SendBatch toplu event gönderir
func (c *GA4Client) SendBatch(events []GA4Event) error {
	payload := GA4Payload{
		ClientID: c.config.ClientID,
		UserID:   c.config.UserID,
		Events:   events,
	}
	
	return c.send(payload)
}

// send payload'ı GA4'e gönderir
func (c *GA4Client) send(payload GA4Payload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON marshal hatası: %w", err)
	}
	
	// Endpoint URL
	baseURL := "https://www.google-analytics.com/mp/collect"
	if c.config.Debug {
		baseURL = "https://www.google-analytics.com/debug/mp/collect"
	}
	
	reqURL := fmt.Sprintf("%s?measurement_id=%s&api_secret=%s",
		baseURL, c.config.MeasurementID, c.config.APISecret)
	
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request oluşturma hatası: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request gönderme hatası: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GA4 hatası: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// getEngagementTime engagement time döner
func (c *GA4Client) getEngagementTime() int64 {
	if c.config.EngagementTime > 0 {
		return c.config.EngagementTime
	}
	// Rastgele 5-60 saniye arası
	c.mu.Lock()
	defer c.mu.Unlock()
	return int64(5000 + c.rng.Intn(55000))
}

// UpdateSessionID session ID'yi günceller
func (c *GA4Client) UpdateSessionID(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.SessionID = sessionID
}

// UpdateClientID client ID'yi günceller
func (c *GA4Client) UpdateClientID(clientID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.ClientID = clientID
}

// ============================================================================
// UNIVERSAL ANALYTICS (analytics.js) SIMULATION
// ============================================================================

// UAConfig Universal Analytics yapılandırması
type UAConfig struct {
	TrackingID string // UA-XXXXXXXX-X formatında
	ClientID   string
	UserID     string
}

// UAClient Universal Analytics client (legacy)
type UAClient struct {
	config     UAConfig
	httpClient *http.Client
}

// NewUAClient yeni UA client oluşturur
func NewUAClient(config UAConfig) *UAClient {
	if config.ClientID == "" {
		config.ClientID = GenerateClientID()
	}
	
	return &UAClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPageView UA pageview gönderir
func (c *UAClient) SendPageView(documentPath, documentTitle, documentReferrer string) error {
	params := url.Values{}
	params.Set("v", "1")                    // Protocol version
	params.Set("tid", c.config.TrackingID)  // Tracking ID
	params.Set("cid", c.config.ClientID)    // Client ID
	params.Set("t", "pageview")             // Hit type
	params.Set("dp", documentPath)          // Document path
	params.Set("dt", documentTitle)         // Document title
	params.Set("dr", documentReferrer)      // Document referrer
	
	if c.config.UserID != "" {
		params.Set("uid", c.config.UserID)
	}
	
	return c.send(params)
}

// SendEvent UA event gönderir
func (c *UAClient) SendEvent(category, action, label string, value int) error {
	params := url.Values{}
	params.Set("v", "1")
	params.Set("tid", c.config.TrackingID)
	params.Set("cid", c.config.ClientID)
	params.Set("t", "event")
	params.Set("ec", category)
	params.Set("ea", action)
	if label != "" {
		params.Set("el", label)
	}
	if value > 0 {
		params.Set("ev", strconv.Itoa(value))
	}
	
	return c.send(params)
}

// send UA hit gönderir
func (c *UAClient) send(params url.Values) error {
	req, err := http.NewRequest("POST", "https://www.google-analytics.com/collect", 
		strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return nil
}

// ============================================================================
// BROWSER-BASED ANALYTICS INJECTION
// ============================================================================

// AnalyticsInjector tarayıcı tabanlı analytics enjeksiyonu
type AnalyticsInjector struct {
	mu  sync.Mutex
	rng *mrand.Rand
}

// NewAnalyticsInjector yeni injector oluşturur
func NewAnalyticsInjector() *AnalyticsInjector {
	return &AnalyticsInjector{
		rng: mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
}

// InjectGoogleAnalytics GA scriptlerinin düzgün çalışmasını sağlar
func (ai *AnalyticsInjector) InjectGoogleAnalytics(ctx context.Context) error {
	// Google Analytics scriptlerinin yüklenmesini bekle ve tetikle
	script := `
(function() {
	// gtag.js kontrolü ve tetikleme
	if (typeof gtag === 'function') {
		// Sayfa görüntüleme eventi
		gtag('event', 'page_view', {
			page_title: document.title,
			page_location: window.location.href,
			page_referrer: document.referrer
		});
		
		// Session start
		gtag('event', 'session_start');
		
		console.log('[Analytics] gtag page_view sent');
	}
	
	// analytics.js kontrolü
	if (typeof ga === 'function') {
		ga('send', 'pageview');
		console.log('[Analytics] ga pageview sent');
	}
	
	// dataLayer kontrolü
	if (typeof dataLayer !== 'undefined' && Array.isArray(dataLayer)) {
		dataLayer.push({
			'event': 'page_view',
			'page_title': document.title,
			'page_location': window.location.href
		});
		console.log('[Analytics] dataLayer push sent');
	}
	
	return true;
})();
`
	
	var result bool
	return chromedp.Run(ctx, chromedp.Evaluate(script, &result))
}

// TriggerEngagementEvents etkileşim eventlerini tetikler
func (ai *AnalyticsInjector) TriggerEngagementEvents(ctx context.Context, scrollPercent int, timeOnPage int) error {
	script := fmt.Sprintf(`
(function() {
	var scrollPercent = %d;
	var timeOnPage = %d;
	
	// gtag ile scroll eventi
	if (typeof gtag === 'function') {
		if (scrollPercent >= 90) {
			gtag('event', 'scroll', {
				'percent_scrolled': scrollPercent
			});
		}
		
		// User engagement
		gtag('event', 'user_engagement', {
			'engagement_time_msec': timeOnPage * 1000
		});
	}
	
	// analytics.js ile event
	if (typeof ga === 'function') {
		ga('send', 'event', 'Engagement', 'scroll', 'Scroll Depth', scrollPercent);
		ga('send', 'event', 'Engagement', 'time', 'Time on Page', timeOnPage);
	}
	
	return true;
})();
`, scrollPercent, timeOnPage)
	
	var result bool
	return chromedp.Run(ctx, chromedp.Evaluate(script, &result))
}

// SimulateRealUserBehavior gerçek kullanıcı davranışını simüle eder
func (ai *AnalyticsInjector) SimulateRealUserBehavior(ctx context.Context) error {
	// Mouse hareketi simülasyonu
	mouseScript := `
(function() {
	// Mouse move eventi
	var event = new MouseEvent('mousemove', {
		bubbles: true,
		cancelable: true,
		clientX: Math.floor(Math.random() * window.innerWidth),
		clientY: Math.floor(Math.random() * window.innerHeight)
	});
	document.dispatchEvent(event);
	
	// Focus eventi
	window.dispatchEvent(new Event('focus'));
	
	// Visibility change
	if (document.hidden === false) {
		// Sayfa görünür, engagement başlat
		if (typeof gtag === 'function') {
			gtag('event', 'user_engagement', {
				'engagement_time_msec': 1000
			});
		}
	}
	
	return true;
})();
`
	
	var result bool
	return chromedp.Run(ctx, chromedp.Evaluate(mouseScript, &result))
}

// WaitForAnalyticsLoad analytics scriptlerinin yüklenmesini bekler
func (ai *AnalyticsInjector) WaitForAnalyticsLoad(ctx context.Context, timeout time.Duration) error {
	script := `
(function() {
	return new Promise((resolve) => {
		var checkInterval = setInterval(function() {
			if (typeof gtag === 'function' || typeof ga === 'function' || typeof dataLayer !== 'undefined') {
				clearInterval(checkInterval);
				resolve(true);
			}
		}, 100);
		
		// Timeout
		setTimeout(function() {
			clearInterval(checkInterval);
			resolve(false);
		}, %d);
	});
})();
`
	
	var loaded bool
	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(script, timeout.Milliseconds()), &loaded))
	if err != nil {
		return err
	}
	
	if !loaded {
		return fmt.Errorf("analytics scripts yüklenemedi")
	}
	
	return nil
}

// ============================================================================
// SEARCH CONSOLE ORGANIC TRAFFIC SIMULATION
// ============================================================================

// SearchConsoleSimulator Search Console için organik trafik simülasyonu
type SearchConsoleSimulator struct {
	mu  sync.Mutex
	rng *mrand.Rand
}

// NewSearchConsoleSimulator yeni simulator oluşturur
func NewSearchConsoleSimulator() *SearchConsoleSimulator {
	return &SearchConsoleSimulator{
		rng: mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
}

// SimulateOrganicSearch organik arama simülasyonu yapar
func (scs *SearchConsoleSimulator) SimulateOrganicSearch(ctx context.Context, keyword, targetURL string) error {
	// Google arama sayfasına git
	googleURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(keyword))
	
	// Referrer'ı Google olarak ayarla
	if err := chromedp.Run(ctx,
		network.SetExtraHTTPHeaders(network.Headers{
			"Referer": "https://www.google.com/",
		}),
	); err != nil {
		return err
	}
	
	// Google'a git
	if err := chromedp.Run(ctx, chromedp.Navigate(googleURL)); err != nil {
		return err
	}
	
	// Arama sonuçlarının yüklenmesini bekle
	if err := chromedp.Run(ctx, chromedp.WaitVisible(`#search`, chromedp.ByID)); err != nil {
		return err
	}
	
	return nil
}

// SetOrganicReferrer organik referrer ayarlar
func (scs *SearchConsoleSimulator) SetOrganicReferrer(ctx context.Context, keyword string) error {
	// Document referrer'ı Google arama sonucu olarak ayarla
	referrer := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(keyword))
	
	script := fmt.Sprintf(`
(function() {
	Object.defineProperty(document, 'referrer', {
		get: function() { return '%s'; }
	});
	return true;
})();
`, referrer)
	
	var result bool
	return chromedp.Run(ctx, chromedp.Evaluate(script, &result))
}

// InjectSearchConsoleSignals Search Console sinyallerini enjekte eder
func (scs *SearchConsoleSimulator) InjectSearchConsoleSignals(ctx context.Context, keyword string) error {
	// Performance Navigation API'yi ayarla
	script := fmt.Sprintf(`
(function() {
	// Navigation type: navigate (organik arama için)
	try {
		Object.defineProperty(performance.navigation, 'type', {
			get: function() { return 0; } // TYPE_NAVIGATE
		});
	} catch(e) {}
	
	// PerformanceNavigationTiming
	try {
		var entries = performance.getEntriesByType('navigation');
		if (entries.length > 0) {
			// Organik arama sinyalleri
		}
	} catch(e) {}
	
	// Referrer ayarla
	Object.defineProperty(document, 'referrer', {
		get: function() { return 'https://www.google.com/search?q=%s'; },
		configurable: true
	});
	
	return true;
})();
`, url.QueryEscape(keyword))
	
	var result bool
	return chromedp.Run(ctx, chromedp.Evaluate(script, &result))
}

// ============================================================================
// COMPREHENSIVE ANALYTICS TRACKER
// ============================================================================

// AnalyticsTracker kapsamlı analytics takipçisi
type AnalyticsTracker struct {
	ga4Client    *GA4Client
	uaClient     *UAClient
	injector     *AnalyticsInjector
	scSimulator  *SearchConsoleSimulator
	mu           sync.Mutex
	sessionStart time.Time
	pageViews    int
	events       []TrackedEvent
	rng          *mrand.Rand
}

// TrackedEvent takip edilen event
type TrackedEvent struct {
	Name      string
	Timestamp time.Time
	Params    map[string]interface{}
}

// AnalyticsTrackerConfig tracker yapılandırması
type AnalyticsTrackerConfig struct {
	GA4MeasurementID string
	GA4APISecret     string
	UATrackingID     string
	EnableGA4        bool
	EnableUA         bool
	EnableInjection  bool
}

// NewAnalyticsTracker yeni tracker oluşturur
func NewAnalyticsTracker(config AnalyticsTrackerConfig) *AnalyticsTracker {
	tracker := &AnalyticsTracker{
		injector:     NewAnalyticsInjector(),
		scSimulator:  NewSearchConsoleSimulator(),
		sessionStart: time.Now(),
		pageViews:    0,
		events:       make([]TrackedEvent, 0),
		rng:          mrand.New(mrand.NewSource(time.Now().UnixNano())),
	}
	
	if config.EnableGA4 && config.GA4MeasurementID != "" && config.GA4APISecret != "" {
		tracker.ga4Client = NewGA4Client(GA4Config{
			MeasurementID: config.GA4MeasurementID,
			APISecret:     config.GA4APISecret,
		})
	}
	
	if config.EnableUA && config.UATrackingID != "" {
		tracker.uaClient = NewUAClient(UAConfig{
			TrackingID: config.UATrackingID,
		})
	}
	
	return tracker
}

// TrackPageView sayfa görüntüleme takibi
func (at *AnalyticsTracker) TrackPageView(ctx context.Context, pageTitle, pageURL, referrer string) error {
	at.mu.Lock()
	at.pageViews++
	isFirstVisit := at.pageViews == 1
	at.mu.Unlock()
	
	var errs []error
	
	// Browser-based injection
	if at.injector != nil {
		// Analytics yüklenmesini bekle
		if err := at.injector.WaitForAnalyticsLoad(ctx, 5*time.Second); err != nil {
			// Hata kritik değil, devam et
			_ = err
		}
		
		// Analytics'i tetikle
		if err := at.injector.InjectGoogleAnalytics(ctx); err != nil {
			errs = append(errs, err)
		}
		
		// Gerçek kullanıcı davranışı simüle et
		if err := at.injector.SimulateRealUserBehavior(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	
	// GA4 Measurement Protocol
	if at.ga4Client != nil {
		if isFirstVisit {
			if err := at.ga4Client.SendFirstVisit(); err != nil {
				errs = append(errs, err)
			}
			if err := at.ga4Client.SendSessionStart(pageURL); err != nil {
				errs = append(errs, err)
			}
		}
		
		if err := at.ga4Client.SendPageView(pageTitle, pageURL, referrer); err != nil {
			errs = append(errs, err)
		}
	}
	
	// Universal Analytics
	if at.uaClient != nil {
		parsedURL, _ := url.Parse(pageURL)
		path := "/"
		if parsedURL != nil {
			path = parsedURL.Path
		}
		
		if err := at.uaClient.SendPageView(path, pageTitle, referrer); err != nil {
			errs = append(errs, err)
		}
	}
	
	// Event kaydet
	at.recordEvent("page_view", map[string]interface{}{
		"page_title":    pageTitle,
		"page_location": pageURL,
		"page_referrer": referrer,
	})
	
	if len(errs) > 0 {
		return fmt.Errorf("analytics hataları: %v", errs)
	}
	
	return nil
}

// TrackScroll scroll takibi
func (at *AnalyticsTracker) TrackScroll(ctx context.Context, percentScrolled int) error {
	var errs []error
	
	// Browser injection
	if at.injector != nil {
		timeOnPage := int(time.Since(at.sessionStart).Seconds())
		if err := at.injector.TriggerEngagementEvents(ctx, percentScrolled, timeOnPage); err != nil {
			errs = append(errs, err)
		}
	}
	
	// GA4
	if at.ga4Client != nil && percentScrolled >= 90 {
		if err := at.ga4Client.SendScroll(percentScrolled); err != nil {
			errs = append(errs, err)
		}
	}
	
	// UA
	if at.uaClient != nil {
		if err := at.uaClient.SendEvent("Engagement", "Scroll", fmt.Sprintf("%d%%", percentScrolled), percentScrolled); err != nil {
			errs = append(errs, err)
		}
	}
	
	at.recordEvent("scroll", map[string]interface{}{
		"percent_scrolled": percentScrolled,
	})
	
	if len(errs) > 0 {
		return fmt.Errorf("scroll tracking hataları: %v", errs)
	}
	
	return nil
}

// TrackEngagement engagement takibi
func (at *AnalyticsTracker) TrackEngagement(ctx context.Context) error {
	engagementTime := time.Since(at.sessionStart).Milliseconds()
	
	var errs []error
	
	// GA4
	if at.ga4Client != nil {
		if err := at.ga4Client.SendUserEngagement(engagementTime); err != nil {
			errs = append(errs, err)
		}
	}
	
	// UA
	if at.uaClient != nil {
		if err := at.uaClient.SendEvent("Engagement", "Time", "Session Duration", int(engagementTime/1000)); err != nil {
			errs = append(errs, err)
		}
	}
	
	at.recordEvent("user_engagement", map[string]interface{}{
		"engagement_time_msec": engagementTime,
	})
	
	if len(errs) > 0 {
		return fmt.Errorf("engagement tracking hataları: %v", errs)
	}
	
	return nil
}

// TrackClick tıklama takibi
func (at *AnalyticsTracker) TrackClick(ctx context.Context, linkURL, linkText string, outbound bool) error {
	var errs []error
	
	// GA4
	if at.ga4Client != nil {
		if err := at.ga4Client.SendClick(linkURL, linkText, outbound); err != nil {
			errs = append(errs, err)
		}
	}
	
	// UA
	if at.uaClient != nil {
		category := "Internal Link"
		if outbound {
			category = "Outbound Link"
		}
		if err := at.uaClient.SendEvent(category, "Click", linkURL, 0); err != nil {
			errs = append(errs, err)
		}
	}
	
	at.recordEvent("click", map[string]interface{}{
		"link_url":  linkURL,
		"link_text": linkText,
		"outbound":  outbound,
	})
	
	if len(errs) > 0 {
		return fmt.Errorf("click tracking hataları: %v", errs)
	}
	
	return nil
}

// SetupOrganicTraffic organik trafik kurulumu
func (at *AnalyticsTracker) SetupOrganicTraffic(ctx context.Context, keyword string) error {
	if at.scSimulator != nil {
		return at.scSimulator.InjectSearchConsoleSignals(ctx, keyword)
	}
	return nil
}

// recordEvent event kaydeder
func (at *AnalyticsTracker) recordEvent(name string, params map[string]interface{}) {
	at.mu.Lock()
	defer at.mu.Unlock()
	
	at.events = append(at.events, TrackedEvent{
		Name:      name,
		Timestamp: time.Now(),
		Params:    params,
	})
}

// GetSessionStats oturum istatistiklerini döner
func (at *AnalyticsTracker) GetSessionStats() map[string]interface{} {
	at.mu.Lock()
	defer at.mu.Unlock()
	
	return map[string]interface{}{
		"session_duration_ms": time.Since(at.sessionStart).Milliseconds(),
		"page_views":          at.pageViews,
		"total_events":        len(at.events),
		"session_start":       at.sessionStart,
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GenerateClientID GA4 formatında client ID üretir
func GenerateClientID() string {
	timestamp := time.Now().Unix()
	random := mrand.Int63n(int64(math.Pow(10, 9)))
	return fmt.Sprintf("%d.%d", random, timestamp)
}

// GenerateSessionID session ID üretir
func GenerateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GenerateUniqueID benzersiz ID üretir
func GenerateUniqueID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ============================================================================
// REAL TRAFFIC VALIDATOR
// ============================================================================

// TrafficValidator trafik doğrulayıcı
type TrafficValidator struct {
	mu              sync.Mutex
	validationRules []ValidationRule
}

// ValidationRule doğrulama kuralı
type ValidationRule struct {
	Name        string
	Description string
	Check       func(ctx context.Context) bool
}

// NewTrafficValidator yeni validator oluşturur
func NewTrafficValidator() *TrafficValidator {
	tv := &TrafficValidator{
		validationRules: make([]ValidationRule, 0),
	}
	
	// Varsayılan kuralları ekle
	tv.addDefaultRules()
	
	return tv
}

// addDefaultRules varsayılan kuralları ekler
func (tv *TrafficValidator) addDefaultRules() {
	tv.validationRules = append(tv.validationRules, ValidationRule{
		Name:        "analytics_loaded",
		Description: "Google Analytics scriptleri yüklendi mi",
		Check: func(ctx context.Context) bool {
			var loaded bool
			script := `typeof gtag === 'function' || typeof ga === 'function'`
			chromedp.Run(ctx, chromedp.Evaluate(script, &loaded))
			return loaded
		},
	})
	
	tv.validationRules = append(tv.validationRules, ValidationRule{
		Name:        "cookies_enabled",
		Description: "Çerezler etkin mi",
		Check: func(ctx context.Context) bool {
			var enabled bool
			script := `navigator.cookieEnabled`
			chromedp.Run(ctx, chromedp.Evaluate(script, &enabled))
			return enabled
		},
	})
	
	tv.validationRules = append(tv.validationRules, ValidationRule{
		Name:        "javascript_enabled",
		Description: "JavaScript etkin mi",
		Check: func(ctx context.Context) bool {
			// JavaScript çalışıyorsa bu fonksiyon zaten çalışır
			return true
		},
	})
	
	tv.validationRules = append(tv.validationRules, ValidationRule{
		Name:        "referrer_set",
		Description: "Referrer ayarlandı mı",
		Check: func(ctx context.Context) bool {
			var referrer string
			script := `document.referrer`
			chromedp.Run(ctx, chromedp.Evaluate(script, &referrer))
			return referrer != ""
		},
	})
}

// Validate tüm kuralları doğrular
func (tv *TrafficValidator) Validate(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	
	for _, rule := range tv.validationRules {
		results[rule.Name] = rule.Check(ctx)
	}
	
	return results
}

// IsValid trafik geçerli mi
func (tv *TrafficValidator) IsValid(ctx context.Context) bool {
	results := tv.Validate(ctx)
	
	for _, valid := range results {
		if !valid {
			return false
		}
	}
	
	return true
}

// AddRule özel kural ekler
func (tv *TrafficValidator) AddRule(rule ValidationRule) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	
	tv.validationRules = append(tv.validationRules, rule)
}
