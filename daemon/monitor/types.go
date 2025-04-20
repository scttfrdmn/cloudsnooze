// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"time"
)

// SystemMetrics represents a complete set of system measurements
type SystemMetrics struct {
	Timestamp     time.Time   `json:"timestamp"`
	CPUPercent    float64     `json:"cpu_percent"`
	MemoryPercent float64     `json:"memory_percent"`
	NetworkKBps   float64     `json:"network_kbps"`
	DiskIOKBps    float64     `json:"disk_io_kbps"`
	InputIdleSecs int         `json:"input_idle_secs"`
	GPUMetrics    []GPUMetric `json:"gpu_metrics,omitempty"`
	IdleStatus    bool        `json:"idle_status"` // true if system is idle
	IdleReason    string      `json:"idle_reason,omitempty"`
}

// GPUMetric represents metrics for a single GPU
type GPUMetric struct {
	Type        string  `json:"type"` // "NVIDIA", "AMD", etc.
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Utilization float64 `json:"utilization"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryTotal uint64  `json:"memory_total"`
	Temperature float64 `json:"temperature,omitempty"`
}

// SnoozeEvent represents a stopping action
type SnoozeEvent struct {
	Timestamp    time.Time         `json:"timestamp"`
	InstanceID   string            `json:"instance_id"`
	InstanceType string            `json:"instance_type"`
	Region       string            `json:"region"`
	Reason       string            `json:"reason"`
	Metrics      SystemMetrics     `json:"metrics"`
	Tags         map[string]string `json:"tags,omitempty"`
	NaptimeMins  int               `json:"naptime_mins"`
}