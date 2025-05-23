<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Restart Logic for External Tools

This document provides detailed guidance on implementing restart logic for instances that have been stopped by CloudSnooze.

## New Restart Capability

CloudSnooze now supports explicit restart authorization for external tools through additional tags. When an instance is stopped, CloudSnooze can be configured to tag it with a `RestartAllowed` flag and optionally specify which external service IDs are allowed to perform restarts.

## Overview

While CloudSnooze focuses on stopping idle instances, external tools like provisioners may need to implement logic to restart these instances when:

1. Users need to access the instance
2. Scheduled jobs need to run
3. System maintenance requires the instance to be online

## Implementation Patterns

### 1. On-Demand Restart

The simplest pattern where instances are restarted when explicitly requested:

```
User/Service Request → Check if CloudSnooze-stopped → Verify restart authorization → Restart → Update Tags
```

#### Example Implementation (AWS)

```python
import boto3

def restart_cloudsnooze_instance(instance_id, service_id, tag_prefix='CloudSnooze'):
    ec2 = boto3.client('ec2')
    
    # Check if instance was stopped by CloudSnooze
    response = ec2.describe_tags(
        Filters=[
            {'Name': 'resource-id', 'Values': [instance_id]},
            {'Name': 'key', 'Values': [f'{tag_prefix}:Status']},
            {'Name': 'value', 'Values': ['Stopped']}
        ]
    )
    
    if not response['Tags']:
        return False, "Instance not stopped by CloudSnooze"
        
    # Check if restart is allowed
    restart_allowed_response = ec2.describe_tags(
        Filters=[
            {'Name': 'resource-id', 'Values': [instance_id]},
            {'Name': 'key', 'Values': [f'{tag_prefix}:RestartAllowed']},
            {'Name': 'value', 'Values': ['true']}
        ]
    )
    
    if not restart_allowed_response['Tags']:
        return False, "Restart not allowed for this instance"
        
    # Check if specific restarters are defined
    allowed_restarters_response = ec2.describe_tags(
        Filters=[
            {'Name': 'resource-id', 'Values': [instance_id]},
            {'Name': 'key', 'Values': [f'{tag_prefix}:AllowedRestarters']}
        ]
    )
    
    # If specific restarters are defined, check if this service is allowed
    if allowed_restarters_response['Tags']:
        allowed_restarters = allowed_restarters_response['Tags'][0]['Value'].split(',')
        if service_id not in allowed_restarters:
            return False, f"Service {service_id} not authorized to restart this instance"
    
    # Start the instance
    try:
        ec2.start_instances(InstanceIds=[instance_id])
        
        # Update tags
        ec2.create_tags(
            Resources=[instance_id],
            Tags=[
                {'Key': f'{tag_prefix}:Status', 'Value': 'Running'},
                {'Key': f'{tag_prefix}:RestartTimestamp', 'Value': datetime.now().isoformat()},
                {'Key': f'{tag_prefix}:RestartReason', 'Value': 'User requested restart'},
                {'Key': f'{tag_prefix}:RestartedBy', 'Value': service_id}
            ]
        )
        
        return True, "Instance restarted successfully"
    except Exception as e:
        return False, f"Error restarting instance: {str(e)}"
```

### 2. Scheduled Restart

For instances that need to run scheduled jobs:

```
Scheduled Event → Find Matching Stopped Instances → Restart → Run Job → Allow to Stop Again
```

#### Example Implementation

```python
def schedule_instance_restart(schedule_expression, instance_tags, tag_prefix='CloudSnooze'):
    ec2 = boto3.client('ec2')
    
    # Find instances matching tags that were stopped by CloudSnooze
    response = ec2.describe_instances(
        Filters=[
            {'Name': 'tag:YourScheduleTag', 'Values': [schedule_expression]},
            {'Name': f'tag:{tag_prefix}:Status', 'Values': ['Stopped']},
            {'Name': 'instance-state-name', 'Values': ['stopped']}
        ]
    )
    
    restarted_instances = []
    for reservation in response['Reservations']:
        for instance in reservation['Instances']:
            instance_id = instance['InstanceId']
            
            # Restart the instance
            ec2.start_instances(InstanceIds=[instance_id])
            
            # Update tags
            ec2.create_tags(
                Resources=[instance_id],
                Tags=[
                    {'Key': f'{tag_prefix}:Status', 'Value': 'Running'},
                    {'Key': f'{tag_prefix}:RestartTimestamp', 'Value': datetime.now().isoformat()},
                    {'Key': f'{tag_prefix}:RestartReason', 'Value': f'Scheduled event: {schedule_expression}'}
                ]
            )
            
            restarted_instances.append(instance_id)
    
    return restarted_instances
```

