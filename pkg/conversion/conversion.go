package conversion

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// EventType dönüşüm olay tipi
type EventType string

const (
	EventPageView       EventType = "page_view"
	EventScroll         EventType = "scroll"
	EventClick          EventType = "click"
	EventFormSubmit     EventType = "form_submit"
	EventFormStart      EventType = "form_start"
	EventVideoStart     EventType = "video_start"
	EventVideoProgress  EventType = "video_progress"
	EventVideoComplete  EventType = "video_complete"
	EventFileDownload   EventType = "file_download"
	EventOutboundClick  EventType = "outbound_click"
	EventSiteSearch     EventType = "site_search"
	EventAddToCart      EventType = "add_to_cart"
	EventPurchase       EventType = "purchase"
	EventSignUp         EventType = "sign_up"
	EventLogin          EventType = "login"
	EventShare          EventType = "share"
	EventCustom         EventType = "custom"
)

// ConversionEvent dönüşüm olayı
type ConversionEvent struct {
	Type       EventType
	Category   string
	Action     string
	Label      string
	Value      float64
	Parameters map[string]interface{}
}

// ConversionConfig dönüşüm simülasyonu yapılandırması
type ConversionConfig struct {
	GA4Enabled       bool
	GA4MeasurementID string
	GTMEnabled       bool
	GTMContainerID   string
	Events           []ConversionEvent
	ScrollMilestones []int // Scroll yüzdeleri (25, 50, 75, 100)
}

// ConversionSimulator dönüşüm simülatörü
type ConversionSimulator struct {
	config ConversionConfig
	mu     sync.Mutex
	rng    *rand.Rand
}

// NewConversionSimulator yeni dönüşüm simülatörü oluşturur
func NewConversionSimulator(config ConversionConfig) *ConversionSimulator {
	if len(config.ScrollMilestones) == 0 {
		config.ScrollMilestones = []int{25, 50, 75, 90}
	}
	
	return &ConversionSimulator{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SendGA4Event GA4 event gönderir
func (cs *ConversionSimulator) SendGA4Event(ctx context.Context, event ConversionEvent) error {
	if !cs.config.GA4Enabled || cs.config.GA4MeasurementID == "" {
		return nil
	}
	
	// GA4 event parametrelerini oluştur
	params := make(map[string]interface{})
	if event.Category != "" {
		params["event_category"] = event.Category
	}
	if event.Label != "" {
		params["event_label"] = event.Label
	}
	if event.Value > 0 {
		params["value"] = event.Value
	}
	
	// Ek parametreleri ekle
	for k, v := range event.Parameters {
		params[k] = v
	}
	
	// gtag event script'i oluştur
	script := fmt.Sprintf(`
		(function() {
			if (typeof gtag === 'function') {
				gtag('event', '%s', %s);
				return true;
			}
			return false;
		})();
	`, event.Action, mapToJS(params))
	
	var result bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &result)); err != nil {
		return fmt.Errorf("GA4 event gönderilemedi: %w", err)
	}
	
	return nil
}

// SendDataLayerEvent dataLayer event gönderir (GTM için)
func (cs *ConversionSimulator) SendDataLayerEvent(ctx context.Context, event ConversionEvent) error {
	params := make(map[string]interface{})
	params["event"] = string(event.Type)
	
	if event.Category != "" {
		params["eventCategory"] = event.Category
	}
	if event.Action != "" {
		params["eventAction"] = event.Action
	}
	if event.Label != "" {
		params["eventLabel"] = event.Label
	}
	if event.Value > 0 {
		params["eventValue"] = event.Value
	}
	
	for k, v := range event.Parameters {
		params[k] = v
	}
	
	script := fmt.Sprintf(`
		(function() {
			window.dataLayer = window.dataLayer || [];
			window.dataLayer.push(%s);
			return true;
		})();
	`, mapToJS(params))
	
	var result bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &result)); err != nil {
		return fmt.Errorf("dataLayer event gönderilemedi: %w", err)
	}
	
	return nil
}

