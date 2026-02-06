// Package worker provides distributed worker architecture with auto-scaling,
// priority queue management, and failure recovery for traffic simulation.
package worker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// JobStatus represents the current status of a job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusRunning
	JobStatusCompleted
	JobStatusFailed
	JobStatusRetrying
	JobStatusCancelled
)

// Job represents a unit of work to be executed
type Job struct {
	ID          string
	Type        string
	Priority    JobPriority
	Payload     interface{}
	Status      JobStatus
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Error       error
	Result      interface{}
	Timeout     time.Duration
	OnComplete  func(*Job)
	OnError     func(*Job, error)
}

// NewJob creates a new job with default settings
func NewJob(jobType string, payload interface{}) *Job {
	id := make([]byte, 8)
	rand.Read(id)
	return &Job{
		ID:         hex.EncodeToString(id),
		Type:       jobType,
		Priority:   PriorityNormal,
		Payload:    payload,
		Status:     JobStatusPending,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		Timeout:    90 * time.Second,
	}
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// WorkerPool manages a pool of workers for job execution
type WorkerPool struct {
	mu              sync.RWMutex
	workers         []*Worker
	minWorkers      int
	maxWorkers      int
	currentWorkers  int32
	jobQueue        *PriorityQueue
	handlers        map[string]JobHandler
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	metrics         *PoolMetrics
	autoScaler      *AutoScaler
	circuitBreaker  *CircuitBreaker
	running         bool
}

// Worker represents a single worker in the pool
type Worker struct {
	ID       string
	pool     *WorkerPool
	jobChan  chan *Job
	quit     chan struct{}
	busy     int32
	jobCount int64
}

// PoolMetrics tracks worker pool performance metrics
type PoolMetrics struct {
	mu                sync.RWMutex
	TotalJobs         int64
	CompletedJobs     int64
	FailedJobs        int64
	RetryJobs         int64
	ActiveWorkers     int32
	QueueSize         int64
	AvgProcessingTime float64
	processingTimes   []time.Duration
	LastUpdated       time.Time
}

// PoolConfig configuration for worker pool
type PoolConfig struct {
	MinWorkers       int
	MaxWorkers       int
	QueueSize        int
	EnableAutoScale  bool
	ScaleUpThreshold float64 // Queue utilization threshold to scale up
	ScaleDownThreshold float64 // Queue utilization threshold to scale down
	ScaleInterval    time.Duration
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	cpuCount := runtime.NumCPU()
	return PoolConfig{
		MinWorkers:              cpuCount,
		MaxWorkers:              cpuCount * 4,
		QueueSize:               10000,
		EnableAutoScale:         true,
		ScaleUpThreshold:        0.8,
		ScaleDownThreshold:      0.2,
		ScaleInterval:           5 * time.Second,
		CircuitBreakerThreshold: 10,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config PoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:    make([]*Worker, 0, config.MaxWorkers),
		minWorkers: config.MinWorkers,
		maxWorkers: config.MaxWorkers,
		jobQueue:   NewPriorityQueue(config.QueueSize),
		handlers:   make(map[string]JobHandler),
		ctx:        ctx,
		cancel:     cancel,
		metrics:    &PoolMetrics{},
		circuitBreaker: NewCircuitBreaker(
			config.CircuitBreakerThreshold,
			config.CircuitBreakerTimeout,
		),
	}

	if config.EnableAutoScale {
		pool.autoScaler = NewAutoScaler(pool, AutoScalerConfig{
			MinWorkers:         config.MinWorkers,
			MaxWorkers:         config.MaxWorkers,
			ScaleUpThreshold:   config.ScaleUpThreshold,
			ScaleDownThreshold: config.ScaleDownThreshold,
			ScaleInterval:      config.ScaleInterval,
		})
	}

	return pool
}

// RegisterHandler registers a job handler for a specific job type
func (p *WorkerPool) RegisterHandler(jobType string, handler JobHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[jobType] = handler
}

// Start starts the worker pool
func (p *WorkerPool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("worker pool already running")
	}

	// Start minimum number of workers
	for i := 0; i < p.minWorkers; i++ {
		worker := p.createWorker()
		p.workers = append(p.workers, worker)
		go worker.run()
	}
	atomic.StoreInt32(&p.currentWorkers, int32(p.minWorkers))

	// Start job dispatcher
	go p.dispatch()

	// Start auto-scaler if enabled
	if p.autoScaler != nil {
		go p.autoScaler.Start(p.ctx)
	}

	p.running = true
	return nil
}

