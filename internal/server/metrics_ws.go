package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"eroshit/pkg/metrics"

	"github.com/gorilla/websocket"
)

// MetricsWebSocket handles real-time metrics streaming via WebSocket
type MetricsWebSocket struct {
	collector   *metrics.MetricsCollector
	hub         *MetricsHub
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
	broadcastCh chan MetricsEvent
}

// MetricsHub manages WebSocket connections for metrics
type MetricsHub struct {
	mu       sync.RWMutex
	conns    map[*websocket.Conn]chan MetricsEvent
	typeSubs map[string]map[*websocket.Conn]bool // Event type -> connections
}

// MetricsEvent represents a WebSocket event
type MetricsEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// HitEvent data for metrics:hit
type HitEvent struct {
	URL          string        `json:"url"`
	Proxy        string        `json:"proxy"`
	ResponseTime time.Duration `json:"response_time_ms"`
	Success      bool          `json:"success"`
	SessionID    string        `json:"session_id,omitempty"`
}

// ProxyStatusEvent data for metrics:proxy_status
type ProxyStatusEvent struct {
	Proxy     string  `json:"proxy"`
	Status    string  `json:"status"` // "active", "failed", "removed"
	Latency   float64 `json:"latency_ms,omitempty"`
	FailCount int     `json:"fail_count,omitempty"`
}

// PerformanceEvent data for metrics:performance
type PerformanceEvent struct {
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	SuccessRate     float64 `json:"success_rate"`
	ErrorRate       float64 `json:"error_rate"`
	HitRate         float64 `json:"hit_rate_per_min"`
	ActiveSessions  int64   `json:"active_sessions"`
}

// SessionEvent data for metrics:session
type SessionEvent struct {
	SessionID    string        `json:"session_id"`
	Action       string        `json:"action"` // "started", "ended", "bounce"
	Duration     time.Duration `json:"duration_ms,omitempty"`
	PagesVisited int           `json:"pages_visited,omitempty"`
}

// NewMetricsHub creates a new metrics hub
func NewMetricsHub() *MetricsHub {
	return &MetricsHub{
		conns:    make(map[*websocket.Conn]chan MetricsEvent),
		typeSubs: make(map[string]map[*websocket.Conn]bool),
	}
}

// Register registers a new connection
func (h *MetricsHub) Register(conn *websocket.Conn, eventTypes []string) chan MetricsEvent {
	ch := make(chan MetricsEvent, 128)
	h.mu.Lock()
	h.conns[conn] = ch

	// Register for specific event types
	if len(eventTypes) > 0 {
		for _, et := range eventTypes {
			if h.typeSubs[et] == nil {
				h.typeSubs[et] = make(map[*websocket.Conn]bool)
			}
			h.typeSubs[et][conn] = true
		}
	} else {
		// Subscribe to all event types
		allTypes := []string{"metrics:hit", "metrics:proxy_status", "metrics:performance", "metrics:session"}
		for _, et := range allTypes {
			if h.typeSubs[et] == nil {
				h.typeSubs[et] = make(map[*websocket.Conn]bool)
			}
			h.typeSubs[et][conn] = true
		}
	}
	h.mu.Unlock()
	return ch
}

// Unregister removes a connection
func (h *MetricsHub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	if ch, ok := h.conns[conn]; ok {
		// Remove from type subscriptions
		for _, subs := range h.typeSubs {
			delete(subs, conn)
		}
		close(ch)
		delete(h.conns, conn)
	}
	h.mu.Unlock()
}

// Broadcast sends event to all subscribed connections
func (h *MetricsHub) Broadcast(event MetricsEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	// Get connections subscribed to this event type
	subs := h.typeSubs[event.Type]
	h.mu.RUnlock()

	// Send to subscribed connections
	for conn := range subs {
		h.mu.RLock()
		ch, ok := h.conns[conn]
		h.mu.RUnlock()
		if !ok {
			continue
		}

		select {
		case ch <- event:
		default:
			// Channel full, drop message
		}
	}

	_ = payload // Use payload for direct write if needed
}

