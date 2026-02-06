package distributed

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMasterWorkerIntegration(t *testing.T) {
	// Start master
	masterConfig := MasterConfig{
		BindAddr:          "127.0.0.1:18080",
		SecretKey:         "test-secret",
		MaxWorkers:        10,
		TaskTimeout:       5 * time.Second,
		HeartbeatInterval: 1 * time.Second,
	}

	master := NewMaster(masterConfig)

	// Start master in background
	go func() {
		if err := master.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("Master stopped: %v", err)
		}
	}()

	// Wait for master to start
	time.Sleep(500 * time.Millisecond)

	defer master.Stop()

	// Test 1: Check master status
	t.Run("MasterStatus", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:18080/api/v1/master/status")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var stats MasterStats
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			t.Fatalf("Failed to decode stats: %v", err)
		}

		if stats.TotalTasks != 0 {
			t.Errorf("Expected 0 tasks, got %d", stats.TotalTasks)
		}
	})

	// Test 2: Submit task via API
	t.Run("SubmitTask", func(t *testing.T) {
		task := Task{
			URL:       "http://example.com",
			SessionID: "test-session-1",
		}

		data, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", "http://127.0.0.1:18080/api/v1/master/task/submit", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-secret")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["status"] != "submitted" {
			t.Errorf("Expected status 'submitted', got %s", result["status"])
		}
	})

	// Test 3: List tasks
	t.Run("ListTasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:18080/api/v1/master/tasks", nil)
		req.Header.Set("Authorization", "Bearer test-secret")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to list tasks: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var tasks []*Task
		if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
			t.Fatalf("Failed to decode tasks: %v", err)
		}

		if len(tasks) == 0 {
			t.Error("Expected at least one task")
		}
	})

	// Test 4: Worker registration
	t.Run("WorkerRegistration", func(t *testing.T) {
		worker := WorkerInfo{
			ID:             "test-worker-1",
			Hostname:       "test-host",
			MaxConcurrency: 5,
			Version:        "1.0.0",
		}

		data, _ := json.Marshal(worker)
		req, _ := http.NewRequest("POST", "http://127.0.0.1:18080/api/v1/worker/register", strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-secret")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to register worker: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check workers list
		req2, _ := http.NewRequest("GET", "http://127.0.0.1:18080/api/v1/master/workers", nil)
		req2.Header.Set("Authorization", "Bearer test-secret")

		resp2, err := http.DefaultClient.Do(req2)
		if err != nil {
			t.Fatalf("Failed to list workers: %v", err)
		}
		defer resp2.Body.Close()

		var workers []*WorkerInfo
		if err := json.NewDecoder(resp2.Body).Decode(&workers); err != nil {
			t.Fatalf("Failed to decode workers: %v", err)
		}

		found := false
		for _, w := range workers {
			if w.ID == "test-worker-1" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find registered worker")
		}
	})
}

func TestWorkerTaskProcessing(t *testing.T) {
	// Start master
	masterConfig := MasterConfig{
		BindAddr:          "127.0.0.1:18081",
		SecretKey:         "",
		MaxWorkers:        10,
		TaskTimeout:       10 * time.Second,
		HeartbeatInterval: 1 * time.Second,
	}

	master := NewMaster(masterConfig)
	go master.Start()
	time.Sleep(500 * time.Millisecond)
	defer master.Stop()

	// Create a test processor
	processor := func(ctx context.Context, task *Task) (*TaskResult, error) {
		return &TaskResult{
			Success:      true,
			StatusCode:   200,
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}, nil
	}

	// Create and start worker
	workerConfig := WorkerConfig{
		MasterURL:      "http://127.0.0.1:18081",
		MaxConcurrency: 2,
		Hostname:       "test-worker",
		Version:        "1.0.0",
	}

	worker := NewWorker(workerConfig, processor)
	go worker.Start()
	defer worker.Stop()

	// Wait for worker to register
	time.Sleep(1 * time.Second)

	// Submit a task
	task := &Task{
		URL:       "http://example.com/test",
		SessionID: "test-session",
	}

	if err := master.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Wait for task to be processed
	time.Sleep(3 * time.Second)

	// Check stats
	stats := master.GetStats()
	if stats.TotalTasks != 1 {
		t.Errorf("Expected 1 total task, got %d", stats.TotalTasks)
	}

	// Task should be completed or still pending (depending on timing)
	if stats.CompletedTasks+stats.FailedTasks+stats.PendingTasks != stats.TotalTasks {
		t.Errorf("Task count mismatch: total=%d, completed=%d, failed=%d, pending=%d",
			stats.TotalTasks, stats.CompletedTasks, stats.FailedTasks, stats.PendingTasks)
	}
}

func TestAuthMiddleware(t *testing.T) {
	masterConfig := MasterConfig{
		BindAddr:  "127.0.0.1:18082",
		SecretKey: "secret123",
	}

	master := NewMaster(masterConfig)
	go master.Start()
	time.Sleep(300 * time.Millisecond)
	defer master.Stop()

	// Test without auth - should fail
	t.Run("NoAuth", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:18082/api/v1/master/workers")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	// Test with wrong auth - should fail
	t.Run("WrongAuth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:18082/api/v1/master/workers", nil)
		req.Header.Set("Authorization", "Bearer wrong-secret")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	// Test with correct auth - should succeed
	t.Run("CorrectAuth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:18082/api/v1/master/workers", nil)
		req.Header.Set("Authorization", "Bearer secret123")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestTaskLifecycle(t *testing.T) {
	// Create a task and verify its lifecycle
	task := &Task{
		ID:        "test-task-1",
		URL:       "http://example.com",
		SessionID: "session-1",
		Status:    TaskPending,
		CreatedAt: time.Now(),
	}

	if task.Status != TaskPending {
		t.Error("Expected initial status to be pending")
	}

	// Simulate assignment
	now := time.Now()
	task.Status = TaskAssigned
	task.AssignedAt = &now
	task.WorkerID = "worker-1"

	if task.Status != TaskAssigned {
		t.Error("Expected status to be assigned")
	}

	// Simulate completion
	now = time.Now()
	task.Status = TaskCompleted
	task.CompletedAt = &now
	task.Result = &TaskResult{
		Success:    true,
		StatusCode: 200,
	}

	if task.Status != TaskCompleted {
		t.Error("Expected status to be completed")
	}

	if !task.Result.Success {
		t.Error("Expected result to be success")
	}
}

func TestWorkerHealth(t *testing.T) {
	// Healthy worker
	healthyWorker := &WorkerInfo{
		LastHeartbeat: time.Now(),
		Status:        "active",
	}

	if !healthyWorker.IsHealthy() {
		t.Error("Expected worker to be healthy")
	}

	// Stale heartbeat
	staleWorker := &WorkerInfo{
		LastHeartbeat: time.Now().Add(-60 * time.Second),
		Status:        "active",
	}

	if staleWorker.IsHealthy() {
		t.Error("Expected worker to be unhealthy (stale heartbeat)")
	}

	// Offline status
	offlineWorker := &WorkerInfo{
		LastHeartbeat: time.Now(),
		Status:        "offline",
	}

	if offlineWorker.IsHealthy() {
		t.Error("Expected worker to be unhealthy (offline status)")
	}
}
