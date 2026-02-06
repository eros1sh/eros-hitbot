// Package analytics provides Google Analytics 4 deep integration
// with comprehensive event tracking, user properties, and e-commerce support
package analytics

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// GA4DeepIntegration provides comprehensive GA4 integration
type GA4DeepIntegration struct {
	MeasurementID     string
	ClientID          string
	UserID            string
	SessionID         string
	SessionNumber     int
	EngagementTime    int64
	FirstVisit        bool
	UserProperties    map[string]interface{}
	CustomDimensions  map[string]string
	CustomMetrics     map[string]float64
	ConsentMode       ConsentMode
	DebugMode         bool
}

// ConsentMode represents GA4 consent mode settings
type ConsentMode struct {
	AdStorage           string // granted, denied
	AnalyticsStorage    string
	FunctionalityStorage string
	PersonalizationStorage string
	SecurityStorage     string
	AdUserData          string
	AdPersonalization   string
}

// GA4DeepEvent represents a GA4 event with extended fields for deep integration
// Note: GA4Event is defined in analytics.go, this extends it with additional fields
type GA4DeepEvent struct {
	Name       string                 `json:"name"`
	Params     map[string]interface{} `json:"params"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id,omitempty"`
	ClientID   string                 `json:"client_id"`
	SessionID  string                 `json:"session_id"`
}

