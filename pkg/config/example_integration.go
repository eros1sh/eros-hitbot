// +build ignore

// This file demonstrates how to integrate the config reloader with your application.
// It's not meant to be compiled directly.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	configpkg "eroshit/pkg/config"
	"go.uber.org/zap"
)

// ZapLogger implements config.Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() *ZapLogger {
	logger, _ := zap.NewProduction()
	return &ZapLogger{logger: logger}
}

func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, zap.Any("fields", fields))
}

// SimulationManager manages running simulations
type SimulationManager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex
	cfg       *configpkg.Config
}

func NewSimulationManager() *SimulationManager {
	return &SimulationManager{}
}

func (sm *SimulationManager) Start(cfg *configpkg.Config) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.isRunning {
		log.Println("Simulation already running")
		return
	}
	
	sm.cfg = cfg
	sm.isRunning = true
	sm.ctx, sm.cancel = context.WithCancel(context.Background())
	
	sm.wg.Add(1)
	go sm.run()
	
	log.Printf("Simulation started: target=%s, hits/min=%d", cfg.TargetDomain, cfg.HitsPerMinute)
}

func (sm *SimulationManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.isRunning {
		return
	}
	
	log.Println("Stopping simulation gracefully...")
	sm.cancel()
	
	// Wait with timeout
	done := make(chan struct{})
	go func() {
		sm.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Println("Simulation stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Simulation stop timeout, forcing exit")
	}
	
	sm.isRunning = false
}

func (sm *SimulationManager) Restart(cfg *configpkg.Config) {
	log.Println("Restarting simulation with new config...")
	sm.Stop()
	time.Sleep(100 * time.Millisecond) // Brief pause
	sm.Start(cfg)
}

func (sm *SimulationManager) run() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.mu.Lock()
			cfg := sm.cfg
			sm.mu.Unlock()
			
			if cfg != nil {
				log.Printf("Simulating visit to %s", cfg.TargetDomain)
			}
		}
	}
}

func (sm *SimulationManager) IsRunning() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.isRunning
}

func main() {
	// Initialize logger
	logger := NewZapLogger()
	
	// Create simulation manager
	simManager := NewSimulationManager()
	
	// Create reloader
	reloader := configpkg.NewReloader("config.yaml")
	reloader.SetLogger(logger)
	
	// Register callback for config changes
	reloader.OnChange(func(newCfg *configpkg.Config) {
		log.Println("=== Config changed! ===")
		
		// Show what changed
		oldCfg := reloader.GetConfig()
		diff := configpkg.Diff(oldCfg, newCfg)
		for field, values := range diff {
			log.Printf("  %s: %v -> %v", field, values.Old, values.New)
		}
		
		// Restart simulation if running
		if simManager.IsRunning() {
			simManager.Restart(newCfg)
		} else {
			simManager.Start(newCfg)
		}
		
		log.Println("======================")
	})
	
	// Start reloader
	if err := reloader.Start(); err != nil {
		log.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()
	
	// Start initial simulation
	if cfg := reloader.GetConfig(); cfg != nil {
		simManager.Start(cfg)
	}
	
	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Application running. Edit config.yaml to trigger reload.")
	log.Println("Press Ctrl+C to exit.")
	
	<-sigCh
	log.Println("Shutting down...")
	
	// Cleanup
	simManager.Stop()
	
	log.Println("Goodbye!")
}
