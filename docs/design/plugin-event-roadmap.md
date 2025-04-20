<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Event-Driven Plugin Framework Roadmap

This document outlines the implementation plan for the CloudSnooze Event-Driven Plugin Framework, which enables plugins to respond to external events such as AWS Spot instance interruption notifications.

## Overview

The Event-Driven Plugin Framework extends CloudSnooze beyond passive monitoring to active event handling. This allows CloudSnooze to trigger appropriate actions when cloud providers or system events require rapid response.

## Implementation Phases

### Phase 1: Core Event Framework (Q3 2025)

1. **Event System Architecture**
   - Design event types, priorities, and data structures
   - Implement event subscription and dispatch mechanisms
   - Create event queue and processing pipeline

2. **Event Source Integration**
   - AWS EC2 metadata service monitoring for Spot interrupts
   - System event sources (disk, memory, network)
   - Basic event simulation for testing

3. **Plugin Interface Extensions**
   - Extend plugin interface for event handling
   - Implement event handler registration and discovery
   - Develop event handler prioritization system

4. **Testing Framework**
   - Create event simulation tools for testing
   - Implement event processing unit tests
   - Develop plugin event handling verification system

### Phase 2: AWS Spot Instance Handler (Q4 2025)

1. **Spot Interrupt Detection**
   - EC2 instance metadata service monitoring
   - Instance termination notice parsing
   - Time-to-termination calculation

2. **Graceful Shutdown Framework**
   - Application notification system
   - Configurable shutdown sequence
   - Shutdown verification and reporting

3. **Official Plugins**
   - Database flushing plugins (MySQL, PostgreSQL, Redis)
   - Application checkpoint plugins (common frameworks)
   - VM state saving plugins

4. **Documentation and Examples**
   - Comprehensive developer guide
   - Plugin implementation examples
   - Best practices for shutdown scripting

### Phase 3: Additional Cloud Providers (Q1 2026)

1. **Azure Integration**
   - Azure Scheduled Events API monitoring
   - VM termination event handling
   - Azure-specific plugins

2. **GCP Integration**
   - Preemptible VM monitoring
   - GCE instance termination events
   - GCP-specific plugins

3. **Cross-Provider Abstraction**
   - Unified event model across providers
   - Provider-agnostic plugin development
   - Event translation and normalization

### Phase 4: Advanced Features (Q2-Q3 2026)

1. **Event Analytics**
   - Historical event tracking
   - Pattern recognition for preemptions
   - Predictive event analysis

2. **Multi-Instance Coordination**
   - Event propagation across instances
   - Coordinated shutdown for distributed systems
   - Leader election for orchestrated shutdowns

3. **Cost Optimization Strategies**
   - Automated bidding strategies for Spot instances
   - Instance type recommendations based on interruption patterns
   - Workload scheduling around interruption patterns

## Specific Implementation Milestones

### AWS Spot Instance Handler

1. **Metadata Service Monitoring**
   - Poll EC2 metadata service at `/latest/meta-data/spot/instance-action`
   - Parse JSON response for action and time
   - Calculate remaining time before termination

2. **Plugin Execution Flow**
   ```
   1. CloudSnooze Daemon detects Spot termination notice
   2. Event Manager creates Spot interrupt event
   3. Plugin Manager finds plugins registered for this event
   4. Plugins execute in priority order:
      a. High priority: Critical data saving
      b. Medium priority: Application state checkpointing
      c. Low priority: Notification and logging
   5. Results are collected and logged
   6. Instance prepares for termination
   ```

3. **Configuration Example**
   ```json
   {
     "event_handlers": {
       "cloud.spot.interrupt": {
         "enabled": true,
         "polling_interval_seconds": 5,
         "plugins": [
           {
             "name": "mysql-flusher",
             "priority": 100,
             "timeout_seconds": 30,
             "config": {
               "connection_string": "user:pass@tcp(localhost:3306)/db"
             }
           },
           {
             "name": "app-checkpointer",
             "priority": 80,
             "timeout_seconds": 45,
             "config": {
               "app_endpoint": "http://localhost:8080/checkpoint",
               "verify_success": true
             }
           },
           {
             "name": "slack-notifier",
             "priority": 10,
             "config": {
               "webhook_url": "https://hooks.slack.com/services/...",
               "channel": "#cloud-events"
             }
           }
         ]
       }
     }
   }
   ```

### Implementing Cloud Event Sources

