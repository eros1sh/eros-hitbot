package sitemap

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	maxURLs        = 500
	maxChildMaps   = 10
)

// URLSet sitemap urlset (tek sitemap)
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

// SitemapIndex sitemap index (alt sitemaplere link)
type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

// Fetch baseURL'den sitemap.xml veya robots.txt'teki Sitemap ile URL listesi döner.
// Aynı domain'deki, http(s) URL'leri döner; en fazla maxURLs kadar.
func Fetch(baseURL string, client *http.Client) ([]string, error) {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	domain := base.Hostname()

	// Önce /sitemap.xml dene
	u := baseURL + "/sitemap.xml"
	urls, err := fetchAndParse(client, u, base, domain, 0)
	if err == nil && len(urls) > 0 {
		return urls, nil
	}

	// robots.txt'te Sitemap: var mı?
	robotsURL := baseURL + "/robots.txt"
	req, _ := http.NewRequest(http.MethodGet, robotsURL, nil)
	req.Header.Set("User-Agent", "ErosHitBot/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
			smapURL := strings.TrimSpace(line[8:])
			if smapURL != "" {
				urls, err = fetchAndParse(client, smapURL, base, domain, 0)
				if err == nil && len(urls) > 0 {
					return urls, nil
				}
			}
		}
	}
	return nil, nil
}

func fetchAndParse(client *http.Client, u string, base *url.URL, domain string, depth int) ([]string, error) {
	if depth > 2 {
		return nil, nil
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ErosHitBot/1.0")
	req.Header.Set("Accept", "application/xml, text/xml, */*")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}

	// Sitemap index mi?
	var idx SitemapIndex
	if err := xml.Unmarshal(data, &idx); err == nil && len(idx.Sitemaps) > 0 {
		var all []string
		for i, s := range idx.Sitemaps {
			if i >= maxChildMaps {
				break
			}
			loc := strings.TrimSpace(s.Loc)
			if loc == "" {
				continue
			}
			child, _ := fetchAndParse(client, loc, base, domain, depth+1)
			all = append(all, child...)
			if len(all) >= maxURLs {
				break
			}
		}
		return dedupeAndFilter(all, domain), nil
	}

	// URL set
	var set URLSet
	if err := xml.Unmarshal(data, &set); err != nil {
		return nil, err
	}
	var out []string
	for _, u := range set.URLs {
		loc := strings.TrimSpace(u.Loc)
		if loc == "" {
			continue
		}
		parsed, err := url.Parse(loc)
		if err != nil {
			continue
		}
		if parsed.Hostname() != domain {
			continue
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			continue
		}
		out = append(out, loc)
		if len(out) >= maxURLs {
			break
		}
	}
	return dedupeAndFilter(out, domain), nil
}

func dedupeAndFilter(urls []string, domain string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		parsed, err := url.Parse(u)
		if err != nil {
			continue
		}
		if parsed.Hostname() != domain {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}
