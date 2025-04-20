// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package accelerator

import (
	"testing"

	"github.com/scttfrdmn/cloudsnooze/daemon/common"
)

// MockGPUMonitor is a mock implementation of the GPUMonitor interface for testing
type MockGPUMonitor struct {
	available bool
	metrics   []common.GPUMetrics
	err       error
}

func (m *MockGPUMonitor) IsAvailable() bool {
	return m.available
}

func (m *MockGPUMonitor) GetMetrics() ([]common.GPUMetrics, error) {
	return m.metrics, m.err
}

func TestGPUServiceImplementsAcceleratorInterface(t *testing.T) {
	// This test verifies that GPUService correctly implements the AcceleratorInterface
	var service common.AcceleratorInterface = NewGPUService()
	
	// If this compiles, the interface is implemented correctly
	if service == nil {
		t.Error("Service should not be nil")
	}
}

func TestGPUServiceInitialize(t *testing.T) {
	service := NewGPUService()
	err := service.Initialize()
	if err != nil {
		t.Errorf("Initialize() returned error: %v", err)
	}
}

func TestGPUServiceGetMetricsWithNoAvailableMonitors(t *testing.T) {
	// Create a service with a mock monitor that is not available
	service := &GPUService{
		monitors: []GPUMonitor{
			&MockGPUMonitor{available: false},
		},
	}
	
	// Get metrics should return empty slice
	metrics, err := service.GetMetrics()
	if err != nil {
		t.Errorf("GetMetrics() returned error with no available monitors: %v", err)
	}
	
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics slice, got %d metrics", len(metrics))
	}
}

func TestGPUServiceGetMetricsWithAvailableMonitors(t *testing.T) {
	// Create mock GPU metrics
	mockMetrics := []common.GPUMetrics{
		{
			ID:          "0",
			Utilization: 75.0,
			MemoryUsed:  4 * 1024 * 1024 * 1024,
			MemoryTotal: 8 * 1024 * 1024 * 1024,
			Temperature: 70.0,
			Vendor:      "NVIDIA",
			Model:       "RTX 3080",
		},
	}
	
	// Create a service with a mock monitor that is available
	service := &GPUService{
		monitors: []GPUMonitor{
			&MockGPUMonitor{
				available: true,
				metrics:   mockMetrics,
				err:       nil,
			},
		},
	}
	
	// Get metrics should return the mock metrics
	metrics, err := service.GetMetrics()
	if err != nil {
		t.Errorf("GetMetrics() returned error: %v", err)
	}
	
	if len(metrics) != 1 {
		t.Errorf("Expected 1 GPU metric, got %d", len(metrics))
	}
	
	if metrics[0].ID != "0" {
		t.Errorf("Expected GPU ID '0', got '%s'", metrics[0].ID)
	}
	
	if metrics[0].Utilization != 75.0 {
		t.Errorf("Expected utilization 75.0, got %f", metrics[0].Utilization)
	}
	
	if metrics[0].Vendor != "NVIDIA" {
		t.Errorf("Expected vendor 'NVIDIA', got '%s'", metrics[0].Vendor)
	}
}

func TestGPUServiceGetUtilization(t *testing.T) {
	// Create mock GPU metrics with different utilizations
	mockMetrics := []common.GPUMetrics{
		{
			ID:          "0",
			Utilization: 60.0,
		},
		{
			ID:          "1",
			Utilization: 80.0,
		},
	}
	
	// Create a service with a mock monitor
	service := &GPUService{
		monitors: []GPUMonitor{
			&MockGPUMonitor{
				available: true,
				metrics:   mockMetrics,
				err:       nil,
			},
		},
	}
	
	// Get average utilization
	utilization, err := service.GetUtilization()
	if err != nil {
		t.Errorf("GetUtilization() returned error: %v", err)
	}
	
	// Expected average: (60.0 + 80.0) / 2 = 70.0
	if utilization != 70.0 {
		t.Errorf("Expected average utilization 70.0, got %f", utilization)
	}
}