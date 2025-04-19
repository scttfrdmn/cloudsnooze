# CloudSnooze

<p align="center">
  <img src="docs/images/logo-placeholder.png" alt="CloudSnooze Logo" width="200"/>
</p>

<p align="center">
  <strong>Automatically stop idle cloud instances and save costs</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#how-it-works">How It Works</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#building-from-source">Building</a> •
  <a href="#license">License</a>
</p>

## Features

- **Low Resource Usage**: Lightweight Go daemon with minimal impact on the monitored instance
- **Comprehensive Monitoring**: Tracks CPU, memory, network, disk I/O, user input, and GPU activity
- **Real User Activity Detection**: Monitors actual keyboard and mouse usage, not just logins
- **Cloud Provider Agnostic**: Initially supports AWS, with design for future expansion
- **Cross-Architecture Support**: Works on both x86_64 and ARM64 instances
- **Multiple Interfaces**: CLI tool, GUI application, and daemon
- **Instance Tagging**: Records when and why instances were stopped
- **Enhanced Logging**: Multiple logging options for visibility and tracking

## How It Works

CloudSnooze monitors system resource usage and automatically stops instances when all metrics remain below specified thresholds for a defined period (the "naptime"). This saves costs by ensuring you only pay for compute resources when they're actually needed.

The system consists of three components:
1. **Daemon (`snoozed`)** - Monitors resources and stops instances
2. **CLI (`snooze`)** - Command-line interface for management
3. **GUI (`snooze-gui`)** - Graphical interface for visual monitoring

<p align="center">
  <img src="docs/images/workflow-placeholder.png" alt="CloudSnooze Workflow" width="600"/>
</p>

## Installation

### From GitHub Releases

1. **Download the appropriate package for your system**:
   - For Debian/Ubuntu (x86_64): `cloudsnooze_1.0.0_amd64.deb`
   - For Debian/Ubuntu (ARM64): `cloudsnooze_1.0.0_arm64.deb` 
   - For RHEL/Fedora/Amazon Linux (x86_64): `cloudsnooze-1.0.0-1.x86_64.rpm`
   - For RHEL/Fedora/Amazon Linux (ARM64): `cloudsnooze-1.0.0-1.aarch64.rpm`

2. **Install the package**:
   ```bash
   # Debian/Ubuntu
   sudo dpkg -i cloudsnooze_1.0.0_*.deb
   
   # RHEL/Fedora/Amazon Linux
   sudo rpm -i cloudsnooze-1.0.0-1.*.rpm
   ```

3. **Configure AWS IAM permissions** (required for AWS instances):
   See the [IAM Configuration Guide](docs/iam-policy-guide.md)

4. **Enable and start the service**:
   ```bash
   sudo systemctl enable snoozed
   sudo systemctl start snoozed
   ```

## Quick Start

After installation, CloudSnooze runs with default settings that work for most scenarios. Here's how to verify and customize:

1. **Check status**:
   ```bash
   snooze status
   ```

2. **View default configuration**:
   ```bash
   snooze config list
   ```

3. **Adjust thresholds** (if needed):
   ```bash
   # Set CPU threshold to 5%
   snooze config set cpu-threshold 5.0
   
   # Set memory threshold to a higher value
   snooze config set memory-threshold 40.0
   
   # Adjust naptime to 20 minutes
   snooze config set naptime 20
   ```

4. **View logs**:
   ```bash
   snooze logs
   ```

## Documentation

- [Overview](docs/overview.md) - Project overview and architecture
- [Data Model](docs/data-model.md) - Core data structures
- [Command Structure](docs/command-structure.md) - CLI commands and usage
- [Deployment Templates](docs/deployment-template.md) - CloudFormation, Terraform, etc.
- [IAM Configuration Guide](docs/iam-policy-guide.md) - AWS permissions setup
- [User Guide](docs/user-guide.md) - Detailed usage instructions
- [Development Guide](docs/development-guide.md) - Information for contributors

## Building from Source

To build CloudSnooze from source:

```bash
# Clone repository
git clone https://github.com/scttfrdmn/cloudsnooze.git
cd cloudsnooze

# Build daemon
cd daemon
go build -o snoozed
cd ..

# Build CLI
cd cli
go build -o snooze
cd ..

# Build GUI (requires Node.js)
cd ui
npm install
npm run build
cd ..
```

See the [Development Guide](docs/development-guide.md) for detailed build instructions and requirements.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all the contributors who have helped with development
- Inspired by the need to save cloud costs automatically
- Built with Go and Electron
