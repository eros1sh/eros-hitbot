// Master Node CLI for ErosHit Distributed Mode
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"eroshit/pkg/distributed"
)

func main() {
	var (
		bindAddr   = flag.String("bind", "0.0.0.0:8080", "Master bind address")
		secretKey  = flag.String("secret", "", "Secret key for worker authentication")
		configFile = flag.String("config", "", "Config file to load tasks from")
	)
	flag.Parse()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           ErosHit - Distributed Master Node                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create master
	config := distributed.MasterConfig{
		BindAddr:          *bindAddr,
		SecretKey:         *secretKey,
		MaxWorkers:        100,
		TaskTimeout:       5 * time.Minute,
		HeartbeatInterval: 10 * time.Second,
	}

	master := distributed.NewMaster(config)

	// Handle shutdown gracefully
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[Master] Shutting down...")
		master.Stop()
		cancel()
	}()

	// Load tasks from config if provided
	if *configFile != "" {
		go loadTasksFromFile(master, *configFile)
	}

	// Start interactive console in background
	go interactiveConsole(master)

	// Print status URL
	fmt.Printf("[Master] Listening on http://%s\n", *bindAddr)
	fmt.Printf("[Master] Status: http://%s/api/v1/master/status\n", *bindAddr)
	fmt.Printf("[Master] Workers: http://%s/api/v1/master/workers\n", *bindAddr)
	fmt.Printf("[Master] Tasks: http://%s/api/v1/master/tasks\n", *bindAddr)
	fmt.Printf("[Master] Stats: http://%s/api/v1/master/stats\n", *bindAddr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Start master (blocking)
	if err := master.Start(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "[Master] Error: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
	fmt.Println("[Master] Stopped")
}

func interactiveConsole(master *distributed.Master) {
	time.Sleep(1 * time.Second) // Wait for master to start

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("master> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "help":
			printHelp()
		case "status", "stats":
			printStats(master)
		case "submit":
			if len(parts) < 2 {
				fmt.Println("Usage: submit <url>")
				continue
			}
			submitTask(master, parts[1])
		case "batch":
			if len(parts) < 3 {
				fmt.Println("Usage: batch <url> <count>")
				continue
			}
			count := 1
			fmt.Sscanf(parts[2], "%d", &count)
			submitBatch(master, parts[1], count)
		case "workers":
			printWorkers(master)
		case "tasks":
			printTasks(master)
		case "quit", "exit":
			fmt.Println("Use Ctrl+C to stop the master")
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  help           - Show this help")
	fmt.Println("  status/stats   - Show master statistics")
	fmt.Println("  submit <url>   - Submit a single task")
	fmt.Println("  batch <url> <n> - Submit n tasks for URL")
	fmt.Println("  workers        - List connected workers")
	fmt.Println("  tasks          - List recent tasks")
	fmt.Println("  quit/exit      - Exit (same as Ctrl+C)")
}

func printStats(master *distributed.Master) {
	stats := master.GetStats()
	data, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(data))
}

func submitTask(master *distributed.Master, url string) {
	task := &distributed.Task{
		URL:     url,
		SessionID: fmt.Sprintf("session_%d", time.Now().Unix()),
	}

	if err := master.SubmitTask(task); err != nil {
		fmt.Printf("Error submitting task: %v\n", err)
		return
	}

	fmt.Printf("Task submitted: %s\n", task.ID)
}

func submitBatch(master *distributed.Master, url string, count int) {
	var tasks []*distributed.Task
	baseSession := fmt.Sprintf("session_%d", time.Now().Unix())

	for i := 0; i < count; i++ {
		task := &distributed.Task{
			URL:       url,
			SessionID: fmt.Sprintf("%s_%d", baseSession, i),
		}
		tasks = append(tasks, task)
	}

	if err := master.SubmitTasks(tasks); err != nil {
		fmt.Printf("Error submitting tasks: %v\n", err)
		return
	}

	fmt.Printf("Submitted %d tasks\n", count)
}

func printWorkers(master *distributed.Master) {
	workers := master.GetHealthyWorkers()
	if len(workers) == 0 {
		fmt.Println("No healthy workers connected")
		return
	}

	fmt.Printf("%-20s %-15s %-10s %-10s %-10s %-10s\n", 
		"ID", "Hostname", "Status", "Active", "Total", "Success")
	fmt.Println(strings.Repeat("-", 80))

	for _, w := range workers {
		fmt.Printf("%-20s %-15s %-10s %-10d %-10d %-10d\n",
			truncate(w.ID, 20),
			truncate(w.Hostname, 15),
			w.Status,
			w.ActiveTasks,
			w.TotalTasks,
			w.SuccessCount,
		)
	}
}

func printTasks(master *distributed.Master) {
	// This would need a method to get recent tasks from master
	fmt.Println("Use HTTP API: GET /api/v1/master/tasks")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func loadTasksFromFile(master *distributed.Master, filename string) {
	// Load tasks from a JSON config file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("[Master] Warning: Could not load config file: %v\n", err)
		return
	}

	var config struct {
		URLs     []string `json:"urls"`
		Tasks    []struct {
			URL       string `json:"url"`
			SessionID string `json:"session_id"`
			Count     int    `json:"count"`
		} `json:"tasks"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("[Master] Warning: Invalid config file: %v\n", err)
		return
	}

	total := 0

	// Submit simple URLs
	for _, url := range config.URLs {
		task := &distributed.Task{
			URL:       url,
			SessionID: fmt.Sprintf("session_%d", time.Now().UnixNano()),
		}
		if err := master.SubmitTask(task); err == nil {
			total++
		}
	}

	// Submit complex tasks
	for _, t := range config.Tasks {
		count := t.Count
		if count <= 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			sessionID := t.SessionID
			if sessionID == "" {
				sessionID = fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), i)
			}
			task := &distributed.Task{
				URL:       t.URL,
				SessionID: sessionID,
			}
			if err := master.SubmitTask(task); err == nil {
				total++
			}
		}
	}

	fmt.Printf("[Master] Loaded %d tasks from %s\n", total, filename)
}
