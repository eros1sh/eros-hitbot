package proxy

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultProxySourceURLs varsayılan public proxy listesi (tekil dosya URL'leri; checker ile kullanılır)
var DefaultProxySourceURLs = []string{
	"https://raw.githubusercontent.com/officialputuid/KangProxy/refs/heads/KangProxy/https/https.txt",
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/refs/heads/master/http.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/refs/heads/master/https.txt",
	"https://raw.githubusercontent.com/dpangestuw/Free-Proxy/refs/heads/main/http_proxies.txt",
	"https://raw.githubusercontent.com/elliottophellia/proxylist/refs/heads/master/results/http/global/http_checked.txt",
	"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt",
	"https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/https/data.txt",
}

// DefaultGitHubRepos varsayılan GitHub proxy repoları; Proxy Çek (varsayılan) bu repolardan tüm .txt indirir, test yok
var DefaultGitHubRepos = []string{
	"https://github.com/officialputuid/KangProxy",
	"https://github.com/TheSpeedX/PROXY-List",
	"https://github.com/mmpx12/proxy-list",
	"https://github.com/dpangestuw/Free-Proxy",
	"https://github.com/elliottophellia/proxylist",
	"https://github.com/TheSpeedX/SOCKS-List",
	"https://github.com/proxifly/free-proxy-list",
}

// ip:port veya host:port (IPv4)
var reHostPort = regexp.MustCompile(`(?i)^(?:https?://)?([0-9a-z._-]+):(\d+)$`)

// ParseProxyLine satırdan ProxyConfig üretir. "http://IP:PORT" veya "IP:PORT" kabul eder.
func ParseProxyLine(line string) (*ProxyConfig, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil, false
	}
	// Önce http(s):// ile başlıyorsa kaldır, host:port yakala
	if u, err := url.Parse(line); err == nil && u.Host != "" {
		host := u.Hostname()
		portStr := u.Port()
		if portStr == "" {
			if u.Scheme == "https" {
				portStr = "443"
			} else {
				portStr = "80"
			}
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return nil, false
		}
		if host == "" {
			return nil, false
		}
		protocol := u.Scheme
		if protocol != "http" && protocol != "https" {
			protocol = "http"
		}
		return &ProxyConfig{Host: host, Port: port, Protocol: protocol}, true
	}
	// host:port (IPv4 veya hostname)
	matches := reHostPort.FindStringSubmatch(line)
	if len(matches) == 3 {
		port, err := strconv.Atoi(matches[2])
		if err != nil || port < 1 || port > 65535 {
			return nil, false
		}
		return &ProxyConfig{Host: strings.TrimSpace(matches[1]), Port: port, Protocol: "http"}, true
	}
	return nil, false
}

// Fetcher proxy listesi çeker ve parse eder
type Fetcher struct {
	Client  *http.Client
	Sources []string
}

// NewFetcher varsayılan client ile fetcher oluşturur
func NewFetcher(sources []string) *Fetcher {
	if len(sources) == 0 {
		sources = DefaultProxySourceURLs
	}
	return &Fetcher{
		Client: &http.Client{
			Timeout: 25 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				IdleConnTimeout:     15 * time.Second,
				DisableCompression:  false,
			},
		},
		Sources: sources,
	}
}

// FetchAll tüm kaynaklardan proxy listesi çeker, birleştirir ve tekrarsız döner
func (f *Fetcher) FetchAll(ctx context.Context) ([]*ProxyConfig, error) {
	all := make([]*ProxyConfig, 0, 4096)
	var listMu sync.Mutex
	var wg sync.WaitGroup
	for _, rawURL := range f.Sources {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			list, _ := f.fetchOne(ctx, u)
			if len(list) == 0 {
				return
			}
			listMu.Lock()
			all = append(all, list...)
			listMu.Unlock()
		}(rawURL)
	}
	wg.Wait()

	seen := make(map[string]*ProxyConfig)
	for _, p := range all {
		k := p.Key()
		if _, ok := seen[k]; !ok {
			seen[k] = p
		}
	}
	out := make([]*ProxyConfig, 0, len(seen))
	for _, p := range seen {
		out = append(out, p)
	}
	return out, nil
}

func (f *Fetcher) fetchOne(ctx context.Context, rawURL string) ([]*ProxyConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var list []*ProxyConfig
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		if p, ok := ParseProxyLine(sc.Text()); ok {
			list = append(list, p)
		}
	}
	return list, sc.Err()
}
