package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// CurrentVersion uygulamanın mevcut sürümü
	CurrentVersion = "2.5.0"

	// gitHubReleaseURL GitHub releases API endpoint
	gitHubReleaseURL = "https://api.github.com/repos/eros1sh/eros-hitbot/releases/latest"

	// checkTimeout API isteği timeout süresi
	checkTimeout = 5 * time.Second
)

// ReleaseInfo GitHub release bilgisi
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// UpdateResult güncelleme kontrolü sonucu
type UpdateResult struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseNotes   string
}

// CheckForUpdate GitHub'dan en son sürümü kontrol eder.
// Hata durumunda sessizce nil döner — uygulama akışını bloke etmez.
func CheckForUpdate() *UpdateResult {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubReleaseURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "eroshit-updater/"+CurrentVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	latestVersion := normalizeVersion(release.TagName)
	currentVersion := normalizeVersion(CurrentVersion)

	result := &UpdateResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   release.Body,
	}

	if compareVersions(latestVersion, currentVersion) > 0 {
		result.Available = true
	}

	return result
}

// CheckForUpdateAsync güncelleme kontrolünü arka planda yapar
func CheckForUpdateAsync() <-chan *UpdateResult {
	ch := make(chan *UpdateResult, 1)
	go func() {
		ch <- CheckForUpdate()
	}()
	return ch
}

// normalizeVersion "v" prefix'ini kaldırır ve boşlukları temizler
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

// compareVersions semantic versioning karşılaştırması yapar
// a > b ise 1, a < b ise -1, eşitse 0 döner
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB)
		}

		if numA > numB {
			return 1
		}
		if numA < numB {
			return -1
		}
	}

	return 0
}

// PrintUpdateNotice güncelleme bildirimini renkli formatta yazdırır
func PrintUpdateNotice(result *UpdateResult, lang string) {
	if result == nil || !result.Available {
		return
	}

	yellow := "\033[33m"
	green := "\033[32m"
	cyan := "\033[36m"
	reset := "\033[0m"
	bold := "\033[1m"

	fmt.Println()
	if lang == "tr" {
		fmt.Printf("  %s%s╔══════════════════════════════════════════════════╗%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s║          GÜNCELLEME MEVCUT!                     ║%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s╠══════════════════════════════════════════════════╣%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s║%s  Mevcut sürüm  : %sv%s%s\n", bold, yellow, reset, green, result.CurrentVersion, reset)
		fmt.Printf("  %s%s║%s  Son sürüm     : %s%sv%s%s%s\n", bold, yellow, reset, bold, green, result.LatestVersion, reset, reset)
		fmt.Printf("  %s%s║%s  İndir         : %s%s%s\n", bold, yellow, reset, cyan, result.ReleaseURL, reset)
		fmt.Printf("  %s%s╚══════════════════════════════════════════════════╝%s\n", bold, yellow, reset)
	} else {
		fmt.Printf("  %s%s╔══════════════════════════════════════════════════╗%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s║          UPDATE AVAILABLE!                      ║%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s╠══════════════════════════════════════════════════╣%s\n", bold, yellow, reset)
		fmt.Printf("  %s%s║%s  Current version : %sv%s%s\n", bold, yellow, reset, green, result.CurrentVersion, reset)
		fmt.Printf("  %s%s║%s  Latest version  : %s%sv%s%s%s\n", bold, yellow, reset, bold, green, result.LatestVersion, reset, reset)
		fmt.Printf("  %s%s║%s  Download        : %s%s%s\n", bold, yellow, reset, cyan, result.ReleaseURL, reset)
		fmt.Printf("  %s%s╚══════════════════════════════════════════════════╝%s\n", bold, yellow, reset)
	}
	fmt.Println()
}
