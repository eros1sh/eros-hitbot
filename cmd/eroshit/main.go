// ErosHit - Etik SEO ve Performans Test Aracı
// Varsayılan: Modern web arayüzü. -cli ile konsol modu.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"eroshit/internal/config"
	"eroshit/internal/server"
	"eroshit/internal/simulator"
	"eroshit/pkg/banner"
	"eroshit/pkg/configfiles"
	"eroshit/pkg/i18n"
	"eroshit/pkg/sysinfo"
	"eroshit/pkg/useragent"
)

// Global language variable
var currentLang = "tr"

func main() {
	cliMode := flag.Bool("cli", false, "Konsol (CLI) modunda çalıştır")
	port := flag.Int("port", 8754, "Web arayüzü portu")
	showSysInfo := flag.Bool("sysinfo", false, "Sistem bilgilerini göster (neofetch benzeri)")
	autoOptimize := flag.Bool("optimize", false, "Otomatik optimizasyon profili uygula")
	flag.Parse()

	// Dil seçimi - her modda ilk adım
	currentLang = promptLang()

	// Sistem bilgisi modu
	if *showSysInfo {
		showSystemInfo(currentLang)
		return
	}

	if *cliMode {
		runCLI(*autoOptimize, currentLang)
		return
	}

	runGUI(*port, currentLang)
}

// showSystemInfo displays neofetch-style system information
func showSystemInfo(lang string) {
	fmt.Println()
	fmt.Println("  " + i18n.T(lang, i18n.MsgDetectingSystem))
	
	info := sysinfo.Detect()
	fmt.Print(info.PrintBannerWithLocale(lang))
	
	profile := info.GenerateOptimizationProfileWithLocale(lang)
	fmt.Print(profile.PrintProfileWithLocale(lang))
	
	fmt.Println()
	fmt.Println("  " + i18n.T(lang, i18n.MsgOptApplyHint))
	fmt.Println()
}

// promptSettingsChoice asks user to choose between recommended or manual settings
// Returns the optimization profile if user chooses recommended, nil otherwise
func promptSettingsChoice(lang string, profile *sysinfo.OptimizationProfile) bool {
	fmt.Println()
	fmt.Println("  " + i18n.T(lang, i18n.MsgRecommendedSettings))
	fmt.Println("  " + i18n.T(lang, i18n.MsgManualSettings))
	fmt.Println()
	fmt.Print("  " + i18n.T(lang, i18n.MsgSettingsQuestion))
	
	rd := bufio.NewReader(os.Stdin)
	line, err := rd.ReadString('\n')
	if err != nil {
		return true // Default to recommended
	}
	line = strings.TrimSpace(strings.TrimSuffix(line, "\n"))
	
	// Default is 1 (recommended)
	if line == "" || line == "1" {
		return true
	}
	return false
}

func runGUI(port int, lang string) {
	// Config dosyalarını exe klasöründe topla (agents, config, operaagent)
	if exeDir, err := getExeDir(); err == nil {
		configfiles.EnsureInDir(exeDir)
	}

	// Sistem bilgilerini göster
	fmt.Println()
	fmt.Println("  " + i18n.T(lang, i18n.MsgDetectingSystem))
	
	info := sysinfo.Detect()
	fmt.Print(info.PrintBannerWithLocale(lang))
	
	profile := info.GenerateOptimizationProfileWithLocale(lang)
	fmt.Print(profile.PrintProfileWithLocale(lang))
	
	// Kullanıcıya seçenek sun
	useRecommended := promptSettingsChoice(lang, profile)
	
	// URL parametrelerini hazırla
	urlParams := fmt.Sprintf("lang=%s", lang)
	
	if useRecommended {
		fmt.Println()
		fmt.Println("  " + i18n.T(lang, i18n.MsgApplyingOptimization))
		fmt.Printf("     - %s %d\n", i18n.T(lang, i18n.MsgOptMaxConcurrent), profile.MaxConcurrentVisits)
		fmt.Printf("     - %s %d\n", i18n.T(lang, i18n.MsgOptHitsPerMinute), profile.HitsPerMinute)
		fmt.Println()
		
		// Önerilen ayarları URL parametresi olarak ekle
		urlParams = fmt.Sprintf("lang=%s&maxConcurrent=%d&hpm=%d&poolMin=%d&poolMax=%d&autoOptimized=true",
			lang, profile.MaxConcurrentVisits, profile.HitsPerMinute, profile.BrowserPoolMin, profile.BrowserPoolMax)
	}

	srv, err := server.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgServerError, err)+"\n")
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", port)
	baseURL := "http://127.0.0.1" + addr
	fullURL := baseURL + "?" + urlParams

	// Terminal banner - seçilen dile göre
	printBanner(fullURL, lang)

	fmt.Println("  " + i18n.T(lang, i18n.MsgOpeningBrowser))
	go openBrowser(fullURL, lang)
	time.Sleep(500 * time.Millisecond)

	// HTTP Server with graceful shutdown
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Routes(),
	}

	// Graceful shutdown için sinyal dinle
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println()
		fmt.Println(i18n.T(lang, i18n.MsgServerShutdown))
		
		// 5 saniye timeout ile graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgShutdownError, err)+"\n")
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgServerError, err)+"\n")
		os.Exit(1)
	}
	
	fmt.Println(i18n.T(lang, i18n.MsgServerShutdownComplete))
}

