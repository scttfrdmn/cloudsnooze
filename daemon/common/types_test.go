// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"
)

func TestGPUMetricsStructure(t *testing.T) {
	// Create a sample GPU metrics object
	gpu := GPUMetrics{
		ID:          "1",
		Utilization: 45.5,
		MemoryUsed:  1024 * 1024 * 1024, // 1 GB
		MemoryTotal: 8 * 1024 * 1024 * 1024, // 8 GB
		Temperature: 65.0,
		Vendor:      "NVIDIA",
		Model:       "Tesla T4",
	}

	// Validate fields
	if gpu.ID != "1" {
		t.Errorf("GPU ID expected '1', got '%s'", gpu.ID)
	}

	if gpu.Utilization != 45.5 {
		t.Errorf("GPU Utilization expected 45.5, got %f", gpu.Utilization)
	}

	if gpu.MemoryUsed != 1024*1024*1024 {
		t.Errorf("GPU MemoryUsed expected 1GB, got %d bytes", gpu.MemoryUsed)
	}

	if gpu.MemoryTotal != 8*1024*1024*1024 {
		t.Errorf("GPU MemoryTotal expected 8GB, got %d bytes", gpu.MemoryTotal)
	}

	if gpu.Temperature != 65.0 {
		t.Errorf("GPU Temperature expected 65.0, got %f", gpu.Temperature)
	}

	if gpu.Vendor != "NVIDIA" {
		t.Errorf("GPU Vendor expected 'NVIDIA', got '%s'", gpu.Vendor)
	}

	if gpu.Model != "Tesla T4" {
		t.Errorf("GPU Model expected 'Tesla T4', got '%s'", gpu.Model)
	}
}

func TestSystemMetricsStructure(t *testing.T) {
	// Create a sample system metrics object with GPU metrics
	gpu := GPUMetrics{
		ID:          "1",
		Utilization: 45.5,
		MemoryUsed:  1024 * 1024 * 1024,
		MemoryTotal: 8 * 1024 * 1024 * 1024,
		Temperature: 65.0,
		Vendor:      "NVIDIA",
		Model:       "Tesla T4",
	}

	metrics := SystemMetrics{
		CPUUsage:       25.0,
		MemoryUsage:    40.0,
		DiskIORate:     15.5,
		NetworkRate:    3000.0,
		IdleTime:       300,
		GPUMetrics:     []GPUMetrics{gpu},
		LastInputTime:  1714500000,
		CollectionTime: 1714500300,
	}

	// Validate fields
	if metrics.CPUUsage != 25.0 {
		t.Errorf("CPU usage expected 25.0, got %f", metrics.CPUUsage)
	}

	if metrics.MemoryUsage != 40.0 {
		t.Errorf("Memory usage expected 40.0, got %f", metrics.MemoryUsage)
	}

	if metrics.DiskIORate != 15.5 {
		t.Errorf("Disk IO rate expected 15.5, got %f", metrics.DiskIORate)
	}

	if metrics.NetworkRate != 3000.0 {
		t.Errorf("Network rate expected 3000.0, got %f", metrics.NetworkRate)
	}

	if metrics.IdleTime != 300 {
		t.Errorf("Idle time expected 300, got %d", metrics.IdleTime)
	}

	if len(metrics.GPUMetrics) != 1 {
		t.Errorf("Expected 1 GPU, got %d", len(metrics.GPUMetrics))
	}

	if metrics.LastInputTime != 1714500000 {
		t.Errorf("Last input time expected 1714500000, got %d", metrics.LastInputTime)
	}

	if metrics.CollectionTime != 1714500300 {
		t.Errorf("Collection time expected 1714500300, got %d", metrics.CollectionTime)
	}

	// Validate GPU metrics within system metrics
	if metrics.GPUMetrics[0].ID != "1" {
		t.Errorf("GPU ID within system metrics expected '1', got '%s'", metrics.GPUMetrics[0].ID)
	}

	if metrics.GPUMetrics[0].Utilization != 45.5 {
		t.Errorf("GPU Utilization within system metrics expected 45.5, got %f", metrics.GPUMetrics[0].Utilization)
	}
}

func TestMonitorResultStructure(t *testing.T) {
	// Create a sample monitor result
	result := MonitorResult{
		IsIdle:     true,
		IdleReason: "CPU usage below threshold",
		Metrics:    25.5, // Example metric value
		Error:      nil,
	}

	// Validate fields
	if !result.IsIdle {
		t.Errorf("IsIdle expected true, got false")
	}

	if result.IdleReason != "CPU usage below threshold" {
		t.Errorf("IdleReason expected 'CPU usage below threshold', got '%s'", result.IdleReason)
	}

	if value, ok := result.Metrics.(float64); !ok || value != 25.5 {
		t.Errorf("Metrics expected 25.5, got %v", result.Metrics)
	}

	if result.Error != nil {
		t.Errorf("Error expected nil, got %v", result.Error)
	}
}