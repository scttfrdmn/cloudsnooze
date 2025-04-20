<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Integration Roadmap

This document outlines the plan for integrating CloudSnooze with external communication platforms and automation services, enhancing the project's flexibility and utility.

## Overview

CloudSnooze provides powerful cloud instance management, but its effectiveness is enhanced when integrated with external services for notifications, automation, and data flow. This roadmap details how CloudSnooze will expand its integration capabilities to support platforms like Slack, Microsoft Teams, and automation hubs like Zapier, Make.com (formerly Integromat), and n8n.

## Integration Types

The integration strategy involves three primary methods:

1. **Webhook-Based Integration**: Send and receive data via HTTP webhooks
2. **Direct API Integration**: Purpose-built connectors for specific services
3. **Integration Hub Support**: Ready-made templates for automation platforms

## Implementation Phases

### Phase 1: Core Webhook Architecture (Q3 2025)

1. **Webhook Emission Framework**
   - Design a flexible event-based webhook system
   - Implement configurable HTTP POST requests for system events
   - Create payload templates for different event types
   - Develop security features (HMAC signing, API keys)

2. **Notification Events**
   - Instance start/stop events
   - Threshold crossing alerts
   - Error and warning notifications
   - Cost-saving summaries

3. **Configuration System**
   - Add webhook configuration to main config
   - Implement webhook management via CLI
   - Add webhook testing capabilities

```json
// Example webhook configuration
{
  "webhooks": {
    "instance_stopped": {
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "payload_template": "slack_instance_stopped",
      "enabled": true,
      "retry_count": 3,
      "include_metrics": true
    }
  }
}
```

### Phase 2: Direct Platform Integrations (Q4 2025)

1. **Slack Integration with Interactive Controls**
   - OAuth-based authentication
   - Rich message formatting with metrics displays
   - **Interactive components for two-way communication:**
     - Action buttons for direct instance control (start/stop/restart)
     - Custom commands via Slack Bot (`/snooze stop i-1234567890abcdef0`)
     - Interactive dialogs for configuration changes
     - Approval workflows for critical actions
   - Persistent context for conversation threads
   - Channel configuration by event type
   - User permission mapping between Slack and CloudSnooze

2. **Microsoft Teams Integration with Interactive Features**
   - Teams connector support with bot functionality
   - Adaptive card templates for metrics visualization
   - **Interactive components for two-way control:**
     - Action buttons for direct instance management
     - Command support within Teams chat (`@CloudSnooze restart i-1234567890abcdef0`)
     - Interactive forms for configuration changes
     - Actionable notifications with approval workflows
   - Tab integration for persistent dashboard
   - Integration with Teams permissions model
   - Message threading for context preservation

3. **Email Notification System**
   - SMTP-based notifications
   - HTML email templates with metrics
   - Customizable alert thresholds
   - Digest options (daily/weekly summaries)

### Phase 3: Integration Hub Support (Q1 2026)

1. **Zapier Integration**
   - Create official Zapier app
   - Define triggers (instance events, threshold alerts)
   - Implement actions (instance control, configuration)
   - Provide pre-built Zap templates

2. **Make.com (Integromat) Support**
   - Develop Make.com app module
   - Custom scenario templates
   - Data transformation helpers
   - Scheduling and conditional flows

3. **n8n Compatibility**
   - Node development for n8n
   - Credential handling
   - Workflow templates for common use cases
   - Self-hosted integration guidance

### Phase 4: Advanced Integration Features (Q2 2026)

1. **Bidirectional Control API**
   - Remote instance management
   - Configuration updates via API
   - Secure authentication system
   - Rate limiting and access control

2. **Custom Integration Framework**
   - Plugin system for third-party integrations
   - Integration SDK with documentation
   - Testing framework for custom integrations
   - Marketplace for community integrations

3. **Data Export Options**
   - CSV/JSON data export of metrics
   - Integration with monitoring platforms
   - Cost analysis data for BI tools
   - Historical data warehousing options

