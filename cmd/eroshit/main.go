// ErosHit - Etik SEO ve Performans Test Aracı
// Varsayılan: Modern web arayüzü. -cli ile konsol modu.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
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
	"eroshit/pkg/useragent"
)

func main() {
	cliMode := flag.Bool("cli", false, "Konsol (CLI) modunda çalıştır")
	port := flag.Int("port", 8754, "Web arayüzü portu")
	flag.Parse()

	if *cliMode {
		runCLI()
		return
	}

	runGUI(*port)
}

func runGUI(port int) {
	// Config dosyalarını exe klasöründe topla (agents, config, operaagent)
	if exeDir, err := getExeDir(); err == nil {
		configfiles.EnsureInDir(exeDir)
	}

	// Dil seçimi - web sayfası açılmadan önce
	lang := promptLang()

	srv, err := server.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Sunucu başlatılamadı: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", port)
	baseURL := "http://127.0.0.1" + addr
	url := baseURL + "?lang=" + lang

	// Terminal banner - seçilen dile göre
	printBanner(url, lang)

	go openBrowser(url)
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

		fmt.Println("\n🛑 Sunucu kapatılıyor...")
		
		// 5 saniye timeout ile graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown hatası: %v\n", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Sunucu hatası: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Sunucu başarıyla kapatıldı.")
}

func promptLang() string {
	fmt.Println()
	fmt.Println("  Dil Seçin / Select Language:")
	fmt.Println("  1 = Türkçe")
	fmt.Println("  2 = English")
	fmt.Print("  Seçim (1/2) [1]: ")
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
	if lang == "en" {
		fmt.Println("╔════════════════════════════════════════════════╗")
		fmt.Println("║           ErosHit - Web Interface              ║")
		fmt.Println("╠════════════════════════════════════════════════╣")
		fmt.Printf("║  Open in browser: %-28s ║\n", url)
		fmt.Println("║  Press Ctrl+C to stop                          ║")
		fmt.Println("╚════════════════════════════════════════════════╝")
	} else {
		fmt.Println("╔════════════════════════════════════════════════╗")
		fmt.Println("║           ErosHit - Web Arayüzü                ║")
		fmt.Println("╠════════════════════════════════════════════════╣")
		fmt.Printf("║  Tarayıcınızda açın: %-26s ║\n", url)
		fmt.Println("║  Durdurmak için Ctrl+C                          ║")
		fmt.Println("╚════════════════════════════════════════════════╝")
	}
	fmt.Println()
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

func runCLI() {
	configPath := flag.String("config", "config.json", "Config dosyası (config.json)")
	targetDomain := flag.String("domain", "", "Hedef domain")
	maxPages := flag.Int("pages", 5, "Max sayfa")
	durationMinutes := flag.Int("duration", 60, "Süre (dakika)")
	hitsPerMinute := flag.Int("hpm", 35, "İstek/dakika")
	flag.Parse()

	cfg, err := config.LoadFromJSON(*configPath)
	if err != nil {
		cfg = &config.Config{
			TargetDomain:    *targetDomain,
			MaxPages:        *maxPages,
			DurationMinutes: *durationMinutes,
			HitsPerMinute:   *hitsPerMinute,
			OutputDir:       "./reports",
			ExportFormat:    "both",
		}
		cfg.ApplyDefaults()
		cfg.ComputeDerived()
	}

	if *targetDomain != "" {
		cfg.TargetDomain = *targetDomain
	}
	if *maxPages > 0 {
		cfg.MaxPages = *maxPages
	}
	if *durationMinutes > 0 {
		cfg.DurationMinutes = *durationMinutes
	}
	if *hitsPerMinute > 0 {
		cfg.HitsPerMinute = *hitsPerMinute
	}
	cfg.ApplyDefaults()
	cfg.ComputeDerived()

	if cfg.TargetDomain == "" || cfg.TargetDomain == "example.com" {
		fmt.Println("UYARI: target_domain gerekli. -domain example.com veya config.json'da belirtin.")
		fmt.Println("Örnek: eroshit -cli -domain mysite.com -pages 5 -duration 60 -hpm 35")
		os.Exit(1)
	}

	agentLoader := useragent.LoadFromDirs([]string{".", ".."})
	sim, err := simulator.New(cfg, agentLoader, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigChan; cancel() }()

	if err := sim.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Simülasyon hatası: %v\n", err)
		os.Exit(1)
	}
}
