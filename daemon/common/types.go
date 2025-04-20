// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package common

// SystemMetrics contains all metrics collected from the system
type SystemMetrics struct {
    CPUUsage        float64
    MemoryUsage     float64
    DiskIORate      float64
    NetworkRate     float64
    IdleTime        int64
    GPUMetrics      []GPUMetrics
    LastInputTime   int64
    CollectionTime  int64
}

// GPUMetrics contains metrics specific to GPU devices
type GPUMetrics struct {
    ID              string
    Utilization     float64
    MemoryUsed      uint64
    MemoryTotal     uint64
    Temperature     float64
    Vendor          string
    Model           string
}

// CloudProvider defines the interface for cloud providers
type CloudProvider interface {
    // VerifyPermissions checks if the daemon has sufficient permissions
    VerifyPermissions() (bool, error)
    
    // GetInstanceInfo retrieves information about the current instance
    GetInstanceInfo() (*InstanceInfo, error)
    
    // StopInstance stops the current instance
    StopInstance(reason string, metrics SystemMetrics) error
    
    // TagInstance adds tags to the current instance
    TagInstance(tags map[string]string) error
    
    // GetExternalTags checks for tags from external systems that might control this instance
    GetExternalTags() (map[string]string, error)
}

// InstanceInfo contains information about the current cloud instance
type InstanceInfo struct {
    ID         string
    Type       string
    Region     string
    Provider   string
    LaunchTime string
    Tags       map[string]string
}

// MonitorResult represents the result of a monitor check
type MonitorResult struct {
    IsIdle      bool
    IdleReason  string
    Metrics     interface{}
    Error       error
}

// MonitorInterface defines the common interface for all system monitors
type MonitorInterface interface {
    Initialize() error
    Check() MonitorResult
    GetName() string
    GetThreshold() float64
    SetThreshold(threshold float64) error
}

// AcceleratorInterface defines the common interface for GPU/accelerator monitors
type AcceleratorInterface interface {
    Initialize() error
    GetMetrics() ([]GPUMetrics, error)
    GetUtilization() (float64, error)
}