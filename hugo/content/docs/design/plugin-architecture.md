# CloudSnooze Plugin Architecture

CloudSnooze uses a plugin architecture to provide extensibility and modularity, particularly for cloud provider integrations. This document describes the plugin architecture and how to develop new plugins.

## Overview

The plugin system allows CloudSnooze to be extended with new functionality without modifying the core code. Currently, the primary use case is for cloud providers, allowing CloudSnooze to work with different cloud platforms.

## Plugin Registry

At the core of the plugin architecture is the plugin registry, which keeps track of all available plugins. Plugins self-register when loaded, making them available to the application.

## Plugin Types

Currently, CloudSnooze supports the following plugin types:

- **Cloud Provider Plugins**: Implement cloud provider-specific logic for detecting, stopping, and tagging instances

## Plugin Interface

All plugins must implement the `Plugin` interface defined in `daemon/plugin/plugin.go`:

```go
type Plugin interface {
    // Info returns plugin metadata
    Info() PluginInfo
    
    // Init initializes the plugin with configuration
    Init(config interface{}) error
    
    // Start starts the plugin
    Start() error
    
    // Stop gracefully stops the plugin
    Stop() error
    
    // IsRunning returns true if the plugin is running
    IsRunning() bool
}
```

## Plugin Metadata

Each plugin provides metadata through the `PluginInfo` structure:

```go
type PluginInfo struct {
    ID           string            // Unique identifier
    Name         string            // Human-readable name
    Type         string            // Plugin type (e.g., "cloud-provider")
    Version      string            // Version string
    Capabilities map[string]bool   // Capabilities this plugin supports
    Author       string            // Plugin author
    Website      string            // Plugin website or repository
    Dependencies []string          // IDs of plugins this plugin depends on
}
```

## Cloud Provider Plugins

Cloud provider plugins must implement the `CloudProviderPlugin` interface defined in `daemon/plugin/cloud/provider.go`:

```go
type CloudProviderPlugin interface {
    plugin.Plugin
    
    // CreateProvider creates a new provider instance with the given configuration
    CreateProvider(config interface{}) (common.CloudProvider, error)
    
    // CanDetect returns true if this plugin can detect if it's running on this cloud provider
    CanDetect() bool
    
    // Detect tries to detect if the current environment is running on this cloud provider
    Detect() (bool, error)
}
```

## Plugin Loading

Plugins can be loaded in two ways:

1. **Built-in Plugins**: These are compiled directly into the binary and self-register via their `init()` functions
2. **External Plugins**: These are loaded from shared libraries (.so files) in a configured plugins directory

## Plugin Configuration

Plugin behavior is configured through the application configuration:

```json
{
  "provider_type": "aws",       // Cloud provider to use (empty for auto-detection)
  "plugins_enabled": true,      // Whether to use the plugin system
  "plugins_dir": "/etc/cloudsnooze/plugins"  // Directory to load external plugins from
}
```

## Plugin Manifests

External plugins can include a manifest file (`manifest.json`) with metadata:

```json
{
  "id": "aws",
  "name": "AWS Cloud Provider",
  "type": "cloud-provider",
  "version": "1.0.0",
  "capabilities": {
    "tagging": true,
    "tag-polling": true,
    "restart": true
  },
  "author": "CloudSnooze Contributors",
  "website": "https://github.com/scttfrdmn/cloudsnooze",
  "dependencies": []
}
```

## Creating a Cloud Provider Plugin

To create a new cloud provider plugin:

1. Create a new package in `daemon/plugin/cloud/<provider-name>/`
2. Implement the `CloudProviderPlugin` interface
3. Create a provider implementation of the `common.CloudProvider` interface
4. Register your plugin in the `init()` function
5. Create a manifest.json file with plugin metadata

Example implementation for a new provider:

```go
package myprovider

import (
    "github.com/scttfrdmn/cloudsnooze/daemon/common"
    "github.com/scttfrdmn/cloudsnooze/daemon/plugin"
    cloudplugin "github.com/scttfrdmn/cloudsnooze/daemon/plugin/cloud"
)

// MyProvider implements the CloudProviderPlugin interface
type MyProvider struct {
    initialized bool
    running     bool
    config      interface{}
}

// Register the plugin
func init() {
    plugin.Registry.Register(NewMyProvider())
}

// NewMyProvider creates a new provider plugin
func NewMyProvider() *MyProvider {
    return &MyProvider{}
}

// Info returns plugin metadata
func (p *MyProvider) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        ID:          "myprovider",
        Name:        "My Cloud Provider",
        Type:        plugin.TypeCloudProvider,
        Version:     "1.0.0",
        Capabilities: map[string]bool{
            "tagging": true,
        },
        Author:   "You",
        Website:  "https://example.com",
    }
}

// Implement other required methods...
```

## Using Plugins via CLI

You can list installed plugins using the CLI:

```
snooze plugins
```

For JSON output:

```
snooze plugins --json
```

## Future Extensions

The plugin architecture is designed to be extended beyond cloud providers. Future plugin types might include:

- Monitoring plugins for custom metrics
- Notification plugins for alerting
- Authentication plugins
- Analytics plugins for cost savings reporting
EOF < /dev/null