## Specific Implementation Examples

### Slack Interactive Integration Example

```go
// slack.go
package integrations

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/scttfrdmn/cloudsnooze/daemon/common"
)

type SlackMessage struct {
    Text        string        `json:"text,omitempty"`
    Blocks      []SlackBlock  `json:"blocks,omitempty"`
    Attachments []interface{} `json:"attachments,omitempty"`
}

type SlackBlock struct {
    Type      string        `json:"type"`
    Text      *SlackText    `json:"text,omitempty"`
    Fields    []SlackText   `json:"fields,omitempty"`
    Elements  []interface{} `json:"elements,omitempty"`
    BlockID   string        `json:"block_id,omitempty"`
}

type SlackText struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type SlackButton struct {
    Type     string    `json:"type"`
    Text     SlackText `json:"text"`
    ActionID string    `json:"action_id"`
    Value    string    `json:"value"`
    Style    string    `json:"style,omitempty"` // primary, danger
}

// SendSlackInstanceStoppedNotification sends an interactive notification to Slack
// with buttons that allow direct control of the instance
func SendSlackInstanceStoppedNotification(webhook string, metrics common.SystemMetrics, instanceID, region, reason string) error {
    message := SlackMessage{
        Blocks: []SlackBlock{
            {
                Type: "header",
                Text: &SlackText{
                    Type: "plain_text",
                    Text: "CloudSnooze: Instance Stopped 💤",
                },
            },
            {
                Type: "section",
                Text: &SlackText{
                    Type: "mrkdwn",
                    Text: fmt.Sprintf("Instance `%s` has been stopped due to inactivity.", instanceID),
                },
            },
            {
                Type: "section",
                Fields: []SlackText{
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Stop Reason:*\n%s", reason)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Stop Time:*\n%s", time.Now().Format(time.RFC1123))},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Region:*\n%s", region)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Instance ID:*\n%s", instanceID)},
                },
            },
            {
                Type: "section",
                Fields: []SlackText{
                    {Type: "mrkdwn", Text: fmt.Sprintf("*CPU Usage:*\n%.2f%%", metrics.CPUUsage)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Memory Usage:*\n%.2f%%", metrics.MemoryUsage)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Network:*\n%.2f KB/s", metrics.NetworkRate)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Disk I/O:*\n%.2f KB/s", metrics.DiskIORate)},
                },
            },
            {
                Type: "actions",
                BlockID: "instance_actions_" + instanceID,
                Elements: []interface{}{
                    SlackButton{
                        Type: "button",
                        Text: SlackText{
                            Type: "plain_text",
                            Text: "Start Instance",
                        },
                        ActionID: "start_instance",
                        Value:    instanceID + "|" + region,
                        Style:    "primary",
                    },
                    SlackButton{
                        Type: "button",
                        Text: SlackText{
                            Type: "plain_text",
                            Text: "View Details",
                        },
                        ActionID: "view_details",
                        Value:    instanceID + "|" + region,
                    },
                    SlackButton{
                        Type: "button",
                        Text: SlackText{
                            Type: "plain_text",
                            Text: "Change Thresholds",
                        },
                        ActionID: "change_thresholds",
                        Value:    instanceID + "|" + region,
                    },
                },
            },
            {
                Type: "context",
                Elements: []interface{}{
                    SlackText{
                        Type: "mrkdwn",
                        Text: fmt.Sprintf("You can also use Slack commands like `/snooze start %s` or `/snooze status`", instanceID),
                    },
                },
            },
        },
    }
    
    payload, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("error marshaling Slack message: %w", err)
    }
    
    resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return fmt.Errorf("error sending Slack notification: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("non-200 response from Slack: %d", resp.StatusCode)
    }
    
    return nil
}

// HandleSlackInteraction processes interactive callbacks from Slack
func HandleSlackInteraction(payload []byte) (string, error) {
    // Parse the interaction payload
    var interaction struct {
        Type string `json:"type"`
        Actions []struct {
            ActionID string `json:"action_id"`
            Value    string `json:"value"`
        } `json:"actions"`
        User struct {
            ID string `json:"id"`
            Username string `json:"username"`
        } `json:"user"`
    }
    
    if err := json.Unmarshal(payload, &interaction); err != nil {
        return "", fmt.Errorf("error parsing interaction: %w", err)
    }
    
    // Only process button actions for now
    if interaction.Type != "block_actions" || len(interaction.Actions) == 0 {
        return "Unsupported interaction type", nil
    }
    
    action := interaction.Actions[0]
    
    // Parse the value which contains instanceID and region
    parts := strings.Split(action.Value, "|")
    if len(parts) != 2 {
        return "Invalid action value format", nil
    }
    
    instanceID := parts[0]
    region := parts[1]
    
    switch action.ActionID {
    case "start_instance":
        // Call AWS API to start the instance
        err := startInstance(region, instanceID)
        if err != nil {
            return fmt.Sprintf("Failed to start instance %s: %v", instanceID, err), nil
        }
        return fmt.Sprintf("Instance %s is starting. This may take a few minutes.", instanceID), nil
        
    case "view_details":
        // Fetch and return detailed information about the instance
        details, err := getInstanceDetails(region, instanceID)
        if err != nil {
            return fmt.Sprintf("Failed to get details for instance %s: %v", instanceID, err), nil
        }
        return fmt.Sprintf("*Instance Details*\n%s", details), nil
        
    case "change_thresholds":
        // This would typically open a dialog, but for simplicity we'll just acknowledge
        return "Opening threshold configuration dialog...", nil
        
    default:
        return "Unsupported action", nil
    }
}
```

