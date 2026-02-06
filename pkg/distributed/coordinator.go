// Package distributed - Master-Worker Distributed Mode for ErosHit
package distributed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"eroshit/pkg/behavior"
	"eroshit/pkg/proxy"
)

// TaskStatus task durumu
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskAssigned  TaskStatus = "assigned"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// Task bir ziyaret task'ı
type Task struct {
	ID          string                   `json:"id"`
	URL         string                   `json:"url"`
	Proxy       *proxy.ProxyConfig       `json:"proxy,omitempty"`
	Profile     *behavior.BehaviorProfile `json:"profile,omitempty"`
	SessionID   string                   `json:"session_id"`
	Status      TaskStatus               `json:"status"`
	WorkerID    string                   `json:"worker_id,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	AssignedAt  *time.Time               `json:"assigned_at,omitempty"`
	CompletedAt *time.Time               `json:"completed_at,omitempty"`
	Result      *TaskResult              `json:"result,omitempty"`
}

// TaskResult task sonucu
type TaskResult struct {
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	PageTitle    string        `json:"page_title,omitempty"`
	StatusCode   int           `json:"status_code"`
	Timestamp    time.Time     `json:"timestamp"`
}

// WorkerInfo worker bilgisi
type WorkerInfo struct {
	ID             string    `json:"id"`
	Hostname       string    `json:"hostname"`
	IPAddress      string    `json:"ip_address"`
	MaxConcurrency int       `json:"max_concurrency"`
	ActiveTasks    int       `json:"active_tasks"`
	TotalTasks     int64     `json:"total_tasks"`
	SuccessCount   int64     `json:"success_count"`
	FailedCount    int64     `json:"failed_count"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	Status         string    `json:"status"`
	Version        string    `json:"version"`
}

// IsHealthy worker'ın sağlıklı olup olmadığını kontrol eder
func (w *WorkerInfo) IsHealthy() bool {
	return time.Since(w.LastHeartbeat) < 30*time.Second && w.Status == "active"
}

// MasterConfig master yapılandırması
type MasterConfig struct {
	BindAddr      string
	SecretKey     string
	MaxWorkers    int
	TaskTimeout   time.Duration
	HeartbeatInterval time.Duration
}

// DefaultMasterConfig varsayılan master config
func DefaultMasterConfig() MasterConfig {
	return MasterConfig{
		BindAddr:          "0.0.0.0:8080",
		SecretKey:         "",
		MaxWorkers:        100,
		TaskTimeout:       5 * time.Minute,
		HeartbeatInterval: 10 * time.Second,
	}
}

