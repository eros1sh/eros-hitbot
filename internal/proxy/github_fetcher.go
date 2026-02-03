package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

// GitHub API (minimal) yanıtları
type githubRepo struct {
	DefaultBranch string `json:"default_branch"`
}

type githubTree struct {
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
}

var githubRepoURLRe = regexp.MustCompile(`(?i)github\.com[/:]([^/]+)/([^/]+?)(?:/.*)?$`)

// parseGitHubRepoURL "https://github.com/owner/repo" veya "owner/repo" -> owner, repo
func parseGitHubRepoURL(raw string) (owner, repo string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	// owner/repo formatı
	if !strings.Contains(raw, "github") {
		parts := strings.SplitN(raw, "/", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			repo = strings.TrimSuffix(parts[1], ".git")
			return parts[0], repo, true
		}
		return "", "", false
	}
	matches := githubRepoURLRe.FindStringSubmatch(raw)
	if len(matches) != 3 {
		return "", "", false
	}
	owner = strings.TrimSpace(matches[1])
	repo = strings.TrimSpace(matches[2])
	repo = strings.TrimSuffix(repo, ".git")
	if owner == "" || repo == "" {
		return "", "", false
	}
	return owner, repo, true
}

// FetchFromGitHubRepos verilen GitHub repo URL'lerinden tüm .txt dosyalarını indirir,
// proxy satırlarını parse edip tek listede birleştirir (tekrarsız).
func FetchFromGitHubRepos(ctx context.Context, repoURLs []string, client *http.Client) ([]*ProxyConfig, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 15 * time.Second,
			},
		}
	}
	all := make([]*ProxyConfig, 0, 8192)
	seen := make(map[string]struct{})

	for _, rawURL := range repoURLs {
		owner, repo, ok := parseGitHubRepoURL(rawURL)
		if !ok {
			continue
		}
		branch, err := getDefaultBranch(ctx, client, owner, repo)
		if err != nil {
			continue
		}
		paths, err := getRepoFilePaths(ctx, client, owner, repo, branch)
		if err != nil {
			continue
		}
		for _, p := range paths {
			if !strings.HasSuffix(strings.ToLower(p), ".txt") {
				continue
			}
			list, err := fetchRawFile(ctx, client, owner, repo, branch, p)
			if err != nil {
				continue
			}
			for _, cfg := range list {
				k := cfg.Key()
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = struct{}{}
				all = append(all, cfg)
			}
		}
	}
	return all, nil
}

func getDefaultBranch(ctx context.Context, client *http.Client, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var r githubRepo
	if json.NewDecoder(resp.Body).Decode(&r) != nil {
		return "", fmt.Errorf("decode repo")
	}
	if r.DefaultBranch != "" {
		return r.DefaultBranch, nil
	}
	return "main", nil
}

func getRepoFilePaths(ctx context.Context, client *http.Client, owner, repo, branch string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var t githubTree
	if json.NewDecoder(resp.Body).Decode(&t) != nil {
		return nil, fmt.Errorf("decode tree")
	}
	var paths []string
	for _, e := range t.Tree {
		if e.Type == "blob" && e.Path != "" {
			paths = append(paths, e.Path)
		}
	}
	return paths, nil
}

func fetchRawFile(ctx context.Context, client *http.Client, owner, repo, branch, filePath string) ([]*ProxyConfig, error) {
	escapedPath := strings.TrimPrefix(path.Clean("/"+filePath), "/")
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		owner, repo, branch, escapedPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ErosHit/1)")
	resp, err := client.Do(req)
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