func promptLang() string {
	fmt.Println()
	fmt.Println("  " + i18n.T("tr", i18n.MsgSelectLanguage))
	fmt.Println("  1 = " + i18n.T("tr", i18n.MsgLanguageTurkish))
	fmt.Println("  2 = " + i18n.T("en", i18n.MsgLanguageEnglish))
	fmt.Print("  " + i18n.T("tr", i18n.MsgSelection))
	rd := bufio.NewReader(os.Stdin)
	line, err := rd.ReadString('\n')
	if err != nil {
		return "tr"
	}
	line = strings.TrimSpace(strings.TrimSuffix(line, "\n"))
	if line == "2" {
		return "en"
	}
	return "tr"
}

func getExeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

func printBanner(url, lang string) {
	banner.PrintRainbow(banner.ErosHitASCII)
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Printf("║           %s              ║\n", i18n.T(lang, i18n.MsgWebInterface))
	fmt.Println("╠════════════════════════════════════════════════╣")
	fmt.Printf("║  %s ║\n", padRight(i18n.T(lang, i18n.MsgOpenBrowser, url), 45))
	fmt.Printf("║  %-45s ║\n", i18n.T(lang, i18n.MsgStopHint))
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()
}

// padRight pads a string to the right with spaces (rune-safe for UTF-8)
func padRight(s string, length int) string {
	runes := []rune(s)
	if len(runes) >= length {
		return string(runes[:length])
	}
	return s + strings.Repeat(" ", length-len(runes))
}

// SECURITY FIX: URL validation to prevent command injection
func openBrowser(rawURL string, lang string) {
	// Validate URL to prevent command injection
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgInvalidURL, err)+"\n")
		return
	}
	
	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgSecurityHTTPOnly)+"\n")
		return
	}
	
	// Validate host is localhost
	host := parsedURL.Hostname()
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgSecurityLocalhost)+"\n")
		return
	}
	
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	case "darwin":
		cmd = exec.Command("open", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "  "+i18n.T(lang, i18n.MsgError, err)+"\n")
	}
}