// Master distributed mode master node
type Master struct {
	config MasterConfig

	// Task queue
	taskQueue   chan *Task
	tasks       map[string]*Task
	tasksMu     sync.RWMutex

	// Workers
	workers     map[string]*WorkerInfo
	workersMu   sync.RWMutex

	// Statistics
	totalTasks     int64
	completedTasks int64
	failedTasks    int64

	// HTTP server
	server  *http.Server
	running int32

	// Context
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewMaster yeni master oluşturur
func NewMaster(config MasterConfig) *Master {
	if config.TaskTimeout == 0 {
		config.TaskTimeout = 5 * time.Minute
	}
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Master{
		config:    config,
		taskQueue: make(chan *Task, 10000),
		tasks:     make(map[string]*Task),
		workers:   make(map[string]*WorkerInfo),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start master'ı başlatır
func (m *Master) Start() error {
	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return fmt.Errorf("master already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/worker/register", m.authMiddleware(m.handleWorkerRegister))
	mux.HandleFunc("/api/v1/worker/heartbeat", m.authMiddleware(m.handleWorkerHeartbeat))
	mux.HandleFunc("/api/v1/worker/task/request", m.authMiddleware(m.handleTaskRequest))
	mux.HandleFunc("/api/v1/worker/task/complete", m.authMiddleware(m.handleTaskComplete))
	mux.HandleFunc("/api/v1/worker/task/fail", m.authMiddleware(m.handleTaskFail))
	mux.HandleFunc("/api/v1/master/status", m.handleMasterStatus)
	mux.HandleFunc("/api/v1/master/workers", m.authMiddleware(m.handleListWorkers))
	mux.HandleFunc("/api/v1/master/tasks", m.authMiddleware(m.handleListTasks))
	mux.HandleFunc("/api/v1/master/task/submit", m.authMiddleware(m.handleSubmitTask))
	mux.HandleFunc("/api/v1/master/stats", m.authMiddleware(m.handleStats))

	m.server = &http.Server{
		Addr:    m.config.BindAddr,
		Handler: mux,
	}

	// Cleanup goroutine
	go m.cleanupLoop()

	fmt.Printf("[Master] Starting on %s\n", m.config.BindAddr)
	return m.server.ListenAndServe()
}

// Stop master'ı durdurur
func (m *Master) Stop() error {
	if !atomic.CompareAndSwapInt32(&m.running, 1, 0) {
		return nil
	}

	m.cancel()
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}

// SubmitTask yeni task gönderir
func (m *Master) SubmitTask(task *Task) error {
	if atomic.LoadInt32(&m.running) == 0 {
		return fmt.Errorf("master not running")
	}

	task.ID = generateTaskID()
	task.Status = TaskPending
	task.CreatedAt = time.Now()

	m.tasksMu.Lock()
	m.tasks[task.ID] = task
	m.tasksMu.Unlock()

	atomic.AddInt64(&m.totalTasks, 1)

	select {
	case m.taskQueue <- task:
		return nil
	case <-m.ctx.Done():
		return fmt.Errorf("master shutting down")
	default:
		return fmt.Errorf("task queue full")
	}
}

// SubmitTasks çoklu task gönderir
func (m *Master) SubmitTasks(tasks []*Task) error {
	for _, task := range tasks {
		if err := m.SubmitTask(task); err != nil {
			return err
		}
	}
	return nil
}

// GetStats istatistikleri döner
func (m *Master) GetStats() MasterStats {
	return MasterStats{
		TotalTasks:     atomic.LoadInt64(&m.totalTasks),
		CompletedTasks: atomic.LoadInt64(&m.completedTasks),
		FailedTasks:    atomic.LoadInt64(&m.failedTasks),
		PendingTasks:   int64(len(m.taskQueue)),
		ActiveWorkers:  int64(len(m.GetHealthyWorkers())),
	}
}

// GetHealthyWorkers sağlıklı worker'ları döner
func (m *Master) GetHealthyWorkers() []*WorkerInfo {
	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	var healthy []*WorkerInfo
	for _, w := range m.workers {
		if w.IsHealthy() {
			healthy = append(healthy, w)
		}
	}
	return healthy
}

// HTTP Handlers

func (m *Master) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.config.SecretKey != "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer "+m.config.SecretKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

func (m *Master) handleWorkerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var worker WorkerInfo
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	worker.LastHeartbeat = time.Now()
	worker.Status = "active"

	m.workersMu.Lock()
	m.workers[worker.ID] = &worker
	m.workersMu.Unlock()

	fmt.Printf("[Master] Worker registered: %s (%s)\n", worker.ID, worker.Hostname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "registered",
		"worker_id": worker.ID,
	})
}

func (m *Master) handleWorkerHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkerID     string `json:"worker_id"`
		ActiveTasks  int    `json:"active_tasks"`
		TotalTasks   int64  `json:"total_tasks"`
		SuccessCount int64  `json:"success_count"`
		FailedCount  int64  `json:"failed_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.workersMu.Lock()
	if worker, ok := m.workers[req.WorkerID]; ok {
		worker.LastHeartbeat = time.Now()
		worker.ActiveTasks = req.ActiveTasks
		worker.TotalTasks = req.TotalTasks
		worker.SuccessCount = req.SuccessCount
		worker.FailedCount = req.FailedCount
	}
	m.workersMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (m *Master) handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkerID string `json:"worker_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	select {
	case task := <-m.taskQueue:
		now := time.Now()
		task.Status = TaskAssigned
		task.WorkerID = req.WorkerID
		task.AssignedAt = &now

		m.tasksMu.Lock()
		m.tasks[task.ID] = task
		m.tasksMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	case <-time.After(5 * time.Second):
		w.WriteHeader(http.StatusNoContent)
	}
}