// SimulateScrollEvents scroll milestone event'leri gönderir
func (cs *ConversionSimulator) SimulateScrollEvents(ctx context.Context) error {
	for _, milestone := range cs.config.ScrollMilestones {
		// Scroll yap
		script := fmt.Sprintf(`
			(function() {
				var docHeight = Math.max(
					document.body.scrollHeight,
					document.documentElement.scrollHeight
				);
				var targetScroll = (docHeight - window.innerHeight) * %d / 100;
				window.scrollTo({top: targetScroll, behavior: 'smooth'});
				return true;
			})();
		`, milestone)
		
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			continue
		}
		
		// Bekle
		time.Sleep(time.Duration(500+cs.rng.Intn(1000)) * time.Millisecond)
		
		// Scroll event gönder
		event := ConversionEvent{
			Type:     EventScroll,
			Category: "engagement",
			Action:   "scroll",
			Label:    fmt.Sprintf("%d%%", milestone),
			Value:    float64(milestone),
			Parameters: map[string]interface{}{
				"percent_scrolled": milestone,
			},
		}
		
		if err := cs.SendGA4Event(ctx, event); err != nil {
			// Event hatası kritik değil
			_ = err
		}
	}
	
	return nil
}

// SimulateFormInteraction form etkileşimi simüle eder
func (cs *ConversionSimulator) SimulateFormInteraction(ctx context.Context, formSelector string) error {
	// Form başlangıç event'i
	startEvent := ConversionEvent{
		Type:     EventFormStart,
		Category: "engagement",
		Action:   "form_start",
		Label:    formSelector,
	}
	
	if err := cs.SendGA4Event(ctx, startEvent); err != nil {
		_ = err
	}
	
	// Form alanlarını bul ve doldur
	script := fmt.Sprintf(`
		(function() {
			var form = document.querySelector('%s');
			if (!form) return false;
			
			var inputs = form.querySelectorAll('input[type="text"], input[type="email"], textarea');
			inputs.forEach(function(input) {
				input.focus();
				input.dispatchEvent(new Event('focus', {bubbles: true}));
			});
			
			return true;
		})();
	`, formSelector)
	
	var result bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &result)); err != nil {
		return err
	}
	
	// Bekle (form doldurma simülasyonu)
	time.Sleep(time.Duration(2000+cs.rng.Intn(3000)) * time.Millisecond)
	
	return nil
}

// SimulateVideoWatch video izleme simüle eder
func (cs *ConversionSimulator) SimulateVideoWatch(ctx context.Context, videoSelector string, watchPercent int) error {
	// Video başlangıç event'i
	startEvent := ConversionEvent{
		Type:     EventVideoStart,
		Category: "video",
		Action:   "video_start",
		Label:    videoSelector,
	}
	
	if err := cs.SendGA4Event(ctx, startEvent); err != nil {
		_ = err
	}
	
	// Video progress event'leri
	milestones := []int{25, 50, 75, 100}
	for _, milestone := range milestones {
		if milestone > watchPercent {
			break
		}
		
		// Bekle
		time.Sleep(time.Duration(1000+cs.rng.Intn(2000)) * time.Millisecond)
		
		progressEvent := ConversionEvent{
			Type:     EventVideoProgress,
			Category: "video",
			Action:   "video_progress",
			Label:    fmt.Sprintf("%d%%", milestone),
			Value:    float64(milestone),
			Parameters: map[string]interface{}{
				"video_percent": milestone,
			},
		}
		
		if err := cs.SendGA4Event(ctx, progressEvent); err != nil {
			_ = err
		}
	}
	
	// Video tamamlandı
	if watchPercent >= 100 {
		completeEvent := ConversionEvent{
			Type:     EventVideoComplete,
			Category: "video",
			Action:   "video_complete",
			Label:    videoSelector,
		}
		
		if err := cs.SendGA4Event(ctx, completeEvent); err != nil {
			_ = err
		}
	}
	
	return nil
}

// SimulateClick tıklama simüle eder
func (cs *ConversionSimulator) SimulateClick(ctx context.Context, selector string, eventLabel string) error {
	// Elemente scroll yap
	if err := chromedp.Run(ctx, chromedp.ScrollIntoView(selector)); err != nil {
		return err
	}
	
	// Bekle
	time.Sleep(time.Duration(300+cs.rng.Intn(500)) * time.Millisecond)
	
	// Tıkla
	if err := chromedp.Run(ctx, chromedp.Click(selector)); err != nil {
		return err
	}
	
	// Click event gönder
	clickEvent := ConversionEvent{
		Type:     EventClick,
		Category: "engagement",
		Action:   "click",
		Label:    eventLabel,
	}
	
	return cs.SendGA4Event(ctx, clickEvent)
}

