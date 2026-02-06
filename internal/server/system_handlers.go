package server

import (
	"encoding/json"
	"net/http"
	"runtime"

	"eroshit/pkg/network"
	"eroshit/pkg/system"
	"eroshit/pkg/stealth"
)

// handleSystemInfo returns system information
type SystemInfoResponse struct {
	CPU struct {
		Count    int  `json:"count"`
		Affinity bool `json:"affinity_enabled"`
	} `json:"cpu"`
	NUMA struct {
		Nodes   int  `json:"nodes"`
		Enabled bool `json:"enabled"`
	} `json:"numa"`
	Memory struct {
		Total     uint64 `json:"total"`
		Available uint64 `json:"available"`
	} `json:"memory"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	optimizer := system.GetGlobalOptimizer()
	
	info := SystemInfoResponse{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	
	info.CPU.Count = optimizer.GetCPUAffinity().GetCPUCount()
	info.CPU.Affinity = optimizer.GetCPUAffinity().IsEnabled()
	info.NUMA.Nodes = optimizer.GetNUMAManager().GetNodeCount()
	info.NUMA.Enabled = optimizer.GetNUMAManager().IsEnabled()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleSystemOptimize applies system optimizations
func (s *Server) handleSystemOptimize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EnableCPUAffinity bool `json:"enable_cpu_affinity"`
		EnableNUMA        bool `json:"enable_numa"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	optimizer := system.GetGlobalOptimizer()
	
	if req.EnableCPUAffinity {
		optimizer.GetCPUAffinity().Enable()
	} else {
		optimizer.GetCPUAffinity().Disable()
	}
	
	if req.EnableNUMA {
		optimizer.GetNUMAManager().Enable()
	} else {
		optimizer.GetNUMAManager().Disable()
	}

	if err := optimizer.Optimize(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "optimized"})
}

// handleNetworkConfig returns/updates network configuration
type NetworkConfigResponse struct {
	ConnectionPool struct {
		Enabled         bool `json:"enabled"`
		MaxIdleConns    int  `json:"max_idle_conns"`
		MaxConnsPerHost int  `json:"max_conns_per_host"`
	} `json:"connection_pool"`
	HTTP3 struct {
		Enabled bool `json:"enabled"`
	} `json:"http3"`
	TCPFastOpen struct {
		Enabled bool `json:"enabled"`
	} `json:"tcp_fast_open"`
}

func (s *Server) handleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pool := network.GetGlobalPool()
		_ = pool.GetMetrics()
		
		config := NetworkConfigResponse{}
		config.ConnectionPool.Enabled = true
		config.ConnectionPool.MaxIdleConns = 100
		config.ConnectionPool.MaxConnsPerHost = 20
		config.HTTP3.Enabled = false
		config.TCPFastOpen.Enabled = false

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		var req NetworkConfigResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Reset global pool with new config
		poolConfig := network.DefaultPoolConfig()
		poolConfig.MaxIdleConns = req.ConnectionPool.MaxIdleConns
		poolConfig.MaxConnsPerHost = req.ConnectionPool.MaxConnsPerHost
		poolConfig.EnableHTTP3 = req.HTTP3.Enabled
		
		network.ResetGlobalPool(poolConfig)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleVMStatus returns VM spoofing status
func (s *Server) handleVMStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	spoofer := stealth.NewVMFingerprintSpoofer(stealth.DefaultVMConfig())
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":             spoofer.IsEnabled(),
		"vm_type":             spoofer.Config.VMType,
		"hide_indicators":     spoofer.Config.HideVMIndicators,
		"spoof_hardware_ids":  spoofer.Config.SpoofHardwareIDs,
		"detection_score":     spoofer.GetVMDetectionScore(),
		"is_running_in_vm":    stealth.IsRunningInVM(),
	})
}

// handleVMScore returns VM detection score
func (s *Server) handleVMScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	spoofer := stealth.NewVMFingerprintSpoofer(stealth.DefaultVMConfig())
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"detection_score": spoofer.GetVMDetectionScore(),
	})
}