### Zapier Integration Example

Zapier Trigger: `instance_stopped`

Trigger payload:
```json
{
  "event_type": "instance_stopped",
  "instance_id": "i-1234567890abcdef0",
  "timestamp": "2025-06-15T14:23:45Z",
  "region": "us-east-1",
  "metrics": {
    "cpu_usage": 1.2,
    "memory_usage": 3.5,
    "network_traffic": 0.1,
    "disk_io": 0.02
  },
  "reason": "System idle for 30 minutes",
  "cost_saved_estimate": 0.125,
  "instance_type": "t3.medium"
}
```

### CLI Command Additions

```
# New CLI integration commands
snooze integrations list                             # List all configured integrations
snooze integrations add slack WEBHOOK_URL            # Add Slack integration
snooze integrations add teams WEBHOOK_URL            # Add Microsoft Teams integration
snooze integrations add email [options]              # Add email notifications
snooze integrations test slack                       # Test Slack integration
snooze integrations remove slack                     # Remove Slack integration
snooze integrations export zapier                    # Generate Zapier integration files
```

## User Benefits

1. **Enhanced Visibility**
   - Real-time notifications when instances are stopped
   - Alerts when thresholds are approaching
   - Cost-saving reports and insights

2. **Remote Management**
   - Control instances through familiar interfaces
   - Automate responses to specific conditions
   - Approval workflows for critical instances

3. **Workflow Automation**
   - Connect instance management to other systems
   - Trigger workflows based on cloud events
   - Chain actions across multiple services

4. **Organizational Awareness**
   - Keep teams informed of cost-saving activities
   - Provide management with cost reduction metrics
   - Alert developers when their instances are affected

## Architecture Diagram

