<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze API Reference

This document provides detailed information about CloudSnooze's APIs for integration with other tools and systems.

<p align="center">
  <img src="/images/api-architecture.svg" alt="CloudSnooze API Architecture" width="700"/>
</p>

## Socket API

CloudSnooze provides a Unix socket-based API for local communication with other applications running on the same instance.

### Socket Location

By default: `/var/run/cloudsnooze.sock`

This can be configured with the `--socket` command-line parameter when starting the daemon.

### Protocol

The socket API uses a simple JSON-based request/response protocol:

1. **Request**: A JSON object with `command` and `params` fields
2. **Response**: A JSON object with the result or error information

### Authentication

The socket is protected by filesystem permissions. By default, only root and members of the `cloudsnooze` group have access.

### Commands

#### STATUS

Gets the current status of the system and idle detection.

**Request:**
```json
{
  "command": "STATUS",
  "params": {}
}
```

**Response:**
```json
{
  "metrics": {
    "timestamp": "2023-04-19T14:23:45Z",
    "cpu_percent": 5.2,
    "memory_percent": 22.7,
    "network_kbps": 12.3,
    "disk_io_kbps": 0.5,
    "input_idle_secs": 125,
    "gpu_metrics": [
      {
        "type": "NVIDIA",
        "id": 0,
        "name": "Tesla T4",
        "utilization": 0.05,
        "memory_used": 123456789,
        "memory_total": 16000000000,
        "temperature": 42.5
      }
    ],
    "idle_status": false,
    "idle_reason": "Input activity detected"
  },
  "idle_since": null,
  "should_snooze": false,
  "snooze_reason": "System is not idle",
  "version": "0.1.0",
  "instance_info": {
    "id": "i-01234567890abcdef",
    "type": "t3.medium",
    "region": "us-east-1",
    "provider": "aws",
    "tags": {}
  }
}
```

#### CONFIG_GET

Retrieves the current configuration.

**Request:**
```json
{
  "command": "CONFIG_GET",
  "params": {}
}
```

**Response:**
```json
{
  "check_interval_seconds": 60,
  "naptime_minutes": 30,
  "cpu_threshold_percent": 10.0,
  "memory_threshold_percent": 30.0,
  "network_threshold_kbps": 50.0,
  "disk_io_threshold_kbps": 100.0,
  "input_idle_threshold_secs": 900,
  "gpu_monitoring_enabled": true,
  "gpu_threshold_percent": 5.0,
  "aws_region": "us-east-1",
  "enable_instance_tags": true,
  "tagging_prefix": "CloudSnooze",
  "detailed_instance_tags": true,
  "tag_polling_enabled": true,
  "tag_polling_interval_secs": 60,
  "enable_restart_flag": true,
  "allowed_restarter_ids": ["UserPortal", "JobScheduler"],
  "logging": {
    "log_level": "info",
    "enable_file_logging": true,
    "log_file_path": "/var/log/cloudsnooze.log",
    "enable_syslog": false,
    "enable_cloudwatch": false,
    "cloudwatch_log_group": "CloudSnooze"
  },
  "monitoring_mode": "basic"
}
```

#### CONFIG_SET

Updates configuration values. Currently a placeholder for future implementation.

**Request:**
```json
{
  "command": "CONFIG_SET",
  "params": {
    "naptime_minutes": 45,
    "cpu_threshold_percent": 15.0
  }
}
```

**Response:**
```json
{
  "updated": false,
  "message": "Not implemented yet"
}
```

#### HISTORY

Retrieves historical snooze events. Currently a placeholder for future implementation.

**Request:**
```json
{
  "command": "HISTORY",
  "params": {
    "limit": 10
  }
}
```

**Response:**
```json
[]
```

## Tag-Based API

CloudSnooze also exposes a tag-based "API" through the instance tags it manages.

### Tag Format

All CloudSnooze tags are prefixed with the configured prefix (default: `CloudSnooze`) followed by a colon and the tag name.

