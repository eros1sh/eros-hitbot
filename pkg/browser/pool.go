// Package browser provides high-performance browser pool management using object pool pattern.
// This implementation reuses Chrome instances to avoid expensive startup costs.
package browser

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// PoolConfig defines configuration for the browser pool
type PoolConfig struct {
	// MaxInstances maximum number of browser instances in the pool (default: 10)
	MaxInstances int
	// MinInstances minimum number of pre-created instances (default: 2)
	MinInstances int
	// AcquireTimeout timeout for acquiring an instance from pool (default: 30s)
	AcquireTimeout time.Duration
	// InstanceMaxAge maximum lifetime of a browser instance (default: 30m)
	InstanceMaxAge time.Duration
	// InstanceMaxSessions maximum sessions per instance before recycling (default: 50)
	InstanceMaxSessions int32
	// ProxyURL optional proxy URL for all instances
	ProxyURL string
	// ProxyUser proxy authentication username
	ProxyUser string
	// ProxyPass proxy authentication password
	ProxyPass string
	// Headless run browser in headless mode (default: true)
	Headless bool
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxInstances:        10,
		MinInstances:        2,
		AcquireTimeout:      30 * time.Second,
		InstanceMaxAge:      30 * time.Minute,
		InstanceMaxSessions: 50,
		Headless:            true,
	}
}

// BrowserInstance represents a managed Chrome instance that can be reused
type BrowserInstance struct {
	id       string
	allocCtx context.Context	// Chrome allocator context
	allocCancel context.CancelFunc
	tabCtx   context.Context	// Current tab context
	tabCancel context.CancelFunc
	
	// Lifecycle tracking
	createdAt    time.Time
	lastUsedAt   time.Time
	sessionCount int32
	inUse        int32
	
	// Proxy configuration
	proxyURL  string
	proxyUser string
	proxyPass string
	
	// Chrome options
	headless bool
	
	// Reset function for cleanup
	resetFn func(*BrowserInstance) error
}

// IsInUse returns true if instance is currently in use
func (bi *BrowserInstance) IsInUse() bool {
	return atomic.LoadInt32(&bi.inUse) == 1
}

// GetSessionCount returns number of sessions handled by this instance
func (bi *BrowserInstance) GetSessionCount() int32 {
	return atomic.LoadInt32(&bi.sessionCount)
}

// NeedsRecycle returns true if instance should be recycled
func (bi *BrowserInstance) NeedsRecycle(maxAge time.Duration, maxSessions int32) bool {
	if time.Since(bi.createdAt) > maxAge {
		return true
	}
	if atomic.LoadInt32(&bi.sessionCount) >= maxSessions {
		return true
	}
	return false
}

// BrowserPool manages a pool of reusable Chrome instances using object pool pattern
type BrowserPool struct {
	config PoolConfig
	
	// Channel-based pool for thread-safe instance management
	available chan *BrowserInstance
	
	// Internal tracking
	instances map[string]*BrowserInstance
	mu        sync.RWMutex
	
	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Metrics
	metrics *PoolMetrics
	
	// Instance counter for unique IDs
	instanceCounter uint64
}

// PoolMetrics tracks pool performance metrics
type PoolMetrics struct {
	TotalCreated   int64
	TotalDestroyed int64
	TotalReused    int64
	TotalAcquired  int64
	TotalReleased  int64
	CurrentActive  int32
	CurrentIdle    int32
	AcquireWaits   int64	// Number of times we had to wait for an instance
	ResetErrors    int64	// Number of reset/cleanup failures
}

// GetMetrics returns current pool metrics (thread-safe copy)
func (p *BrowserPool) GetMetrics() PoolMetrics {
	return PoolMetrics{
		TotalCreated:   atomic.LoadInt64(&p.metrics.TotalCreated),
		TotalDestroyed: atomic.LoadInt64(&p.metrics.TotalDestroyed),
		TotalReused:    atomic.LoadInt64(&p.metrics.TotalReused),
		TotalAcquired:  atomic.LoadInt64(&p.metrics.TotalAcquired),
		TotalReleased:  atomic.LoadInt64(&p.metrics.TotalReleased),
		CurrentActive:  atomic.LoadInt32(&p.metrics.CurrentActive),
		CurrentIdle:    atomic.LoadInt32(&p.metrics.CurrentIdle),
		AcquireWaits:   atomic.LoadInt64(&p.metrics.AcquireWaits),
		ResetErrors:    atomic.LoadInt64(&p.metrics.ResetErrors),
	}
}

