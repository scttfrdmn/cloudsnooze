// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUMonitor handles CPU usage monitoring
type CPUMonitor struct {
	lastCheckTime time.Time
	lastUsage     float64
}

// NewCPUMonitor creates a new CPU monitor
func NewCPUMonitor() *CPUMonitor {
	return &CPUMonitor{
		lastCheckTime: time.Now(),
	}
}

// GetUsage returns the current CPU usage percentage
func (m *CPUMonitor) GetUsage() (float64, error) {
	// Get CPU usage over a short interval (100ms)
	percentages, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		return 0, err
	}

	// Get the average across all CPUs
	var total float64
	for _, p := range percentages {
		total += p
	}
	avgUsage := total / float64(len(percentages))

	// Update last check data
	m.lastCheckTime = time.Now()
	m.lastUsage = avgUsage

	return avgUsage, nil
}