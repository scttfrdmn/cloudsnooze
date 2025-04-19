# CloudSnooze Packaging Guide

This document details the packaging and release process for CloudSnooze.

## Versioning Scheme

CloudSnooze follows Semantic Versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Incompatible API changes
- **MINOR**: Added functionality in a backwards-compatible manner
- **PATCH**: Backwards-compatible bug fixes

### Release Tags

Each release version is tagged in git using the format: `v{MAJOR}.{MINOR}.{PATCH}`

Examples:
- `v0.1.0` - Initial release
- `v0.1.1` - Bug fixes
- `v0.2.0` - New features

## Package Naming Conventions

### DEB Packages (Debian/Ubuntu)

Format: `cloudsnooze_{VERSION}_{ARCH}.deb`

Examples:
- `cloudsnooze_0.1.0_amd64.deb`
- `cloudsnooze_0.1.0_arm64.deb`

Latest symlink:
- `cloudsnooze-latest_amd64.deb`
- `cloudsnooze-latest_arm64.deb`

### RPM Packages (RHEL/CentOS/Fedora)

Format: `cloudsnooze-{VERSION}-{RELEASE}.{ARCH}.rpm`

Examples:
- `cloudsnooze-0.1.0-1.x86_64.rpm`
- `cloudsnooze-0.1.0-1.aarch64.rpm`

Latest symlink:
- `cloudsnooze-latest.x86_64.rpm`
- `cloudsnooze-latest.aarch64.rpm`

## Building Packages

The packaging scripts in `/packaging` directory automate the build process:

### Prerequisites

- Go 1.16 or later
- Required for DEB: `dpkg-deb`
- Required for RPM: `rpmbuild`

### Building DEB Packages

```bash
cd packaging/deb
./build.sh
```

Output will be in: `./build/cloudsnooze_{VERSION}_{ARCH}.deb`

### Building RPM Packages

```bash
cd packaging/rpm
./build.sh
```

Output will be in: `./build/dist/cloudsnooze-{VERSION}-1.{ARCH}.rpm`

### Building All Packages

The master release script handles version bumping and package building:

```bash
cd packaging
./release.sh --patch  # Increment patch version
./release.sh --minor  # Increment minor version
./release.sh --major  # Increment major version
./release.sh --release  # Use current version
```

Output packages will be in: `./packaging/release/`

## GitHub Releases

To create a GitHub release with the packaged files:

1. Run the release script to build all packages:
   ```bash
   cd packaging
   ./release.sh --patch
   ```

2. Push the git tag created by the script:
   ```bash
   git push origin v0.1.0
   ```

3. Create a new release on GitHub:
   - Visit: https://github.com/scttfrdmn/cloudsnooze/releases/new
   - Select the tag just pushed
   - Add release notes
   - Upload all package files from `./packaging/release/`
   - For the "latest" release, also check "Set as latest release"

## Package Contents

Each package includes:

- `/usr/bin/snoozed` - Main daemon executable
- `/usr/bin/snooze` - CLI executable
- `/etc/snooze/snooze.json` - Default configuration
- `/usr/lib/systemd/system/snoozed.service` - Systemd service file
- `/usr/share/doc/cloudsnooze/` - Documentation

## Maintaining Multiple Architectures

The build scripts automatically detect the current architecture. To build for another architecture:

### Cross-compiling for ARM64

```bash
# For DEB
cd packaging/deb
GOOS=linux GOARCH=arm64 PKG_ARCH=arm64 ./build.sh

# For RPM 
cd packaging/rpm
GOOS=linux GOARCH=arm64 PKG_ARCH=aarch64 ./build.sh
```

## Testing Packages

Before release, test packages on appropriate distributions:

- DEB: Test on Ubuntu 20.04 LTS and Debian 11
- RPM: Test on Amazon Linux 2 and RHEL/CentOS 8

Testing script:
```bash
# Test installation
sudo dpkg -i cloudsnooze_0.1.0_amd64.deb  # or
sudo rpm -i cloudsnooze-0.1.0-1.x86_64.rpm

# Verify service starts
systemctl status snoozed

# Test CLI
snooze status

# Test uninstallation
sudo dpkg -r cloudsnooze  # or
sudo rpm -e cloudsnooze
```

## `-latest` Packages

The `-latest` packages are symlinks that point to the most recent stable version's actual package file. These symlinks are automatically created by the build scripts and are uploaded to the GitHub release.

Usage in installation scripts:
```bash
# Always get the latest version
curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/latest/cloudsnooze-latest_amd64.deb
sudo dpkg -i cloudsnooze-latest_amd64.deb
```