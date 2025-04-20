<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Plugin Architecture

This document describes the technical architecture for the CloudSnooze Plugin System, which extends CloudSnooze's idle detection capabilities through a flexible plugin framework.

## Architecture Overview

The CloudSnooze Plugin System uses a hybrid approach combining:

1. **Native Go Plugins**: For performance-critical, trusted plugins
2. **External Process Plugins**: For third-party, language-agnostic plugins with stronger isolation

This dual approach balances performance, security, and flexibility while enabling a rich ecosystem of extensions.

<p align="center">
  <img src="/images/plugin-architecture.svg" alt="CloudSnooze Plugin Architecture" width="700"/>
</p>

## Core Components

### 1. Plugin Interface

All plugins must implement this interface:

```go
// SnoozePluginInterface defines the contract all plugins must implement
type SnoozePluginInterface interface {
    // Info returns plugin metadata
    Info() PluginInfo
    
    // CheckIdle determines if the system is idle according to this plugin
    // Returns: (isIdle bool, idleReason string, error)
    CheckIdle(metrics SystemMetrics) (bool, string, error)
    
    // Configure allows runtime configuration of the plugin
    Configure(config map[string]interface{}) error
    
    // Initialize sets up the plugin with initial configuration
    Initialize() error
    
    // Shutdown allows plugins to clean up resources
    Shutdown() error
}
```

### 2. Plugin Manager

The Plugin Manager handles:

- Plugin discovery and loading
- Plugin lifecycle management (init, shutdown)
- Plugin execution and error handling
- Configuration management

```go
type PluginManager struct {
    // Loaded plugins
    plugins map[string]SnoozePluginInterface
    
    // Plugin execution statistics
    stats map[string]*PluginStats
    
    // Configuration
    config PluginManagerConfig
    
    // Process manager for external plugins
    procManager *ProcessManager
}

func (pm *PluginManager) LoadPlugin(path string) error
func (pm *PluginManager) UnloadPlugin(name string) error
func (pm *PluginManager) EnablePlugin(name string) error
func (pm *PluginManager) DisablePlugin(name string) error
func (pm *PluginManager) ConfigurePlugin(name string, config map[string]interface{}) error
func (pm *PluginManager) CheckAllPlugins(metrics SystemMetrics) (map[string]PluginResult, error)
```

### 3. Plugin Metadata

Plugin metadata provides information about each plugin:

```go
type PluginInfo struct {
    // Basic plugin information
    Name            string   // Plugin identifier
    DisplayName     string   // Human-readable name
    Version         string   // Semver version
    Author          string   // Author information
    Description     string   // Plugin description
    Category        string   // Plugin category
    Tags            []string // Searchable tags
    
    // Plugin source information
    Repository      string   // Source code location
    Homepage        string   // Documentation URL
    Official        bool     // If part of official repository
    
    // Configuration and requirements
    ConfigSchema    map[string]ConfigField // Configuration schema
    Dependencies    []Dependency           // Required dependencies
    MinHostVersion  string                 // Minimum CloudSnooze version
    
    // Plugin implementation details
    Type            PluginType // Native/External
    Language        string     // Implementation language
    ProcessModel    string     // Execution model
}

type ConfigField struct {
    Type        string      // Data type (string, int, bool, etc.)
    Description string      // Field description
    Default     interface{} // Default value
    Required    bool        // If required
    Constraints string      // Validation constraints
}

type Dependency struct {
    Name    string // Dependency name
    Version string // Version constraint
}
```

### 4. Plugin Registry

The registry manages available plugins:

```go
type PluginRegistry struct {
    // Official plugins
    OfficialPlugins []PluginInfo
    
    // Community plugins
    CommunityPlugins []PluginInfo
    
    // External repositories
    ExternalRepositories []RepositoryInfo
}

type RepositoryInfo struct {
    Name        string // Repository name
    URL         string // Repository URL
    Description string // Repository description 
    Trusted     bool   // If trusted
}
```

### 5. IPC Protocol (External Plugins)

External plugins communicate using a simple JSON-RPC protocol:

```json
// Request format (from CloudSnooze to plugin)
{
  "method": "CheckIdle",
  "params": {
    "metrics": {
      "cpu_usage": 1.5,
      "memory_usage": 25.0,
      "network_traffic": 0.1,
      "disk_io": 0.5,
      "gpu_usage": 0.0,
      "timestamp": 1650000000
    }
  },
  "id": 12345
}

// Response format (from plugin to CloudSnooze)
{
  "result": {
    "is_idle": true,
    "reason": "Database has no active connections",
    "details": {
      "connection_count": 0,
      "last_query_timestamp": 1649999000
    }
  },
  "id": 12345
}
```

