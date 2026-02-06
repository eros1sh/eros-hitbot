# Eros Hit Bot

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-2.4.0-red)](https://github.com/eros1sh/eros-hitbot/releases)
[![Dashboard](https://img.shields.io/badge/Dashboard-Live-brightgreen)](http://localhost:8754)
[![Prometheus](https://img.shields.io/badge/Metrics-Prometheus-orange)](https://prometheus.io)

> **💼 Premium Version & Rental Services Available**  
> **🇹🇷 Ücretli Sürüm ve Kiralama Hizmetleri Mevcut**
> 
> **English:** For premium version, rental services, or specialized solutions (including betting sites), please contact me via Telegram.  
> **Türkçe:** Ücretli sürüm, kiralama hizmetleri veya özel çözümler (bahis siteleri dahil) için lütfen Telegram üzerinden benimle iletişime geçin.
> 
> **📱 Telegram:** Contact me via Telegram for inquiries

**Parasitic SEO traffic simulation tool** — Simulate organic search traffic, boost search engine rankings, and verify analytics (GA4/GTM) through realistic, keyword-driven web visits.

<p align="center">
  <img src="assets/web.png" alt="Web Interface" width="600"/>
</p>

---

## Table of Contents

- [Overview](#overview)
- [What's New in v2.4.0](#-whats-new-in-v240)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Feature Details](#feature-details)
- [API Documentation](#api-documentation)
- [Performance Benchmarks](#performance-benchmarks)
- [Changelog](#changelog)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Eros Hit Bot is an open-source **parasitic SEO** tool that generates simulated organic traffic to improve your website's search engine rankings. By mimicking real user behavior—including search referrers, varied fingerprints, and analytics events—it helps your target pages accumulate engagement signals that search engines interpret as organic interest.

### How It Works

1. **Keyword-driven traffic** — You define target keywords; the bot simulates visits originating from Google, Bing, and other search engines.
2. **Realistic fingerprinting** — Each visit uses unique user agents, screen sizes, timezones, and canvas/WebGL fingerprints to avoid detection.
3. **Analytics integration** — GA4/GTM events (page views, scrolls, clicks) are triggered so your analytics reflect the traffic.
4. **Parallel execution** — Multiple browser contexts run concurrently for higher throughput.

---

## 🚀 What's New in v2.4.0

### ✨ New Features

- **Real-time Dashboard & Metrics** - Prometheus-compatible metrics with live WebSocket streaming
- **Advanced Session Management** - Cookie/LocalStorage/IndexedDB persistence with returning visitor simulation
- **Smart Proxy Rotation** - 7 selection strategies (weighted, fastest, success-rate, geo-based)
- **Behavioral Fingerprint Randomization** - Unique visitor profiles for realistic behavior
- **Structured Logging (Zap)** - High-performance JSON logging with rotation
- **Distributed Mode (Master-Worker)** - Scale across multiple machines
- **Configuration Hot-Reload** - Auto-reload config without restart
- **HTTP/3 QUIC Support** - Next-generation protocol for faster connections with 0-RTT support
- **Advanced Connection Pooling** - Keep-alive optimization with connection reuse and HTTP/2 support
- **TCP Fast Open** - Reduce TCP handshake latency (Linux-only optimization)
- **CPU Affinity** - Pin threads to specific CPU cores for better cache utilization
- **NUMA Awareness** - NUMA-aware memory management for high-end servers
- **VM Fingerprint Spoofing** - Hide VirtualBox, VMware, Hyper-V traces and spoof hardware IDs

### 🔧 Improvements

- **Browser Pool Pattern** - ~30x faster visit initiation through pre-allocated Chrome instances
- **Memory Optimization** - sync.Pool usage for reduced GC pressure
- **Adaptive Rate Limiting** - Token bucket algorithm with dynamic adjustment
- **Circuit Breaker Pattern** - Automatic failure detection and recovery

### 🛡️ Enhanced Startup Flow

- **Automatic System Detection** — On startup, the application detects your system specifications (CPU, RAM, disk, GPU) using neofetch-style display
- **Smart Optimization Profiles** — Based on detected hardware, the system generates optimized settings profiles (Low, Medium, High, Ultra)
- **User Choice** — After language selection and system detection, users can choose between:
  - **Recommended Settings** — Apply auto-optimized settings based on system capabilities
  - **Manual Configuration** — Open web interface with default settings for manual configuration

### 🌐 Complete i18n Support

- **Full Localization** — All text moved to the i18n system with complete multi-language translations (Turkish and English supported)
- **Consistent Language Experience** — Selected language is applied throughout the entire application lifecycle

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ErosHit v2.4.0                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        Web Interface (Port 8754)                     │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐ │   │
│  │  │   Metrics   │ │   Config    │ │   Proxy     │ │   Dashboard     │ │   │
│  │  │   (WebSocket)│ │ (Hot-Reload)│ │  (7 Modes)  │ │  (Real-time)   │ │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        API Layer                                     │   │
│  │     REST API    │    WebSocket    │   Prometheus    │    SSE       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Core Engine                                      │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────────────┐ │   │
│  │  │  Browser  │  │  Session  │  │   Proxy   │  │  Rate Limiter     │ │   │
│  │  │   Pool    │  │  Manager  │  │  Selector │  │ (Token Bucket)    │ │   │
│  │  │(~30x perf)│  │(Persistence)│ │(7 strategies)│  (Adaptive)      │ │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────────────┘ │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────────────┐ │   │
│  │  │  Circuit  │  │  Behavior │  │   TLS     │  │   Zap Logger      │ │   │
│  │  │  Breaker  │  │  Profile  │  │  Fingerprint│  │ (JSON/Console)   │ │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Chrome/CDP Layer                                  │   │
│  │         Headless Chrome Instances (Pre-allocated Pool)              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                   Target Website                                     │   │
│  │         GA4/GTM Events │ Scroll │ Clicks │ Session Simulation        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         Distributed Mode (Optional)                          │
│  ┌──────────────┐                      ┌──────────────┐                     │
│  │    Master    │◄────────────────────►│   Worker 1   │                     │
│  │   (Coordinator)                     │ (Browser)    │                     │
│  └──────────────┘                      └──────────────┘                     │
│         ▲                              ┌──────────────┐                     │
│         └──────────────────────────────►│   Worker 2   │                     │
│         (Task Distribution)             │ (Browser)    │                     │
│                                         └──────────────┘                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Features

| Feature | Description | Version |
|---------|-------------|---------|
| **Parasitic SEO** | Simulate organic search traffic to boost rankings | 1.0+ |
| **Keyword targeting** | Define custom keywords; traffic appears to come from search | 1.0+ |
| **GSC Integration** | Fetch real queries from Google Search Console API | 2.3.0+ |
| **Multi-browser** | Up to 50 concurrent headless Chrome instances | 1.0+ |
| **Proxy support** | HTTP proxy with authentication (rotation ready) | 1.0+ |
| **Multi-proxy** | Add multiple private proxies with rotation modes | 2.3.0+ |
| **Public proxy** | Auto-fetch and test proxies from public lists | 2.0+ |
| **GA4/GTM** | Automatic page_view, scroll, session_start, user_engagement events | 2.0+ |
| **Custom Dimensions** | Send custom GA4 dimensions and metrics | 2.3.0+ |
| **Headless bypass** | Comprehensive stealth techniques to reduce bot detection | 2.0+ |
| **Canvas fingerprinting** | Unique canvas/WebGL/Audio noise per visit | 2.0+ |
| **TLS Fingerprinting** | JA3/JA4 fingerprint randomization | 2.3.0+ |
| **Device emulation** | Mobile, tablet, desktop with brand-specific profiles | 2.2.0+ |
| **Geo targeting** | Country, timezone, and language spoofing | 2.2.0+ |
| **Bounce Rate Control** | Target specific bounce rate with multi-page visits | 2.3.0+ |
| **Session Depth** | Simulate realistic session depth (2-5 pages) | 2.3.0+ |
| **Returning Visitors** | Simulate returning visitors with persistent client_id | 2.3.0+ |
| **Exit Page Control** | Define which pages end sessions | 2.3.0+ |
| **Browser Profiles** | Persist browser profiles, cookies, localStorage | 2.3.0+ |
| **Sitemap.xml** | Optional: use sitemap URLs + weighted homepage traffic | 2.0+ |
| **i18n** | Multi-language UI + logs (Turkish and English) | 2.1.0+ |
| **Reports** | CSV, JSON, and HTML dashboard export | 1.0+ |
| **Quality metrics** | Real-time quality scoring (A+ to F grade) | 2.2.0+ |
| **Real-time Metrics** | Prometheus-compatible metrics with WebSocket streaming | **2.4.0** |
| **Session Persistence** | Cookie/LocalStorage/IndexedDB with encryption | **2.4.0** |
| **Smart Proxy Rotation** | 7 selection strategies | **2.4.0** |
| **Behavior Profiles** | 7 predefined visitor behavior patterns | **2.4.0** |
| **Structured Logging** | Zap-based JSON logging with rotation | **2.4.0** |
| **Distributed Mode** | Master-Worker architecture for scaling | **2.4.0** |
| **Hot-Reload** | Auto-reload config without restart | **2.4.0** |
| **Browser Pool** | ~30x faster visit initiation | **2.4.0** |
| **Circuit Breaker** | Automatic failure detection and recovery | **2.4.0** |

### Screenshots

<p align="center">
  <img src="assets/cli.png" alt="CLI Mode" width="450"/>
  <br/>
  <em>CLI mode with rainbow banner</em>
</p>

<p align="center">
  <img src="assets/test.png" alt="Test Run" width="450"/>
  <br/>
  <em>Simulation metrics and log output</em>
</p>

---

## Installation

### Option 1: Download Pre-built Binary (Recommended)

1. Go to [Releases](https://github.com/eros1sh/eros-hitbot/releases)
2. Download the appropriate build for your platform:

| Platform | File | Notes |
|----------|------|-------|
| Windows (64-bit) | `eros-hitbot-windows-amd64.zip` | Includes custom icon |
| macOS (Intel) | `eros-hitbot-darwin-amd64.zip` | For Intel Macs |
| macOS (Apple Silicon) | `eros-hitbot-darwin-arm64.zip` | For M1/M2/M3 Macs |
| Linux (64-bit) | `eros-hitbot-linux-amd64.tar.gz` | For x86_64 systems |
| Linux (ARM64) | `eros-hitbot-linux-arm64.tar.gz` | For ARM64/aarch64 systems |

3. Extract and run:

**Windows:**
```bash
# Extract zip and double-click eros-hitbot-windows-amd64.exe
```

**macOS/Linux:**
```bash
# Extract and make executable
unzip eros-hitbot-darwin-amd64.zip  # or tar -xzf for Linux
chmod +x eros-hitbot-*
./eros-hitbot-darwin-amd64  # or linux-amd64
```

### Option 2: Build from Source

**Requirements:** Go 1.21+, Chrome/Chromium (for headless mode)

```bash
git clone https://github.com/eros1sh/eros-hitbot.git
cd eros-hitbot
go build -o eroshit ./cmd/eroshit  # Add .exe on Windows
```

---

## Quick Start

### 1. Prepare Configuration

Copy the example config and customize:

```bash
cp config.example.json config.json
cp agents.json.example agents.json      # Optional: add more user agents
cp operaagent.json.example operaagent.json  # Optional
```

### 2. Edit `config.json`

```json
{
  "targetDomain": "your-site.com",
  "fallbackGAID": "G-XXXXXXXXXX",
  "keywords": ["your target keyword", "long tail keyword"],
  "maxPages": 5,
  "durationMinutes": 60,
  "hitsPerMinute": 35,
  "maxConcurrentVisits": 10,
  "deviceType": "mixed",
  "deviceBrands": ["apple", "samsung", "windows"],
  "referrerKeyword": "your brand name",
  "referrerEnabled": true,
  "targetBounceRate": 35,
  "sessionMinPages": 2,
  "sessionMaxPages": 5,
  "returningVisitorRate": 30,
  "enableBrowserProfile": true,
  "enableJa3Randomization": true,
  "enableJa4Randomization": true
}
```

### 3. Run

**Web UI (default):**
```bash
./eroshit
```
Then open `http://127.0.0.1:8754` in your browser.

**CLI mode:**
```bash
./eroshit -cli -domain your-site.com -pages 5 -duration 60 -hpm 35
```

---

## Configuration

### Basic Settings

| Field | Description | Default |
|-------|-------------|---------|
| `targetDomain` | Domain to send traffic to | `example.com` |
| `fallbackGAID` | GA4 Measurement ID | `G-XXXXXXXXXX` |
| `maxPages` | Max pages to crawl per session | 5 |
| `durationMinutes` | Simulation duration (minutes) | 60 |
| `hitsPerMinute` | Request rate (HPM) | 35 |
| `maxConcurrentVisits` | Parallel browser tabs | 10 |
| `outputDir` | Report output directory | `./reports` |
| `exportFormat` | `csv`, `json`, `html`, or `both` | `both` |

### Traffic Simulation

| Field | Description | Default |
|-------|-------------|---------|
| `deviceType` | `desktop`, `mobile`, `tablet`, or `mixed` | `mixed` |
| `deviceBrands` | Array of brands to use | `[]` (all) |
| `scrollStrategy` | `gradual`, `fast`, `reader` | `gradual` |
| `canvasFingerprint` | Enable canvas/WebGL noise | `true` |
| `sendScrollEvent` | Send GA4 scroll events | `true` |
| `minPageDuration` | Minimum page duration (seconds) | 15 |
| `maxPageDuration` | Maximum page duration (seconds) | 120 |
| `minScrollPercent` | Minimum scroll percentage | 25 |
| `maxScrollPercent` | Maximum scroll percentage | 100 |
| `clickProbability` | Click probability (0-100) | 30 |

### Session & Bounce Control

| Field | Description | Default |
|-------|-------------|---------|
| `targetBounceRate` | Target bounce rate percentage | 35 |
| `enableBounceControl` | Enable bounce rate control | `true` |
| `sessionMinPages` | Minimum pages per session | 2 |
| `sessionMaxPages` | Maximum pages per session | 5 |
| `enableSessionDepth` | Enable session depth simulation | `true` |

### Returning Visitor & Exit Pages

| Field | Description | Default |
|-------|-------------|---------|
| `returningVisitorRate` | Returning visitor percentage | 30 |
| `returningVisitorDays` | Days between return visits | 7 |
| `enableReturningVisitor` | Enable returning visitor simulation | `true` |
| `exitPages` | Array of exit page URL patterns | `[]` |
| `enableExitPageControl` | Enable exit page control | `false` |

### Browser Profile Persistence

| Field | Description | Default |
|-------|-------------|---------|
| `browserProfilePath` | Profile storage directory | `./browser_profiles` |
| `maxBrowserProfiles` | Maximum stored profiles | 100 |
| `enableBrowserProfile` | Enable profile persistence | `false` |
| `persistCookies` | Persist cookies | `true` |
| `persistLocalStorage` | Persist localStorage | `true` |

### TLS Fingerprint

| Field | Description | Default |
|-------|-------------|---------|
| `tlsFingerprintMode` | `random`, `chrome`, `firefox`, `safari`, `edge` | `random` |
| `enableJa3Randomization` | Enable JA3 randomization | `true` |
| `enableJa4Randomization` | Enable JA4 randomization | `true` |

### Smart Proxy Rotation (v2.4.0)

| Field | Description | Default |
|-------|-------------|---------|
| `proxyRotationMode` | Selection strategy | `weighted` |
| `proxyRotationInterval` | Requests between rotation | 1 |
| `enableProxyRotation` | Enable rotation | `true` |
| `geoCountries` | Preferred countries (geo mode) | `[]` |

**Available Rotation Modes:**
- `round-robin` — Selects proxies in sequence
- `random` — Randomly selects proxies
- `least-used` — Selects the least used proxy
- `fastest` — Selects the fastest response time proxy
- `success-rate` — Selects the highest success rate proxy
- `geo` — Selects from specific countries
- `weighted` — Combines all metrics (recommended)

### Behavioral Profiles (v2.4.0)

| Profile Type | Description | Use Case |
|--------------|-------------|----------|
| `fast_reader` | Fast reader, short dwell time | General use |
| `detailed_reader` | Detailed reader, long dwell time | Blog content |
| `shopper` | Shopper, browses products | E-commerce |
| `researcher` | Researcher, visits many pages | Information sites |
| `bouncer` | Immediate exit | Bounce rate testing |
| `engaged` | Engaged user, high interaction | Engagement boost |
| `mobile` | Mobile user | Mobile traffic |

### Session Persistence (v2.4.0)

```yaml
# Advanced Session Management
session_persistence: true              # Enable session persistence
session_storage_path: "./sessions"     # Session storage directory
session_encryption: true               # Enable session encryption
session_encryption_key: ""             # Encryption key (empty = auto-generated)
session_ttl_hours: 168                 # 7 days TTL
session_indexeddb_persist: true        # Enable IndexedDB persistence
session_canvas_fingerprint: true       # Use canvas fingerprint
returning_visitor_rate: 30             # Returning visitor rate (%)
```

### Structured Logging (v2.4.0)

```yaml
# Logger configuration
logging:
  level: "info"              # debug, info, warn, error, fatal
  format: "console"          # json or console
  output: "stdout"           # stdout, stderr, or file path
  max_size: 100              # MB before rotation
  max_backups: 5             # number of old files to keep
  max_age: 30                # days to keep old files
  compress: true             # gzip rotated logs
  async: false               # async logging for performance
  async_buffer_size: 1000    # async buffer size
  development: false         # dev mode with stack traces
```

---

## Feature Details

### Real-time Dashboard & Metrics (v2.4.0)

Prometheus-compatible metrics system with WebSocket streaming for real-time monitoring.

**What it does:**
- Exposes metrics in Prometheus format at `/api/metrics`
- Streams real-time updates via WebSocket at `/api/metrics/stream`
- Pre-built Grafana dashboard export at `/api/metrics/dashboard`
- Tracks hits, success rate, response times, proxy performance, bounce rate

**How to use:**

```bash
# Prometheus scrape config
scrape_configs:
  - job_name: 'eroshit'
    static_configs:
      - targets: ['localhost:8754']
    metrics_path: '/api/metrics'
    scrape_interval: 5s
```

```javascript
// WebSocket streaming
const ws = new WebSocket('ws://localhost:8754/api/metrics/stream');
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log(msg.type, msg.data);
};
```

### Advanced Session Management (v2.4.0)

Comprehensive session persistence with Cookie/LocalStorage/IndexedDB support.

**What it does:**
- Persists HTTP cookies across sessions
- Saves localStorage and sessionStorage data
- Stores IndexedDB contents
- AES-256-GCM encryption for sensitive data
- Returning visitor simulation with configurable rate
- Automatic TTL-based cleanup

**How to use:**

```go
// Session manager setup
cfg := session.SessionManagerConfig{
    StoragePath:          "./sessions",
    TTL:                  168 * time.Hour,
    Encrypt:              true,
    ReturningVisitorRate: 30,
}
sm, _ := session.NewSessionManager(cfg)

// Get or create session (30% chance of returning)
sess := sm.GetOrCreateSession()
```

### Browser Pool Pattern (v2.4.0)

High-performance Chrome instance pool for ~30x faster visit initiation.

**What it does:**
- Pre-allocates Chrome instances for instant reuse
- Thread-safe acquisition and release
- Automatic cookie/cache cleanup between visits
- Instance recycling based on age/session count
- Multi-tab support within same browser instance

**Performance Comparison:**

| Metric | Traditional | Pool | Improvement |
|--------|-------------|------|-------------|
| First visit | ~3 sec | ~3 sec | — |
| Subsequent visits | ~3 sec | ~0.1 sec | **30x** |
| Memory/instance | ~150MB | ~150MB | — |
| Max parallelism | Unbounded | 10-20 | **Stable** |

### Smart Proxy Rotation (v2.4.0)

7 selection strategies for optimal proxy usage.

**What it does:**
- Tracks proxy success rate, response time, and usage count
- Weighted scoring algorithm: `Score = (success * 0.4) + (speed * 0.3) + (recency * 0.2) + (random * 0.1)`
- Geo-targeting for specific countries
- Real-time metrics collection

**How to use:**

```go
// Weighted selector (recommended)
selector := proxy.NewSelectorFromString("weighted")

// Geo selector for specific countries
geoSelector := proxy.NewGeoSelector([]string{"US", "DE", "GB"})

// Usage
proxy := selector.Select(pool, metrics)
```

### Distributed Mode (v2.4.0)

Master-Worker architecture for scaling across multiple machines.

**What it does:**
- Central coordinator (Master) distributes tasks
- Workers execute browser visits independently
- Auto-scaling based on queue utilization
- Circuit breaker for failure recovery
- Task prioritization (Low, Normal, High, Critical)

**How to use:**

```bash
# Start Master
./eroshit -master -bind 0.0.0.0:8080 -secret my-secret

# Start Workers
./eroshit -worker -master http://master:8080 -secret my-secret -concurrency 10

# Submit tasks via API
curl -X POST http://master:8080/api/v1/master/task/submit \
  -H "Authorization: Bearer my-secret" \
  -d '{"url": "https://example.com"}'
```

### Configuration Hot-Reload (v2.4.0)

Auto-reload configuration without restarting the application.

**What it does:**
- Watches config file for changes using fsnotify
- 1-second debounce to batch rapid changes
- Graceful simulation restart on config change
- Thread-safe config access

**How to use:**

```go
reloader := configpkg.NewReloader("config.yaml")
reloader.OnChange(func(newCfg *configpkg.Config) {
    log.Println("Config reloaded!")
    // Restart simulation with new config
})
reloader.Start()
```

### HTTP/3 QUIC Support (v2.4.0)

Next-generation HTTP protocol for faster, more reliable connections.

**What it does:**
- Uses QUIC protocol instead of TCP for reduced latency
- 0-RTT connection resumption for instant reconnections
- Better performance on unreliable networks
- Built-in encryption and congestion control

**Config:**

```yaml
enable_http3: true
```

**Build with HTTP/3:**

```bash
go build -tags http3 -o eroshit ./cmd/eroshit
```

### Advanced Connection Pooling (v2.4.0)

Optimized connection reuse for maximum throughput.

**What it does:**
- Keeps TCP connections alive for reuse
- HTTP/2 multiplexing support
- Configurable pool sizes per host
- Automatic connection health checks

**Benefits:**
- ~40% reduction in connection overhead
- Better proxy rotation performance
- Reduced memory allocation

**Config:**

```yaml
enable_connection_pool: true
connection_pool_size: 100
max_conns_per_host: 20
enable_tcp_fast_open: true  # Linux only
```

### TCP Fast Open (v2.4.0)

Eliminate TCP handshake latency on supported platforms.

**What it does:**
- Sends data during initial TCP handshake (TFO cookie)
- Reduces connection establishment by 1 RTT
- Linux only (requires kernel support)

**Config:**

```yaml
enable_tcp_fast_open: true
```

### CPU Affinity (v2.4.0)

Pin threads to specific CPU cores for better cache utilization.

**What it does:**
- Binds goroutines to specific CPU cores
- Reduces context switching overhead
- Better L1/L2 cache hit rates
- Ideal for high-throughput scenarios

**Config:**

```yaml
enable_cpu_affinity: true
cpu_affinity_cores: [0, 1, 2, 3]  # Use specific cores
```

### NUMA Awareness (v2.4.0)

NUMA-aware memory management for high-end servers.

**What it does:**
- Prefers memory allocation from local NUMA nodes
- Reduces cross-socket memory access
- Optimized for dual/quad-socket systems
- Automatic NUMA topology detection

**Config:**

```yaml
enable_numa: true
numa_nodes: [0, 1]  # Use specific NUMA nodes
```

### VM Fingerprint Spoofing (v2.4.0)

Hide virtual machine traces from detection systems.

**What it does:**
- Removes VirtualBox, VMware, Hyper-V indicators
- Spoofs hardware IDs (CPU cores, RAM, etc.)
- Randomizes VM-specific parameters
- Mimics physical machine fingerprints

**Supported VM Types:**
- VirtualBox
- VMware
- Hyper-V
- Parallels
- QEMU
- Xen

**Config:**

```yaml
enable_vm_spoofing: true
vm_type: "none"  # Pretend to be physical machine
hide_vm_indicators: true
spoof_hardware_ids: true
randomize_vm_params: true
```

**Web UI Integration:**
- Real-time VM detection score
- Visual indicator of spoofing effectiveness
- One-click enable/disable

---

## API Documentation

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/config` | GET/POST | Get or update configuration |
| `/api/start` | POST | Start simulation |
| `/api/stop` | POST | Stop simulation |
| `/api/status` | GET | Get current status and metrics |
| `/api/logs` | GET | SSE stream for log messages |
| `/api/ws` | WebSocket | Real-time status and log updates |
| `/health` | GET | Health check endpoint |

### Metrics Endpoints (v2.4.0)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/metrics` | GET | Prometheus format metrics |
| `/api/metrics/json` | GET | JSON format metrics |
| `/api/metrics/stream` | WS | Real-time WebSocket stream |
| `/api/metrics/dashboard` | GET | Grafana dashboard JSON |

### Proxy Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/proxy/fetch` | POST | Fetch and test public proxies |
| `/api/proxy/status` | GET | Get proxy pool status |
| `/api/proxy/live` | GET | Get list of live proxies |
| `/api/proxy/export` | GET | Export live proxies as text file |
| `/api/proxy/test` | POST | Test a single proxy |

### GSC Integration

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/gsc/queries` | POST | Fetch queries from GSC API |

### Distributed Mode Endpoints (v2.4.0)

**Master Endpoints:**

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/master/status` | GET | No | Master status |
| `/api/v1/master/stats` | GET | Yes | Statistics |
| `/api/v1/master/workers` | GET | Yes | Worker list |
| `/api/v1/master/tasks` | GET | Yes | Task list |
| `/api/v1/master/task/submit` | POST | Yes | Submit task |

**Worker Endpoints:**

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/worker/register` | POST | Yes | Register worker |
| `/api/v1/worker/heartbeat` | POST | Yes | Send heartbeat |
| `/api/v1/worker/task/request` | POST | Yes | Request task |
| `/api/v1/worker/task/complete` | POST | Yes | Complete task |
| `/api/v1/worker/task/fail` | POST | Yes | Fail task |

---

## Performance Benchmarks

### Browser Pool Performance

| Metric | v2.3.x | v2.4.0 | Improvement |
|--------|--------|--------|-------------|
| Visit initiation | ~3,000ms | ~100ms | **30x faster** |
| Memory (10 concurrent) | ~1.5GB | ~1.5GB | Stable |
| CPU usage | High spikes | Smooth | **Better** |
| Concurrent stability | Variable | Consistent | **Reliable** |

### Metrics System Overhead

| Metric | Without Metrics | With Metrics | Overhead |
|--------|-----------------|--------------|----------|
| Hits/minute | 1000 | 985 | **~1.5%** |
| Memory | 200MB | 215MB | **+7.5%** |
| Response time | 1.2s | 1.21s | **Negligible** |

### Distributed Mode Scaling

| Workers | Throughput | Latency | Efficiency |
|---------|------------|---------|------------|
| 1 | 100 H/min | 1.2s | 100% |
| 5 | 480 H/min | 1.25s | 96% |
| 10 | 920 H/min | 1.3s | 92% |
| 20 | 1,700 H/min | 1.4s | 85% |

---

## Changelog

### v2.4.0 (Latest)

#### New Features
- **Real-time Dashboard & Metrics** — Prometheus-compatible metrics with WebSocket streaming
- **Advanced Session Management** — Cookie/LocalStorage/IndexedDB persistence with returning visitor simulation
- **Smart Proxy Rotation** — 7 selection strategies (weighted, fastest, success-rate, geo-based)
- **Behavioral Fingerprint Randomization** — 7 predefined visitor behavior profiles
- **Structured Logging (Zap)** — High-performance JSON logging with rotation
- **Distributed Mode (Master-Worker)** — Scale across multiple machines
- **Configuration Hot-Reload** — Auto-reload config without restart

#### Improvements
- **Browser Pool Pattern** — ~30x faster visit initiation
- **Memory optimization** — sync.Pool usage for reduced GC pressure
- **Adaptive rate limiting** — Token bucket with dynamic adjustment
- **Circuit breaker pattern** — Automatic failure detection and recovery
- **Auto-scaling** — Dynamic worker pool adjustment based on load

### v2.3.0
- Google Search Console API integration
- Bounce rate control with multi-page visits
- Session depth simulation
- Custom GA4 dimensions and metrics
- Returning visitor simulation
- Exit page control
- Browser profile persistence
- JA3/JA4 TLS fingerprint randomization
- Multi-private proxy support with rotation modes

### v2.2.0
- Complete UI overhaul with tabbed interface
- Device type and brand filtering
- Page duration and scroll depth settings
- Stealth & anti-detection options
- Enhanced proxy stats with live metrics
- Quality score metric (A+ to F)

### v2.1.0
- Tailwind CSS redesign
- Public proxy integration
- Race condition fixes
- Graceful shutdown
- Go 1.21+ compatibility

### v2.0.0
- Sitemap.xml support
- Homepage weight configuration
- HPM token bucket fix
- Visit timeout increase
- Fetch blocking fix

---

## Usage

### Web Interface

1. Start the program: `./eroshit`
2. Select language (Turkish/English) at prompt
3. Choose between **Recommended Settings** or **Manual Configuration**
4. Browser opens to the control panel
5. Navigate through tabs to configure:
   - **Basic Settings** — Domain, duration, HPM, export format
   - **Traffic Simulation** — Device type, brands, scroll behavior, session depth, bounce rate
   - **SEO & Analytics** — Keywords, referrer, geo location, custom dimensions
   - **GSC & Advanced** — GSC integration, returning visitors, exit pages, browser profiles, TLS fingerprinting
   - **Advanced** — Stealth options, performance, resource blocking
   - **Proxy** — Multi-proxy list, rotation settings, public proxy fetching
6. Click **Start** to run the simulation
7. View live metrics and logs; export reports when done

### CLI Mode

```bash
./eroshit -cli -config config.json
# or
./eroshit -cli -domain example.com -pages 5 -duration 60 -hpm 35
```

### Port

Change the web server port (default 8754):

```bash
./eroshit -port 9000
```

---

## Releases

Pre-built binaries for **Windows**, **macOS**, and **Linux** are available in the [Releases](https://github.com/eros1sh/eros-hitbot/releases) section.

To create a new release: push a version tag (e.g. `v2.4.0`). GitHub Actions will automatically build binaries for all platforms.

| Platform | Architecture | File |
|----------|--------------|------|
| Windows | x64 (amd64) | `eros-hitbot-windows-amd64.zip` |
| macOS | Intel (amd64) | `eros-hitbot-darwin-amd64.zip` |
| macOS | Apple Silicon (arm64) | `eros-hitbot-darwin-arm64.zip` |
| Linux | x64 (amd64) | `eros-hitbot-linux-amd64.tar.gz` |
| Linux | ARM64 (aarch64) | `eros-hitbot-linux-arm64.tar.gz` |

**Note:** The Windows build includes a custom application icon. Ensure Chrome/Chromium is installed for headless browsing on all platforms.

---

## Architecture

```
eros-hitbot/
├── cmd/eroshit/          # Main entry (GUI + CLI + Master + Worker)
├── internal/
│   ├── browser/          # Chromedp hit visitor (stealth, fingerprint, analytics)
│   ├── config/           # Config loading (JSON/YAML) + hot-reload
│   ├── crawler/          # Colly page discovery
│   ├── proxy/            # Proxy fetcher, checker, live pool, selector (7 modes)
│   ├── reporter/         # Metrics, CSV/JSON/HTML export
│   ├── server/           # Web UI + API (tabbed interface)
│   ├── session/          # Session management, returning visitors
│   ├── simulator/        # Orchestrator, parallel workers (optimized)
│   └── worker/           # Distributed worker pool, circuit breaker
├── pkg/
│   ├── analytics/        # GA4/GTM event injection, traffic simulator
│   ├── antidetect/       # Anti-detection techniques
│   ├── banner/           # Rainbow ASCII art
│   ├── behavior/         # Human-like scroll, mouse, profiles (7 types)
│   ├── browser/          # Browser pool (high-performance)
│   ├── canvas/           # Canvas/WebGL/Audio fingerprint
│   ├── config/           # Config hot-reload system
│   ├── distributed/      # Master-Worker coordinator
│   ├── engagement/       # Scroll, dwell, clicks
│   ├── fingerprint/      # UA, platform, screen, advanced FP, TLS (JA3/JA4)
│   ├── geo/              # Geo location simulation
│   ├── gsc/              # Google Search Console API integration
│   ├── i18n/             # Multi-language log messages (Turkish/English)
│   ├── interaction/      # Keyboard, mouse simulation
│   ├── logger/           # Zap structured logging
│   ├── metrics/          # Prometheus-compatible metrics
│   ├── mobile/           # Device profiles (30+ devices)
│   ├── profile/          # Browser profile persistence
│   ├── proxy/            # Smart proxy selector (7 strategies)
│   ├── referrer/         # Search/social referrer chain
│   ├── seo/              # Keywords, organic traffic, SERP
│   ├── session/          # Advanced session management
│   ├── sitemap/          # sitemap.xml fetch & parse
│   ├── stealth/          # Headless detection bypass
│   ├── tls/              # TLS fingerprint randomization
│   ├── useragent/        # agents.json, operaagent.json loader
│   └── utils/            # Utility functions, sync.Pool
├── assets/               # Screenshots for README
├── config.example.json
├── config/
│   └── config.yaml       # YAML config with v2.4.0 features
└── examples/
    └── distributed/      # Distributed mode examples
```

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing`
3. Commit: `git commit -m 'Add amazing feature'`
4. Push: `git push origin feature/amazing`
5. Open a Pull Request

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## Disclaimer

This tool is intended for **testing and research purposes**—e.g., verifying analytics, load testing, and SEO experiments on properties you own or have permission to test. Use responsibly and in accordance with applicable laws and terms of service.
