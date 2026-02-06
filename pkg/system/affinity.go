// CPU Affinity and NUMA Awareness for ErosHit
package system

import (
	"fmt"
	"runtime"
	"sync"
)

// CPUAffinityConfig CPU affinity configuration
type CPUAffinityConfig struct {
	Enabled         bool
	CPUs            []int
	NumaNodes       []int
	EnableAutoPin   bool
	SMTAware        bool
}

// DefaultCPUAffinityConfig returns default config
func DefaultCPUAffinityConfig() CPUAffinityConfig {
	return CPUAffinityConfig{
		Enabled:       false,
		CPUs:          []int{},
		NumaNodes:     []int{},
		EnableAutoPin: false,
		SMTAware:      true,
	}
}

// CPUAffinityManager manages CPU affinity
type CPUAffinityManager struct {
	config           CPUAffinityConfig
	cpuCount         int
	mu               sync.RWMutex
	pinnedGoRoutines map[int][]int
}

// NewCPUAffinityManager creates affinity manager
func NewCPUAffinityManager(config CPUAffinityConfig) *CPUAffinityManager {
	return &CPUAffinityManager{
		config:           config,
		cpuCount:         runtime.NumCPU(),
		pinnedGoRoutines: make(map[int][]int),
	}
}

// SetProcessAffinity sets CPU affinity
func (c *CPUAffinityManager) SetProcessAffinity(cpus []int) error {
	if !c.config.Enabled {
		return nil
	}
	
	runtime.GOMAXPROCS(len(cpus))
	
	c.mu.Lock()
	c.config.CPUs = cpus
	c.mu.Unlock()
	
	return nil
}

// GetRecommendedCPUs returns recommended CPU set
func (c *CPUAffinityManager) GetRecommendedCPUs() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if len(c.config.CPUs) > 0 {
		return c.config.CPUs
	}
	
	cpus := make([]int, c.cpuCount)
	for i := 0; i < c.cpuCount; i++ {
		cpus[i] = i
	}
	
	return cpus
}

// Enable enables CPU affinity
func (c *CPUAffinityManager) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Enabled = true
}

// Disable disables CPU affinity
func (c *CPUAffinityManager) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Enabled = false
}

// IsEnabled returns status
func (c *CPUAffinityManager) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Enabled
}

// GetCPUCount returns number of CPUs
func (c *CPUAffinityManager) GetCPUCount() int {
	return c.cpuCount
}

// NUMAConfig NUMA-aware configuration
type NUMAConfig struct {
	Enabled       bool
	PreferredNode int
	MemoryPolicy  string
}

// DefaultNUMAConfig returns default NUMA config
func DefaultNUMAConfig() NUMAConfig {
	return NUMAConfig{
		Enabled:       false,
		PreferredNode: 0,
		MemoryPolicy:  "preferred",
	}
}

// NUMAManager NUMA awareness manager
type NUMAManager struct {
	config    NUMAConfig
	nodeCount int
	mu        sync.RWMutex
}

// NewNUMAManager creates NUMA manager
func NewNUMAManager(config NUMAConfig) *NUMAManager {
	return &NUMAManager{
		config:    config,
		nodeCount: 1,
	}
}

// SetMemoryPolicy sets NUMA memory policy
func (n *NUMAManager) SetMemoryPolicy(node int, policy string) error {
	if !n.config.Enabled {
		return nil
	}
	
	n.mu.Lock()
	defer n.mu.Unlock()
	
	n.config.PreferredNode = node
	n.config.MemoryPolicy = policy
	
	return nil
}

// GetPreferredNode returns preferred NUMA node
func (n *NUMAManager) GetPreferredNode() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.config.PreferredNode
}

// GetNodeCount returns number of NUMA nodes
func (n *NUMAManager) GetNodeCount() int {
	return n.nodeCount
}

// Enable enables NUMA awareness
func (n *NUMAManager) Enable() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.config.Enabled = true
}

// Disable disables NUMA awareness
func (n *NUMAManager) Disable() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.config.Enabled = false
}

// IsEnabled returns status
func (n *NUMAManager) IsEnabled() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.config.Enabled
}

// SystemOptimizer combines CPU and NUMA optimization
type SystemOptimizer struct {
	cpuAffinity *CPUAffinityManager
	numa        *NUMAManager
	mu          sync.RWMutex
}

// NewSystemOptimizer creates system optimizer
func NewSystemOptimizer(cpuConfig CPUAffinityConfig, numaConfig NUMAConfig) *SystemOptimizer {
	return &SystemOptimizer{
		cpuAffinity: NewCPUAffinityManager(cpuConfig),
		numa:        NewNUMAManager(numaConfig),
	}
}

// Optimize applies optimizations
func (s *SystemOptimizer) Optimize() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.cpuAffinity.IsEnabled() {
		cpus := s.cpuAffinity.GetRecommendedCPUs()
		if err := s.cpuAffinity.SetProcessAffinity(cpus); err != nil {
			return fmt.Errorf("failed to set CPU affinity: %w", err)
		}
	}
	
	if s.numa.IsEnabled() {
		if err := s.numa.SetMemoryPolicy(s.numa.GetPreferredNode(), s.numa.config.MemoryPolicy); err != nil {
			return fmt.Errorf("failed to set NUMA policy: %w", err)
		}
	}
	
	return nil
}

// GetCPUAffinity returns CPU affinity manager
func (s *SystemOptimizer) GetCPUAffinity() *CPUAffinityManager {
	return s.cpuAffinity
}

// GetNUMAManager returns NUMA manager
func (s *SystemOptimizer) GetNUMAManager() *NUMAManager {
	return s.numa
}

// Global optimizer
var globalOptimizer *SystemOptimizer
var optimizerOnce sync.Once

// GetGlobalOptimizer returns singleton
func GetGlobalOptimizer() *SystemOptimizer {
	optimizerOnce.Do(func() {
		globalOptimizer = NewSystemOptimizer(
			DefaultCPUAffinityConfig(),
			DefaultNUMAConfig(),
		)
	})
	return globalOptimizer
}
