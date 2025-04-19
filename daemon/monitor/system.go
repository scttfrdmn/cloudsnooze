package monitor

import (
	"fmt"
	"time"
	
	"github.com/scttfrdmn/cloudsnooze/daemon/accelerator"
)

// SystemMonitor coordinates all monitoring activities
type SystemMonitor struct {
	cpuMonitor     *CPUMonitor
	memoryMonitor  *MemoryMonitor
	networkMonitor *NetworkMonitor
	diskMonitor    *DiskMonitor
	inputMonitor   *InputMonitor
	
	// Thresholds from configuration
	cpuThreshold    float64
	memoryThreshold float64
	networkThreshold float64
	diskThreshold   float64
	inputThreshold  int
	gpuThreshold    float64
	
	// Tracking data
	idleSince          *time.Time
	napTimeMinutes     int
	lastMetrics        SystemMetrics
	checkIntervalMs    int
	
	// GPU monitoring
	gpuMonitoringEnabled bool
	gpuService           *accelerator.GPUService
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor(cpuThreshold, memoryThreshold, networkThreshold, diskThreshold, gpuThreshold float64, 
	inputThreshold, napTimeMinutes, checkIntervalMs int, gpuMonitoringEnabled bool) *SystemMonitor {
	
	return &SystemMonitor{
		cpuMonitor:      NewCPUMonitor(),
		memoryMonitor:   NewMemoryMonitor(),
		networkMonitor:  NewNetworkMonitor(checkIntervalMs),
		diskMonitor:     NewDiskMonitor(checkIntervalMs),
		inputMonitor:    NewInputMonitor(),
		
		cpuThreshold:    cpuThreshold,
		memoryThreshold: memoryThreshold,
		networkThreshold: networkThreshold,
		diskThreshold:   diskThreshold,
		inputThreshold:  inputThreshold,
		gpuThreshold:    gpuThreshold,
		
		napTimeMinutes:   napTimeMinutes,
		checkIntervalMs:  checkIntervalMs,
		
		gpuMonitoringEnabled: gpuMonitoringEnabled,
		gpuService:           accelerator.NewGPUService(),
	}
}

// CollectMetrics gathers all system metrics and evaluates idle status
func (m *SystemMonitor) CollectMetrics() (SystemMetrics, error) {
	metrics := SystemMetrics{
		Timestamp: time.Now(),
	}
	
	// Collect CPU metrics
	cpuUsage, err := m.cpuMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting CPU metrics: %v", err)
	}
	metrics.CPUPercent = cpuUsage
	
	// Collect memory metrics
	memoryUsage, err := m.memoryMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting memory metrics: %v", err)
	}
	metrics.MemoryPercent = memoryUsage
	
	// Collect network metrics
	networkUsage, err := m.networkMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting network metrics: %v", err)
	}
	metrics.NetworkKBps = networkUsage
	
	// Collect disk metrics
	diskUsage, err := m.diskMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting disk metrics: %v", err)
	}
	metrics.DiskIOKBps = diskUsage
	
	// Collect input activity metrics
	inputIdleSecs, err := m.inputMonitor.GetIdleSeconds()
	if err != nil {
		// Just log and continue, don't fail the entire collection
		fmt.Printf("Warning: Failed to get input metrics: %v\n", err)
		inputIdleSecs = 0
	}
	metrics.InputIdleSecs = inputIdleSecs
	
	// Collect GPU metrics if enabled
	if m.gpuMonitoringEnabled {
		gpuMetrics, err := m.gpuService.GetAllMetrics()
		if err != nil {
			// Just log and continue
			fmt.Printf("Warning: Failed to get GPU metrics: %v\n", err)
		} else {
			metrics.GPUMetrics = gpuMetrics
		}
	}
	
	// Determine idle status based on all metrics
	metrics.IdleStatus = false
	var reasons []string
	
	if cpuUsage < m.cpuThreshold {
		reasons = append(reasons, fmt.Sprintf("CPU usage %.1f%% below threshold %.1f%%", cpuUsage, m.cpuThreshold))
	} else {
		metrics.IdleStatus = false
		m.idleSince = nil
		metrics.IdleReason = fmt.Sprintf("CPU usage %.1f%% above threshold %.1f%%", cpuUsage, m.cpuThreshold)
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	if memoryUsage < m.memoryThreshold {
		reasons = append(reasons, fmt.Sprintf("Memory usage %.1f%% below threshold %.1f%%", memoryUsage, m.memoryThreshold))
	} else {
		metrics.IdleStatus = false
		m.idleSince = nil
		metrics.IdleReason = fmt.Sprintf("Memory usage %.1f%% above threshold %.1f%%", memoryUsage, m.memoryThreshold)
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	if networkUsage < m.networkThreshold {
		reasons = append(reasons, fmt.Sprintf("Network usage %.1f KB/s below threshold %.1f KB/s", networkUsage, m.networkThreshold))
	} else {
		metrics.IdleStatus = false
		m.idleSince = nil
		metrics.IdleReason = fmt.Sprintf("Network usage %.1f KB/s above threshold %.1f KB/s", networkUsage, m.networkThreshold)
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	if diskUsage < m.diskThreshold {
		reasons = append(reasons, fmt.Sprintf("Disk I/O %.1f KB/s below threshold %.1f KB/s", diskUsage, m.diskThreshold))
	} else {
		metrics.IdleStatus = false
		m.idleSince = nil
		metrics.IdleReason = fmt.Sprintf("Disk I/O %.1f KB/s above threshold %.1f KB/s", diskUsage, m.diskThreshold)
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check input idle time if threshold is set
	if m.inputThreshold > 0 && inputIdleSecs < m.inputThreshold {
		metrics.IdleStatus = false
		m.idleSince = nil
		metrics.IdleReason = fmt.Sprintf("Input idle time %d secs below threshold %d secs", inputIdleSecs, m.inputThreshold)
		m.lastMetrics = metrics
		return metrics, nil
	} else if m.inputThreshold > 0 {
		reasons = append(reasons, fmt.Sprintf("Input idle time %d secs above threshold %d secs", inputIdleSecs, m.inputThreshold))
	}
	
	// Check GPU usage if enabled
	if m.gpuMonitoringEnabled && len(metrics.GPUMetrics) > 0 {
		for _, gpu := range metrics.GPUMetrics {
			if gpu.Utilization > m.gpuThreshold {
				metrics.IdleStatus = false
				m.idleSince = nil
				metrics.IdleReason = fmt.Sprintf("GPU %d (%s) usage %.1f%% above threshold %.1f%%", 
					gpu.ID, gpu.Name, gpu.Utilization, m.gpuThreshold)
				m.lastMetrics = metrics
				return metrics, nil
			}
		}
		
		// All GPUs are below threshold
		reasons = append(reasons, fmt.Sprintf("All GPUs below usage threshold %.1f%%", m.gpuThreshold))
	}
	
	// If we got here, all metrics are below thresholds
	metrics.IdleStatus = true
	
	// Handle idle state tracking
	if m.idleSince == nil {
		now := time.Now()
		m.idleSince = &now
		metrics.IdleReason = "System just became idle"
	} else {
		idleDuration := time.Since(*m.idleSince)
		metrics.IdleReason = fmt.Sprintf("System idle for %s", idleDuration.Round(time.Second))
	}
	
	m.lastMetrics = metrics
	return metrics, nil
}

// ShouldSnooze determines if the instance should be snoozed based on idle time
func (m *SystemMonitor) ShouldSnooze() (bool, string) {
	if m.idleSince == nil {
		return false, "System is not idle"
	}
	
	idleDuration := time.Since(*m.idleSince)
	idleMinutes := int(idleDuration.Minutes())
	
	if idleMinutes >= m.napTimeMinutes {
		return true, fmt.Sprintf("System idle for %d minutes (threshold: %d minutes)", 
			idleMinutes, m.napTimeMinutes)
	}
	
	return false, fmt.Sprintf("System idle for %d minutes, waiting for %d minutes",
		idleMinutes, m.napTimeMinutes)
}

// GetLastMetrics returns the most recently collected metrics
func (m *SystemMonitor) GetLastMetrics() SystemMetrics {
	return m.lastMetrics
}

// GetIdleSince returns the time when the system became idle
func (m *SystemMonitor) GetIdleSince() *time.Time {
	return m.idleSince
}

// ResetIdleState resets the idle state tracking
func (m *SystemMonitor) ResetIdleState() {
	m.idleSince = nil
}