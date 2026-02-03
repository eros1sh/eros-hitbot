# Eros Hit Bot

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-2.3.0-red)](https://github.com/eros1sh/eros-hitbot/releases)

**Parasitic SEO traffic simulation tool** — Simulate organic search traffic, boost search engine rankings, and verify analytics (GA4/GTM) through realistic, keyword-driven web visits.

<p align="center">
  <img src="assets/web.png" alt="Web Interface" width="600"/>
</p>


---

## Table of Contents

- [Overview](#overview)
- [What's New in 2.3.0](#whats-new-in-230)
- [What's New in 2.2.0](#whats-new-in-220)
- [What's New in 2.1.0](#whats-new-in-210)
- [What's New in 2.0](#whats-new-in-20)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Releases](#releases)
- [Architecture](#architecture)
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

## What's New in 2.3.0

### Google Search Console Integration
- **GSC API Integration** — Fetch real search queries from Google Search Console API
- **Query-based Traffic** — Use actual GSC queries to simulate organic traffic with real keywords
- **Property URL Configuration** — Configure GSC property URL and API key for authentication
- **Query Fetching UI** — One-click button to fetch and display GSC queries in the interface

### Bounce Rate Control
- **Target Bounce Rate** — Set a target bounce rate percentage (e.g., 35%)
- **Automatic Multi-page Visits** — Visitors automatically browse multiple pages to achieve target bounce rate
- **Real-time Bounce Metric** — New bounce rate metric displayed in the dashboard

### Session Depth Simulation
- **Pages per Session** — Configure minimum and maximum pages per session (2-5 default)
- **Realistic Session Behavior** — Simulate natural user browsing patterns with multiple page visits
- **Session Depth Toggle** — Enable/disable session depth simulation

### Custom Dimensions & Metrics (GA4)
- **Custom Dimensions** — Send custom dimensions to GA4 via JSON configuration
- **Custom Metrics** — Send custom metrics to GA4 via JSON configuration
- **Flexible Configuration** — Define any custom dimension/metric as JSON objects

### Returning Visitor Simulation
- **Returning Visitor Rate** — Configure percentage of returning visitors (default 30%)
- **Visit Interval** — Set days between return visits (default 7 days)
- **Client ID Persistence** — Same client_id used for returning visitor simulation

### Exit Page Control
- **Exit Page Patterns** — Define URL patterns where sessions should end
- **Controlled Session Flow** — Sessions continue on non-exit pages, end on exit pages
- **Pattern-based Matching** — Support for URL patterns (e.g., /contact, /thank-you)

### Browser Profile Persistence
- **Profile Storage** — Save browser profiles to disk for reuse
- **Cookie Persistence** — Persist cookies across sessions
- **LocalStorage Persistence** — Persist localStorage data across sessions
- **Max Profile Limit** — Configure maximum number of stored profiles (default 100)
- **Profile Directory** — Configurable profile storage directory

### TLS Fingerprint Randomization
- **JA3 Randomization** — Randomize JA3 TLS fingerprints to avoid detection
- **JA4 Randomization** — Randomize JA4 TLS fingerprints for enhanced stealth
- **Fingerprint Modes** — Choose from Random, Chrome, Firefox, Safari, or Edge fingerprint profiles

### Multi Private Proxy Support
- **Dynamic Proxy List** — Add multiple private proxies with + button
- **Proxy Rotation Modes** — Round Robin, Random, Least Used, or Fastest rotation
- **Rotation Interval** — Configure how often to rotate proxies (per request)
- **Individual Proxy Testing** — Test each proxy individually with one-click button
- **Proxy Removal** — Remove individual proxies from the list

### UI Improvements
- **New GSC & Advanced Tab** — Dedicated tab for GSC integration and advanced analytics features
- **7-Column Metrics Grid** — Added bounce rate metric to the dashboard
- **Enhanced Proxy Tab** — Reorganized with multi-proxy support and rotation settings
- **Improved Form Persistence** — All new fields automatically saved to localStorage

---

## What's New in 2.2.0

### Complete UI Overhaul
- **Tabbed Interface** — Organized settings into 5 logical tabs: Basic Settings, Traffic Simulation, SEO & Analytics, Advanced, and Proxy
- **All Features Accessible** — Every backend feature is now controllable from the web interface
- **Full i18n Support** — Complete Turkish and English translations for all new features

### Traffic Simulation Tab
- **Device Type Selection** — Choose between Desktop, Mobile, Tablet, or Mixed device emulation
- **Device Brand Filtering** — Filter by specific brands: Apple, Samsung, Google, Xiaomi, Huawei, OnePlus, Windows, Mac, Linux
- **Page Duration Control** — Configure minimum and maximum page visit duration (15-120 seconds default)
- **Scroll Depth Settings** — Set minimum and maximum scroll percentages (25-100% default)
- **Click Probability** — Configure the probability of simulated clicks (0-100%)
- **Behavior Simulation Toggles** — Enable/disable mouse movement, keyboard simulation, click simulation, and focus/blur events

### SEO & Analytics Tab
- **Referrer Settings** — Configure referrer keyword and source (Google, Bing, Yahoo, DuckDuckGo, Mixed, Direct)
- **Keywords Management** — Multi-line keyword input with comma/newline separation
- **Analytics Events Control** — Toggle individual GA4 events: Page View, Session Start, User Engagement, First Visit
- **Geo Location Settings** — Configure country, timezone, and language for traffic simulation

### Advanced Tab
- **Stealth & Anti-Detect Options** — 8 toggleable stealth features:
  - Webdriver hiding
  - Chrome automation hiding
  - Plugin spoofing
  - WebGL noise injection
  - Audio fingerprint noise
  - Canvas fingerprint noise
  - Timezone spoofing
  - Language spoofing
- **Performance Settings** — Visit timeout, page load wait time, retry count configuration
- **Resource Blocking** — Toggle blocking for images, CSS, fonts, and media

### Proxy Tab
- **Enhanced Proxy Stats** — Real-time display of live, queued, added, and removed proxy counts
- **Proxy Export** — One-click export of live proxies to text file
- **Improved UI** — Better organized private and public proxy sections

### Metrics & Quality
- **Quality Score** — New metric showing traffic quality grade (A+ to F) based on success rate
- **6-Column Metrics Grid** — Added quality score to the live metrics display

### Technical Improvements
- **Optimized Tab Switching** — Smooth transitions between configuration tabs
- **Auto-save on Change** — All form fields automatically save to localStorage
- **Clear Log Button** — Quick log clearing functionality
- **Version Badge** — Updated to v2.2.0 in header

---

## What's New in 2.1.0

### UI/UX Improvements
- **Tailwind CSS redesign** — Modern, cleaner interface with improved visual hierarchy and responsive grid layouts
- **Public proxy integration** — Public proxy settings moved into the main Proxy Settings section for better organization
- **Live proxy metric** — New metric card showing real-time count of active public proxies (purple theme)
- **Loading states** — "Fetch & Test Proxies" button now shows a spinning animation during proxy testing with real-time status polling
- **GitHub link** — Repository link added to the header with version badge

### Bug Fixes & Code Quality
- **Fixed deprecated `rand.Seed()`** — Migrated to `math/rand/v2` and `crypto/rand` for Go 1.22+ compatibility
- **Fixed race condition in crawler** — Added mutex protection for `startTime` map in request/response handlers
- **Fixed race condition in LivePool** — `GetNext()` now uses single lock for thread-safe proxy rotation
- **Fixed nil pointer dereference** — Added null checks for visitor in simulator's public proxy mode
- **Improved error handling** — All chromedp operations now properly handle errors (non-critical errors are logged but don't break flow)
- **Fixed memory leak** — Reporter log channel now properly closes; added `Close()` method to prevent goroutine leaks
- **Graceful shutdown** — HTTP server now handles SIGINT/SIGTERM with 5-second timeout for clean shutdown

### Technical Changes
- **Go 1.21+ compatible** — Uses standard `math/rand` package for broad compatibility
- **Improved proxy status API** — `/api/proxy/status` now returns `checking` flag for UI polling

---

## What's New in 2.0

- **Sitemap.xml support** — Optional checkbox to use sitemap. If the site has `sitemap.xml` (or a Sitemap entry in `robots.txt`), URLs are loaded from there; otherwise page discovery (crawler) is used.
- **Homepage weight** — When sitemap is enabled, a configurable percentage of requests go to the homepage and the rest are distributed randomly across sitemap URLs (default 60% homepage).
- **HPM (hits per minute) fix** — Token bucket enforces the actual HPM limit; a new visit starts as soon as a slot is free, and all parallel slots fill quickly at startup.
- **Visit timeout** — Per-visit timeout increased from 30s to 90s for slow pages; reduces `context deadline exceeded` and many `ERR_TIMED_OUT` errors.
- **Fetch blocking fix** — Document and Script are never blocked; only Image, Stylesheet, Font, and Media are blocked, fixing `ERR_BLOCKED_BY_CLIENT`.
- **Windows-only release** — Build and release target Windows (amd64) only; simpler download and setup.

---

## Features

| Feature | Description |
|---------|-------------|
| **Parasitic SEO** | Simulate organic search traffic to boost rankings |
| **Keyword targeting** | Define custom keywords; traffic appears to come from search |
| **GSC Integration** | Fetch real queries from Google Search Console API |
| **Multi-browser** | Up to 50 concurrent headless Chrome instances |
| **Proxy support** | HTTP proxy with authentication (rotation ready) |
| **Multi-proxy** | Add multiple private proxies with rotation modes |
| **Public proxy** | Auto-fetch and test proxies from public lists |
| **GA4/GTM** | Automatic page_view, scroll, session_start, user_engagement events |
| **Custom Dimensions** | Send custom GA4 dimensions and metrics |
| **Headless bypass** | Comprehensive stealth techniques to reduce bot detection |
| **Canvas fingerprinting** | Unique canvas/WebGL/Audio noise per visit |
| **TLS Fingerprinting** | JA3/JA4 fingerprint randomization |
| **Device emulation** | Mobile, tablet, desktop with brand-specific profiles |
| **Geo targeting** | Country, timezone, and language spoofing |
| **Bounce Rate Control** | Target specific bounce rate with multi-page visits |
| **Session Depth** | Simulate realistic session depth (2-5 pages) |
| **Returning Visitors** | Simulate returning visitors with persistent client_id |
| **Exit Page Control** | Define which pages end sessions |
| **Browser Profiles** | Persist browser profiles, cookies, localStorage |
| **Sitemap.xml** | Optional: use sitemap URLs + weighted homepage traffic |
| **i18n** | Turkish and English UI + logs |
| **Reports** | CSV, JSON, and HTML dashboard export |
| **Quality metrics** | Real-time quality scoring (A+ to F grade) |

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

### SEO & Referrer

| Field | Description | Default |
|-------|-------------|---------|
| `keywords` | Target keywords (comma-separated in UI) | `[]` |
| `referrerKeyword` | Keyword for Google referrer | `""` |
| `referrerEnabled` | Enable referrer simulation | `false` |
| `referrerSource` | `google`, `bing`, `yahoo`, `duckduckgo`, `mixed`, `direct` | `google` |
| `useSitemap` | Use sitemap.xml for URLs | `false` |
| `sitemapHomepageWeight` | % of requests to homepage (0–100) | 60 |

### GSC Integration

| Field | Description | Default |
|-------|-------------|---------|
| `gscPropertyUrl` | GSC property URL | `""` |
| `gscApiKey` | GSC API service account key | `""` |
| `enableGscIntegration` | Enable GSC integration | `false` |
| `useGscQueries` | Use fetched GSC queries | `false` |

### Custom Dimensions & Metrics

| Field | Description | Default |
|-------|-------------|---------|
| `customDimensions` | JSON object of custom dimensions | `{}` |
| `customMetrics` | JSON object of custom metrics | `{}` |
| `enableCustomDimensions` | Enable custom dimensions/metrics | `false` |

### Proxy Settings

| Field | Description | Default |
|-------|-------------|---------|
| `privateProxies` | Array of private proxy objects | `[]` |
| `proxyRotationMode` | `round-robin`, `random`, `least-used`, `fastest` | `round-robin` |
| `proxyRotationInterval` | Requests between rotation | 1 |
| `enableProxyRotation` | Enable proxy rotation | `true` |
| `usePublicProxy` | Use auto-fetched public proxies | `false` |
| `checkerWorkers` | Parallel proxy checker workers | 25 |

### User Agent Files

- **agents.json** — Full UA + headers (Chrome-style sec-ch-ua, accept-language, etc.)
- **operaagent.json** — Simple UA strings only

Use the `.example` files as templates. Add as many user agents as needed for variety.

---

## Usage

### Web Interface

1. Start the program: `./eroshit`
2. Select language (TR/EN) at prompt
3. Browser opens to the control panel
4. Navigate through tabs to configure:
   - **Basic Settings** — Domain, duration, HPM, export format
   - **Traffic Simulation** — Device type, brands, scroll behavior, session depth, bounce rate
   - **SEO & Analytics** — Keywords, referrer, geo location, custom dimensions
   - **GSC & Advanced** — GSC integration, returning visitors, exit pages, browser profiles, TLS fingerprinting
   - **Advanced** — Stealth options, performance, resource blocking
   - **Proxy** — Multi-proxy list, rotation settings, public proxy fetching
5. Click **Start** to run the simulation
6. View live metrics and logs; export reports when done

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

To create a new release: push a version tag (e.g. `v2.3.0`). GitHub Actions will automatically build binaries for all platforms.

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
├── cmd/eroshit/          # Main entry (GUI + CLI)
├── internal/
│   ├── browser/          # Chromedp hit visitor (stealth, fingerprint, analytics)
│   ├── config/           # Config loading (JSON/YAML)
│   ├── crawler/          # Colly page discovery
│   ├── proxy/            # Proxy fetcher, checker, live pool, rotation
│   ├── reporter/         # Metrics, CSV/JSON/HTML export
│   ├── server/           # Web UI + API (tabbed interface)
│   ├── session/          # Session management, returning visitors
│   └── simulator/        # Orchestrator, parallel workers
├── pkg/
│   ├── analytics/        # GA4/GTM event injection, traffic simulator, custom dimensions
│   ├── antidetect/       # Anti-detection techniques
│   ├── banner/           # Rainbow ASCII art
│   ├── behavior/         # Human-like scroll, mouse, profiles
│   ├── canvas/           # Canvas/WebGL/Audio fingerprint
│   ├── configfiles/      # Config file bootstrapping
│   ├── conversion/       # Conversion tracking
│   ├── delay/            # Request interval, token bucket (HPM)
│   ├── engagement/       # Scroll, dwell, clicks
│   ├── fingerprint/      # UA, platform, screen, advanced FP, TLS (JA3/JA4)
│   ├── geo/              # Geo location simulation
│   ├── gsc/              # Google Search Console API integration
│   ├── i18n/             # TR/EN log messages
│   ├── interaction/      # Keyboard, mouse simulation
│   ├── mobile/           # Device profiles (30+ devices)
│   ├── profile/          # Browser profile persistence
│   ├── referrer/         # Search/social referrer chain
│   ├── seo/              # Keywords, organic traffic, SERP
│   ├── serp/             # Search engine results simulation
│   ├── sitemap/          # sitemap.xml fetch & parse
│   ├── stealth/          # Headless detection bypass
│   ├── tls/              # TLS fingerprint randomization (JA3/JA4)
│   ├── useragent/        # agents.json, operaagent.json loader
│   └── utils/            # Utility functions
├── assets/               # Screenshots for README
├── config.example.json
├── agents.json.example
└── operaagent.json.example
```

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/config` | GET/POST | Get or update configuration |
| `/api/start` | POST | Start simulation |
| `/api/stop` | POST | Stop simulation |
| `/api/status` | GET | Get current status and metrics |
| `/api/logs` | GET | SSE stream for log messages |
| `/api/ws` | WebSocket | Real-time status and log updates |
| `/api/proxy/fetch` | POST | Fetch and test public proxies |
| `/api/proxy/status` | GET | Get proxy pool status |
| `/api/proxy/live` | GET | Get list of live proxies |
| `/api/proxy/export` | GET | Export live proxies as text file |
| `/api/proxy/test` | POST | Test a single proxy |
| `/api/gsc/queries` | POST | Fetch queries from GSC API |
| `/health` | GET | Health check endpoint |

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