For AWS Spot instances, the implementation will:

1. Create a background goroutine that periodically polls:
   ```go
   func pollSpotInterruptNotices(ctx context.Context, eventCh chan<- Event) {
       ticker := time.NewTicker(5 * time.Second)
       defer ticker.Stop()
       
       for {
           select {
           case <-ctx.Done():
               return
           case <-ticker.C:
               resp, err := http.Get("http://169.254.169.254/latest/meta-data/spot/instance-action")
               if err != nil || resp.StatusCode != 200 {
                   // No notice yet
                   continue
               }
               
               defer resp.Body.Close()
               var notice SpotInterruptNotice
               if err := json.NewDecoder(resp.Body).Decode(&notice); err != nil {
                   log.Printf("Error decoding spot interrupt notice: %v", err)
                   continue
               }
               
               // Create and dispatch event
               event := Event{
                   Type:      EventSpotInterrupt,
                   Source:    "aws.ec2.metadata-service",
                   Timestamp: time.Now(),
                   ID:        uuid.New().String(),
                   Severity:  SeverityUrgent,
                   Data: map[string]interface{}{
                       "action":         notice.Action,
                       "time":           notice.Time,
                       "timeRemaining":  time.Until(notice.Time).Seconds(),
                       "instanceID":     getInstanceID(),
                       "instanceType":   getInstanceType(),
                   },
               }
               
               eventCh <- event
           }
       }
   }
   ```

2. Register this poller with the Event Manager:
   ```go
   func (em *EventManager) Start(ctx context.Context) error {
       em.eventCh = make(chan Event, 100)
       
       // Start event processing goroutine
       go em.processEvents(ctx)
       
       // Start event sources
       if em.config.AWS.SpotInterruptMonitoring.Enabled {
           go pollSpotInterruptNotices(ctx, em.eventCh)
       }
       
       // Add other event sources...
       
       return nil
   }
   ```

## Third-Party Integration Examples

### 1. Database Flushing on Spot Termination

```go
// Example plugin for MySQL database flushing
func (p *MySQLFlusher) HandleEvent(event plugin.Event) (bool, map[string]interface{}, error) {
    if event.Type != plugin.EventSpotInterrupt {
        return false, nil, nil
    }
    
    // Connect to database
    db, err := sql.Open("mysql", p.connectionString)
    if err != nil {
        return false, nil, fmt.Errorf("failed to connect to MySQL: %w", err)
    }
    defer db.Close()
    
    // Flush tables with read lock
    _, err = db.Exec("FLUSH TABLES WITH READ LOCK")
    if err != nil {
        return false, nil, fmt.Errorf("failed to flush tables: %w", err)
    }
    
    // Unlock tables after a brief pause to ensure flush completed
    defer func() {
        _, _ = db.Exec("UNLOCK TABLES")
    }()
    
    return true, map[string]interface{}{
        "status": "mysql_flushed",
        "tables_flushed": true,
    }, nil
}
```

### 2. Application Checkpoint Trigger

```python
# Example external plugin for application checkpoint
class AppCheckpointer:
    def __init__(self):
        self.app_endpoint = "http://localhost:8080/api/checkpoint"
        self.timeout_seconds = 30
        
    def handle_event(self, event):
        if event.get("type") != "cloud.spot.interrupt":
            return {"handled": False}
            
        try:
            # Call application checkpoint API
            response = requests.post(
                self.app_endpoint,
                json={
                    "reason": "spot_termination",
                    "time_remaining": event["data"]["timeRemaining"]
                },
                timeout=self.timeout_seconds
            )
            
            if response.status_code == 200:
                return {
                    "handled": True,
                    "status": "checkpoint_triggered",
                    "app_response": response.json()
                }
            else:
                return {
                    "handled": False,
                    "error": f"Checkpoint API returned status {response.status_code}"
                }
                
        except Exception as e:
            return {
                "handled": False,
                "error": str(e)
            }
```

## Success Metrics and Goals

1. **Reliability Metrics**
   - 99.9% successful handling of Spot termination events
   - <5 second detection-to-action latency
   - <1% plugin execution failures

2. **Community Engagement**
   - 10+ event handler plugins available at launch
   - 5+ third-party plugins within 3 months
   - Developer documentation with >90% satisfaction

3. **Business Impact**
   - 30%+ cost savings through safe Spot instance usage
   - <1% data loss incidents on Spot terminations
   - 50%+ reduction in manual intervention for spot terminations