func (m *Master) handleTaskComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TaskID string     `json:"task_id"`
		Result TaskResult `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now()

	m.tasksMu.Lock()
	if task, ok := m.tasks[req.TaskID]; ok {
		task.Status = TaskCompleted
		task.CompletedAt = &now
		task.Result = &req.Result
	}
	m.tasksMu.Unlock()

	atomic.AddInt64(&m.completedTasks, 1)
	w.WriteHeader(http.StatusOK)
}

func (m *Master) handleTaskFail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TaskID string `json:"task_id"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.tasksMu.Lock()
	if task, ok := m.tasks[req.TaskID]; ok {
		task.Status = TaskFailed
		now := time.Now()
		task.CompletedAt = &now
		task.Result = &TaskResult{
			Success:   false,
			Error:     req.Error,
			Timestamp: now,
		}
	}
	m.tasksMu.Unlock()

	atomic.AddInt64(&m.failedTasks, 1)
	w.WriteHeader(http.StatusOK)
}

func (m *Master) handleMasterStatus(w http.ResponseWriter, r *http.Request) {
	stats := m.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (m *Master) handleListWorkers(w http.ResponseWriter, r *http.Request) {
	m.workersMu.RLock()
	workers := make([]*WorkerInfo, 0, len(m.workers))
	for _, w := range m.workers {
		workers = append(workers, w)
	}
	m.workersMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

func (m *Master) handleListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	m.tasksMu.RLock()
	tasks := make([]*Task, 0)
	for _, t := range m.tasks {
		if status == "" || string(t.Status) == status {
			tasks = append(tasks, t)
		}
	}
	m.tasksMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (m *Master) handleSubmitTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := m.SubmitTask(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "submitted",
		"task_id": task.ID,
	})
}

func (m *Master) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := m.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (m *Master) cleanupLoop() {
	ticker := time.NewTicker(m.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupStaleWorkers()
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Master) cleanupStaleWorkers() {
	m.workersMu.Lock()
	defer m.workersMu.Unlock()

	now := time.Now()
	for id, worker := range m.workers {
		if now.Sub(worker.LastHeartbeat) > 2*m.config.HeartbeatInterval {
			worker.Status = "offline"
			fmt.Printf("[Master] Worker marked offline: %s\n", id)
		}
	}
}

// MasterStats master istatistikleri
type MasterStats struct {
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	PendingTasks   int64 `json:"pending_tasks"`
	ActiveWorkers  int64 `json:"active_workers"`
}

// ==================== WORKER ====================

// WorkerConfig worker yapılandırması
type WorkerConfig struct {
	MasterURL      string
	SecretKey      string
	MaxConcurrency int
	Hostname       string
	Version        string
}

// DefaultWorkerConfig varsayılan worker config
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		MasterURL:      "http://localhost:8080",
		MaxConcurrency: 10,
		Version:        "1.0.0",
	}
}

// Worker distributed mode worker node
type Worker struct {
	config WorkerConfig
	ID     string

	// State
	activeTasks  int32
	totalTasks   int64
	successCount int64
	failedCount  int64

	// HTTP client
	client *http.Client

	// Task processor
	taskProcessor TaskProcessor

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	running int32

	// Worker info
	info *WorkerInfo
}

// TaskProcessor task işleme fonksiyonu
type TaskProcessor func(ctx context.Context, task *Task) (*TaskResult, error)

// NewWorker yeni worker oluşturur
func NewWorker(config WorkerConfig, processor TaskProcessor) *Worker {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		config:        config,
		ID:            generateWorkerID(),
		client:        &http.Client{Timeout: 30 * time.Second},
		taskProcessor: processor,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start worker'ı başlatır
func (w *Worker) Start() error {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return fmt.Errorf("worker already running")
	}

	// Register with master
	if err := w.register(); err != nil {
		atomic.StoreInt32(&w.running, 0)
		return fmt.Errorf("failed to register: %w", err)
	}

	fmt.Printf("[Worker] Started: %s -> %s\n", w.ID, w.config.MasterURL)

	// Start heartbeat goroutine
	go w.heartbeatLoop()

	// Start task processing goroutines
	for i := 0; i < w.config.MaxConcurrency; i++ {
		go w.taskLoop()
	}

	<-w.ctx.Done()
	return nil
}

