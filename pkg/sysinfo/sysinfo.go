// Package sysinfo provides system information detection similar to neofetch
// and automatic optimization recommendations based on system capabilities
package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"eroshit/pkg/i18n"
)

// SystemInfo contains comprehensive system information
type SystemInfo struct {
	// OS Information
	OS            string
	OSVersion     string
	Kernel        string
	Architecture  string
	Hostname      string
	Username      string
	
	// Hardware
	CPU           string
	CPUCores      int
	CPUThreads    int
	CPUFrequency  string
	GPU           string
	
	// Memory
	TotalMemory   uint64 // bytes
	FreeMemory    uint64 // bytes
	UsedMemory    uint64 // bytes
	MemoryPercent float64
	
	// Disk
	TotalDisk     uint64
	FreeDisk      uint64
	UsedDisk      uint64
	DiskPercent   float64
	
	// Network
	NetworkInterfaces []string
	PublicIP          string
	
	// Runtime
	GoVersion     string
	NumGoroutines int
	Uptime        time.Duration
	
	// Shell & Terminal
	Shell         string
	Terminal      string
	
	// Display
	Resolution    string
	
	// Timestamps
	DetectedAt    time.Time
}

// OptimizationProfile contains recommended settings based on system capabilities
type OptimizationProfile struct {
	MaxConcurrentVisits int
	HitsPerMinute       int
	BrowserPoolMin      int
	BrowserPoolMax      int
	WorkerQueueSize     int
	EnableAutoScaling   bool
	RecommendedMode     string // "low", "medium", "high", "ultra"
	Warnings            []string
	Recommendations     []string
}

// Detect gathers all system information
func Detect() *SystemInfo {
	info := &SystemInfo{
		DetectedAt:    time.Now(),
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		Architecture:  runtime.GOARCH,
		CPUThreads:    runtime.NumCPU(),
	}
	
	// OS Detection
	info.OS = runtime.GOOS
	info.detectOSDetails()
	
	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}
	
	// Username
	if user := os.Getenv("USER"); user != "" {
		info.Username = user
	} else if user := os.Getenv("USERNAME"); user != "" {
		info.Username = user
	}
	
	// Shell
	if shell := os.Getenv("SHELL"); shell != "" {
		info.Shell = shell
	} else if shell := os.Getenv("COMSPEC"); shell != "" {
		info.Shell = shell
	}
	
	// Terminal
	if term := os.Getenv("TERM"); term != "" {
		info.Terminal = term
	} else if term := os.Getenv("TERM_PROGRAM"); term != "" {
		info.Terminal = term
	}
	
	// Hardware detection
	info.detectCPU()
	info.detectGPU()
	info.detectMemory()
	info.detectDisk()
	info.detectNetwork()
	info.detectUptime()
	info.detectResolution()
	
	return info
}

func (s *SystemInfo) detectOSDetails() {
	switch runtime.GOOS {
	case "linux":
		s.detectLinuxDetails()
	case "darwin":
		s.detectMacOSDetails()
	case "windows":
		s.detectWindowsDetails()
	}
}

func (s *SystemInfo) detectLinuxDetails() {
	// Try to read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				s.OSVersion = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}
	
	// Kernel version
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		s.Kernel = strings.TrimSpace(string(out))
	}
}

func (s *SystemInfo) detectMacOSDetails() {
	if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
		s.OSVersion = "macOS " + strings.TrimSpace(string(out))
	}
	
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		s.Kernel = strings.TrimSpace(string(out))
	}
}

func (s *SystemInfo) detectWindowsDetails() {
	if out, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
		s.OSVersion = strings.TrimSpace(string(out))
	}
	
	// Try to get more detailed Windows version
	if out, err := exec.Command("wmic", "os", "get", "Caption", "/value").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Caption=") {
				s.OSVersion = strings.TrimSpace(strings.TrimPrefix(line, "Caption="))
				break
			}
		}
	}
}

func (s *SystemInfo) detectCPU() {
	switch runtime.GOOS {
	case "linux":
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						s.CPU = strings.TrimSpace(parts[1])
						break
					}
				}
			}
			// Count physical cores
			coreCount := 0
			for _, line := range lines {
				if strings.HasPrefix(line, "processor") {
					coreCount++
				}
			}
			s.CPUCores = coreCount
		}
		
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			s.CPU = strings.TrimSpace(string(out))
		}
		if out, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output(); err == nil {
			if cores, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
				s.CPUCores = cores
			}
		}
		
	case "windows":
		if out, err := exec.Command("wmic", "cpu", "get", "Name", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Name=") {
					s.CPU = strings.TrimSpace(strings.TrimPrefix(line, "Name="))
					break
				}
			}
		}
		if out, err := exec.Command("wmic", "cpu", "get", "NumberOfCores", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "NumberOfCores=") {
					if cores, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "NumberOfCores="))); err == nil {
						s.CPUCores = cores
					}
					break
				}
			}
		}
	}
	
	if s.CPUCores == 0 {
		s.CPUCores = runtime.NumCPU()
	}
}

