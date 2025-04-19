# CLAUDE.md - CloudSnooze Project

## Project Overview

CloudSnooze is a tool for automatically stopping idle cloud instances to save costs. The system monitors various system metrics (CPU, memory, network, disk I/O, user activity, GPU usage) and stops instances when they remain below specified thresholds for a defined period.

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

## Implementation Details

1. **Core Functionality**:
   - Monitoring system resources using Go's built-in libraries
   - GPU monitoring for NVIDIA, AMD, and Intel GPUs
   - Input activity detection (keyboard/mouse)
   - Cloud provider APIs for stopping instances
   - Unix socket communication between components
   - Permission verification at startup

2. **Design Principles**:
   - Minimize dependencies (no CLI dependency)
   - Cross-platform support (x86_64 and ARM64)
   - Low resource utilization
   - Single, clear method for stopping instances
   - Detailed logging and event recording

3. **Features to Implement**:
   - User input activity monitoring
   - Multiple GPU type detection and monitoring
   - Cloud provider API integration
   - Tagging of snoozed instances
   - Comprehensive logging
   - CLI and GUI interfaces
   - Cross-architecture package building

## AWS Components

For AWS integration, implement:
1. IAM permission verification
2. EC2 instance stopping via AWS SDK
3. Instance tagging
4. CloudWatch logging (optional)
5. Instance metadata service access

## Electron GUI

The GUI should include:
1. Real-time monitoring dashboard
2. Configuration interface
3. Historical snooze events
4. Visual representation of metrics
5. Threshold adjustment controls

## Future Extensions

These could be implemented after the core functionality:
1. Support for additional cloud providers
2. Advanced prediction of idle periods
3. Scheduled wake-up functionality
4. Cost savings reporting

## Development Approach

1. Implement the daemon core first
2. Add CLI functionality
3. Develop the GUI
4. Create packaging and distribution
5. Add cloud provider-specific enhancements

## Special Considerations

- AWS credentials and IAM roles need verification at startup
- GPU monitoring requires different approaches for each GPU type
- User input activity tracking requires X11 monitoring or similar
- Package building needs to consider cross-architecture compilation
