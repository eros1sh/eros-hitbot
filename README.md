# Eros Hit Bot

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Parasitic SEO traffic simulation tool** — Simulate organic search traffic, boost search engine rankings, and verify analytics (GA4/GTM) through realistic, keyword-driven web visits.

<p align="center">
  <img src="assets/web.png" alt="Web Interface" width="600"/>
</p>

---

## Table of Contents

- [Overview](#overview)
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

## Features

| Feature | Description |
|---------|-------------|
| **Parasitic SEO** | Simulate organic search traffic to boost rankings |
| **Keyword targeting** | Define custom keywords; traffic appears to come from search |
| **Multi-browser** | Up to 50 concurrent headless Chrome instances |
| **Proxy support** | HTTP proxy with authentication (rotation ready) |
| **GA4/GTM** | Automatic page_view, scroll, and custom event firing |
| **Headless bypass** | Stealth techniques to reduce bot detection |
| **Canvas fingerprinting** | Unique canvas/WebGL/Audio noise per visit |
| **i18n** | Turkish and English UI + logs |
| **Reports** | CSV, JSON, and HTML dashboard export |

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
2. Download **Windows** build: `eros-hitbot-windows-amd64.zip`
3. Extract the zip and run `eros-hitbot-windows-amd64.exe`

### Option 2: Build from Source

**Requirements:** Go 1.21+, Chrome/Chromium (for headless mode)

```bash
git clone https://github.com/eros1sh/eros-hitbot.git
cd eros-hitbot
go build -o eroshit.exe ./cmd/eroshit
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
  "maxConcurrentVisits": 10
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

| Field | Description | Default |
|-------|-------------|---------|
| `targetDomain` | Domain to send traffic to | `example.com` |
| `fallbackGAID` | GA4 Measurement ID | `G-XXXXXXXXXX` |
| `keywords` | Target keywords (comma-separated in UI) | `[]` |
| `maxPages` | Max pages to crawl per session | 5 |
| `durationMinutes` | Simulation duration (minutes) | 60 |
| `hitsPerMinute` | Request rate (HPM) | 35 |
| `maxConcurrentVisits` | Parallel browser tabs | 10 |
| `outputDir` | Report output directory | `./reports` |
| `exportFormat` | `csv`, `json`, `html`, or `both` | `both` |
| `canvasFingerprint` | Enable canvas/WebGL noise | true |
| `scrollStrategy` | `gradual`, `fast`, `reader` | `gradual` |
| `sendScrollEvent` | Send GA4 scroll events | true |
| `PROXY_HOST` | Proxy host (optional) | - |
| `PROXY_PORT` | Proxy port | 3120 |
| `PROXY_USER` | Proxy username | - |
| `PROXY_PASS` | Proxy password | - |

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
4. Set target domain, keywords, and parameters
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

Pre-built **Windows** executable is available in the [Releases](https://github.com/eros1sh/eros-hitbot/releases) section.

To create a new release: push a version tag (e.g. `v1.0.0`). GitHub Actions will build the Windows binary and attach `eros-hitbot-windows-amd64.zip`.

| Platform | File |
|----------|------|
| Windows (amd64) | `eros-hitbot-windows-amd64.zip` |

Download, extract, and run `eros-hitbot-windows-amd64.exe`. Ensure Chrome/Chromium is installed for headless browsing.

---

## Architecture

```
eros-hitbot/
├── cmd/eroshit/          # Main entry (GUI + CLI)
├── internal/
│   ├── browser/          # Chromedp hit visitor (stealth, fingerprint, analytics)
│   ├── config/           # Config loading (JSON/YAML)
│   ├── crawler/          # Colly page discovery
│   ├── reporter/         # Metrics, CSV/JSON/HTML export
│   ├── server/           # Web UI + API
│   └── simulator/        # Orchestrator, parallel workers
├── pkg/
│   ├── analytics/        # GA4/GTM event injection
│   ├── banner/           # Rainbow ASCII art
│   ├── behavior/         # Human-like scroll, mouse
│   ├── canvas/           # Canvas/WebGL/Audio fingerprint
│   ├── configfiles/      # Config file bootstrapping
│   ├── delay/            # Request interval, jitter
│   ├── engagement/       # Scroll, dwell, clicks
│   ├── fingerprint/      # UA, platform, screen
│   ├── i18n/             # TR/EN log messages
│   ├── referrer/         # Search/social referrer chain
│   ├── seo/              # Keywords, organic traffic
│   ├── stealth/          # Headless detection bypass
│   └── useragent/        # agents.json, operaagent.json loader
├── assets/               # Screenshots for README
├── config.example.json
├── agents.json.example
└── operaagent.json.example
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
