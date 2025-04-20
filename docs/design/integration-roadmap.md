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

1. **Slack Integration**
   - OAuth-based authentication
   - Rich message formatting with metrics displays
   - Interactive buttons for instance management
   - Channel configuration by event type

2. **Microsoft Teams Integration**
   - Teams connector support
   - Adaptive card templates for metrics visualization
   - Action buttons for immediate response
   - Tab integration for persistent dashboard

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

### Slack Notification Example

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
    Type   string      `json:"type"`
    Text   *SlackText  `json:"text,omitempty"`
    Fields []SlackText `json:"fields,omitempty"`
}

type SlackText struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

func SendSlackInstanceStoppedNotification(webhook string, metrics common.SystemMetrics, instanceID, reason string) error {
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
                    {Type: "mrkdwn", Text: fmt.Sprintf("*CPU Usage:*\n%.2f%%", metrics.CPUUsage)},
                    {Type: "mrkdwn", Text: fmt.Sprintf("*Memory Usage:*\n%.2f%%", metrics.MemoryUsage)},
                },
            },
            {
                Type: "section",
                Text: &SlackText{
                    Type: "mrkdwn",
                    Text: "Want to restart this instance? Use `snooze restart INSTANCE_ID` or click below:",
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

## Future Expansion Possibilities

- **ChatOps Integration**: Command CloudSnooze directly from chat platforms
- **Voice Assistant Integration**: Alexa and Google Assistant skills
- **Mobile Push Notifications**: Direct push to CloudSnooze mobile app
- **Custom Scripting**: Execute custom scripts on events
- **Incident Management**: Integration with PagerDuty, OpsGenie, etc.
- **Analytics Platform**: Datadog, New Relic, Grafana integrations