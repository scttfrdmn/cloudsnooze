// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package main

// Config represents the complete configuration
type Config struct {
	// General settings
	CheckIntervalSeconds int     `json:"check_interval_seconds"`
	NaptimeMinutes       int     `json:"naptime_minutes"`
	
	// Thresholds
	CPUThresholdPercent    float64 `json:"cpu_threshold_percent"`
	MemoryThresholdPercent float64 `json:"memory_threshold_percent"`
	NetworkThresholdKBps   float64 `json:"network_threshold_kbps"`
	DiskIOThresholdKBps    float64 `json:"disk_io_threshold_kbps"`
	InputIdleThresholdSecs int     `json:"input_idle_threshold_secs"`
	
	// GPU/Accelerator settings
	GPUMonitoringEnabled bool    `json:"gpu_monitoring_enabled"`
	GPUThresholdPercent  float64 `json:"gpu_threshold_percent"`
	
	// Cloud provider settings
	ProviderType         string `json:"provider_type"`       // Which cloud provider to use (empty for auto-detection)
	
	// AWS settings
	AWSRegion          string `json:"aws_region"`
	EnableInstanceTags bool   `json:"enable_instance_tags"`
	TaggingPrefix      string `json:"tagging_prefix"`
	
	// Tag-based monitoring for external tools
	DetailedInstanceTags    bool `json:"detailed_instance_tags"`     // Whether to add detailed tags about the stop reason
	TagPollingEnabled       bool `json:"tag_polling_enabled"`        // Whether to poll for tags from external systems
	TagPollingIntervalSecs  int  `json:"tag_polling_interval_secs"`  // How often to poll for tags (in seconds)
	
	// Logging settings
	Logging LoggingConfig `json:"logging"`
	
	// Advanced settings
	MonitoringMode string `json:"monitoring_mode"` // "basic" or "advanced"
	
	// Plugin settings
	PluginsEnabled bool   `json:"plugins_enabled"`     // Whether to use the plugin system
	PluginsDir     string `json:"plugins_dir"`         // Directory to load external plugins from
}

// LoggingConfig defines logging behavior
type LoggingConfig struct {
	LogLevel           string `json:"log_level"` // "debug", "info", "warn", "error"
	EnableFileLogging  bool   `json:"enable_file_logging"`
	LogFilePath        string `json:"log_file_path"`
	EnableSyslog       bool   `json:"enable_syslog"`
	EnableCloudWatch   bool   `json:"enable_cloudwatch"`
	CloudWatchLogGroup string `json:"cloudwatch_log_group"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		CheckIntervalSeconds:    60,
		NaptimeMinutes:          30,
		CPUThresholdPercent:     10.0,
		MemoryThresholdPercent:  30.0,
		NetworkThresholdKBps:    50.0,
		DiskIOThresholdKBps:     100.0,
		InputIdleThresholdSecs:  900,
		GPUMonitoringEnabled:    true,
		GPUThresholdPercent:     5.0,
		ProviderType:            "",  // Empty for auto-detection
		AWSRegion:               "us-east-1",
		EnableInstanceTags:      true,
		TaggingPrefix:           "CloudSnooze",
		DetailedInstanceTags:    true,
		TagPollingEnabled:       true,
		TagPollingIntervalSecs:  60,  // 1 minute by default
		Logging: LoggingConfig{
			LogLevel:           "info",
			EnableFileLogging:  true,
			LogFilePath:        "/var/log/cloudsnooze.log",
			EnableSyslog:       false,
			EnableCloudWatch:   false,
			CloudWatchLogGroup: "CloudSnooze",
		},
		MonitoringMode: "basic",
		PluginsEnabled: true,
		PluginsDir:     "/etc/cloudsnooze/plugins",
	}
}