// SimulateAddToCart sepete ekleme simüle eder
func (cs *ConversionSimulator) SimulateAddToCart(ctx context.Context, productID string, productName string, price float64) error {
	event := ConversionEvent{
		Type:     EventAddToCart,
		Category: "ecommerce",
		Action:   "add_to_cart",
		Label:    productName,
		Value:    price,
		Parameters: map[string]interface{}{
			"currency": "TRY",
			"items": []map[string]interface{}{
				{
					"item_id":   productID,
					"item_name": productName,
					"price":     price,
					"quantity":  1,
				},
			},
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SimulatePurchase satın alma simüle eder
func (cs *ConversionSimulator) SimulatePurchase(ctx context.Context, transactionID string, value float64, items []map[string]interface{}) error {
	event := ConversionEvent{
		Type:     EventPurchase,
		Category: "ecommerce",
		Action:   "purchase",
		Label:    transactionID,
		Value:    value,
		Parameters: map[string]interface{}{
			"transaction_id": transactionID,
			"currency":       "TRY",
			"value":          value,
			"items":          items,
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SimulateSiteSearch site içi arama simüle eder
func (cs *ConversionSimulator) SimulateSiteSearch(ctx context.Context, searchTerm string) error {
	event := ConversionEvent{
		Type:     EventSiteSearch,
		Category: "engagement",
		Action:   "search",
		Label:    searchTerm,
		Parameters: map[string]interface{}{
			"search_term": searchTerm,
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SimulateSignUp kayıt simüle eder
func (cs *ConversionSimulator) SimulateSignUp(ctx context.Context, method string) error {
	event := ConversionEvent{
		Type:     EventSignUp,
		Category: "engagement",
		Action:   "sign_up",
		Label:    method,
		Parameters: map[string]interface{}{
			"method": method,
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SimulateLogin giriş simüle eder
func (cs *ConversionSimulator) SimulateLogin(ctx context.Context, method string) error {
	event := ConversionEvent{
		Type:     EventLogin,
		Category: "engagement",
		Action:   "login",
		Label:    method,
		Parameters: map[string]interface{}{
			"method": method,
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SimulateShare paylaşım simüle eder
func (cs *ConversionSimulator) SimulateShare(ctx context.Context, method string, contentType string, itemID string) error {
	event := ConversionEvent{
		Type:     EventShare,
		Category: "engagement",
		Action:   "share",
		Label:    method,
		Parameters: map[string]interface{}{
			"method":       method,
			"content_type": contentType,
			"item_id":      itemID,
		},
	}
	
	return cs.SendGA4Event(ctx, event)
}

// SendCustomEvent özel event gönderir
func (cs *ConversionSimulator) SendCustomEvent(ctx context.Context, eventName string, params map[string]interface{}) error {
	event := ConversionEvent{
		Type:       EventCustom,
		Action:     eventName,
		Parameters: params,
	}
	
	return cs.SendGA4Event(ctx, event)
}

// mapToJS map'i JavaScript objesine çevirir
func mapToJS(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}
	
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ", "
		}
		first = false
		
		switch val := v.(type) {
		case string:
			result += fmt.Sprintf("'%s': '%s'", k, val)
		case int, int64, float64:
			result += fmt.Sprintf("'%s': %v", k, val)
		case bool:
			result += fmt.Sprintf("'%s': %t", k, val)
		case []map[string]interface{}:
			result += fmt.Sprintf("'%s': %s", k, sliceMapToJS(val))
		case map[string]interface{}:
			result += fmt.Sprintf("'%s': %s", k, mapToJS(val))
		default:
			result += fmt.Sprintf("'%s': '%v'", k, val)
		}
	}
	result += "}"
	
	return result
}

// sliceMapToJS slice of map'i JavaScript array'ine çevirir
func sliceMapToJS(s []map[string]interface{}) string {
	if len(s) == 0 {
		return "[]"
	}
	
	result := "["
	for i, m := range s {
		if i > 0 {
			result += ", "
		}
		result += mapToJS(m)
	}
	result += "]"
	
	return result
}

// GetEventTypes tüm event tiplerini döner
func GetEventTypes() []EventType {
	return []EventType{
		EventPageView,
		EventScroll,
		EventClick,
		EventFormSubmit,
		EventFormStart,
		EventVideoStart,
		EventVideoProgress,
		EventVideoComplete,
		EventFileDownload,
		EventOutboundClick,
		EventSiteSearch,
		EventAddToCart,
		EventPurchase,
		EventSignUp,
		EventLogin,
		EventShare,
		EventCustom,
	}
}