// DeepEcommerceItem represents an e-commerce item with extended fields
// Note: EcommerceItem is defined in ecommerce.go, this extends it
type DeepEcommerceItem struct {
	ItemID        string  `json:"item_id"`
	ItemName      string  `json:"item_name"`
	ItemBrand     string  `json:"item_brand,omitempty"`
	ItemCategory  string  `json:"item_category,omitempty"`
	ItemCategory2 string  `json:"item_category2,omitempty"`
	ItemCategory3 string  `json:"item_category3,omitempty"`
	ItemCategory4 string  `json:"item_category4,omitempty"`
	ItemCategory5 string  `json:"item_category5,omitempty"`
	ItemVariant   string  `json:"item_variant,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Quantity      int     `json:"quantity,omitempty"`
	Coupon        string  `json:"coupon,omitempty"`
	Index         int     `json:"index,omitempty"`
	Discount      float64 `json:"discount,omitempty"`
	Affiliation   string  `json:"affiliation,omitempty"`
	ItemListID    string  `json:"item_list_id,omitempty"`
	ItemListName  string  `json:"item_list_name,omitempty"`
	LocationID    string  `json:"location_id,omitempty"`
}

// NewGA4DeepIntegration creates a new GA4 deep integration instance
func NewGA4DeepIntegration(measurementID string) *GA4DeepIntegration {
	return &GA4DeepIntegration{
		MeasurementID:    measurementID,
		ClientID:         generateClientID(),
		SessionID:        generateSessionID(),
		SessionNumber:    1,
		EngagementTime:   0,
		FirstVisit:       true,
		UserProperties:   make(map[string]interface{}),
		CustomDimensions: make(map[string]string),
		CustomMetrics:    make(map[string]float64),
		ConsentMode: ConsentMode{
			AdStorage:           "granted",
			AnalyticsStorage:    "granted",
			FunctionalityStorage: "granted",
			PersonalizationStorage: "granted",
			SecurityStorage:     "granted",
			AdUserData:          "granted",
			AdPersonalization:   "granted",
		},
		DebugMode: false,
	}
}

// SetUserID sets the user ID for cross-device tracking
func (g *GA4DeepIntegration) SetUserID(userID string) {
	g.UserID = userID
}

// SetUserProperty sets a user property
func (g *GA4DeepIntegration) SetUserProperty(name string, value interface{}) {
	g.UserProperties[name] = value
}

// SetCustomDimension sets a custom dimension
func (g *GA4DeepIntegration) SetCustomDimension(name string, value string) {
	g.CustomDimensions[name] = value
}

// SetCustomMetric sets a custom metric
func (g *GA4DeepIntegration) SetCustomMetric(name string, value float64) {
	g.CustomMetrics[name] = value
}

// GenerateGA4Script generates comprehensive GA4 JavaScript
func (g *GA4DeepIntegration) GenerateGA4Script() string {
	userPropsJSON := mapToJSON(g.UserProperties)
	customDimsJSON := mapToJSON(stringMapToInterface(g.CustomDimensions))
	customMetricsJSON := mapToJSON(floatMapToInterface(g.CustomMetrics))

	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// GA4 Deep Integration
	const ga4Config = {
		measurementId: '%s',
		clientId: '%s',
		userId: '%s',
		sessionId: '%s',
		sessionNumber: %d,
		firstVisit: %t,
		debugMode: %t,
		userProperties: %s,
		customDimensions: %s,
		customMetrics: %s,
		consent: {
			ad_storage: '%s',
			analytics_storage: '%s',
			functionality_storage: '%s',
			personalization_storage: '%s',
			security_storage: '%s',
			ad_user_data: '%s',
			ad_personalization: '%s'
		}
	};
	
	// Initialize dataLayer
	window.dataLayer = window.dataLayer || [];
	function gtag() { dataLayer.push(arguments); }
	window.gtag = gtag;
	
	// Set consent mode
	gtag('consent', 'default', ga4Config.consent);
	
	// Initialize GA4
	gtag('js', new Date());
	
	// Configure with all parameters
	const configParams = {
		client_id: ga4Config.clientId,
		session_id: ga4Config.sessionId,
		session_number: ga4Config.sessionNumber,
		first_visit: ga4Config.firstVisit,
		debug_mode: ga4Config.debugMode,
		send_page_view: false, // We'll send manually
		cookie_flags: 'SameSite=None;Secure',
		cookie_domain: 'auto',
		cookie_expires: 63072000, // 2 years
		cookie_update: true,
		allow_google_signals: true,
		allow_ad_personalization_signals: true
	};
	
	// Add user ID if set
	if (ga4Config.userId) {
		configParams.user_id = ga4Config.userId;
	}
	
	// Add custom dimensions
	Object.keys(ga4Config.customDimensions).forEach(key => {
		configParams[key] = ga4Config.customDimensions[key];
	});
	
	// Add custom metrics
	Object.keys(ga4Config.customMetrics).forEach(key => {
		configParams[key] = ga4Config.customMetrics[key];
	});
	
	gtag('config', ga4Config.measurementId, configParams);
	
	// Set user properties
	if (Object.keys(ga4Config.userProperties).length > 0) {
		gtag('set', 'user_properties', ga4Config.userProperties);
	}
	
	// Engagement time tracking
	let engagementStartTime = Date.now();
	let totalEngagementTime = 0;
	let isEngaged = true;
	
	document.addEventListener('visibilitychange', function() {
		if (document.hidden) {
			totalEngagementTime += Date.now() - engagementStartTime;
			isEngaged = false;
		} else {
			engagementStartTime = Date.now();
			isEngaged = true;
		}
	});
	
	// Get current engagement time
	window.__getEngagementTime = function() {
		if (isEngaged) {
			return totalEngagementTime + (Date.now() - engagementStartTime);
		}
		return totalEngagementTime;
	};
	
	// Core event functions
	window.__ga4 = {
		// Page view with enhanced parameters
		pageView: function(params = {}) {
			const defaultParams = {
				page_title: document.title,
				page_location: window.location.href,
				page_path: window.location.pathname,
				page_referrer: document.referrer,
				engagement_time_msec: window.__getEngagementTime(),
				session_id: ga4Config.sessionId,
				session_number: ga4Config.sessionNumber
			};
			
			gtag('event', 'page_view', Object.assign(defaultParams, params));
		},
		
		// Session start
		sessionStart: function(params = {}) {
			gtag('event', 'session_start', Object.assign({
				session_id: ga4Config.sessionId,
				session_number: ga4Config.sessionNumber
			}, params));
		},
		
		// First visit
		firstVisit: function(params = {}) {
			if (ga4Config.firstVisit) {
				gtag('event', 'first_visit', Object.assign({
					session_id: ga4Config.sessionId
				}, params));
			}
		},
		
		// User engagement
		userEngagement: function(params = {}) {
			gtag('event', 'user_engagement', Object.assign({
				engagement_time_msec: window.__getEngagementTime(),
				session_id: ga4Config.sessionId
			}, params));
		},
		
		// Scroll tracking
		scroll: function(percentScrolled, params = {}) {
			gtag('event', 'scroll', Object.assign({
				percent_scrolled: percentScrolled,
				engagement_time_msec: window.__getEngagementTime()
			}, params));
		},
		
		// Click tracking
		click: function(linkUrl, linkText, outbound = false, params = {}) {
			gtag('event', 'click', Object.assign({
				link_url: linkUrl,
				link_text: linkText,
				outbound: outbound
			}, params));
		},
		
		// File download
		fileDownload: function(fileName, fileExtension, linkUrl, params = {}) {
			gtag('event', 'file_download', Object.assign({
				file_name: fileName,
				file_extension: fileExtension,
				link_url: linkUrl
			}, params));
		},
		
		// Video events
		videoStart: function(videoTitle, videoUrl, videoProvider, params = {}) {
			gtag('event', 'video_start', Object.assign({
				video_title: videoTitle,
				video_url: videoUrl,
				video_provider: videoProvider
			}, params));
		},
		
		videoProgress: function(videoTitle, videoPercent, params = {}) {
			gtag('event', 'video_progress', Object.assign({
				video_title: videoTitle,
				video_percent: videoPercent
			}, params));
		},
		
		videoComplete: function(videoTitle, params = {}) {
			gtag('event', 'video_complete', Object.assign({
				video_title: videoTitle
			}, params));
		},
		
		// Form events
		formStart: function(formId, formName, params = {}) {
			gtag('event', 'form_start', Object.assign({
				form_id: formId,
				form_name: formName
			}, params));
		},
		
		formSubmit: function(formId, formName, params = {}) {
			gtag('event', 'form_submit', Object.assign({
				form_id: formId,
				form_name: formName
			}, params));
		},
		
		// Search
		search: function(searchTerm, params = {}) {
			gtag('event', 'search', Object.assign({
				search_term: searchTerm
			}, params));
		},
		
		// View search results
		viewSearchResults: function(searchTerm, params = {}) {
			gtag('event', 'view_search_results', Object.assign({
				search_term: searchTerm
			}, params));
		},
		
		// Custom event
		event: function(eventName, params = {}) {
			params.engagement_time_msec = window.__getEngagementTime();
			params.session_id = ga4Config.sessionId;
			gtag('event', eventName, params);
		},
		
		// E-commerce: View item list
		viewItemList: function(items, listId, listName, params = {}) {
			gtag('event', 'view_item_list', Object.assign({
				item_list_id: listId,
				item_list_name: listName,
				items: items
			}, params));
		},
		
		// E-commerce: Select item
		selectItem: function(items, listId, listName, params = {}) {
			gtag('event', 'select_item', Object.assign({
				item_list_id: listId,
				item_list_name: listName,
				items: items
			}, params));
		},
		
		// E-commerce: View item
		viewItem: function(items, currency, value, params = {}) {
			gtag('event', 'view_item', Object.assign({
				currency: currency,
				value: value,
				items: items
			}, params));
		},
		
		// E-commerce: Add to cart
		addToCart: function(items, currency, value, params = {}) {
			gtag('event', 'add_to_cart', Object.assign({
				currency: currency,
				value: value,
				items: items
			}, params));
		},
		
		// E-commerce: Remove from cart
		removeFromCart: function(items, currency, value, params = {}) {
			gtag('event', 'remove_from_cart', Object.assign({
				currency: currency,
				value: value,
				items: items
			}, params));
		},
		
		// E-commerce: View cart
		viewCart: function(items, currency, value, params = {}) {
			gtag('event', 'view_cart', Object.assign({
				currency: currency,
				value: value,
				items: items
			}, params));
		},
		
		// E-commerce: Begin checkout
		beginCheckout: function(items, currency, value, coupon, params = {}) {
			gtag('event', 'begin_checkout', Object.assign({
				currency: currency,
				value: value,
				coupon: coupon,
				items: items
			}, params));
		},
		
		// E-commerce: Add shipping info
		addShippingInfo: function(items, currency, value, shippingTier, params = {}) {
			gtag('event', 'add_shipping_info', Object.assign({
				currency: currency,
				value: value,
				shipping_tier: shippingTier,
				items: items
			}, params));
		},
		
		// E-commerce: Add payment info
		addPaymentInfo: function(items, currency, value, paymentType, params = {}) {
			gtag('event', 'add_payment_info', Object.assign({
				currency: currency,
				value: value,
				payment_type: paymentType,
				items: items
			}, params));
		},
		
		// E-commerce: Purchase
		purchase: function(transactionId, items, currency, value, tax, shipping, coupon, params = {}) {
			gtag('event', 'purchase', Object.assign({
				transaction_id: transactionId,
				currency: currency,
				value: value,
				tax: tax,
				shipping: shipping,
				coupon: coupon,
				items: items
			}, params));
		},
		
		// E-commerce: Refund
		refund: function(transactionId, items, currency, value, params = {}) {
			gtag('event', 'refund', Object.assign({
				transaction_id: transactionId,
				currency: currency,
				value: value,
				items: items
			}, params));
		},
		
		// E-commerce: View promotion
		viewPromotion: function(promotionId, promotionName, creativeName, creativeSlot, items, params = {}) {
			gtag('event', 'view_promotion', Object.assign({
				promotion_id: promotionId,
				promotion_name: promotionName,
				creative_name: creativeName,
				creative_slot: creativeSlot,
				items: items
			}, params));
		},
		
		// E-commerce: Select promotion
		selectPromotion: function(promotionId, promotionName, creativeName, creativeSlot, items, params = {}) {
			gtag('event', 'select_promotion', Object.assign({
				promotion_id: promotionId,
				promotion_name: promotionName,
				creative_name: creativeName,
				creative_slot: creativeSlot,
				items: items
			}, params));
		},
		
		// Lead generation
		generateLead: function(currency, value, params = {}) {
			gtag('event', 'generate_lead', Object.assign({
				currency: currency,
				value: value
			}, params));
		},
		
		// Sign up
		signUp: function(method, params = {}) {
			gtag('event', 'sign_up', Object.assign({
				method: method
			}, params));
		},
		
		// Login
		login: function(method, params = {}) {
			gtag('event', 'login', Object.assign({
				method: method
			}, params));
		},
		
		// Share
		share: function(method, contentType, itemId, params = {}) {
			gtag('event', 'share', Object.assign({
				method: method,
				content_type: contentType,
				item_id: itemId
			}, params));
		},
		
		// Exception
		exception: function(description, fatal = false, params = {}) {
			gtag('event', 'exception', Object.assign({
				description: description,
				fatal: fatal
			}, params));
		},
		
		// Timing
		timing: function(name, value, eventCategory, eventLabel, params = {}) {
			gtag('event', 'timing_complete', Object.assign({
				name: name,
				value: value,
				event_category: eventCategory,
				event_label: eventLabel
			}, params));
		},
		
		// Set user property
		setUserProperty: function(name, value) {
			const props = {};
			props[name] = value;
			gtag('set', 'user_properties', props);
		},
		
		// Set user ID
		setUserId: function(userId) {
			gtag('config', ga4Config.measurementId, {
				user_id: userId
			});
		},
		
		// Get client ID
		getClientId: function() {
			return ga4Config.clientId;
		},
		
		// Get session ID
		getSessionId: function() {
			return ga4Config.sessionId;
		}
	};
	
	// Auto-track scroll depth
	let maxScrollDepth = 0;
	const scrollThresholds = [25, 50, 75, 90, 100];
	const scrolledThresholds = new Set();
	
	window.addEventListener('scroll', function() {
		const scrollHeight = document.documentElement.scrollHeight - window.innerHeight;
		const scrollPercent = Math.round((window.scrollY / scrollHeight) * 100);
		
		if (scrollPercent > maxScrollDepth) {
			maxScrollDepth = scrollPercent;
			
			scrollThresholds.forEach(threshold => {
				if (scrollPercent >= threshold && !scrolledThresholds.has(threshold)) {
					scrolledThresholds.add(threshold);
					window.__ga4.scroll(threshold);
				}
			});
		}
	}, { passive: true });
	
	// Auto-track outbound links
	document.addEventListener('click', function(e) {
		const link = e.target.closest('a');
		if (link && link.href) {
			const url = new URL(link.href);
			if (url.hostname !== window.location.hostname) {
				window.__ga4.click(link.href, link.textContent, true);
			}
		}
	});
	
	// Send initial events
	window.__ga4.sessionStart();
	window.__ga4.firstVisit();
	window.__ga4.pageView();
	
	// Send user engagement periodically
	setInterval(function() {
		if (!document.hidden) {
			window.__ga4.userEngagement();
		}
	}, 10000);
	
	// Send final engagement on page unload
	window.addEventListener('beforeunload', function() {
		window.__ga4.userEngagement();
	});
	
	console.log('[GA4 Deep] Integration initialized for ' + ga4Config.measurementId);
})();
`, g.MeasurementID, g.ClientID, g.UserID, g.SessionID, g.SessionNumber, g.FirstVisit, g.DebugMode,
		userPropsJSON, customDimsJSON, customMetricsJSON,
		g.ConsentMode.AdStorage, g.ConsentMode.AnalyticsStorage, g.ConsentMode.FunctionalityStorage,
		g.ConsentMode.PersonalizationStorage, g.ConsentMode.SecurityStorage,
		g.ConsentMode.AdUserData, g.ConsentMode.AdPersonalization)
}

