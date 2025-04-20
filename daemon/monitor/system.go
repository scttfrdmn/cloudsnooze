// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"fmt"
	"time"
	
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
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
	lastMetrics        common.SystemMetrics
	checkIntervalMs    int
	
	// GPU monitoring
	gpuMonitoringEnabled bool
	gpuService           common.AcceleratorInterface
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor(cpuThreshold, memoryThreshold, networkThreshold, diskThreshold, gpuThreshold float64, 
	inputThreshold, napTimeMinutes, checkIntervalMs int, gpuMonitoringEnabled bool) *SystemMonitor {
	
	// Import from the accelerator package is now accessed via factory
	// to avoid circular dependencies
	var gpuService common.AcceleratorInterface
	
	// For now, we'll create the accelerator in another function to break the import cycle
	// Typically we would use a factory or dependency injection pattern
	
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
		gpuService:           gpuService, // Will be set later via SetGPUService
	}
}

// SetGPUService sets the GPU monitoring service
// This is used to break circular dependencies
func (m *SystemMonitor) SetGPUService(service common.AcceleratorInterface) {
	m.gpuService = service
}

// CollectMetrics gathers all system metrics and evaluates idle status
func (m *SystemMonitor) CollectMetrics() (common.SystemMetrics, error) {
	metrics := common.SystemMetrics{
		CollectionTime: time.Now().Unix(),
	}
	
	// Collect CPU metrics
	cpuUsage, err := m.cpuMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting CPU metrics: %v", err)
	}
	metrics.CPUUsage = cpuUsage
	
	// Collect memory metrics
	memoryUsage, err := m.memoryMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting memory metrics: %v", err)
	}
	metrics.MemoryUsage = memoryUsage
	
	// Collect network metrics
	networkUsage, err := m.networkMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting network metrics: %v", err)
	}
	metrics.NetworkRate = networkUsage
	
	// Collect disk metrics
	diskUsage, err := m.diskMonitor.GetUsage()
	if err != nil {
		return metrics, fmt.Errorf("error collecting disk metrics: %v", err)
	}
	metrics.DiskIORate = diskUsage
	
	// Collect input activity metrics
	inputIdleSecs, err := m.inputMonitor.GetIdleSeconds()
	if err != nil {
		// Just log and continue, don't fail the entire collection
		fmt.Printf("Warning: Failed to get input metrics: %v\n", err)
		inputIdleSecs = 0
	}
	metrics.LastInputTime = time.Now().Unix() - int64(inputIdleSecs)
	
	// Collect GPU metrics if enabled
	if m.gpuMonitoringEnabled && m.gpuService != nil {
		gpuMetrics, err := m.gpuService.GetMetrics()
		if err != nil {
			// Just log and continue
			fmt.Printf("Warning: Failed to get GPU metrics: %v\n", err)
		} else {
			metrics.GPUMetrics = gpuMetrics
		}
	}
	
	// Check CPU usage - if above threshold, system is not idle
	if cpuUsage >= m.cpuThreshold {
		m.idleSince = nil
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check memory usage
	if memoryUsage >= m.memoryThreshold {
		m.idleSince = nil
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check network usage
	if networkUsage >= m.networkThreshold {
		m.idleSince = nil
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check disk usage
	if diskUsage >= m.diskThreshold {
		m.idleSince = nil
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check input idle time if threshold is set
	if m.inputThreshold > 0 && inputIdleSecs < m.inputThreshold {
		m.idleSince = nil
		m.lastMetrics = metrics
		return metrics, nil
	}
	
	// Check GPU usage if enabled
	if m.gpuMonitoringEnabled && len(metrics.GPUMetrics) > 0 {
		for _, gpu := range metrics.GPUMetrics {
			if gpu.Utilization > m.gpuThreshold {
				m.idleSince = nil
				m.lastMetrics = metrics
				return metrics, nil
			}
		}
	}
	
	// At this point, the system is idle (all metrics below thresholds)
	// Update idle state tracking
	if m.idleSince == nil {
		now := time.Now()
		m.idleSince = &now
	}
	
	// Set idle time in metrics
	idleDuration := time.Since(*m.idleSince)
	metrics.IdleTime = idleDuration.Milliseconds() / 1000 // Convert to seconds
	
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
func (m *SystemMonitor) GetLastMetrics() common.SystemMetrics {
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