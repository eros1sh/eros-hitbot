// Worker Node CLI for ErosHit Distributed Mode
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eroshit/internal/config"
	"eroshit/pkg/distributed"
	"eroshit/pkg/proxy"
	"eroshit/pkg/useragent"
)

func main() {
	var (
		masterURL      = flag.String("master", "http://localhost:8080", "Master URL")
		secretKey      = flag.String("secret", "", "Secret key for authentication")
		maxConcurrency = flag.Int("concurrency", 10, "Max concurrent tasks")
		configPath     = flag.String("config", "config.json", "Config file path")
	)
	flag.Parse()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           ErosHit - Distributed Worker Node                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load local config for simulation settings
	cfg, err := config.LoadFromJSON(*configPath)
	if err != nil {
		fmt.Printf("[Worker] Warning: Could not load config: %v\n", err)
		fmt.Println("[Worker] Using default config")
		cfg = &config.Config{
			TargetDomain:        "example.com",
			MaxConcurrentVisits: *maxConcurrency,
		}
		cfg.ApplyDefaults()
		cfg.ComputeDerived()
	}

	// Create task processor
	processor := createTaskProcessor(cfg)

	// Create worker
	workerConfig := distributed.WorkerConfig{
		MasterURL:      *masterURL,
		SecretKey:      *secretKey,
		MaxConcurrency: *maxConcurrency,
		Hostname:       getHostname(),
		Version:        "1.0.0",
	}

	worker := distributed.NewWorker(workerConfig, processor)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[Worker] Shutting down...")
		worker.Stop()
	}()

	// Print info
	fmt.Printf("[Worker] ID: %s\n", worker.ID)
	fmt.Printf("[Worker] Master: %s\n", *masterURL)
	fmt.Printf("[Worker] Concurrency: %d\n", *maxConcurrency)
	fmt.Printf("[Worker] Hostname: %s\n", getHostname())
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Start worker (blocking)
	if err := worker.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[Worker] Error: %v\n", err)
		os.Exit(1)
	}
}

func createTaskProcessor(cfg *config.Config) distributed.TaskProcessor {
	// Load user agents
	agentLoader := useragent.LoadFromDirs([]string{".", "..", "./agents"})

	return func(ctx context.Context, task *distributed.Task) (*distributed.TaskResult, error) {
		start := time.Now()

		fmt.Printf("[Worker] Processing task: %s -> %s\n", task.ID, task.URL)

		// Create a minimal simulator for this task
		result := &distributed.TaskResult{
			Timestamp: start,
		}

		// SECURITY FIX: Validate task URL before processing
		if task.URL == "" {
			result.Success = false
			result.Error = "empty URL"
			result.ResponseTime = time.Since(start)
			return result, fmt.Errorf("empty URL")
		}

		// Build visit config
		visitCfg := &config.Config{
			TargetDomain:        extractDomain(task.URL),
			MaxPages:            1,
			DurationMinutes:     1,
			HitsPerMinute:       60,
			MaxConcurrentVisits: 1,
			OutputDir:           "./reports",
			ExportFormat:        "none", // No reporting in worker mode
		}
		visitCfg.ApplyDefaults()
		visitCfg.ComputeDerived()

		// Use proxy if provided
		if task.Proxy != nil {
			visitCfg.ProxyHost = task.Proxy.Host
			visitCfg.ProxyPort = task.Proxy.Port
			visitCfg.ProxyUser = task.Proxy.Username
			visitCfg.ProxyPass = task.Proxy.Password
			visitCfg.ProxyEnabled = true
			visitCfg.ComputeDerived()
		}

		// Run single visit
		visitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		// Simple HTTP GET instead of full simulation for distributed mode
		// This is faster and more suitable for workers
		client := createHTTPClient(task.Proxy, cfg)
		req, err := http.NewRequestWithContext(visitCtx, "GET", task.URL, nil)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.ResponseTime = time.Since(start)
			return result, err
		}

		// Set headers to mimic browser
		req.Header.Set("User-Agent", agentLoader.Random())
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")

		resp, err := client.Do(req)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.ResponseTime = time.Since(start)
			fmt.Printf("[Worker] Task failed: %s - %v\n", task.ID, err)
			return result, err
		}
		defer resp.Body.Close()

		result.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
		result.StatusCode = resp.StatusCode
		result.ResponseTime = time.Since(start)

		if result.Success {
			fmt.Printf("[Worker] Task completed: %s - %d (%v)\n",
				task.ID, resp.StatusCode, result.ResponseTime)
		} else {
			fmt.Printf("[Worker] Task failed: %s - %d\n", task.ID, resp.StatusCode)
		}

		// Suppress visitCfg unused warning (used for config setup)
		_ = visitCfg

		return result, nil
	}
}

func createHTTPClient(proxyCfg *proxy.ProxyConfig, cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxyCfg != nil && proxyCfg.Host != "" {
		// Note: In production, you'd set up the proxy transport here
		// For now, we use direct connection
		_ = proxyCfg
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// SECURITY FIX: Safe domain extraction using net/url package
// Prevents panic from unsafe string slicing and validates URL format
func extractDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	
	// Parse URL safely using net/url package
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// Fallback: try to extract domain manually but safely
		return extractDomainFallback(rawURL)
	}
	
	// Return hostname (without port)
	host := parsedURL.Hostname()
	if host != "" {
		return host
	}
	
	// If no host found, try fallback
	return extractDomainFallback(rawURL)
}

// extractDomainFallback safely extracts domain without panicking
func extractDomainFallback(rawURL string) string {
	// Remove protocol prefix safely
	s := rawURL
	if len(s) > 7 && s[:7] == "http://" {
		s = s[7:]
	} else if len(s) > 8 && s[:8] == "https://" {
		s = s[8:]
	}
	
	// Find first slash
	for i, c := range s {
		if c == '/' || c == '?' || c == '#' {
			return s[:i]
		}
	}
	
	// Remove port if present
	for i, c := range s {
		if c == ':' {
			return s[:i]
		}
	}
	
	return s
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