// ConnectionCount returns number of active connections
func (h *MetricsHub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

// NewMetricsWebSocket creates a new metrics WebSocket handler
func NewMetricsWebSocket(collector *metrics.MetricsCollector) *MetricsWebSocket {
	mws := &MetricsWebSocket{
		collector:   collector,
		hub:         NewMetricsHub(),
		broadcastCh: make(chan MetricsEvent, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow same-origin requests
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				allowedOrigins := []string{
					"http://127.0.0.1",
					"http://localhost",
					"https://127.0.0.1",
					"https://localhost",
				}
				for _, allowed := range allowedOrigins {
					if len(origin) >= len(allowed) && origin[:len(allowed)] == allowed {
						return true
					}
				}
				return false
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	// Start broadcaster
	go mws.broadcaster()

	// Start periodic performance updates
	go mws.periodicUpdates()

	return mws
}

// HandleWebSocket upgrades HTTP to WebSocket and handles the connection
func (mws *MetricsWebSocket) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse event types from query params
	eventTypes := r.URL.Query()["type"]

	conn, err := mws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ch := mws.hub.Register(conn, eventTypes)
	defer mws.hub.Unregister(conn)

	// Send initial snapshot
	snapshot := mws.collector.GetSnapshot()
	initEvent := MetricsEvent{
		Type:      "metrics:snapshot",
		Timestamp: time.Now(),
		Data:      snapshot,
	}
	if err := conn.WriteJSON(initEvent); err != nil {
		return
	}

	// Writer goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range ch {
			if err := conn.WriteJSON(event); err != nil {
				return
			}
		}
	}()

	// Reader loop (for ping/pong and subscriptions)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Handle subscription changes
		var cmd struct {
			Action string   `json:"action"`
			Types  []string `json:"types"`
		}
		if err := json.Unmarshal(msg, &cmd); err == nil {
			switch cmd.Action {
			case "subscribe":
				mws.updateSubscription(conn, cmd.Types)
			case "ping":
				conn.WriteJSON(MetricsEvent{
					Type:      "pong",
					Timestamp: time.Now(),
				})
			}
		}
	}

	<-done
}

// updateSubscription updates event type subscriptions for a connection
func (mws *MetricsWebSocket) updateSubscription(conn *websocket.Conn, eventTypes []string) {
	mws.hub.mu.Lock()
	defer mws.hub.mu.Unlock()

	// Remove from all current subscriptions
	for _, subs := range mws.hub.typeSubs {
		delete(subs, conn)
	}

	// Add to new subscriptions
	for _, et := range eventTypes {
		if mws.hub.typeSubs[et] == nil {
			mws.hub.typeSubs[et] = make(map[*websocket.Conn]bool)
		}
		mws.hub.typeSubs[et][conn] = true
	}
}

// broadcaster sends events from broadcastCh to hub
func (mws *MetricsWebSocket) broadcaster() {
	for event := range mws.broadcastCh {
		mws.hub.Broadcast(event)
	}
}

// periodicUpdates sends periodic performance updates
func (mws *MetricsWebSocket) periodicUpdates() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		snapshot := mws.collector.GetSnapshot()
		mws.BroadcastPerformance(PerformanceEvent{
			AvgResponseTime: 0, // TODO: Calculate from collector histogram
			SuccessRate:     snapshot.SuccessRate,
			ErrorRate:       snapshot.ErrorRate,
			HitRate:         snapshot.HitRatePerMin,
			ActiveSessions:  snapshot.ActiveSessions,
		})
	}
}

// BroadcastHit broadcasts a hit event
func (mws *MetricsWebSocket) BroadcastHit(event HitEvent) {
	mws.broadcastCh <- MetricsEvent{
		Type:      "metrics:hit",
		Timestamp: time.Now(),
		Data:      event,
	}
}

// BroadcastProxyStatus broadcasts a proxy status event
func (mws *MetricsWebSocket) BroadcastProxyStatus(event ProxyStatusEvent) {
	mws.broadcastCh <- MetricsEvent{
		Type:      "metrics:proxy_status",
		Timestamp: time.Now(),
		Data:      event,
	}
}

