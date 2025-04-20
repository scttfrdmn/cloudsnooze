<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Development Roadmap

This document outlines the planned development roadmap for CloudSnooze. The roadmap is organized into phases, with each phase building upon the previous one.

## Completed Features

- ✅ Core system monitoring (CPU, memory, network, disk I/O)
- ✅ Configuration structure and loading
- ✅ Basic daemon architecture
- ✅ Socket API infrastructure
- ✅ AWS instance metadata integration
- ✅ Input activity monitoring (Linux, macOS)
- ✅ GPU monitoring (NVIDIA, AMD)
- ✅ Tag-based monitoring for external tools
- ✅ Documentation for external tool integration

## Next Steps

### Phase 1: Core Functionality Completion

1. **Implement Pluggable Cloud Providers**
   - Develop a flexible provider architecture for extensibility
   - Create a standardized plugin interface for cloud providers
   - Implement dynamic loading of provider plugins
   - Add configuration options for provider selection
   - Document the provider API for third-party implementations
   - See [Cloud Provider Architecture](design/cloud-provider-architecture.md) for details

2. **Implement Unit Testing Framework**
   - Set up testing infrastructure for Go components
   - Add tests for core monitoring modules
   - Implement integration tests for the system as a whole
   - Create CI/CD pipelines for automated testing

3. **Complete AWS SDK Integration**
   - Replace placeholder code with actual AWS SDK calls
   - Implement instance stopping functionality
   - Add proper tag management through the SDK
   - Implement IAM permission verification

4. **Enhance CLI Command Implementation**
   - Complete the `status` command with live data
   - Implement configuration management commands
   - Add history retrieval functionality
   - Create help and documentation commands

5. **Add Logging Implementation**
   - Implement file-based logging with rotation
   - Add syslog integration
   - Implement CloudWatch logging for AWS
   - Add proper error handling throughout the codebase

### Phase 2: Packaging and Distribution

6. **Create Packaging Scripts**
   - Develop DEB packaging for Debian-based systems
   - Implement RPM packaging for Red Hat-based systems
   - Create Homebrew formula for macOS
   - Implement Chocolatey package for Windows
   - Build Windows MSI installer using WiX
   - Add detailed installation instructions
   - Create standardized configuration templates

7. **Service Integration**
   - Finalize systemd service for Linux
   - Add launchd service for macOS
   - Implement Windows service
   - Add proper signal handling
   - Create service management documentation
   - Ensure automatic startup across platforms

### Phase 3: User Experience

8. **Develop Electron GUI**
   - Create basic UI layout
   - Implement real-time monitoring dashboard
   - Add configuration management interface
   - Develop historical data visualization

9. **Document Installation Process**
   - Create comprehensive installation guides
   - Add configuration walkthroughs
   - Develop troubleshooting documentation
   - Create user manual

### Phase 4: Expansion

10. **Plugin Framework Implementation**
    - Create plugin architecture for extensible idle detection
    - Develop plugin manager for discovery and lifecycle management
    - Implement both native Go and external process plugins
    - Provide SDK and examples for plugin developers
    - See [Plugin Architecture](design/plugin-architecture.md) for details

11. **Event-Driven Plugin Framework**
    - Implement cloud event monitoring (AWS Spot interrupts, etc.)
    - Develop event dispatch and handling system
    - Create plugins for graceful shutdowns and data preservation
    - Add prioritized event handling for critical operations
    - See [Event Framework Roadmap](design/plugin-event-roadmap.md) for details

12. **Expand Cloud Provider Support**
    - Add GCP integration
    - Implement Azure support
    - Create abstraction layer for multi-cloud deployments
    - Test and document cross-cloud functionality

13. **External Integration Framework**
    - Implement webhook system for event notifications
    - Add direct integrations with Slack and Microsoft Teams
    - Develop connectors for automation platforms (Zapier, Make.com)
    - Create extensible API for third-party systems
    - See [Integration Roadmap](design/integration-roadmap.md) for details

14. **Advanced Features**
    - Implement predictive idle detection
    - Add scheduled operation policies
    - Create cost savings reports
    - Develop administrator dashboard

## Future Considerations

- Multi-instance coordination and orchestrated shutdowns
- Cost optimization suggestions and automatic instance bidding
- Instance right-sizing recommendations based on usage patterns
- Integration with cost management tools
- Mobile app for remote monitoring and control
- REST API for third-party integrations
- Webhook notifications for important events
- Team collaboration with permission levels
- Machine learning for predictive idle detection and event forecasting
- Infrastructure as Code integration (Terraform, CloudFormation)
- Tagged instance groups with different policies
- Instance resume functionality
- Cost analytics dashboard
- Advanced event correlation for complex cloud environments
- Plugin marketplace with rating system
- Cross-instance event propagation for distributed systems
- Custom event source plugins for application-specific monitoring
- Automated application-aware recovery after spot interruptions

## Contributing

If you're interested in contributing to CloudSnooze, please focus on the tasks in the current phase. Check the GitHub issues for specific tasks that need attention.

## Roadmap Status

This roadmap was last updated on: April 20, 2025

Please note that this roadmap is subject to change based on user feedback and development priorities.