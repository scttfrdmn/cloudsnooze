// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// NetworkMonitor handles network usage monitoring
type NetworkMonitor struct {
	lastCheckTime   time.Time
	lastBytesRecv   uint64
	lastBytesSent   uint64
	lastUsageKBps   float64
	checkIntervalMs int
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(checkIntervalMs int) *NetworkMonitor {
	// Get initial stats
	ioStats, _ := net.IOCounters(false)
	var initialBytesRecv, initialBytesSent uint64
	if len(ioStats) > 0 {
		initialBytesRecv = ioStats[0].BytesRecv
		initialBytesSent = ioStats[0].BytesSent
	}

	return &NetworkMonitor{
		lastCheckTime:   time.Now(),
		lastBytesRecv:   initialBytesRecv,
		lastBytesSent:   initialBytesSent,
		checkIntervalMs: checkIntervalMs,
	}
}

// GetUsage returns the current network I/O in KB/s
func (m *NetworkMonitor) GetUsage() (float64, error) {
	// Get current stats
	ioStats, err := net.IOCounters(false)
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
	var currentBytesRecv, currentBytesSent uint64
	if len(ioStats) > 0 {
		currentBytesRecv = ioStats[0].BytesRecv
		currentBytesSent = ioStats[0].BytesSent
	}

	bytesRecvDiff := currentBytesRecv - m.lastBytesRecv
	bytesSentDiff := currentBytesSent - m.lastBytesSent
	totalBytesDiff := bytesRecvDiff + bytesSentDiff

	// Calculate KB/s
	kbps := float64(totalBytesDiff) / elapsedSecs / 1024.0

	// Update last check data
	m.lastCheckTime = currentTime
	m.lastBytesRecv = currentBytesRecv
	m.lastBytesSent = currentBytesSent
	m.lastUsageKBps = kbps

	return kbps, nil
}