// BroadcastPerformance broadcasts a performance event
func (mws *MetricsWebSocket) BroadcastPerformance(event PerformanceEvent) {
	mws.broadcastCh <- MetricsEvent{
		Type:      "metrics:performance",
		Timestamp: time.Now(),
		Data:      event,
	}
}

// BroadcastSession broadcasts a session event
func (mws *MetricsWebSocket) BroadcastSession(event SessionEvent) {
	mws.broadcastCh <- MetricsEvent{
		Type:      "metrics:session",
		Timestamp: time.Now(),
		Data:      event,
	}
}

// ConnectionCount returns number of connected clients
func (mws *MetricsWebSocket) ConnectionCount() int {
	return mws.hub.ConnectionCount()
}

// Close cleans up resources
func (mws *MetricsWebSocket) Close() {
	close(mws.broadcastCh)
}

// ========================================
// HTTP Handlers
// ========================================

// MetricsHandler returns Prometheus format metrics
func MetricsHandler(collector *metrics.MetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		collector.MetricsHandler().ServeHTTP(w, r)
	}
}

// MetricsJSONHandler returns metrics in JSON format
func MetricsJSONHandler(collector *metrics.MetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(collector.GetSnapshot())
	}
}

// DashboardHandler returns Grafana dashboard JSON
func DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		dashboard := generateGrafanaDashboard()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=eroshit-dashboard.json")
		json.NewEncoder(w).Encode(dashboard)
	}
}