// Stop stops the worker pool gracefully
func (p *WorkerPool) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = false
	p.mu.Unlock()

	// Cancel context to signal all workers to stop
	p.cancel()

	// Stop all workers
	p.mu.RLock()
	for _, worker := range p.workers {
		close(worker.quit)
	}
	p.mu.RUnlock()

	// Wait for all workers to finish
	p.wg.Wait()

	return nil
}

// Submit submits a job to the pool
func (p *WorkerPool) Submit(job *Job) error {
	if !p.running {
		return fmt.Errorf("worker pool not running")
	}

	// Check circuit breaker
	if !p.circuitBreaker.Allow() {
		return fmt.Errorf("circuit breaker open, too many failures")
	}

	return p.jobQueue.Push(job)
}

// SubmitAndWait submits a job and waits for completion
func (p *WorkerPool) SubmitAndWait(ctx context.Context, job *Job) error {
	done := make(chan struct{})
	var jobErr error

	originalOnComplete := job.OnComplete
	job.OnComplete = func(j *Job) {
		if originalOnComplete != nil {
			originalOnComplete(j)
		}
		close(done)
	}

	originalOnError := job.OnError
	job.OnError = func(j *Job, err error) {
		jobErr = err
		if originalOnError != nil {
			originalOnError(j, err)
		}
		close(done)
	}

	if err := p.Submit(job); err != nil {
		return err
	}

	select {
	case <-done:
		return jobErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetMetrics returns current pool metrics
func (p *WorkerPool) GetMetrics() PoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return PoolMetrics{
		TotalJobs:         atomic.LoadInt64(&p.metrics.TotalJobs),
		CompletedJobs:     atomic.LoadInt64(&p.metrics.CompletedJobs),
		FailedJobs:        atomic.LoadInt64(&p.metrics.FailedJobs),
		RetryJobs:         atomic.LoadInt64(&p.metrics.RetryJobs),
		ActiveWorkers:     atomic.LoadInt32(&p.currentWorkers),
		QueueSize:         int64(p.jobQueue.Len()),
		AvgProcessingTime: p.metrics.AvgProcessingTime,
		LastUpdated:       time.Now(),
	}
}

// GetQueueUtilization returns current queue utilization (0.0 - 1.0)
func (p *WorkerPool) GetQueueUtilization() float64 {
	return float64(p.jobQueue.Len()) / float64(p.jobQueue.Cap())
}

// GetWorkerCount returns current number of workers
func (p *WorkerPool) GetWorkerCount() int {
	return int(atomic.LoadInt32(&p.currentWorkers))
}

// ScaleUp adds more workers to the pool
func (p *WorkerPool) ScaleUp(count int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	added := 0
	for i := 0; i < count; i++ {
		if int(atomic.LoadInt32(&p.currentWorkers)) >= p.maxWorkers {
			break
		}
		worker := p.createWorker()
		p.workers = append(p.workers, worker)
		go worker.run()
		atomic.AddInt32(&p.currentWorkers, 1)
		added++
	}
	return added
}

// ScaleDown removes workers from the pool
func (p *WorkerPool) ScaleDown(count int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	removed := 0
	for i := len(p.workers) - 1; i >= 0 && removed < count; i-- {
		if int(atomic.LoadInt32(&p.currentWorkers)) <= p.minWorkers {
			break
		}
		worker := p.workers[i]
		if atomic.LoadInt32(&worker.busy) == 0 {
			close(worker.quit)
			p.workers = p.workers[:i]
			atomic.AddInt32(&p.currentWorkers, -1)
			removed++
		}
	}
	return removed
}

func (p *WorkerPool) createWorker() *Worker {
	id := make([]byte, 4)
	rand.Read(id)
	return &Worker{
		ID:      hex.EncodeToString(id),
		pool:    p,
		jobChan: make(chan *Job, 1),
		quit:    make(chan struct{}),
	}
}

func (p *WorkerPool) dispatch() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			job, err := p.jobQueue.Pop(p.ctx)
			if err != nil {
				continue
			}

			// Find available worker
			dispatched := false
			p.mu.RLock()
			for _, worker := range p.workers {
				if atomic.LoadInt32(&worker.busy) == 0 {
					select {
					case worker.jobChan <- job:
						dispatched = true
					default:
						continue
					}
					if dispatched {
						break
					}
				}
			}
			p.mu.RUnlock()

			// If no worker available, put job back in queue
			if !dispatched {
				p.jobQueue.Push(job)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (w *Worker) run() {
	w.pool.wg.Add(1)
	defer w.pool.wg.Done()

	for {
		select {
		case <-w.quit:
			return
		case job := <-w.jobChan:
			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job *Job) {
	atomic.StoreInt32(&w.busy, 1)
	defer atomic.StoreInt32(&w.busy, 0)

	atomic.AddInt64(&w.pool.metrics.TotalJobs, 1)
	job.Status = JobStatusRunning
	job.StartedAt = time.Now()

	// Get handler for job type
	w.pool.mu.RLock()
	handler, exists := w.pool.handlers[job.Type]
	w.pool.mu.RUnlock()

	if !exists {
		job.Status = JobStatusFailed
		job.Error = fmt.Errorf("no handler registered for job type: %s", job.Type)
		atomic.AddInt64(&w.pool.metrics.FailedJobs, 1)
		w.pool.circuitBreaker.RecordFailure()
		if job.OnError != nil {
			job.OnError(job, job.Error)
		}
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(w.pool.ctx, job.Timeout)
	defer cancel()

	// Execute job with retry logic
	var err error
	for attempt := 0; attempt <= job.MaxRetries; attempt++ {
		if attempt > 0 {
			job.Status = JobStatusRetrying
			job.RetryCount = attempt
			atomic.AddInt64(&w.pool.metrics.RetryJobs, 1)
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			time.Sleep(backoff)
		}

		err = handler(ctx, job)
		if err == nil {
			break
		}

		// Check if context was cancelled
		if ctx.Err() != nil {
			break
		}
	}

	job.CompletedAt = time.Now()
	processingTime := job.CompletedAt.Sub(job.StartedAt)

	// Update metrics
	w.pool.metrics.mu.Lock()
	w.pool.metrics.processingTimes = append(w.pool.metrics.processingTimes, processingTime)
	if len(w.pool.metrics.processingTimes) > 100 {
		w.pool.metrics.processingTimes = w.pool.metrics.processingTimes[1:]
	}
	var total time.Duration
	for _, t := range w.pool.metrics.processingTimes {
		total += t
	}
	w.pool.metrics.AvgProcessingTime = float64(total.Milliseconds()) / float64(len(w.pool.metrics.processingTimes))
	w.pool.metrics.mu.Unlock()

	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err
		atomic.AddInt64(&w.pool.metrics.FailedJobs, 1)
		w.pool.circuitBreaker.RecordFailure()
		if job.OnError != nil {
			job.OnError(job, err)
		}
	} else {
		job.Status = JobStatusCompleted
		atomic.AddInt64(&w.pool.metrics.CompletedJobs, 1)
		w.pool.circuitBreaker.RecordSuccess()
		if job.OnComplete != nil {
			job.OnComplete(job)
		}
	}

	atomic.AddInt64(&w.jobCount, 1)
}

// PriorityQueue is a thread-safe priority queue for jobs
type PriorityQueue struct {
	mu       sync.Mutex
	cond     *sync.Cond
	queues   map[JobPriority][]*Job
	size     int
	capacity int
	closed   bool
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(capacity int) *PriorityQueue {
	pq := &PriorityQueue{
		queues:   make(map[JobPriority][]*Job),
		capacity: capacity,
	}
	pq.cond = sync.NewCond(&pq.mu)
	return pq
}

// Push adds a job to the queue
func (pq *PriorityQueue) Push(job *Job) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.closed {
		return fmt.Errorf("queue is closed")
	}

	if pq.size >= pq.capacity {
		return fmt.Errorf("queue is full")
	}

	pq.queues[job.Priority] = append(pq.queues[job.Priority], job)
	pq.size++
	pq.cond.Signal()
	return nil
}

// Pop removes and returns the highest priority job
func (pq *PriorityQueue) Pop(ctx context.Context) (*Job, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for pq.size == 0 && !pq.closed {
		// Wait with context awareness
		done := make(chan struct{})
		go func() {
			pq.cond.Wait()
			close(done)
		}()

		pq.mu.Unlock()
		select {
		case <-ctx.Done():
			pq.mu.Lock()
			return nil, ctx.Err()
		case <-done:
			pq.mu.Lock()
		}
	}

	if pq.closed && pq.size == 0 {
		return nil, fmt.Errorf("queue is closed")
	}

	// Get highest priority job
	for priority := PriorityCritical; priority >= PriorityLow; priority-- {
		if len(pq.queues[priority]) > 0 {
			job := pq.queues[priority][0]
			pq.queues[priority] = pq.queues[priority][1:]
			pq.size--
			return job, nil
		}
	}

	return nil, fmt.Errorf("no jobs available")
}

// Len returns the number of jobs in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.size
}

// Cap returns the capacity of the queue
func (pq *PriorityQueue) Cap() int {
	return pq.capacity
}

// Close closes the queue
func (pq *PriorityQueue) Close() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.closed = true
	pq.cond.Broadcast()
}

// AutoScaler automatically scales workers based on load
type AutoScaler struct {
	pool   *WorkerPool
	config AutoScalerConfig
}

// AutoScalerConfig configuration for auto-scaler
type AutoScalerConfig struct {
	MinWorkers         int
	MaxWorkers         int
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
	ScaleInterval      time.Duration
}

// NewAutoScaler creates a new auto-scaler
func NewAutoScaler(pool *WorkerPool, config AutoScalerConfig) *AutoScaler {
	return &AutoScaler{
		pool:   pool,
		config: config,
	}
}

// Start starts the auto-scaler
func (as *AutoScaler) Start(ctx context.Context) {
	ticker := time.NewTicker(as.config.ScaleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			as.evaluate()
		}
	}
}

func (as *AutoScaler) evaluate() {
	utilization := as.pool.GetQueueUtilization()
	currentWorkers := as.pool.GetWorkerCount()

	if utilization > as.config.ScaleUpThreshold && currentWorkers < as.config.MaxWorkers {
		// Scale up
		toAdd := (as.config.MaxWorkers - currentWorkers) / 2
		if toAdd < 1 {
			toAdd = 1
		}
		as.pool.ScaleUp(toAdd)
	} else if utilization < as.config.ScaleDownThreshold && currentWorkers > as.config.MinWorkers {
		// Scale down
		toRemove := (currentWorkers - as.config.MinWorkers) / 2
		if toRemove < 1 {
			toRemove = 1
		}
		as.pool.ScaleDown(toRemove)
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	failures        int
	threshold       int
	timeout         time.Duration
	state           CircuitState
	lastFailureTime time.Time
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		state:     CircuitClosed,
	}
}

// Allow checks if a request should be allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return true
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failures = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