func (s *SystemInfo) detectGPU() {
	switch runtime.GOOS {
	case "linux":
		// Try lspci
		if out, err := exec.Command("lspci").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "vga") || strings.Contains(strings.ToLower(line), "3d") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						s.GPU = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
		
	case "darwin":
		if out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Chipset Model:") {
					s.GPU = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Chipset Model:"))
					break
				}
			}
		}
		
	case "windows":
		if out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "Name", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Name=") {
					s.GPU = strings.TrimSpace(strings.TrimPrefix(line, "Name="))
					break
				}
			}
		}
	}
}

func (s *SystemInfo) detectMemory() {
	switch runtime.GOOS {
	case "linux":
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if kb, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							s.TotalMemory = kb * 1024
						}
					}
				} else if strings.HasPrefix(line, "MemAvailable:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if kb, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
							s.FreeMemory = kb * 1024
						}
					}
				}
			}
		}
		
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
			if bytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64); err == nil {
				s.TotalMemory = bytes
			}
		}
		// Get free memory from vm_stat
		if out, err := exec.Command("vm_stat").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			var freePages, inactivePages uint64
			for _, line := range lines {
				if strings.Contains(line, "Pages free:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
							freePages = pages
						}
					}
				} else if strings.Contains(line, "Pages inactive:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
							inactivePages = pages
						}
					}
				}
			}
			s.FreeMemory = (freePages + inactivePages) * 4096 // Page size is 4KB
		}
		
	case "windows":
		if out, err := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
					if kb, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "TotalVisibleMemorySize=")), 10, 64); err == nil {
						s.TotalMemory = kb * 1024
					}
					break
				}
			}
		}
		if out, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "FreePhysicalMemory=") {
					if kb, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "FreePhysicalMemory=")), 10, 64); err == nil {
						s.FreeMemory = kb * 1024
					}
					break
				}
			}
		}
	}
	
	if s.TotalMemory > 0 {
		s.UsedMemory = s.TotalMemory - s.FreeMemory
		s.MemoryPercent = float64(s.UsedMemory) / float64(s.TotalMemory) * 100
	}
}

func (s *SystemInfo) detectDisk() {
	// Get current working directory disk info
	wd, _ := os.Getwd()
	
	switch runtime.GOOS {
	case "linux", "darwin":
		if out, err := exec.Command("df", "-B1", wd).Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 4 {
					if total, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						s.TotalDisk = total
					}
					if used, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
						s.UsedDisk = used
					}
					if free, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
						s.FreeDisk = free
					}
				}
			}
		}
		
	case "windows":
		// Get drive letter from working directory
		if len(wd) >= 2 && wd[1] == ':' {
			drive := wd[:2]
			if out, err := exec.Command("wmic", "logicaldisk", "where", fmt.Sprintf("DeviceID='%s'", drive), "get", "Size,FreeSpace", "/value").Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "Size=") {
						if size, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "Size=")), 10, 64); err == nil {
							s.TotalDisk = size
						}
					} else if strings.HasPrefix(line, "FreeSpace=") {
						if free, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(line, "FreeSpace=")), 10, 64); err == nil {
							s.FreeDisk = free
						}
					}
				}
				s.UsedDisk = s.TotalDisk - s.FreeDisk
			}
		}
	}
	
	if s.TotalDisk > 0 {
		s.DiskPercent = float64(s.UsedDisk) / float64(s.TotalDisk) * 100
	}
}

func (s *SystemInfo) detectNetwork() {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("ip", "link", "show").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, ": ") && !strings.Contains(line, "lo:") {
					parts := strings.Split(line, ": ")
					if len(parts) >= 2 {
						s.NetworkInterfaces = append(s.NetworkInterfaces, parts[1])
					}
				}
			}
		}
		
	case "darwin":
		if out, err := exec.Command("networksetup", "-listallhardwareports").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Device:") {
					s.NetworkInterfaces = append(s.NetworkInterfaces, strings.TrimSpace(strings.TrimPrefix(line, "Device:")))
				}
			}
		}
		
	case "windows":
		if out, err := exec.Command("netsh", "interface", "show", "interface").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for idx, line := range lines {
				if idx > 2 && strings.TrimSpace(line) != "" { // Skip header lines
					fields := strings.Fields(line)
					if len(fields) >= 4 {
						s.NetworkInterfaces = append(s.NetworkInterfaces, fields[len(fields)-1])
					}
				}
			}
		}
	}
}