```
                                 ┌─────────────────┐
                                 │                 │
                                 │   CloudSnooze   │
                                 │                 │
                                 └─────┬───────────┘
                                       │
                                       │ Events
                                       │
                              ┌────────▼─────────┐
                              │                  │
                              │ Integration Hub  │
                              │                  │
                              └──┬──────┬────┬───┘
                                 │      │    │
                 ┌───────────────┘      │    └──────────────┐
                 │                      │                   │
        ┌────────▼─────────┐   ┌────────▼─────────┐ ┌───────▼────────┐
        │                  │   │                  │ │                │
        │  Direct Platform │   │   Webhooks      │ │  Automation    │
        │  Integrations    │   │                  │ │  Platforms     │
        │                  │   │                  │ │                │
        └──┬──────┬────────┘   └───────┬──────────┘ └──┬────────┬───┘
           │      │                    │               │        │
     ┌─────▼─┐ ┌──▼────┐        ┌─────▼─────┐     ┌───▼───┐ ┌──▼────┐
     │       │ │       │        │           │     │       │ │       │
     │ Slack │ │ Teams │        │  Custom   │     │ Zapier│ │Make.com│
     │       │ │       │        │  Webhooks │     │       │ │       │
     └───────┘ └───────┘        └───────────┘     └───────┘ └───────┘
```

## Success Metrics and KPIs

- **Integration Adoption Rate**: 50% of users configuring at least one integration
- **Notification Delivery**: 99.9% successful delivery of notifications
- **Integration Platform Coverage**: Support for 5+ major platforms
- **User Satisfaction**: 90%+ approval rating for integration features
- **Time to Value**: <10 minutes to set up first integration

## Timeline Summary

- **Q3 2025**: Core webhook architecture (10 weeks)
- **Q4 2025**: Direct platform integrations (12 weeks)
- **Q1 2026**: Integration hub support (10 weeks)
- **Q2 2026**: Advanced integration features (12 weeks)

## Resource Requirements

- **Development**: 2 developers (part-time)
- **Design**: 1 designer for UI components
- **Documentation**: Technical writer for integration guides
- **Testing**: QA engineer for cross-platform testing
- **External Services**: Development accounts for Slack, Zapier, etc.

## ChatOps Integration (Two-Way Communication)

CloudSnooze will implement a comprehensive ChatOps approach, allowing users to not only receive notifications but also control instances and modify configurations directly from chat platforms. This creates a seamless operational workflow where monitoring, alerts, and actions all happen in the same conversation context.

### Key ChatOps Capabilities

1. **Conversational Commands**
   - Natural language processing for simple commands
   - Structured commands for complex operations
   - Context-aware responses based on previous messages
   - Command suggestions and auto-completion

2. **Interactive Controls**
   - One-click buttons for common actions (start/stop/restart)
   - Drop-down selectors for choosing instances or regions
   - Form-based inputs for threshold configuration
   - Progress indicators for long-running operations

3. **Conversation Flow Examples**

```
User: @CloudSnooze list idle instances
Bot:  Found 3 instances that are currently idle:
      • i-1234abcd (us-east-1): CPU 2%, Memory 15%, idle for 45 minutes
      • i-5678efgh (us-west-2): CPU 1%, Memory 12%, idle for 30 minutes
      • i-9012ijkl (eu-west-1): CPU 3%, Memory 18%, idle for 25 minutes
      Would you like to stop any of these instances?

User: stop the first one
Bot:  I'll stop instance i-1234abcd in us-east-1.
      [Stop Instance] [View Details] [Cancel]

User: [clicks Stop Instance]
Bot:  ✅ Instance i-1234abcd has been stopped successfully.
      Estimated cost savings: $0.12/hour
```

4. **Multi-User Collaboration**
   - Shared visibility of instance state and actions
   - Role-based access control tied to chat platform roles
   - Action audit trail within the conversation
   - Approval workflows for critical actions

### Implementation Architecture

The ChatOps integration will use a stateful bot implementation that maintains conversation context and understands user intent across message exchanges. This allows for natural, flowing interactions rather than isolated command/response pairs.

## Future Expansion Possibilities

- **Voice Assistant Integration**: Alexa and Google Assistant skills
- **Mobile Push Notifications**: Direct push to CloudSnooze mobile app
- **Custom Scripting**: Execute custom scripts on events
- **Incident Management**: Integration with PagerDuty, OpsGenie, etc.
- **Analytics Platform**: Datadog, New Relic, Grafana integrations