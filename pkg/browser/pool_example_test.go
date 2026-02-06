// Package browser provides usage examples for the browser pool
package browser

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// Example_basicUsage demonstrates basic pool usage
func Example_basicUsage() {
	// Create pool with default configuration
	config := DefaultPoolConfig()
	config.MaxInstances = 5
	config.MinInstances = 2

	pool, err := NewBrowserPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Acquire instance from pool
	ctx := context.Background()
	instance, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Use the instance
	tabCtx := instance.GetContext()
	err = chromedp.Run(tabCtx,
		chromedp.Navigate("https://example.com"),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Navigation error: %v", err)
	}

	// Release instance back to pool (will be reset automatically)
	pool.Release(instance)

	fmt.Println("Basic usage completed")
}

// Example_parallelUsage demonstrates parallel usage with multiple workers
func Example_parallelUsage() {
	config := DefaultPoolConfig()
	config.MaxInstances = 10
	config.MinInstances = 3

	pool, err := NewBrowserPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	urls := []string{
		"https://example.com",
		"https://example.org",
		"https://example.net",
	}

	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(targetURL string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			instance, err := pool.Acquire(ctx)
			if err != nil {
				log.Printf("Failed to acquire browser: %v", err)
				return
			}
			defer pool.Release(instance)

			tabCtx := instance.GetContext()
			err = chromedp.Run(tabCtx,
				chromedp.Navigate(targetURL),
				chromedp.WaitReady("body", chromedp.ByQuery),
			)
			if err != nil {
				log.Printf("Failed to visit %s: %v", targetURL, err)
				return
			}

			fmt.Printf("Successfully visited: %s\n", targetURL)
		}(url)
	}

	wg.Wait()
}

// Example_withProxy demonstrates pool usage with proxy configuration
func Example_withProxy() {
	config := PoolConfig{
		MaxInstances:   5,
		MinInstances:   2,
		AcquireTimeout: 30 * time.Second,
		ProxyURL:       "http://proxy.example.com:8080",
		ProxyUser:      "username",
		ProxyPass:      "password",
		Headless:       true,
	}

	pool, err := NewBrowserPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	instance, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Release(instance)

	// Use instance with proxy
	tabCtx := instance.GetContext()
	err = chromedp.Run(tabCtx,
		chromedp.Navigate("https://api.ipify.org?format=json"),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		log.Printf("Navigation error: %v", err)
	}

	fmt.Println("Proxy usage completed")
}

// Example_metrics demonstrates monitoring pool metrics
func Example_metrics() {
	config := DefaultPoolConfig()
	config.MaxInstances = 5

	pool, err := NewBrowserPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Do some work
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		instance, err := pool.Acquire(ctx)
		cancel()

		if err != nil {
			log.Printf("Acquire failed: %v", err)
			continue
		}

		// Simulate work
		time.Sleep(100 * time.Millisecond)
		pool.Release(instance)
	}

	// Get metrics
	metrics := pool.GetMetrics()
	fmt.Printf("Pool Metrics:\n")
	fmt.Printf("  Total Created: %d\n", metrics.TotalCreated)
	fmt.Printf("  Total Destroyed: %d\n", metrics.TotalDestroyed)
	fmt.Printf("  Total Reused: %d\n", metrics.TotalReused)
	fmt.Printf("  Current Active: %d\n", metrics.CurrentActive)
	fmt.Printf("  Current Idle: %d\n", metrics.CurrentIdle)
	fmt.Printf("  Acquire Waits: %d\n", metrics.AcquireWaits)
}

// Example_forceReset demonstrates deep cleanup when needed
func Example_forceReset() {
	pool, err := NewBrowserPool(DefaultPoolConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	instance, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Do some work that might store data
	tabCtx := instance.GetContext()
	chromedp.Run(tabCtx,
		chromedp.Navigate("https://example.com"),
		chromedp.Evaluate(`localStorage.setItem('test', 'value')`, nil),
	)

	// Force deep reset before releasing
	if err := pool.ForceReset(instance); err != nil {
		log.Printf("Force reset failed: %v", err)
	}

	pool.Release(instance)
	fmt.Println("Force reset completed")
}

// Example_multiTab demonstrates creating multiple tabs in same instance
func Example_multiTab() {
	pool, err := NewBrowserPool(DefaultPoolConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()
	instance, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Release(instance)

	// First tab
	tabCtx1 := instance.GetContext()
	chromedp.Run(tabCtx1,
		chromedp.Navigate("https://example.com"),
		chromedp.Sleep(1*time.Second),
	)

	// Create new tab in same browser instance
	tabCtx2, tabCancel2, err := instance.CreateNewTab(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tabCancel2()

	chromedp.Run(tabCtx2,
		chromedp.Navigate("https://example.org"),
		chromedp.Sleep(1*time.Second),
	)

	fmt.Println("Multi-tab usage completed")
}

// Example_poolWithTimeout demonstrates handling pool exhaustion
func Example_poolWithTimeout() {
	config := PoolConfig{
		MaxInstances:   2,
		AcquireTimeout: 5 * time.Second,
	}

	pool, err := NewBrowserPool(config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Acquire both instances
	inst1, _ := pool.Acquire(context.Background())
	inst2, _ := pool.Acquire(context.Background())

	// This will timeout since pool is exhausted
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = pool.Acquire(ctx)
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
	}

	// Release instances
	pool.Release(inst1)
	pool.Release(inst2)
}
