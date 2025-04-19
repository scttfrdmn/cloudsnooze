# CloudSnooze Integration Guide for External Tools

This document describes how external tools like provisioners, schedulers, and monitoring systems can integrate with CloudSnooze to detect and respond to instance stop events.

## Overview

CloudSnooze is a tool for automatically stopping idle cloud instances. External tools may need to:

1. Detect when CloudSnooze has stopped an instance
2. Understand why an instance was stopped
3. Gather metrics that led to the stop decision
4. Possibly restart instances when needed

## Integration Methods

### 1. Tag-Based Integration (Recommended)

CloudSnooze adds detailed tags to instances before stopping them. External tools can poll these tags to detect and analyze stop events.

#### Tag Structure

All CloudSnooze tags use the configured prefix (default: `CloudSnooze`):

| Tag Key | Description | Example Value |
|---------|-------------|---------------|
| `CloudSnooze:Status` | Current instance status | `Stopped` |
| `CloudSnooze:StopTimestamp` | When the instance was stopped | `2023-04-19T14:23:45Z` |
| `CloudSnooze:StopReason` | Why the instance was stopped | `System idle for 30 minutes (threshold: 30 minutes)` |

#### Detailed Metrics Tags (When Enabled)

| Tag Key | Description | Example Value |
|---------|-------------|---------------|
| `CloudSnooze:CPUPercent` | CPU usage at stop time | `2.50` |
| `CloudSnooze:MemoryPercent` | Memory usage at stop time | `15.20` |
| `CloudSnooze:NetworkKBps` | Network usage at stop time | `0.80` |
| `CloudSnooze:DiskIOKBps` | Disk I/O at stop time | `0.30` |
| `CloudSnooze:InputIdleSecs` | Input idle time in seconds | `1800` |
| `CloudSnooze:GPUPercent` | Average GPU usage (if present) | `0.05` |
| `CloudSnooze:GPUCount` | Number of GPUs detected | `2` |
| `CloudSnooze:NaptimeMinutes` | Configured idle time threshold | `30` |
| `CloudSnooze:InstanceType` | Instance type | `t3.medium` |
| `CloudSnooze:Region` | Instance region | `us-east-1` |

### 2. API Socket Communication (Advanced)

CloudSnooze exposes a Unix socket API that can be accessed by other applications running on the same instance.

#### Socket Location
By default: `/var/run/cloudsnooze.sock`

#### Available Commands
- `STATUS`: Get current system status and metrics
- `CONFIG_GET`: Retrieve current configuration
- `HISTORY`: Get historical snooze events

## Implementation Guide

### 1. Tag Polling Implementation

For external tools like provisioners, implement a tag polling mechanism with these steps:

```python
# Example in Python using boto3
import boto3
import time

ec2 = boto3.client('ec2')

def check_cloudsnooze_status(instance_id, tag_prefix='CloudSnooze'):
    response = ec2.describe_tags(
        Filters=[
            {'Name': 'resource-id', 'Values': [instance_id]},
            {'Name': 'key', 'Values': [f'{tag_prefix}:Status']}
        ]
    )
    
    if response['Tags']:
        status = response['Tags'][0]['Value']
        if status == 'Stopped':
            # Instance was stopped by CloudSnooze
            # Get additional tags for details
            all_tags = ec2.describe_tags(
                Filters=[
                    {'Name': 'resource-id', 'Values': [instance_id]},
                    {'Name': 'key', 'Values': [f'{tag_prefix}:*']}
                ]
            )
            
            # Convert tags to a dictionary for easier access
            tag_dict = {tag['Key']: tag['Value'] for tag in all_tags['Tags']}
            
            return {
                'status': 'stopped',
                'timestamp': tag_dict.get(f'{tag_prefix}:StopTimestamp'),
                'reason': tag_dict.get(f'{tag_prefix}:StopReason'),
                'metrics': {
                    'cpu': tag_dict.get(f'{tag_prefix}:CPUPercent'),
                    'memory': tag_dict.get(f'{tag_prefix}:MemoryPercent'),
                    # Add other metrics as needed
                }
            }
    
    return {'status': 'running'}

# Usage in a polling loop
def monitor_instances(instances, interval=60):
    while True:
        for instance_id in instances:
            status = check_cloudsnooze_status(instance_id)
            if status['status'] == 'stopped':
                print(f"Instance {instance_id} was stopped by CloudSnooze")
                print(f"Reason: {status['reason']}")
                print(f"Time: {status['timestamp']}")
                
                # Implement your response logic here:
                # - Log the event
                # - Update billing information
                # - Potentially restart the instance if needed
                
        time.sleep(interval)
```

### 2. Socket API Integration

For applications running on the same instance:

```go
// Example in Go
package main

import (
    "encoding/json"
    "net"
    "fmt"
)

func getCloudSnoozeStatus() (map[string]interface{}, error) {
    conn, err := net.Dial("unix", "/var/run/cloudsnooze.sock")
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    
    // Create command structure
    cmd := map[string]interface{}{
        "command": "STATUS",
        "params": map[string]interface{}{},
    }
    
    cmdBytes, err := json.Marshal(cmd)
    if err != nil {
        return nil, err
    }
    
    // Send command
    _, err = conn.Write(cmdBytes)
    if err != nil {
        return nil, err
    }
    
    // Read response
    buf := make([]byte, 4096)
    n, err := conn.Read(buf)
    if err != nil {
        return nil, err
    }
    
    // Parse response
    var response map[string]interface{}
    err = json.Unmarshal(buf[:n], &response)
    if err != nil {
        return nil, err
    }
    
    return response, nil
}
```

## Best Practices

1. **Poll at Appropriate Intervals**: 
   - The recommended default polling interval is 1 minute.
   - Adjust based on your requirements and instance count.

2. **Consider a Hierarchical Approach**:
   - For large fleets, use tag-based filtering to only check instances with CloudSnooze tags.
   - Implement exponential backoff for polling stopped instances.

3. **Handle Restart Carefully**:
   - When restarting instances, preserve the CloudSnooze tags but update the status.
   - Add your own tag to indicate the instance was restarted by your tool.

4. **Track Cost Savings**:
   - Use the stop timestamps to calculate exact savings.
   - The detailed metrics can help justify the stop decisions.

5. **Avoid API Socket for Cross-Instance Communication**:
   - The socket API is only intended for local communication.
   - For fleet-wide management, use tag-based polling.

## Security Considerations

1. **IAM Permissions**:
   - Ensure your external tool has at minimum the following permissions:
     - `ec2:DescribeTags`
     - `ec2:DescribeInstances`
   - For restart functionality, also add:
     - `ec2:StartInstances`

2. **Tag Modification**:
   - Only CloudSnooze should modify its own tags.
   - Your tool should read but not modify CloudSnooze tags.

3. **Socket API Permissions**:
   - The Unix socket is protected by file permissions.
   - Make sure your application has appropriate filesystem access.

## Troubleshooting

1. **Missing Tags**:
   - Verify CloudSnooze has `enable_instance_tags` set to `true`.
   - Check if `detailed_instance_tags` is enabled for metric details.

2. **Inconsistent Data**:
   - Check if CloudSnooze logs show successful tagging.
   - Verify the IAM role has permission to create and modify tags.

## Additional Resources

- CloudSnooze Configuration: See `config/snooze.json`
- API Documentation: See `docs/api/socket.md`

## Version Compatibility

This integration guide is compatible with CloudSnooze v0.1.0 and above.