func (s *SystemInfo) detectUptime() {
	switch runtime.GOOS {
	case "linux":
		if data, err := os.ReadFile("/proc/uptime"); err == nil {
			parts := strings.Fields(string(data))
			if len(parts) >= 1 {
				if seconds, err := strconv.ParseFloat(parts[0], 64); err == nil {
					s.Uptime = time.Duration(seconds) * time.Second
				}
			}
		}
		
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "kern.boottime").Output(); err == nil {
			// Parse: { sec = 1234567890, usec = 123456 }
			str := string(out)
			if idx := strings.Index(str, "sec = "); idx != -1 {
				str = str[idx+6:]
				if idx := strings.Index(str, ","); idx != -1 {
					if bootTime, err := strconv.ParseInt(str[:idx], 10, 64); err == nil {
						s.Uptime = time.Since(time.Unix(bootTime, 0))
					}
				}
			}
		}
		
	case "windows":
		if out, err := exec.Command("wmic", "os", "get", "LastBootUpTime", "/value").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "LastBootUpTime=") {
					timeStr := strings.TrimSpace(strings.TrimPrefix(line, "LastBootUpTime="))
					if len(timeStr) >= 14 {
						// Format: 20240101120000.000000+180
						year, _ := strconv.Atoi(timeStr[0:4])
						month, _ := strconv.Atoi(timeStr[4:6])
						day, _ := strconv.Atoi(timeStr[6:8])
						hour, _ := strconv.Atoi(timeStr[8:10])
						min, _ := strconv.Atoi(timeStr[10:12])
						sec, _ := strconv.Atoi(timeStr[12:14])
						bootTime := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
						s.Uptime = time.Since(bootTime)
					}
					break
				}
			}
		}
	}
}

func (s *SystemInfo) detectResolution() {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("xrandr", "--current").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "*") {
					fields := strings.Fields(line)
					if len(fields) >= 1 {
						s.Resolution = fields[0]
						break
					}
				}
			}
		}
		
	case "darwin":
		if out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Resolution:") {
					s.Resolution = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Resolution:"))
					break
				}
			}
		}
		
	case "windows":
		if out, err := exec.Command("wmic", "path", "Win32_VideoController", "get", "CurrentHorizontalResolution,CurrentVerticalResolution", "/value").Output(); err == nil {
			var width, height string
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "CurrentHorizontalResolution=") {
					width = strings.TrimSpace(strings.TrimPrefix(line, "CurrentHorizontalResolution="))
				} else if strings.HasPrefix(line, "CurrentVerticalResolution=") {
					height = strings.TrimSpace(strings.TrimPrefix(line, "CurrentVerticalResolution="))
				}
			}
			if width != "" && height != "" {
				s.Resolution = width + "x" + height
			}
		}
	}
}

// GenerateOptimizationProfile creates optimization recommendations based on system capabilities
func (s *SystemInfo) GenerateOptimizationProfile() *OptimizationProfile {
	return s.GenerateOptimizationProfileWithLocale("tr")
}

