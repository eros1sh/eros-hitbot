# ErosHit Metrics System

Real-time metrics collection and dashboard system for ErosHit.

## Features

- **Prometheus-compatible metrics**: All metrics are exposed in Prometheus format
- **WebSocket streaming**: Real-time metrics via WebSocket
- **Grafana dashboard**: Pre-built dashboard JSON export
- **HTTP API**: REST endpoints for metrics access

## Quick Start

### 1. Metrics Collector

```go
import "eroshit/pkg/metrics"

// Get the global collector
collector := metrics.GetGlobalCollector()

// Record metrics
collector.RecordHit()
collector.RecordResponseTime(duration)
collector.RecordSuccess(proxy)
collector.RecordFailure(proxy)
collector.SetActiveSessions(count)
collector.SetActiveProxies(count)
```

### 2. API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/metrics` | GET | Prometheus format metrics |
| `/api/metrics/json` | GET | JSON format metrics |
| `/api/metrics/stream` | WS | Real-time WebSocket stream |
| `/api/metrics/dashboard` | GET | Grafana dashboard JSON |

### 3. WebSocket Events

Connect to `/api/metrics/stream` to receive real-time events:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/metrics/stream');

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log(msg.type, msg.data);
};
```

#### Event Types

- `metrics:hit` - Hit completed
- `metrics:proxy_status` - Proxy status changed
- `metrics:performance` - Performance metrics update
- `metrics:session` - Session event
- `metrics:snapshot` - Initial full snapshot

#### Subscribe to specific events

```javascript
// Subscribe to specific event types
const ws = new WebSocket('ws://localhost:8080/api/metrics/stream?type=metrics:hit&type=metrics:performance');
```

### 4. Prometheus Integration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'eroshit'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/metrics'
    scrape_interval: 5s
```

### 5. Grafana Setup

1. Download dashboard JSON:
   ```bash
   curl http://localhost:8080/api/metrics/dashboard -o eroshit-dashboard.json
   ```

2. Import to Grafana:
   - Go to Dashboards → Import
   - Upload JSON file
   - Select Prometheus datasource

## Metrics Reference

### Counters

| Metric | Description |
|--------|-------------|
| `eroshit_hits_total` | Total number of hits |
| `eroshit_proxy_success_total{proxy}` | Successful requests per proxy |
| `eroshit_proxy_failure_total{proxy}` | Failed requests per proxy |

### Gauges

| Metric | Description |
|--------|-------------|
| `eroshit_hit_rate_per_minute` | Current hit rate per minute |
| `eroshit_active_sessions` | Number of active sessions |
| `eroshit_active_proxies` | Number of active proxies |
| `eroshit_queue_size` | Current queue size |
| `eroshit_success_rate` | Success rate (0-1) |
| `eroshit_bounce_rate` | Bounce rate (0-1) |
| `eroshit_error_rate` | Error rate (0-1) |

### Histograms

| Metric | Description |
|--------|-------------|
| `eroshit_response_time_seconds` | Response time distribution |
| `eroshit_proxy_latency_seconds{proxy}` | Proxy latency per proxy |

## Example: Custom Dashboard Widget

```html
<div id="metrics">
  <div>Hits: <span id="hits">0</span></div>
  <div>Hit Rate: <span id="rate">0</span>/min</div>
  <div>Success Rate: <span id="success">0</span>%</div>
</div>

<script>
const ws = new WebSocket('ws://localhost:8080/api/metrics/stream');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  
  if (msg.type === 'metrics:snapshot') {
    document.getElementById('hits').textContent = msg.data.total_hits;
    document.getElementById('rate').textContent = msg.data.hit_rate_per_min.toFixed(1);
    document.getElementById('success').textContent = (msg.data.success_rate * 100).toFixed(1);
  }
};
</script>
```

## Integration with Simulator

```go
import "eroshit/pkg/metrics"

// In your simulator code
func (s *Simulator) recordHit(result HitResult) {
    hooks := metrics.NewSimulatorHooks(s.metricsCollector)
    
    timer := hooks.StartTimer(result.Proxy)
    defer timer.Stop(result.Success)
    
    // ... perform hit ...
}
```