### Standard Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `CloudSnooze:Status` | Current status | `Running` or `Stopped` |
| `CloudSnooze:StopTimestamp` | When the instance was stopped | `2023-04-19T14:23:45Z` |
| `CloudSnooze:StopReason` | Why the instance was stopped | `System idle for 30 minutes` |
| `CloudSnooze:RestartAllowed` | Whether external tools can restart | `true` |
| `CloudSnooze:AllowedRestarters` | Comma-separated list of service IDs allowed to restart | `UserPortal,JobScheduler` |

### Extended Tags

When detailed tagging is enabled, additional tags provide metrics information:

| Tag | Description | Example |
|-----|-------------|---------|
| `CloudSnooze:CPUPercent` | CPU usage at stop time | `2.50` |
| `CloudSnooze:MemoryPercent` | Memory usage at stop time | `15.20` |
| `CloudSnooze:NetworkKBps` | Network usage at stop time | `0.80` |
| `CloudSnooze:DiskIOKBps` | Disk I/O at stop time | `0.30` |
| `CloudSnooze:InputIdleSecs` | Input idle time in seconds | `1800` |
| `CloudSnooze:GPUPercent` | Average GPU usage (if present) | `0.05` |
| `CloudSnooze:GPUCount` | Number of GPUs detected | `2` |

### Tag Lifecycle

1. **Instance Running**: 
   - `CloudSnooze:Status` = `Running`

2. **Instance Stopped by CloudSnooze**:
   - `CloudSnooze:Status` = `Stopped`
   - `CloudSnooze:StopTimestamp` added
   - `CloudSnooze:StopReason` added
   - `CloudSnooze:RestartAllowed` added (if enabled)
   - `CloudSnooze:AllowedRestarters` added (if configured)
   - Detailed metrics tags added (if enabled)

3. **Instance Restarted by External Tool**:
   - External tool should verify it's in the `AllowedRestarters` list (if specified)
   - External tool should update `CloudSnooze:Status` to `Running`
   - External tool should add `CloudSnooze:RestartTimestamp`
   - External tool should add `CloudSnooze:RestartReason`
   - External tool should add `CloudSnooze:RestartedBy` with its service ID

## Programming Examples

### Socket API in Go

```go
package main

import (
    "encoding/json"
    "net"
    "fmt"
)

type StatusResponse struct {
    Metrics struct {
        Timestamp    string  `json:"timestamp"`
        CPUPercent   float64 `json:"cpu_percent"`
        MemoryPercent float64 `json:"memory_percent"`
        NetworkKBps  float64 `json:"network_kbps"`
        DiskIOKBps   float64 `json:"disk_io_kbps"`
        InputIdleSecs int    `json:"input_idle_secs"`
        IdleStatus   bool    `json:"idle_status"`
        IdleReason   string  `json:"idle_reason"`
    } `json:"metrics"`
    IdleSince    string `json:"idle_since"`
    ShouldSnooze bool   `json:"should_snooze"`
    SnoozeReason string `json:"snooze_reason"`
    Version      string `json:"version"`
}

func getStatus() (*StatusResponse, error) {
    conn, err := net.Dial("unix", "/var/run/cloudsnooze.sock")
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    
    request := map[string]interface{}{
        "command": "STATUS",
        "params": map[string]interface{}{},
    }
    
    requestBytes, err := json.Marshal(request)
    if err != nil {
        return nil, err
    }
    
    _, err = conn.Write(requestBytes)
    if err != nil {
        return nil, err
    }
    
    buf := make([]byte, 4096)
    n, err := conn.Read(buf)
    if err != nil {
        return nil, err
    }
    
    var response StatusResponse
    err = json.Unmarshal(buf[:n], &response)
    if err != nil {
        return nil, err
    }
    
    return &response, nil
}

func main() {
    status, err := getStatus()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("CPU: %.2f%%, Memory: %.2f%%\n", 
        status.Metrics.CPUPercent, 
        status.Metrics.MemoryPercent)
    
    if status.ShouldSnooze {
        fmt.Printf("Instance should be snoozed: %s\n", status.SnoozeReason)
    } else {
        fmt.Printf("Instance is active: %s\n", status.SnoozeReason)
    }
}
```