// GenerateOptimizationProfileWithLocale creates optimization recommendations with i18n support
func (s *SystemInfo) GenerateOptimizationProfileWithLocale(locale string) *OptimizationProfile {
	profile := &OptimizationProfile{
		EnableAutoScaling: true,
		Recommendations:   []string{},
		Warnings:          []string{},
	}
	
	// Memory-based recommendations
	memoryGB := float64(s.TotalMemory) / (1024 * 1024 * 1024)
	
	if memoryGB < 4 {
		profile.RecommendedMode = "low"
		profile.MaxConcurrentVisits = 3
		profile.HitsPerMinute = 15
		profile.BrowserPoolMin = 1
		profile.BrowserPoolMax = 3
		profile.WorkerQueueSize = 1000
		profile.Warnings = append(profile.Warnings, i18n.T(locale, i18n.MsgOptWarnLowRAM, memoryGB))
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecResourceBlock))
	} else if memoryGB < 8 {
		profile.RecommendedMode = "medium"
		profile.MaxConcurrentVisits = 5
		profile.HitsPerMinute = 25
		profile.BrowserPoolMin = 2
		profile.BrowserPoolMax = 5
		profile.WorkerQueueSize = 5000
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecMediumSystem))
	} else if memoryGB < 16 {
		profile.RecommendedMode = "high"
		profile.MaxConcurrentVisits = 10
		profile.HitsPerMinute = 40
		profile.BrowserPoolMin = 3
		profile.BrowserPoolMax = 10
		profile.WorkerQueueSize = 10000
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecGoodSystem))
	} else {
		profile.RecommendedMode = "ultra"
		profile.MaxConcurrentVisits = 20
		profile.HitsPerMinute = 60
		profile.BrowserPoolMin = 5
		profile.BrowserPoolMax = 20
		profile.WorkerQueueSize = 50000
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecPowerfulSystem))
	}
	
	// CPU-based adjustments
	if s.CPUThreads < 4 {
		profile.MaxConcurrentVisits = min(profile.MaxConcurrentVisits, 3)
		profile.BrowserPoolMax = min(profile.BrowserPoolMax, 3)
		profile.Warnings = append(profile.Warnings, i18n.T(locale, i18n.MsgOptWarnLowCPU, s.CPUThreads))
	} else if s.CPUThreads >= 8 {
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecStrongCPU, s.CPUThreads))
	}
	
	// Memory usage warning
	if s.MemoryPercent > 80 {
		profile.Warnings = append(profile.Warnings, i18n.T(locale, i18n.MsgOptWarnHighMemory, s.MemoryPercent))
		profile.MaxConcurrentVisits = max(1, profile.MaxConcurrentVisits-2)
	}
	
	// Disk space warning
	freeDiskGB := float64(s.FreeDisk) / (1024 * 1024 * 1024)
	if freeDiskGB < 5 {
		profile.Warnings = append(profile.Warnings, i18n.T(locale, i18n.MsgOptWarnLowDisk, freeDiskGB))
	}
	
	// OS-specific recommendations
	switch s.OS {
	case "windows":
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecWindows))
	case "linux":
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecLinux))
	case "darwin":
		profile.Recommendations = append(profile.Recommendations, i18n.T(locale, i18n.MsgOptRecMacOS))
	}
	
	return profile
}

// FormatSize formats bytes to human readable format
func FormatSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration to human readable format (default Turkish)
func FormatDuration(d time.Duration) string {
	return FormatDurationWithLocale(d, "tr")
}

// FormatDurationWithLocale formats duration with i18n support
func FormatDurationWithLocale(d time.Duration, locale string) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return i18n.T(locale, i18n.MsgSysDays, days, hours, minutes)
	}
	if hours > 0 {
		return i18n.T(locale, i18n.MsgSysHours, hours, minutes)
	}
	return i18n.T(locale, i18n.MsgSysMinutes, minutes)
}

// PrintBanner prints a neofetch-style system info banner (default Turkish)
func (s *SystemInfo) PrintBanner() string {
	return s.PrintBannerWithLocale("tr")
}

// PrintBannerWithLocale prints a neofetch-style system info banner with i18n support
func (s *SystemInfo) PrintBannerWithLocale(locale string) string {
	var sb strings.Builder
	
	// ASCII Art Logo
	logo := `
    ███████╗██████╗  ██████╗ ███████╗██╗  ██╗██╗████████╗
    ██╔════╝██╔══██╗██╔═══██╗██╔════╝██║  ██║██║╚══██╔══╝
    █████╗  ██████╔╝██║   ██║███████╗███████║██║   ██║   
    ██╔══╝  ██╔══██╗██║   ██║╚════██║██╔══██║██║   ██║   
    ███████╗██║  ██║╚██████╔╝███████║██║  ██║██║   ██║   
    ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝   ╚═╝   
`
	sb.WriteString(logo)
	sb.WriteString("\n")
	
	// System Info
	sb.WriteString(fmt.Sprintf("    \033[1;36m%s\033[0m@\033[1;36m%s\033[0m\n", s.Username, s.Hostname))
	sb.WriteString("    " + strings.Repeat("-", 40) + "\n")
	
	if s.OSVersion != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysOS), s.OSVersion))
	} else {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s %s\n", i18n.T(locale, i18n.MsgSysOS), s.OS, s.Architecture))
	}
	
	if s.Kernel != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysKernel), s.Kernel))
	}
	
	if s.Uptime > 0 {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysUptime), FormatDurationWithLocale(s.Uptime, locale)))
	}
	
	if s.Shell != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysShell), s.Shell))
	}
	
	if s.Resolution != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysResolution), s.Resolution))
	}
	
	if s.Terminal != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysTerminal), s.Terminal))
	}
	
	if s.CPU != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s (%d cores)\n", i18n.T(locale, i18n.MsgSysCPU), s.CPU, s.CPUCores))
	} else {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysCPU), i18n.T(locale, i18n.MsgSysCPUCores, s.CPUCores, s.CPUThreads)))
	}
	
	if s.GPU != "" {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysGPU), s.GPU))
	}
	
	if s.TotalMemory > 0 {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s / %s (%.1f%%)\n", 
			i18n.T(locale, i18n.MsgSysMemory), FormatSize(s.UsedMemory), FormatSize(s.TotalMemory), s.MemoryPercent))
	}
	
	if s.TotalDisk > 0 {
		sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s / %s (%.1f%%)\n", 
			i18n.T(locale, i18n.MsgSysDisk), FormatSize(s.UsedDisk), FormatSize(s.TotalDisk), s.DiskPercent))
	}
	
	sb.WriteString(fmt.Sprintf("    \033[1;33m%s\033[0m %s\n", i18n.T(locale, i18n.MsgSysGoVersion), s.GoVersion))
	
	sb.WriteString("\n")
	
	return sb.String()
}