func runCLI(autoOptimize bool, lang string) {
	// CLI modunda flag'ler zaten parse edildi, yeniden parse etme
	// Sadece positional argümanları kullan
	
	configPath := "config.json"
	targetDomain := ""
	maxPages := 5
	durationMinutes := 60
	hitsPerMinute := 35
	maxConcurrent := 10
	
	// Argümanları manuel parse et (flag zaten parse edildi)
	args := flag.Args()
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-config":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "-domain":
			if i+1 < len(args) {
				targetDomain = args[i+1]
				i++
			}
		case "-pages":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxPages)
				i++
			}
		case "-duration":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &durationMinutes)
				i++
			}
		case "-hpm":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &hitsPerMinute)
				i++
			}
		case "-concurrent":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxConcurrent)
				i++
			}
		}
	}

	// Otomatik optimizasyon modu
	if autoOptimize {
		fmt.Println()
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIAutoOptimize))
		fmt.Println("  " + i18n.T(lang, i18n.MsgDetectingSystem))
		fmt.Println()
		
		info := sysinfo.Detect()
		fmt.Print(info.PrintBannerWithLocale(lang))
		
		profile := info.GenerateOptimizationProfileWithLocale(lang)
		fmt.Print(profile.PrintProfileWithLocale(lang))
		
		// Kullanıcıya sor
		fmt.Print("\n  " + i18n.T(lang, i18n.MsgCLIApplySettings))
		rd := bufio.NewReader(os.Stdin)
		line, err := rd.ReadString('\n')
		
		applyOptimization := true
		if err == nil {
			line = strings.TrimSpace(strings.ToLower(line))
			if lang == "tr" {
				applyOptimization = line != "h" && line != "hayır"
			} else {
				applyOptimization = line != "n" && line != "no"
			}
		}
		
		if !applyOptimization {
			fmt.Println("  " + i18n.T(lang, i18n.MsgOptimizationCancelled))
		} else {
			// Optimizasyon profilini uygula
			maxConcurrent = profile.MaxConcurrentVisits
			hitsPerMinute = profile.HitsPerMinute
			fmt.Println()
			fmt.Println("  " + i18n.T(lang, i18n.MsgOptimizationApplied))
			fmt.Printf("     - %s %d\n", i18n.T(lang, i18n.MsgOptMaxConcurrent), maxConcurrent)
			fmt.Printf("     - %s %d\n", i18n.T(lang, i18n.MsgOptHitsPerMinute), hitsPerMinute)
			fmt.Println()
		}
	}

	cfg, err := config.LoadFromJSON(configPath)
	if err != nil {
		cfg = &config.Config{
			TargetDomain:        targetDomain,
			MaxPages:            maxPages,
			DurationMinutes:     durationMinutes,
			HitsPerMinute:       hitsPerMinute,
			MaxConcurrentVisits: maxConcurrent,
			OutputDir:           "./reports",
			ExportFormat:        "both",
		}
		cfg.ApplyDefaults()
		cfg.ComputeDerived()
	}

	if targetDomain != "" {
		cfg.TargetDomain = targetDomain
	}
	if maxPages > 0 {
		cfg.MaxPages = maxPages
	}
	if durationMinutes > 0 {
		cfg.DurationMinutes = durationMinutes
	}
	if hitsPerMinute > 0 {
		cfg.HitsPerMinute = hitsPerMinute
	}
	if maxConcurrent > 0 {
		cfg.MaxConcurrentVisits = maxConcurrent
	}
	cfg.ApplyDefaults()
	cfg.ComputeDerived()

	if cfg.TargetDomain == "" || cfg.TargetDomain == "example.com" {
		fmt.Println(i18n.T(lang, i18n.MsgWarning, i18n.T(lang, i18n.MsgCLIConfigRequired)))
		fmt.Println(i18n.T(lang, i18n.MsgCLIExample))
		fmt.Println()
		fmt.Println(i18n.T(lang, i18n.MsgCLIFlags))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagCli))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagSysinfo))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagOptimize))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagDomain))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagPages))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagDurationFlag))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagHpm))
		fmt.Println("  " + i18n.T(lang, i18n.MsgCLIFlagConcurrent))
		os.Exit(1)
	}

	// Banner göster
	banner.PrintRainbow(banner.ErosHitASCII)
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Printf("║              %s                            ║\n", i18n.T(lang, i18n.MsgCLIMode))
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  %s ║\n", padRight(i18n.T(lang, i18n.MsgCLITarget, cfg.TargetDomain), 57))
	fmt.Printf("║  %s ║\n", padRight(i18n.T(lang, i18n.MsgCLIDuration, cfg.DurationMinutes, cfg.HitsPerMinute, cfg.MaxConcurrentVisits), 57))
	fmt.Printf("║  %-57s ║\n", i18n.T(lang, i18n.MsgCLIStopHint))
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	agentLoader := useragent.LoadFromDirs([]string{".", ".."})
	sim, err := simulator.New(cfg, agentLoader, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgError, err)+"\n")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigChan; cancel() }()

	if err := sim.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, i18n.T(lang, i18n.MsgSimulationError, err)+"\n")
		os.Exit(1)
	}
}