// generateGrafanaDashboard generates Grafana dashboard JSON
func generateGrafanaDashboard() map[string]interface{} {
	return map[string]interface{}{
		"annotations": map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"builtIn": 1,
					"datasource": map[string]string{
						"type": "grafana",
						"uid":  "-- Grafana --",
					},
					"enable": true,
					"hide":   true,
					"iconColor": "rgba(0, 211, 255, 1)",
					"name": "Annotations & Alerts",
					"type": "dashboard",
				},
			},
		},
		"editable":             true,
		"fiscalYearStartMonth": 0,
		"graphTooltip":         0,
		"id":                   nil,
		"links":                []interface{}{},
		"liveNow":              false,
		"panels": []map[string]interface{}{
			{
				"collapsed": false,
				"gridPos": map[string]int{
					"h": 1,
					"w": 24,
					"x": 0,
					"y": 0,
				},
				"id":   1,
				"panels": []interface{}{},
				"title": "Overview",
				"type":  "row",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "thresholds",
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "green", "value": nil},
							},
						},
						"unit": "none",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 4,
					"w": 4,
					"x": 0,
					"y": 1,
				},
				"id": 2,
				"options": map[string]interface{}{
					"colorMode":   "value",
					"graphMode":   "area",
					"justifyMode": "auto",
					"orientation": "auto",
					"reduceOptions": map[string]interface{}{
						"calcs":  []string{"lastNotNull"},
						"fields": "",
						"values": false,
					},
					"textMode": "auto",
				},
				"pluginVersion": "10.0.0",
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_hits_total",
						"refId":        "A",
						"legendFormat": "Total Hits",
					},
				},
				"title": "Total Hits",
				"type":  "stat",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "palette-classic",
						},
						"custom": map[string]interface{}{
							"axisCenteredZero": false,
							"axisColorMode":    "text",
							"axisLabel":        "",
							"axisPlacement":    "auto",
							"barAlignment":     0,
							"drawStyle":        "line",
							"fillOpacity":      20,
							"gradientMode":     "opacity",
							"hideFrom": map[string]bool{
								"legend":  false,
								"tooltip": false,
								"viz":     false,
							},
							"lineInterpolation": "smooth",
							"lineWidth":         2,
							"pointSize":         5,
							"scaleDistribution": map[string]string{
								"type": "linear",
							},
							"showPoints": "never",
							"spanNulls":  false,
							"stacking": map[string]interface{}{
								"group": "A",
								"mode":  "none",
							},
							"thresholdsStyle": map[string]string{
								"mode": "off",
							},
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "green", "value": nil},
							},
						},
						"unit": "hpm",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 8,
					"w": 8,
					"x": 4,
					"y": 1,
				},
				"id": 3,
				"options": map[string]interface{}{
					"legend": map[string]interface{}{
						"calcs":       []string{},
						"displayMode": "list",
						"placement":   "bottom",
						"showLegend":  true,
					},
					"tooltip": map[string]interface{}{
						"mode": "single",
						"sort": "none",
					},
				},
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_hit_rate_per_minute",
						"refId":        "A",
						"legendFormat": "Hits/min",
					},
				},
				"title": "Hit Rate",
				"type":  "timeseries",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "thresholds",
						},
						"mappings": []interface{}{},
						"max":     1,
						"min":     0,
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "red", "value": nil},
								{"color": "yellow", "value": 0.5},
								{"color": "green", "value": 0.8},
							},
						},
						"unit": "percentunit",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 4,
					"w": 4,
					"x": 12,
					"y": 1,
				},
				"id": 4,
				"options": map[string]interface{}{
					"orientation": "auto",
					"reduceOptions": map[string]interface{}{
						"calcs":  []string{"lastNotNull"},
						"fields": "",
						"values": false,
					},
					"showThresholdLabels": false,
					"showThresholdMarkers": true,
				},
				"pluginVersion": "10.0.0",
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_success_rate",
						"refId":        "A",
						"legendFormat": "Success Rate",
					},
				},
				"title": "Success Rate",
				"type":  "gauge",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "thresholds",
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "blue", "value": nil},
							},
						},
						"unit": "none",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 4,
					"w": 4,
					"x": 16,
					"y": 1,
				},
				"id": 5,
				"options": map[string]interface{}{
					"colorMode":   "value",
					"graphMode":   "area",
					"justifyMode": "auto",
					"orientation": "auto",
					"reduceOptions": map[string]interface{}{
						"calcs":  []string{"lastNotNull"},
						"fields": "",
						"values": false,
					},
					"textMode": "auto",
				},
				"pluginVersion": "10.0.0",
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_active_sessions",
						"refId":        "A",
						"legendFormat": "Active Sessions",
					},
				},
				"title": "Active Sessions",
				"type":  "stat",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "thresholds",
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "purple", "value": nil},
							},
						},
						"unit": "none",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 4,
					"w": 4,
					"x": 20,
					"y": 1,
				},
				"id": 6,
				"options": map[string]interface{}{
					"colorMode":   "value",
					"graphMode":   "area",
					"justifyMode": "auto",
					"orientation": "auto",
					"reduceOptions": map[string]interface{}{
						"calcs":  []string{"lastNotNull"},
						"fields": "",
						"values": false,
					},
					"textMode": "auto",
				},
				"pluginVersion": "10.0.0",
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_active_proxies",
						"refId":        "A",
						"legendFormat": "Active Proxies",
					},
				},
				"title": "Active Proxies",
				"type":  "stat",
			},
			{
				"collapsed": false,
				"gridPos": map[string]int{
					"h": 1,
					"w": 24,
					"x": 0,
					"y": 9,
				},
				"id":   10,
				"panels": []interface{}{},
				"title": "Performance",
				"type":  "row",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "palette-classic",
						},
						"custom": map[string]interface{}{
							"axisCenteredZero": false,
							"axisColorMode":    "text",
							"axisLabel":        "",
							"axisPlacement":    "auto",
							"barAlignment":     0,
							"drawStyle":        "line",
							"fillOpacity":      20,
							"gradientMode":     "opacity",
							"hideFrom": map[string]bool{
								"legend":  false,
								"tooltip": false,
								"viz":     false,
							},
							"lineInterpolation": "smooth",
							"lineWidth":         2,
							"pointSize":         5,
							"scaleDistribution": map[string]string{
								"type": "linear",
							},
							"showPoints": "never",
							"spanNulls":  false,
							"stacking": map[string]interface{}{
								"group": "A",
								"mode":  "none",
							},
							"thresholdsStyle": map[string]string{
								"mode": "off",
							},
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "green", "value": nil},
							},
						},
						"unit": "s",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 8,
					"w": 12,
					"x": 0,
					"y": 10,
				},
				"id": 11,
				"options": map[string]interface{}{
					"legend": map[string]interface{}{
						"calcs":       []string{},
						"displayMode": "list",
						"placement":   "bottom",
						"showLegend":  true,
					},
					"tooltip": map[string]interface{}{
						"mode": "single",
						"sort": "none",
					},
				},
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "histogram_quantile(0.95, sum(rate(eroshit_response_time_seconds_bucket[5m])) by (le))",
						"refId":        "A",
						"legendFormat": "p95 Response Time",
					},
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "histogram_quantile(0.50, sum(rate(eroshit_response_time_seconds_bucket[5m])) by (le))",
						"refId":        "B",
						"legendFormat": "p50 Response Time",
					},
				},
				"title": "Response Time Distribution",
				"type":  "timeseries",
			},
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"fieldConfig": map[string]interface{}{
					"defaults": map[string]interface{}{
						"color": map[string]string{
							"mode": "palette-classic",
						},
						"custom": map[string]interface{}{
							"axisCenteredZero": false,
							"axisColorMode":    "text",
							"axisLabel":        "",
							"axisPlacement":    "auto",
							"barAlignment":     0,
							"drawStyle":        "line",
							"fillOpacity":      20,
							"gradientMode":     "opacity",
							"hideFrom": map[string]bool{
								"legend":  false,
								"tooltip": false,
								"viz":     false,
							},
							"lineInterpolation": "smooth",
							"lineWidth":         2,
							"pointSize":         5,
							"scaleDistribution": map[string]string{
								"type": "linear",
							},
							"showPoints": "never",
							"spanNulls":  false,
							"stacking": map[string]interface{}{
								"group": "A",
								"mode":  "none",
							},
							"thresholdsStyle": map[string]string{
								"mode": "off",
							},
						},
						"mappings": []interface{}{},
						"thresholds": map[string]interface{}{
							"mode": "absolute",
							"steps": []map[string]interface{}{
								{"color": "green", "value": nil},
							},
						},
						"unit": "percentunit",
					},
					"overrides": []interface{}{},
				},
				"gridPos": map[string]int{
					"h": 8,
					"w": 12,
					"x": 12,
					"y": 10,
				},
				"id": 12,
				"options": map[string]interface{}{
					"legend": map[string]interface{}{
						"calcs":       []string{},
						"displayMode": "list",
						"placement":   "bottom",
						"showLegend":  true,
					},
					"tooltip": map[string]interface{}{
						"mode": "single",
						"sort": "none",
					},
				},
				"targets": []map[string]interface{}{
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_error_rate",
						"refId":        "A",
						"legendFormat": "Error Rate",
					},
					{
						"datasource": map[string]string{
							"type": "prometheus",
							"uid":  "${datasource}",
						},
						"expr":         "eroshit_bounce_rate",
						"refId":        "B",
						"legendFormat": "Bounce Rate",
					},
				},
				"title": "Error & Bounce Rates",
				"type":  "timeseries",
			},
		},
		"refresh": "5s",
		"schemaVersion": 38,
		"style":         "dark",
		"tags":          []string{"eroshit", "metrics", "dashboard"},
		"templating": map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"current": map[string]interface{}{
						"selected": false,
						"text":     "Prometheus",
						"value":    "Prometheus",
					},
					"hide":       0,
					"includeAll": false,
					"label":      "Data Source",
					"multi":      false,
					"name":       "datasource",
					"options":    []interface{}{},
					"query":      "prometheus",
					"refresh":    1,
					"regex":      "",
					"skipUrlSync": false,
					"type":       "datasource",
				},
			},
		},
		"time": map[string]interface{}{
			"from": "now-1h",
			"to":   "now",
		},
		"timepicker": map[string]interface{}{
			"refresh_intervals": []string{
				"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d",
			},
			"time_options": []string{
				"5m", "15m", "1h", "6h", "12h", "24h", "2d", "7d", "30d",
			},
		},
		"timezone": "",
		"title":    "ErosHit Metrics Dashboard",
		"uid":      "eroshit-metrics",
		"version":  1,
		"weekStart": "",
	}
}
