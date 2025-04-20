// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskMonitor handles disk I/O monitoring
type DiskMonitor struct {
	lastCheckTime    time.Time
	lastReadBytes    uint64
	lastWriteBytes   uint64
	lastUsageKBps    float64
	checkIntervalMs  int
}

// NewDiskMonitor creates a new disk I/O monitor
func NewDiskMonitor(checkIntervalMs int) *DiskMonitor {
	// Get initial stats
	ioStats, _ := disk.IOCounters()
	
	var initialReadBytes, initialWriteBytes uint64
	for _, stat := range ioStats {
		initialReadBytes += stat.ReadBytes
		initialWriteBytes += stat.WriteBytes
	}

	return &DiskMonitor{
		lastCheckTime:    time.Now(),
		lastReadBytes:    initialReadBytes,
		lastWriteBytes:   initialWriteBytes,
		checkIntervalMs:  checkIntervalMs,
	}
}

// GetUsage returns the current disk I/O in KB/s
func (m *DiskMonitor) GetUsage() (float64, error) {
	// Get current stats
	ioStats, err := disk.IOCounters()
	if err != nil {
		return 0, err
	}

	// Calculate elapsed time since last check
	currentTime := time.Now()
	elapsedSecs := currentTime.Sub(m.lastCheckTime).Seconds()
	if elapsedSecs < 0.001 {
		return m.lastUsageKBps, nil // Return last value if time diff is too small
	}

	// Calculate bytes transferred since last check
	var currentReadBytes, currentWriteBytes uint64
	for _, stat := range ioStats {
		currentReadBytes += stat.ReadBytes
		currentWriteBytes += stat.WriteBytes
	}

	readBytesDiff := currentReadBytes - m.lastReadBytes
	writeBytesDiff := currentWriteBytes - m.lastWriteBytes
	totalBytesDiff := readBytesDiff + writeBytesDiff

	// Calculate KB/s
	kbps := float64(totalBytesDiff) / elapsedSecs / 1024.0

	// Update last check data
	m.lastCheckTime = currentTime
	m.lastReadBytes = currentReadBytes
	m.lastWriteBytes = currentWriteBytes
	m.lastUsageKBps = kbps

	return kbps, nil
}