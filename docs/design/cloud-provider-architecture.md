<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# Cloud Provider Architecture

This document outlines the architecture for implementing pluggable cloud providers in CloudSnooze.

## Overview

The cloud provider architecture enables CloudSnooze to support multiple cloud platforms through a standardized interface. This design allows for:

1. Built-in providers for major cloud platforms (AWS, GCP, Azure)
2. Third-party providers via a plugin mechanism
3. A consistent interface for all cloud operations
4. Dynamic loading of providers at runtime

## Core Components

### Provider Interface

All cloud providers must implement the `CloudProvider` interface defined in the `common` package:

```go
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
```

### Provider Registry

The provider registry manages available providers and handles provider discovery:

```go
// ProviderRegistry manages cloud providers
type ProviderRegistry interface {
    // RegisterProvider adds a provider to the registry
    RegisterProvider(name string, factory ProviderFactory) error
    
    // GetProvider returns a provider by name
    GetProvider(name string, config interface{}) (CloudProvider, error)
    
    // ListProviders returns all registered providers
    ListProviders() []string
    
    // DetectProvider attempts to auto-detect the current cloud provider
    DetectProvider() (string, error)
}
```

### Provider Factory

Each provider is created through a factory:

```go
// ProviderFactory creates provider instances
type ProviderFactory interface {
    // CreateProvider creates a provider with the given config
    CreateProvider(config interface{}) (CloudProvider, error)
    
    // GetConfigType returns the type of config this provider expects
    GetConfigType() reflect.Type
}
```

## Implementation Plan

### Phase 1: Static Provider Refactoring

1. Refactor existing AWS provider to use the new interfaces
2. Implement a basic provider registry
3. Update configuration to specify provider type

### Phase 2: Dynamic Loading

1. Implement plugin discovery mechanism
2. Create plugin loader for external providers
3. Add versioning and compatibility checking

### Phase 3: Additional Built-in Providers

1. Implement GCP provider
2. Implement Azure provider
3. Add provider-specific configuration options

### Phase 4: Third-party Provider SDK

1. Create provider development documentation
2. Build example third-party provider
3. Set up provider testing framework

## Provider Implementation Guidelines

Each cloud provider implementation should:

1. Be contained in its own package
2. Implement all methods in the CloudProvider interface
3. Provide sensible defaults for its configuration
4. Include thorough error handling and logging
5. Be thread-safe
6. Include comprehensive documentation
7. Include unit and integration tests

## Configuration

The system configuration will be extended to include provider-specific sections:

```json
{
  "provider": "aws",
  "provider_configs": {
    "aws": {
      "region": "us-east-1",
      "enable_tags": true,
      "tag_prefix": "CloudSnooze",
      "detailed_tags": true
    },
    "gcp": {
      "project_id": "my-project",
      "zone": "us-central1-a"
    },
    "azure": {
      "resource_group": "my-resources",
      "location": "eastus"
    }
  }
}
```

## Metrics Documentation

The following metrics are monitored by CloudSnooze to determine instance idle state. Each metric has a configurable threshold:

| Metric | Parameter | Description | Default Threshold | Units |
|--------|-----------|-------------|------------------|-------|
| CPU Usage | `cpu_threshold_percent` | Average CPU utilization across all cores | 10.0 | Percentage (0-100) |
| Memory Usage | `memory_threshold_percent` | Percentage of used memory relative to total memory | 30.0 | Percentage (0-100) |
| Network Traffic | `network_threshold_kbps` | Combined ingress/egress network traffic | 50.0 | Kilobytes per second |
| Disk I/O | `disk_io_threshold_kbps` | Combined read/write disk operations | 100.0 | Kilobytes per second |
| Input Activity | `input_idle_threshold_secs` | Time since last keyboard/mouse activity | 900 | Seconds |
| GPU Usage | `gpu_threshold_percent` | Average GPU utilization across all detected GPUs | 5.0 | Percentage (0-100) |

## Plugin Architecture

### Plugin Manifest

Each cloud provider plugin must include a manifest file that contains:

```json
{
  "name": "cloudsnooze-gcp-provider",
  "version": "1.0.0",
  "provider_name": "gcp",
  "compatible_versions": [">=0.1.0"],
  "author": "CloudSnooze Team",
  "entry_point": "main.so",
  "config_schema": {
    "project_id": "string",
    "zone": "string",
    "enable_tags": "bool"
  }
}
```

### Plugin Loading

Plugins will be loaded from a designated plugin directory, typically:
- Linux: `/usr/lib/cloudsnooze/plugins/`
- macOS: `/usr/local/lib/cloudsnooze/plugins/`
- User-specified location via configuration

## Conclusion

The pluggable cloud provider architecture provides flexibility and extensibility for CloudSnooze. It enables supporting multiple cloud platforms while maintaining a consistent internal API, and allows third-party developers to extend the system for specialized use cases.