## Plugin Types

### 1. Native Go Plugins

Native plugins are compiled Go packages loaded directly into the CloudSnooze process:

- Higher performance
- Direct memory access
- Type safety
- Must be compiled against the exact Go version

Example:

```go
// mysql_monitor.go
package main

import "github.com/scttfrdmn/cloudsnooze/plugin"

type MySQLMonitor struct {
    connectionThreshold int
    dsn string
}

func (m *MySQLMonitor) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        Name: "mysql-monitor",
        DisplayName: "MySQL Connection Monitor",
        // ... additional metadata
    }
}

func (m *MySQLMonitor) CheckIdle(metrics plugin.SystemMetrics) (bool, string, error) {
    // Check if MySQL is idle based on connection count
    connections, err := m.getConnectionCount()
    if err != nil {
        return false, "", err
    }
    
    if connections < m.connectionThreshold {
        return true, fmt.Sprintf("MySQL has only %d active connections (below threshold %d)", 
               connections, m.connectionThreshold), nil
    }
    
    return false, "", nil
}

// Implementation of other interface methods...

// This is required for Go plugins
var Plugin MySQLMonitor
```

### 2. External Process Plugins

External plugins run as separate processes and can be written in any language:

- Strong isolation
- Language-agnostic
- Easier debugging
- Potentially higher resource usage

Example (Python):

```python
#!/usr/bin/env python3
# redis_monitor.py

import json
import sys
import redis

class RedisMonitor:
    def __init__(self):
        self.connection_threshold = 5
        self.redis_url = "redis://localhost:6379"
        
    def info(self):
        return {
            "name": "redis-monitor",
            "display_name": "Redis Connection Monitor",
            "version": "1.0.0",
            "author": "CloudSnooze Team",
            "description": "Monitors Redis for active connections",
            # ... additional metadata
        }
    
    def check_idle(self, metrics):
        try:
            r = redis.from_url(self.redis_url)
            info = r.info()
            connected_clients = info.get('connected_clients', 0)
            
            if connected_clients < self.connection_threshold:
                return {
                    "is_idle": True,
                    "reason": f"Redis has only {connected_clients} connections (below threshold {self.connection_threshold})"
                }
            return {"is_idle": False, "reason": ""}
        except Exception as e:
            return {"error": str(e)}
    
    def configure(self, config):
        if 'connection_threshold' in config:
            self.connection_threshold = config['connection_threshold']
        if 'redis_url' in config:
            self.redis_url = config['redis_url']
        return {"success": True}

# Main loop to handle IPC
monitor = RedisMonitor()

for line in sys.stdin:
    try:
        request = json.loads(line)
        method = request.get('method')
        params = request.get('params', {})
        req_id = request.get('id')
        
        if method == 'Info':
            result = monitor.info()
        elif method == 'CheckIdle':
            result = monitor.check_idle(params.get('metrics', {}))
        elif method == 'Configure':
            result = monitor.configure(params.get('config', {}))
        else:
            result = {"error": f"Unknown method: {method}"}
        
        response = {"result": result, "id": req_id}
        print(json.dumps(response))
        sys.stdout.flush()
    except Exception as e:
        error_response = {"error": str(e), "id": req_id if 'req_id' in locals() else None}
        print(json.dumps(error_response))
        sys.stdout.flush()
```

## Event-Driven Plugin Framework

The CloudSnooze Event-Driven Plugin Framework extends the basic plugin architecture to handle bidirectional communication. This allows CloudSnooze to both receive information from plugins (idle detection) and trigger plugin actions in response to external events (such as AWS Spot instance interruption notifications).

### 1. Event Types

```go
// EventType identifies different kinds of events plugins can handle
type EventType string

const (
    // Cloud provider events
    EventSpotInterrupt        EventType = "cloud.spot.interrupt"
    EventInstanceTermination  EventType = "cloud.instance.termination"
    EventAutoscalingScale     EventType = "cloud.autoscaling.scale"
    
    // System events
    EventLowDiskSpace         EventType = "system.disk.low"
    EventHighCPU              EventType = "system.cpu.high"
    EventMemoryPressure       EventType = "system.memory.pressure"
    
    // Custom events
    EventCustom               EventType = "custom"
)
```

### 2. Event Data