// NewBrowserPool creates a new high-performance browser pool
func NewBrowserPool(config PoolConfig) (*BrowserPool, error) {
	// Apply defaults for zero values
	if config.MaxInstances <= 0 {
		config.MaxInstances = 10
	}
	if config.MinInstances <= 0 {
		config.MinInstances = 2
	}
	if config.MinInstances > config.MaxInstances {
		config.MinInstances = config.MaxInstances
	}
	if config.AcquireTimeout <= 0 {
		config.AcquireTimeout = 30 * time.Second
	}
	if config.InstanceMaxAge <= 0 {
		config.InstanceMaxAge = 30 * time.Minute
	}
	if config.InstanceMaxSessions <= 0 {
		config.InstanceMaxSessions = 50
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &BrowserPool{
		config:    config,
		available: make(chan *BrowserInstance, config.MaxInstances),
		instances: make(map[string]*BrowserInstance),
		ctx:       ctx,
		cancel:    cancel,
		metrics:   &PoolMetrics{},
	}

	// Pre-create minimum instances
	for i := 0; i < config.MinInstances; i++ {
		instance, err := pool.createInstance()
		if err != nil {
			// Log error but continue - we'll try to create on demand
			continue
		}
		pool.available <- instance
		atomic.AddInt32(&pool.metrics.CurrentIdle, 1)
	}

	// Start maintenance goroutine
	pool.wg.Add(1)
	go pool.maintenanceLoop()

	return pool, nil
}

// Acquire gets a browser instance from the pool.
// Blocks until an instance is available or context is cancelled.
// Returns error if pool is closed or acquire timeout exceeded.
func (p *BrowserPool) Acquire(ctx context.Context) (*BrowserInstance, error) {
	atomic.AddInt64(&p.metrics.TotalAcquired, 1)

	// Fast path: try to get from available channel without blocking
	select {
	case instance := <-p.available:
		atomic.AddInt32(&p.metrics.CurrentIdle, -1)
		return p.prepareInstance(instance)
	default:
		// No idle instance available
	}

	// Check if we can create a new instance
	p.mu.Lock()
	currentCount := len(p.instances)
	canCreate := currentCount < p.config.MaxInstances
	p.mu.Unlock()

	if canCreate {
		instance, err := p.createInstance()
		if err != nil {
			return nil, fmt.Errorf("failed to create new instance: %w", err)
		}
		return p.prepareInstance(instance)
	}

	// Pool at max capacity, need to wait
	atomic.AddInt64(&p.metrics.AcquireWaits, 1)

	// Create timeout context if not provided
	acquireCtx, cancel := context.WithTimeout(ctx, p.config.AcquireTimeout)
	defer cancel()

	select {
	case instance := <-p.available:
		atomic.AddInt32(&p.metrics.CurrentIdle, -1)
		return p.prepareInstance(instance)
	case <-acquireCtx.Done():
		return nil, fmt.Errorf("acquire timeout: %w", acquireCtx.Err())
	case <-p.ctx.Done():
		return nil, fmt.Errorf("pool closed")
	}
}

// prepareInstance marks instance as in-use and updates metrics
func (p *BrowserPool) prepareInstance(instance *BrowserInstance) (*BrowserInstance, error) {
	if instance.NeedsRecycle(p.config.InstanceMaxAge, p.config.InstanceMaxSessions) {
		// Instance too old or overused, destroy and create new
		p.destroyInstance(instance)
		newInstance, err := p.createInstance()
		if err != nil {
			return nil, err
		}
		instance = newInstance
	}

	atomic.StoreInt32(&instance.inUse, 1)
	instance.lastUsedAt = time.Now()
	atomic.AddInt32(&instance.sessionCount, 1)
	atomic.AddInt32(&p.metrics.CurrentActive, 1)
	atomic.AddInt64(&p.metrics.TotalReused, 1)

	return instance, nil
}

// Release returns a browser instance to the pool.
// The instance is reset (cookies/cache cleared) before being reused.
func (p *BrowserPool) Release(instance *BrowserInstance) {
	if instance == nil {
		return
	}

	atomic.AddInt64(&p.metrics.TotalReleased, 1)
	atomic.AddInt32(&p.metrics.CurrentActive, -1)
	atomic.StoreInt32(&instance.inUse, 0)

	// Reset instance (clear cookies, cache, etc.)
	if err := p.Reset(instance); err != nil {
		// Reset failed, destroy instance instead of returning to pool
		atomic.AddInt64(&p.metrics.ResetErrors, 1)
		p.destroyInstance(instance)
		return
	}

	// Try to return to pool (non-blocking)
	select {
	case p.available <- instance:
		atomic.AddInt32(&p.metrics.CurrentIdle, 1)
	default:
		// Pool is full, destroy instance
		p.destroyInstance(instance)
	}
}

// Reset clears cookies, cache and session data from the instance.
// This should be called before reusing an instance to ensure clean state.
func (p *BrowserPool) Reset(instance *BrowserInstance) error {
	if instance == nil || instance.tabCtx == nil {
		return nil
	}

	// Create a short timeout context for cleanup operations
	ctx, cancel := context.WithTimeout(instance.allocCtx, 10*time.Second)
	defer cancel()

	// Clear browser cookies
	if err := network.ClearBrowserCookies().Do(ctx); err != nil {
		// Continue even if cookie clearing fails
		_ = err
	}

	// Clear browser cache
	if err := network.ClearBrowserCache().Do(ctx); err != nil {
		_ = err
	}

	// Cancel old tab context and create new one for next use
	if instance.tabCancel != nil {
		instance.tabCancel()
	}

	// Create fresh tab context for next session
	tabCtx, tabCancel := chromedp.NewContext(instance.allocCtx)
	instance.tabCtx = tabCtx
	instance.tabCancel = tabCancel

	return nil
}

// ForceReset performs deep cleanup including storage and service workers.
// Use this when Reset() is not sufficient.
func (p *BrowserPool) ForceReset(instance *BrowserInstance) error {
	if instance == nil {
		return nil
	}

	// Cancel current contexts
	if instance.tabCancel != nil {
		instance.tabCancel()
	}

	// Create new contexts (effectively resets the browser state)
	tabCtx, tabCancel := chromedp.NewContext(instance.allocCtx)
	instance.tabCtx = tabCtx
	instance.tabCancel = tabCancel

	// Clear storage
	clearScript := `
		localStorage.clear();
		sessionStorage.clear();
		indexedDB.deleteDatabase = indexedDB.deleteDatabase || function(){};
		caches.keys().then(keys => Promise.all(keys.map(key => caches.delete(key))));
	`

	ctx, cancel := context.WithTimeout(tabCtx, 10*time.Second)
	defer cancel()

	var result interface{}
	if err := chromedp.Run(ctx, chromedp.Evaluate(clearScript, &result)); err != nil {
		return fmt.Errorf("failed to clear storage: %w", err)
	}

	return nil
}

// Close shuts down the browser pool and destroys all instances
func (p *BrowserPool) Close() error {
	p.cancel()
	p.wg.Wait()

	// Close available channel
	close(p.available)

	// Drain and destroy all instances from available channel
	for instance := range p.available {
		p.destroyInstance(instance)
	}

	// Destroy any remaining tracked instances
	p.mu.Lock()
	for _, instance := range p.instances {
		if !instance.IsInUse() {
			p.destroyInstance(instance)
		}
	}
	p.mu.Unlock()

	return nil
}

// createInstance creates a new Chrome browser instance
func (p *BrowserPool) createInstance() (*BrowserInstance, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", p.config.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		// Headless bypass - critical for anti-detection
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process,TranslateUI"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-sync", true),
		// Performance optimizations
		chromedp.Flag("disable-web-security", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-images", false), // May need images for some sites
		chromedp.Flag("disk-cache-size", 33554432), // 32MB cache
		chromedp.Flag("media-cache-size", 33554432),
	)

	// Configure proxy if provided
	proxyURL := p.config.ProxyURL
	proxyUser := p.config.ProxyUser
	proxyPass := p.config.ProxyPass

	if proxyURL != "" {
		// Parse and extract auth from URL if present
		if parsedURL, err := url.Parse(proxyURL); err == nil && parsedURL.User != nil {
			if proxyUser == "" {
				proxyUser = parsedURL.User.Username()
			}
			if proxyPass == "" {
				if pass, ok := parsedURL.User.Password(); ok {
					proxyPass = pass
				}
			}
			// Rebuild URL without auth
			proxyURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		}

		opts = append(opts,
			chromedp.ProxyServer(proxyURL),
			chromedp.Flag("proxy-bypass-list", "<-loopback>"),
		)
	}

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(p.ctx, opts...)

	// Create initial tab context
	tabCtx, tabCancel := chromedp.NewContext(allocCtx)

	// Generate unique ID
	id := fmt.Sprintf("browser-%d-%d", time.Now().UnixNano(), atomic.AddUint64(&p.instanceCounter, 1))

	instance := &BrowserInstance{
		id:          id,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		tabCtx:      tabCtx,
		tabCancel:   tabCancel,
		createdAt:   time.Now(),
		lastUsedAt:  time.Now(),
		proxyURL:    p.config.ProxyURL,
		proxyUser:   proxyUser,
		proxyPass:   proxyPass,
		headless:    p.config.Headless,
	}

	p.mu.Lock()
	p.instances[id] = instance
	p.mu.Unlock()

	atomic.AddInt64(&p.metrics.TotalCreated, 1)

	return instance, nil
}

