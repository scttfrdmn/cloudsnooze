---
title: "CloudSnooze"
linkTitle: "Home"
---

<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->


<p align="center">
  <img src="/images/cloudsnooze_logo.png" alt="CloudSnooze Logo" width="300"/>
</p>

# CloudSnooze

CloudSnooze is a tool for automatically stopping idle cloud instances to save costs. The system monitors various system metrics (CPU, memory, network, disk I/O, user activity, GPU usage) and stops instances when they remain below specified thresholds for a defined period.

<p align="center">
  <a href="https://github.com/scttfrdmn/cloudsnooze/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue" alt="License"></a>
  <a href="https://github.com/scttfrdmn/cloudsnooze/releases"><img src="https://img.shields.io/badge/version-v0.1.0--alpha-orange" alt="Version"></a>
  <a href="https://go.dev/dl/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go Version"></a>
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey" alt="Platform">
  <img src="https://img.shields.io/badge/architecture-x86__64%20%7C%20ARM64-green" alt="Architecture">
  <img src="https://img.shields.io/badge/cloud-AWS%20%7C%20GCP%20%7C%20Azure-4285F4" alt="Cloud Support">
  <img src="https://img.shields.io/badge/status-alpha-red" alt="Status">
</p>

## How It Works

CloudSnooze monitors system resource usage and automatically stops instances when all metrics remain below specified thresholds for a defined period (the "naptime"). This saves costs by ensuring you only pay for compute resources when they're actually needed.

<p align="center">
  <img src="/images/cloudsnooze-workflow.svg" alt="CloudSnooze Workflow" width="700"/>
</p>

## Key Components

1. **Daemon (`snoozed`)**: A Go service that:
   - Monitors system resources
   - Verifies cloud provider permissions
   - Tags instances when stopping (optional)
   - Logs snooze events
   - Stops instances via cloud provider APIs

2. **CLI tool (`snooze`)**: A command-line interface for:
   - Viewing system status
   - Configuring thresholds and options
   - Managing the daemon
   - Viewing snooze history

3. **GUI application (`snooze-gui`)**: An Electron-based interface for:
   - Visual monitoring of system metrics
   - Configuration management
   - Historical data visualization

## Project Structure

```
cloudsnooze/
├── daemon/                  # Go daemon code
│   ├── main.go
│   ├── monitor/             # Monitoring modules
│   ├── accelerator/         # GPU monitoring
│   └── api/                 # Socket API
├── cli/                     # Go CLI code
│   ├── main.go
│   └── cmd/                 # CLI commands
├── ui/                      # Electron GUI
│   ├── main.js              # Electron main process
│   ├── index.html           # GUI interface
│   └── package.json         # Dependencies
├── packaging/               # Package building
│   ├── deb/                 # Debian packaging
│   └── rpm/                 # RPM packaging
├── man/                     # Man pages
├── systemd/                 # Systemd service files
├── config/                  # Default configurations
├── docs/                    # Documentation
├── .github/workflows/       # GitHub Actions
└── scripts/                 # Build scripts
```

## Features

- Multi-metric monitoring (CPU, memory, network, disk, GPU, user input)
- Support for multiple cloud providers (currently AWS)
- Detailed instance tagging
- Restart capability for external tools
- Low resource utilization
- Cross-platform support (x86_64 and ARM64)

## Documentation

For more detailed information, please check:

- [Design Documentation](/docs/design/)
- [Integration Guide](/docs/integration/)
- [Building and Packaging](/docs/building/)
- [Project Roadmap](/docs/roadmap)