```go
// Event represents a structured event that can trigger plugin actions
type Event struct {
    // Event metadata
    Type        EventType           // Type of event
    Source      string              // Event source (e.g., "aws.spot-fleet")
    Timestamp   time.Time           // When the event occurred
    ID          string              // Unique event ID
    
    // Event data
    Data        map[string]interface{} // Event-specific data
    
    // Severity and timing
    Severity    EventSeverity       // Event importance
    Deadline    *time.Time          // Optional deadline for action
}

type EventSeverity int

const (
    SeverityInfo     EventSeverity = 0
    SeverityWarning  EventSeverity = 1
    SeverityUrgent   EventSeverity = 2
    SeverityCritical EventSeverity = 3
)
```

### 3. Event Handling Interface

Plugins can optionally implement this interface to handle events:

```go
// EventHandlerInterface defines methods for plugins that can handle events
type EventHandlerInterface interface {
    // Required methods from base SnoozePluginInterface
    
    // HandleEvent processes an event and takes appropriate action
    // Returns: (handled bool, result map[string]interface{}, error)
    HandleEvent(event Event) (bool, map[string]interface{}, error)
    
    // GetSupportedEvents returns events this plugin can handle
    GetSupportedEvents() []EventType
}
```

### 4. Event Processing

The event processing workflow:

```go
func (pm *PluginManager) HandleEvent(event Event) error {
    // Find plugins that can handle this event
    handlers := pm.findEventHandlers(event.Type)
    
    // Sort by priority
    pm.sortHandlersByPriority(handlers)
    
    // Process through handlers until one succeeds
    for _, handler := range handlers {
        handled, result, err := handler.HandleEvent(event)
        if err != nil {
            log.Printf("Plugin %s failed to handle event: %v", handler.Info().Name, err)
            continue
        }
        
        if handled {
            log.Printf("Event %s handled by plugin %s: %v", 
                event.ID, handler.Info().Name, result)
            return nil
        }
    }
    
    return fmt.Errorf("no plugin successfully handled event %s", event.ID)
}
```

### 5. Example: AWS Spot Instance Interrupt Handler

```go
// spot_interrupt_handler.go
package main

import (
    "context"
    "github.com/scttfrdmn/cloudsnooze/plugin"
    "os/exec"
    "time"
)

type SpotInterruptHandler struct {
    shutdownScriptPath string
    timeoutSeconds     int
    notified           bool
}

func (h *SpotInterruptHandler) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        Name:        "spot-interrupt-handler",
        DisplayName: "AWS Spot Interrupt Handler",
        Version:     "1.0.0",
        Author:      "CloudSnooze Team",
        Description: "Handles AWS Spot instance interruption notifications",
        Category:    "cloud-event",
        Tags:        []string{"aws", "spot", "interrupt"},
        // ... additional metadata
    }
}

func (h *SpotInterruptHandler) GetSupportedEvents() []plugin.EventType {
    return []plugin.EventType{plugin.EventSpotInterrupt}
}

func (h *SpotInterruptHandler) HandleEvent(event plugin.Event) (bool, map[string]interface{}, error) {
    if event.Type != plugin.EventSpotInterrupt {
        return false, nil, nil
    }
    
    // Already handled this event?
    if h.notified {
        return true, map[string]interface{}{"status": "already_handled"}, nil
    }
    
    // Extract information from event
    timeRemaining, ok := event.Data["timeRemaining"].(int)
    if !ok {
        timeRemaining = 120 // Default to 2 minutes if not specified
    }
    
    // Set a context with deadline based on time remaining
    ctx, cancel := context.WithTimeout(context.Background(), 
        time.Duration(min(timeRemaining-15, h.timeoutSeconds))*time.Second)
    defer cancel()
    
    // Execute shutdown script
    cmd := exec.CommandContext(ctx, h.shutdownScriptPath)
    output, err := cmd.CombinedOutput()
    
    h.notified = true
    
    result := map[string]interface{}{
        "script_output": string(output),
        "time_remaining": timeRemaining,
    }
    
    if err != nil {
        result["success"] = false
        result["error"] = err.Error()
        return true, result, nil
    }
    
    result["success"] = true
    return true, result, nil
}

// Implement other SnoozePluginInterface methods...

// This is required for Go plugins
var Plugin SpotInterruptHandler
```

### 6. Event Sources

CloudSnooze monitors several event sources:

1. **Cloud Provider APIs**
   - AWS EC2 Spot Instance Interruption Notices (via metadata service)
   - AWS Auto Scaling Group Events (via SNS/EventBridge)
   - Azure Scheduled Events API
   - GCP Preemptible VM Termination Events

2. **System Event Monitoring**
   - Disk space thresholds
   - CPU/Memory pressure indicators
   - Network connectivity issues

3. **Custom Event Sources**
   - Custom endpoint for external events
   - Integrated applications
   - User-triggered events

### 7. Event Configuration

Events can be configured per plugin:

```json
{
  "plugins": {
    "plugin_configs": {
      "spot-interrupt-handler": {
        "shutdown_script_path": "/usr/local/bin/graceful_shutdown.sh",
        "timeout_seconds": 90,
        "events": {
          "cloud.spot.interrupt": {
            "priority": 100,
            "enabled": true
          }
        }
      },
      "database-flusher": {
        "db_type": "mysql",
        "connection_string": "user:pass@tcp(localhost:3306)/mydb",
        "events": {
          "cloud.spot.interrupt": {
            "priority": 50,
            "enabled": true
          },
          "system.disk.low": {
            "priority": 80,
            "enabled": true
          }
        }
      }
    }
  }
}
```

## Integration with CloudSnooze

### 1. System Monitor Extension

The System Monitor component will be extended to include plugin results:

```go
func (sm *SystemMonitor) ShouldSnooze() (bool, string) {
    // Check built-in monitors
    if shouldSnooze, reason := sm.checkBuiltInMonitors(); shouldSnooze {
        return true, reason
    }
    
    // Check plugins
    pluginResults, err := sm.pluginManager.CheckAllPlugins(sm.lastMetrics)
    if err != nil {
        log.Printf("Error checking plugins: %v", err)
    }
    
    // Evaluate plugin results
    for name, result := range pluginResults {
        if result.IsIdle {
            return true, fmt.Sprintf("Plugin %s: %s", name, result.Reason)
        }
    }
    
    return false, ""
}
```

### 2. Configuration Integration

Plugin configuration will be integrated into CloudSnooze's configuration:

```json
{
  "cpu_threshold_percent": 5.0,
  "memory_threshold_percent": 10.0,
  "network_threshold_kbps": 50.0,
  "disk_io_threshold_kbps": 100.0,
  "naptime_minutes": 30,
  
  "plugins": {
    "enabled": true,
    "directories": [
      "/etc/cloudsnooze/plugins",
      "/usr/local/share/cloudsnooze/plugins"
    ],
    "external_repositories": [
      {
        "name": "community",
        "url": "https://plugins.cloudsnooze.io/community",
        "trusted": true
      }
    ],
    "plugin_configs": {
      "mysql-monitor": {
        "connection_threshold": 5,
        "dsn": "user:pass@tcp(localhost:3306)/dbname"
      },
      "redis-monitor": {
        "connection_threshold": 3,
        "redis_url": "redis://localhost:6379"
      }
    }
  }
}
```

### 3. CLI Integration

The CloudSnooze CLI will be extended with plugin management commands:

```
# Plugin management commands
snooze plugin list                           # List all plugins
snooze plugin info mysql-monitor              # Show plugin details
snooze plugin install redis-monitor           # Install a plugin
snooze plugin uninstall redis-monitor         # Remove a plugin
snooze plugin enable mysql-monitor            # Enable a plugin
snooze plugin disable mysql-monitor           # Disable a plugin
snooze plugin configure mysql-monitor         # Configure a plugin
snooze plugin update-all                      # Update all plugins
snooze plugin search database                 # Search for plugins
```

## Plugin Development Workflow

1. **Create** a new plugin
   - Use provided templates
   - Implement the plugin interface
   - Package with metadata

2. **Test** locally
   - Use the provided testing framework
   - Verify against test metrics
   - Validate configuration handling

3. **Package** for distribution
   - Create release artifacts
   - Sign plugin (for official plugins)
   - Generate documentation

4. **Publish** to registry
   - Submit to official/community repository
   - Provide installation instructions
   - Include usage examples

## Security Considerations

1. **Plugin Isolation**
   - Native plugins run in-process (trusted code only)
   - External plugins run as separate processes
   - Resource limits enforced by ProcessManager

2. **Permission Model**
   - Plugins define required permissions
   - Users must approve permission requests
   - Granular permission control

3. **Verification**
   - Official plugins are signed and verified
   - Community plugins display trust warnings
   - Plugin hashes are verified

4. **Monitoring**
   - Plugin performance is monitored
   - Crashes/hangs are detected
   - Excessive resource usage triggers warnings

## Future Extensions

1. **Remote Plugin Execution**
   - Run plugins on separate servers
   - Aggregate results from multiple sources
   - Centralized management

2. **Plugin Marketplace**
   - Rating and review system
   - Download statistics
   - Featured plugins section

3. **Advanced Plugin Types**
   - Machine learning-based idle detection
   - Predictive analytics plugins
   - Cross-instance coordination plugins
   
4. **Event-Driven Plugin Actions**
   - Cloud provider event integration
   - Custom triggers and reactions
   - Event prioritization system