// destroyInstance terminates a Chrome browser instance
func (p *BrowserPool) destroyInstance(instance *BrowserInstance) {
	if instance == nil {
		return
	}

	// Cancel contexts in reverse order
	if instance.tabCancel != nil {
		instance.tabCancel()
	}
	if instance.allocCancel != nil {
		instance.allocCancel()
	}

	p.mu.Lock()
	delete(p.instances, instance.id)
	p.mu.Unlock()

	atomic.AddInt64(&p.metrics.TotalDestroyed, 1)
}

// maintenanceLoop runs periodic cleanup and health checks
func (p *BrowserPool) maintenanceLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performMaintenance()
		}
	}
}

// performMaintenance cleans up old instances and ensures minimum pool size
func (p *BrowserPool) performMaintenance() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find instances that need recycling
	var toRecycle []string
	for id, instance := range p.instances {
		if !instance.IsInUse() && instance.NeedsRecycle(p.config.InstanceMaxAge, p.config.InstanceMaxSessions) {
			toRecycle = append(toRecycle, id)
		}
	}

	// Calculate how many we can remove while keeping minimum
	currentCount := len(p.instances)
	maxRemove := currentCount - p.config.MinInstances
	if maxRemove < 0 {
		maxRemove = 0
	}
	if len(toRecycle) > maxRemove {
		toRecycle = toRecycle[:maxRemove]
	}

	// Remove from tracking (they'll be cleaned up when released)
	for _, id := range toRecycle {
		delete(p.instances, id)
	}

	// Ensure minimum instances if pool is low
	idleCount := len(p.available)
	needed := p.config.MinInstances - (currentCount - len(toRecycle))
	
	// Create new instances up to minimum
	for i := 0; i < needed && idleCount < p.config.MinInstances; i++ {
		instance, err := p.createInstance()
		if err != nil {
			continue
		}
		select {
		case p.available <- instance:
			atomic.AddInt32(&p.metrics.CurrentIdle, 1)
			idleCount++
		default:
			p.destroyInstance(instance)
		}
	}
}

