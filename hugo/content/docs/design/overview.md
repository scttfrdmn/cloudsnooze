# CloudSnooze - Project Overview

## Introduction

CloudSnooze is a lightweight, efficient solution for automatically stopping idle cloud instances to save costs. It monitors various system metrics and, when all metrics remain below specified thresholds for a defined period, safely stops the instance.

## Key Features

- **Low-resource monitoring**: Written in Go for minimal impact on the monitored instance
- **Multi-metric detection**: Considers CPU, memory, network, disk I/O, user input, and GPU activity
- **User activity awareness**: Detects actual keyboard and mouse usage, not just login sessions
- **Cloud provider agnostic**: Initially supporting AWS, with design for future expansion
- **Cross-architecture support**: Works on both x86_64 and ARM64 instances
- **Multiple interfaces**: CLI tool, GUI application, and daemon
- **Tagging and logging**: Records when and why instances were stopped
- **Permission verification**: Validates cloud provider permissions at startup

## Architecture

CloudSnooze consists of three main components:

1. **Daemon (`snoozed`)**: A background service that continuously monitors system resources and initiates instance stopping when appropriate.

2. **CLI Tool (`snooze`)**: A command-line interface for configuration, monitoring, and control.

3. **GUI Application (`snooze-gui`)**: An Electron-based graphical interface for visual monitoring and configuration.

The components communicate via a Unix socket, allowing for secure local interaction:

```
┌────────────────┐     ┌────────────────┐     ┌────────────────┐
│                │     │                │     │                │
│  snoozed       │◄───►│  snooze        │     │  snooze-gui    │
│  (Daemon)      │     │  (CLI)         │     │  (Electron)    │
│                │     │                │     │                │
└────────┬───────┘     └────────────────┘     └────────┬───────┘
         │                                             │
         │              Unix Socket                    │
         └─────────────────────────────────────────────┘
                              │
                              ▼
                   ┌────────────────────┐
                   │                    │
                   │  Configuration     │
                   │  (/etc/snooze)     │
                   │                    │
                   └────────────────────┘
```

## Workflow

1. The daemon starts and verifies it has the necessary cloud provider permissions
2. System metrics are monitored at regular intervals (configurable)
3. When all metrics remain below their thresholds for the "naptime" duration:
   - The instance is tagged (optional)
   - The event is logged
   - The instance is stopped via the cloud provider API

## Terminology

CloudSnooze uses playful terminology that aligns with its name:

- **Naptime**: The duration an instance must be idle before it's stopped
- **Snooze Thresholds**: Resource usage levels below which an instance is considered idle
- **Wake Triggers**: Conditions that would prevent an instance from being snoozed
- **Snooze History**: Record of when instances were previously stopped

## Implementation Details

- **Language**: Go for the daemon and CLI, JavaScript/Electron for the GUI
- **Configuration**: JSON format stored in `/etc/snooze/snooze.json`
- **Logs**: Multiple options including file, syslog, and CloudWatch
- **Packaging**: RPM and DEB packages for easy installation
- **Systemd Integration**: Automatic startup and reliable service management
- **Cross-compilation**: Supports both x86_64 and ARM64 architectures

## Cloud Provider Requirements

For AWS (initial support):
- IAM role with permissions to stop the instance and apply tags
- Network access to the AWS EC2 API endpoints
- Instance metadata service accessibility

## Future Expansion

- Support for additional cloud providers (Azure, GCP)
- Wake-up scheduling for predictable instance availability
- Integration with cloud provider billing APIs for cost tracking
- Advanced prediction of idle periods based on historical patterns