// GenerateMeasurementProtocolPayload generates a Measurement Protocol payload
func (g *GA4DeepIntegration) GenerateMeasurementProtocolPayload(events []GA4Event) string {
	var eventStrings []string

	for _, event := range events {
		params := make([]string, 0)
		for k, v := range event.Params {
			params = append(params, fmt.Sprintf(`"%s":%v`, k, formatValue(v)))
		}
		eventStrings = append(eventStrings, fmt.Sprintf(`{"name":"%s","params":{%s}}`, event.Name, strings.Join(params, ",")))
	}

	payload := fmt.Sprintf(`{
		"client_id": "%s",
		"user_id": "%s",
		"timestamp_micros": "%d",
		"events": [%s]
	}`, g.ClientID, g.UserID, time.Now().UnixMicro(), strings.Join(eventStrings, ","))

	return payload
}

// GenerateMeasurementProtocolURL generates the Measurement Protocol URL
func (g *GA4DeepIntegration) GenerateMeasurementProtocolURL(apiSecret string) string {
	return fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s",
		url.QueryEscape(g.MeasurementID), url.QueryEscape(apiSecret))
}

// Helper functions
func generateClientID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d.%s", timestamp, hex.EncodeToString(bytes))
}

func generateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func mapToJSON(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf(`"%s":%s`, k, formatValue(v)))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func stringMapToInterface(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func floatMapToInterface(m map[string]float64) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(val, `"`, `\"`))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf(`"%v"`, val)
	}
}

// GenerateRandomEngagementTime generates a random but realistic engagement time
func GenerateRandomEngagementTime() int64 {
	// Between 10 seconds and 5 minutes
	n, _ := rand.Int(rand.Reader, big.NewInt(290000))
	return 10000 + n.Int64()
}

// GenerateRandomSessionNumber generates a random session number
func GenerateRandomSessionNumber(isReturning bool) int {
	if !isReturning {
		return 1
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(10))
	return int(n.Int64()) + 2
}