// Stop worker'ı durdurur
func (w *Worker) Stop() {
	if atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		w.cancel()
	}
}

// IsRunning worker'ın çalışıp çalışmadığını kontrol eder
func (w *Worker) IsRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

func (w *Worker) register() error {
	w.info = &WorkerInfo{
		ID:             w.ID,
		Hostname:       w.config.Hostname,
		MaxConcurrency: w.config.MaxConcurrency,
		Version:        w.config.Version,
		Status:         "active",
	}

	data, _ := json.Marshal(w.info)
	req, err := http.NewRequest("POST", w.config.MasterURL+"/api/v1/worker/register", bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.config.SecretKey)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: %s", resp.Status)
	}

	return nil
}

func (w *Worker) heartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.sendHeartbeat()
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Worker) sendHeartbeat() {
	data, _ := json.Marshal(map[string]interface{}{
		"worker_id":     w.ID,
		"active_tasks":  atomic.LoadInt32(&w.activeTasks),
		"total_tasks":   atomic.LoadInt64(&w.totalTasks),
		"success_count": atomic.LoadInt64(&w.successCount),
		"failed_count":  atomic.LoadInt64(&w.failedCount),
	})

	req, err := http.NewRequest("POST", w.config.MasterURL+"/api/v1/worker/heartbeat", bytes.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if w.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.config.SecretKey)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (w *Worker) taskLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.requestAndProcessTask()
		}
	}
}

func (w *Worker) requestAndProcessTask() {
	// Request task from master
	data, _ := json.Marshal(map[string]string{
		"worker_id": w.ID,
	})

	req, err := http.NewRequest("POST", w.config.MasterURL+"/api/v1/worker/task/request", bytes.NewReader(data))
	if err != nil {
		time.Sleep(5 * time.Second)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if w.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.config.SecretKey)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		time.Sleep(5 * time.Second)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		time.Sleep(2 * time.Second)
		return
	}

	if resp.StatusCode != http.StatusOK {
		time.Sleep(5 * time.Second)
		return
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return
	}

	// Process task
	atomic.AddInt32(&w.activeTasks, 1)
	atomic.AddInt64(&w.totalTasks, 1)

	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	result, err := w.taskProcessor(ctx, &task)
	cancel()

	atomic.AddInt32(&w.activeTasks, -1)

	if err != nil || !result.Success {
		atomic.AddInt64(&w.failedCount, 1)
		w.reportTaskFail(task.ID, err)
	} else {
		atomic.AddInt64(&w.successCount, 1)
		w.reportTaskComplete(task.ID, result)
	}
}

func (w *Worker) reportTaskComplete(taskID string, result *TaskResult) {
	data, _ := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"result":  result,
	})

	req, _ := http.NewRequest("POST", w.config.MasterURL+"/api/v1/worker/task/complete", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if w.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.config.SecretKey)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (w *Worker) reportTaskFail(taskID string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	data, _ := json.Marshal(map[string]string{
		"task_id": taskID,
		"error":   errMsg,
	})

	req, _ := http.NewRequest("POST", w.config.MasterURL+"/api/v1/worker/task/fail", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if w.config.SecretKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.config.SecretKey)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// GetStats worker istatistiklerini döner
func (w *Worker) GetStats() WorkerStats {
	return WorkerStats{
		WorkerID:     w.ID,
		ActiveTasks:  int(atomic.LoadInt32(&w.activeTasks)),
		TotalTasks:   atomic.LoadInt64(&w.totalTasks),
		SuccessCount: atomic.LoadInt64(&w.successCount),
		FailedCount:  atomic.LoadInt64(&w.failedCount),
	}
}

// WorkerStats worker istatistikleri
type WorkerStats struct {
	WorkerID     string `json:"worker_id"`
	ActiveTasks  int    `json:"active_tasks"`
	TotalTasks   int64  `json:"total_tasks"`
	SuccessCount int64  `json:"success_count"`
	FailedCount  int64  `json:"failed_count"`
}

// Helper functions

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func generateWorkerID() string {
	return fmt.Sprintf("worker_%d", time.Now().UnixNano())
}