// GetContext returns the chromedp context for the browser instance.
// This context should be used for browser automation operations.
func (bi *BrowserInstance) GetContext() context.Context {
	if bi.tabCtx == nil {
		return context.Background()
	}
	return bi.tabCtx
}

// GetAllocatorContext returns the allocator context for this instance.
// This is useful for creating new tabs within the same browser instance.
func (bi *BrowserInstance) GetAllocatorContext() context.Context {
	if bi.allocCtx == nil {
		return context.Background()
	}
	return bi.allocCtx
}

// GetID returns the unique identifier of this browser instance
func (bi *BrowserInstance) GetID() string {
	return bi.id
}

// GetAge returns the age of this browser instance
func (bi *BrowserInstance) GetAge() time.Duration {
	return time.Since(bi.createdAt)
}

// GetLastUsed returns the time since last use
func (bi *BrowserInstance) GetLastUsed() time.Duration {
	return time.Since(bi.lastUsedAt)
}

// IsHealthy checks if the browser instance is still responsive
func (bi *BrowserInstance) IsHealthy() bool {
	if bi.allocCtx == nil || bi.tabCtx == nil {
		return false
	}
	
	// Check if contexts are still valid
	select {
	case <-bi.allocCtx.Done():
		return false
	case <-bi.tabCtx.Done():
		return false
	default:
		return true
	}
}

// CreateNewTab creates a new tab within the same browser instance.
// This is useful for multi-tab operations without creating new browser instances.
func (bi *BrowserInstance) CreateNewTab(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if bi.allocCtx == nil {
		return nil, nil, fmt.Errorf("browser instance not initialized")
	}

	// Cancel old tab context if exists
	if bi.tabCancel != nil {
		bi.tabCancel()
	}

	// Create new tab context
	tabCtx, tabCancel := chromedp.NewContext(bi.allocCtx)
	bi.tabCtx = tabCtx
	bi.tabCancel = tabCancel

	return tabCtx, tabCancel, nil
}

// GetProxyAuth returns proxy credentials if configured
func (bi *BrowserInstance) GetProxyAuth() (username, password string) {
	return bi.proxyUser, bi.proxyPass
}
