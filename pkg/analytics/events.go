package analytics

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
)

// EventType analytics event tipi
type EventType string

const (
	EventPageView   EventType = "page_view"
	EventClick      EventType = "click"
	EventScroll     EventType = "scroll"
	EventFormSubmit EventType = "form_submit"
	EventAddToCart  EventType = "add_to_cart"
	EventPurchase  EventType = "purchase"
	EventSignup    EventType = "sign_up"
	EventSearch    EventType = "search"
)

// Event tek bir analytics event
type Event struct {
	Type       EventType
	Category   string
	Action     string
	Label      string
	Value      int
	Parameters map[string]interface{}
}

// Manager analytics yöneticisi
type Manager struct {
	GA4Enabled       bool
	GA4MeasurementID string
	GTMEnabled       bool
	GTMID            string
	FBPixelEnabled   bool
	FBPixelID        string
}

// SendEvent event'i yapılandırılmış platformlara gönderir
func (m *Manager) SendEvent(ctx context.Context, event Event) error {
	var errs []error
	if m.GA4Enabled && m.GA4MeasurementID != "" {
		if err := m.sendGA4Event(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	if m.GTMEnabled && m.GTMID != "" {
		if err := m.sendGTMEvent(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	if m.FBPixelEnabled && m.FBPixelID != "" {
		if err := m.sendFBPixelEvent(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("event errors: %v", errs)
	}
	return nil
}

func (m *Manager) sendGA4Event(ctx context.Context, event Event) error {
	params := m.formatGA4Params(event.Parameters)
	script := fmt.Sprintf(`(function(){
		if(typeof gtag==='function'){
			gtag('event','%s',{'event_category':'%s','event_label':'%s','value':%d%s});
		}
	})();`,
		escapeJS(event.Action),
		escapeJS(event.Category),
		escapeJS(event.Label),
		event.Value,
		params,
	)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

func (m *Manager) sendGTMEvent(ctx context.Context, event Event) error {
	script := fmt.Sprintf(`(function(){
		if(typeof dataLayer!=='undefined'){
			dataLayer.push({
				'event':'%s','eventCategory':'%s','eventAction':'%s','eventLabel':'%s','eventValue':%d
			});
		}
	})();`,
		escapeJS(string(event.Type)),
		escapeJS(event.Category),
		escapeJS(event.Action),
		escapeJS(event.Label),
		event.Value,
	)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

func (m *Manager) sendFBPixelEvent(ctx context.Context, event Event) error {
	fbName := mapToFBEvent(event.Type)
	script := fmt.Sprintf(`(function(){
		if(typeof fbq==='function'){
			fbq('track','%s',{value:%d,currency:'USD'});
		}
	})();`, escapeJS(fbName), event.Value)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

func (m *Manager) formatGA4Params(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("'%s':'%v'", escapeJS(k), v))
	}
	return "," + strings.Join(parts, ",")
}

func mapToFBEvent(t EventType) string {
	m := map[EventType]string{
		EventPageView:   "PageView",
		EventAddToCart:  "AddToCart",
		EventPurchase:   "Purchase",
		EventSignup:     "CompleteRegistration",
		EventSearch:     "Search",
	}
	if s, ok := m[t]; ok {
		return s
	}
	return "CustomEvent"
}

func escapeJS(s string) string {
	return strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "").Replace(s)
}
