package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryMonitor handles memory usage monitoring
type MemoryMonitor struct {
	lastCheckTime time.Time
	lastUsage     float64
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		lastCheckTime: time.Now(),
	}
}

// GetUsage returns the current memory usage percentage
func (m *MemoryMonitor) GetUsage() (float64, error) {
	// Get memory statistics
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	// Update last check data
	m.lastCheckTime = time.Now()
	m.lastUsage = memStats.UsedPercent

	return memStats.UsedPercent, nil
}