### 3. Predictive Restart

A more sophisticated approach that predicts when users will need instances:

```
User Activity Data → Predict Usage Pattern → Preemptively Restart → Update Tags
```

#### Factors to Consider

- Historical usage patterns
- Time of day/week
- User login activity from other services
- Calendar/meeting data

## Coordination with CloudSnooze

To ensure proper coordination with CloudSnooze, external tools should:

1. **Set Expected Usage Tag**: When restarting an instance, set a tag indicating when the instance might become idle again.

2. **Modify Status Tag**: Set `CloudSnooze:Status` to `Running` when restarting.

3. **Add Context Tags**: Include information about why and when the instance was restarted.

### Tag Schema for Restarts

### Tags Set by CloudSnooze

| Tag | Description | Example |
|-----|-------------|----------|
| `CloudSnooze:Status` | Current status | `Stopped` |
| `CloudSnooze:StopTimestamp` | When the instance was stopped | `2023-04-19T15:30:00Z` |
| `CloudSnooze:StopReason` | Why the instance was stopped | `System idle for 30 minutes` |
| `CloudSnooze:RestartAllowed` | Whether external tools can restart | `true` |
| `CloudSnooze:AllowedRestarters` | Comma-separated list of service IDs allowed to restart | `UserPortal,JobScheduler` |

### Tags Set by External Tools

| Tag | Description | Example |
|-----|-------------|----------|
| `CloudSnooze:Status` | Current status (updated) | `Running` |
| `CloudSnooze:RestartTimestamp` | When the instance was restarted | `2023-04-19T15:30:00Z` |
| `CloudSnooze:RestartReason` | Why the instance was restarted | `User login` or `Scheduled job` |
| `CloudSnooze:ExpectedUsageDuration` | How long the instance is expected to be needed | `120` (minutes) |
| `CloudSnooze:RestartedBy` | Service that restarted the instance | `UserPortal` or `JobScheduler` |

## State Machine

The complete lifecycle of an instance with CloudSnooze and an external restart tool:

```
Running → Idle → Stopped by CloudSnooze → Restarted by External Tool → Running → ...
```

### State Transitions

1. **Running to Idle**: CloudSnooze detects inactivity below thresholds
2. **Idle to Stopped**: CloudSnooze stops the instance after naptime
3. **Stopped to Restarting**: External tool initiates restart
4. **Restarting to Running**: Instance becomes available
5. **Running to Monitored**: CloudSnooze resumes monitoring

## Best Practices

1. **Respect Authorization Boundaries**:
   - Only attempt to restart instances where `RestartAllowed` is set to `true`
   - Verify your service ID is in the `AllowedRestarters` list if specified
   - Log authorization failures for security auditing

2. **Respect Idle Detection**:
   - Don't disable CloudSnooze when restarting instances
   - Allow the natural idle detection to work

3. **Throttle Restarts**:
   - Implement cooldown periods to prevent rapid stop/start cycles
   - Consider minimum runtime enforcements

4. **Track Effectiveness**:
   - Log when an instance is restarted
   - Track how long it remains active
   - Analyze if the restart was necessary

5. **User Communication**:
   - Inform users when an instance is restarted
   - Provide context about when it might be stopped again

6. **Optimize Cold Start**:
   - For instances that take time to become fully useful after restart
   - Consider warming caches or preloading data

## Example Architecture

For a complete solution, consider:

1. **Central Management Service**:
   - Maintains state of all CloudSnooze-managed instances
   - Coordinates restart operations

2. **User Portal Integration**:
   - Allows users to see stopped instances
   - Provides one-click restart capability

3. **Scheduler Integration**:
   - Ensures instances are running for scheduled jobs
   - Allows jobs to complete before instances idle out

4. **Monitoring Integration**:
   - Tracks stop/restart patterns
   - Identifies opportunities for optimization

## Performance Considerations

1. **Cold Start Time**:
   - Account for the time needed for instances to fully restart
   - For time-sensitive operations, restart in advance

2. **Resource Bursting**:
   - Be aware that restarting many instances simultaneously can cause resource contention
   - Consider staggered restarts for large fleets

3. **Cost Implications**:
   - Balance between the cost savings of stopping and the overhead of restarting
   - Some instances may be better left running if restart frequency is high