// PrintOptimizationProfile prints optimization recommendations (default Turkish)
func (p *OptimizationProfile) PrintProfile() string {
	return p.PrintProfileWithLocale("tr")
}

// PrintProfileWithLocale prints optimization recommendations with i18n support
func (p *OptimizationProfile) PrintProfileWithLocale(locale string) string {
	var sb strings.Builder
	
	sb.WriteString("\n    ╔══════════════════════════════════════════════════════════╗\n")
	sb.WriteString(fmt.Sprintf("    ║           %s               ║\n", i18n.T(locale, i18n.MsgOptProfileTitle)))
	sb.WriteString("    ╠══════════════════════════════════════════════════════════╣\n")
	
	modeColors := map[string]string{
		"low":    "\033[1;31m",  // Red
		"medium": "\033[1;33m",  // Yellow
		"high":   "\033[1;32m",  // Green
		"ultra":  "\033[1;35m",  // Magenta
	}
	
	modeNames := i18n.GetModeNames(locale)
	
	color := modeColors[p.RecommendedMode]
	modeName := modeNames[p.RecommendedMode]
	
	sb.WriteString(fmt.Sprintf("    ║  %s %s%s\033[0m%s║\n", 
		i18n.T(locale, i18n.MsgOptRecommendedMode), color, modeName, strings.Repeat(" ", 42-len(modeName)-len(i18n.T(locale, i18n.MsgOptRecommendedMode)))))
	sb.WriteString("    ╠══════════════════════════════════════════════════════════╣\n")
	
	maxConcLabel := i18n.T(locale, i18n.MsgOptMaxConcurrent)
	sb.WriteString(fmt.Sprintf("    ║  %s %-*d║\n", maxConcLabel, 56-len(maxConcLabel), p.MaxConcurrentVisits))
	
	hpmLabel := i18n.T(locale, i18n.MsgOptHitsPerMinute)
	sb.WriteString(fmt.Sprintf("    ║  %s %-*d║\n", hpmLabel, 56-len(hpmLabel), p.HitsPerMinute))
	
	poolLabel := i18n.T(locale, i18n.MsgOptBrowserPool)
	poolValue := fmt.Sprintf("%d - %d", p.BrowserPoolMin, p.BrowserPoolMax)
	sb.WriteString(fmt.Sprintf("    ║  %s %-*s║\n", poolLabel, 56-len(poolLabel), poolValue))
	
	queueLabel := i18n.T(locale, i18n.MsgOptWorkerQueue)
	sb.WriteString(fmt.Sprintf("    ║  %s %-*d║\n", queueLabel, 56-len(queueLabel), p.WorkerQueueSize))
	
	if len(p.Warnings) > 0 {
		sb.WriteString("    ╠══════════════════════════════════════════════════════════╣\n")
		for _, warning := range p.Warnings {
			// Truncate if too long
			if len(warning) > 56 {
				warning = warning[:53] + "..."
			}
			sb.WriteString(fmt.Sprintf("    ║  %s%s║\n", warning, strings.Repeat(" ", 58-len(warning))))
		}
	}
	
	if len(p.Recommendations) > 0 {
		sb.WriteString("    ╠══════════════════════════════════════════════════════════╣\n")
		for _, rec := range p.Recommendations {
			// Truncate if too long
			if len(rec) > 56 {
				rec = rec[:53] + "..."
			}
			sb.WriteString(fmt.Sprintf("    ║  %s%s║\n", rec, strings.Repeat(" ", 58-len(rec))))
		}
	}
	
	sb.WriteString("    ╚══════════════════════════════════════════════════════════╝\n")
	
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