### Tag-Based API in Python (AWS)

```python
import boto3
from datetime import datetime

def get_cloudsnooze_status(instance_id, tag_prefix='CloudSnooze'):
    """Get CloudSnooze status for an instance from its tags."""
    ec2 = boto3.client('ec2')
    
    response = ec2.describe_tags(
        Filters=[
            {'Name': 'resource-id', 'Values': [instance_id]},
            {'Name': 'key', 'Values': [f'{tag_prefix}:*']}
        ]
    )
    
    status = {}
    
    for tag in response.get('Tags', []):
        key = tag['Key']
        value = tag['Value']
        
        # Remove prefix from key
        if key.startswith(f'{tag_prefix}:'):
            clean_key = key[len(f'{tag_prefix}:'):]
            status[clean_key] = value
    
    return status

def instance_was_stopped_by_cloudsnooze(instance_id, tag_prefix='CloudSnooze'):
    """Check if instance was stopped by CloudSnooze."""
    status = get_cloudsnooze_status(instance_id, tag_prefix)
    
    return (
        status.get('Status') == 'Stopped' and
        'StopTimestamp' in status and
        'StopReason' in status
    )

def can_restart_instance(instance_id, service_id, tag_prefix='CloudSnooze'):
    """Check if this service is allowed to restart the instance."""
    status = get_cloudsnooze_status(instance_id, tag_prefix)
    
    # Check if restart is allowed at all
    if status.get('RestartAllowed') != 'true':
        return False
        
    # If no specific restarters are defined, any service can restart
    allowed_restarters = status.get('AllowedRestarters', '')
    if not allowed_restarters:
        return True
        
    # Check if our service ID is in the allowed list
    restarter_list = [r.strip() for r in allowed_restarters.split(',')]
    return service_id in restarter_list

def restart_cloudsnooze_instance(instance_id, service_id, reason='Manual restart', tag_prefix='CloudSnooze'):
    """Restart an instance that was stopped by CloudSnooze."""
    if not instance_was_stopped_by_cloudsnooze(instance_id, tag_prefix):
        return False, "Instance was not stopped by CloudSnooze"
        
    # Check if this service is allowed to restart the instance
    if not can_restart_instance(instance_id, service_id, tag_prefix):
        return False, "This service is not authorized to restart this instance"
    
    ec2 = boto3.client('ec2')
    
    # Start the instance
    try:
        ec2.start_instances(InstanceIds=[instance_id])
        
        # Update the status tag
        ec2.create_tags(
            Resources=[instance_id],
            Tags=[
                {'Key': f'{tag_prefix}:Status', 'Value': 'Running'},
                {'Key': f'{tag_prefix}:RestartTimestamp', 'Value': datetime.now().isoformat()},
                {'Key': f'{tag_prefix}:RestartReason', 'Value': reason},
                {'Key': f'{tag_prefix}:RestartedBy', 'Value': service_id},
                # Optional: Set expected usage duration if known
                # {'Key': f'{tag_prefix}:ExpectedUsageDuration', 'Value': '120'}
            ]
        )
        
        return True, "Instance restarted successfully"
    except Exception as e:
        return False, f"Error restarting instance: {str(e)}"
```

## Error Handling

### Socket API Errors

Socket API errors follow this format:

```json
{
  "error": true,
  "message": "Error message describing what went wrong",
  "code": 123
}
```

Common error codes:

- `100`: Invalid command format
- `101`: Unknown command
- `102`: Invalid parameters
- `200`: Internal error

### Tag API Error Handling

When working with tags, consider these error scenarios:

1. **Missing Tags**: Instance may not have CloudSnooze tags if:
   - It was never managed by CloudSnooze
   - Tags were deleted manually
   - Tagging failed due to permissions

2. **Inconsistent State**: Check for timestamp consistency between tags

## Versioning and Compatibility

The API version is tied to the CloudSnooze version:

- `v0.1.0`: Initial API implementation
  - Basic socket commands
  - Standard tag format

Future versions will maintain backward compatibility with existing tag formats and socket commands.