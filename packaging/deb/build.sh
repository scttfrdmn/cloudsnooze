#!/bin/bash
set -e

# CloudSnooze DEB package build script
# This script builds a DEB package for the CloudSnooze daemon and CLI

# Configuration
PKG_NAME="cloudsnooze"
PKG_VERSION=$(grep -oP 'const version = "\K[^"]+' ../../daemon/main.go)
PKG_MAINTAINER="scott freedman <scott@freedman.io>"
PKG_DESCRIPTION="Automatically stop idle cloud instances to save costs"
PKG_DEPENDS="systemd"
PKG_ARCH="amd64"  # or arm64 for ARM builds

# Create staging directory
STAGE_DIR=$(mktemp -d)
PACKAGE_DIR="${STAGE_DIR}/DEBIAN"
mkdir -p "${PACKAGE_DIR}"
BUILD_DIR="./build/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}"
mkdir -p "${BUILD_DIR}"

echo "Building DEB package for CloudSnooze v${PKG_VERSION} (${PKG_ARCH})"

# Create control file
cat > "${PACKAGE_DIR}/control" << EOF
Package: ${PKG_NAME}
Version: ${PKG_VERSION}
Section: utils
Priority: optional
Architecture: ${PKG_ARCH}
Maintainer: ${PKG_MAINTAINER}
Depends: ${PKG_DEPENDS}
Description: ${PKG_DESCRIPTION}
 CloudSnooze monitors CPU, memory, network, disk I/O, GPU, and input activity
 to automatically stop idle cloud instances, saving costs.
EOF

# Create directories
mkdir -p "${STAGE_DIR}/usr/bin"
mkdir -p "${STAGE_DIR}/etc/snooze"
mkdir -p "${STAGE_DIR}/lib/systemd/system"
mkdir -p "${STAGE_DIR}/usr/share/doc/cloudsnooze"
mkdir -p "${STAGE_DIR}/usr/share/man/man1"

# Build the daemon
echo "Building daemon..."
(cd ../../daemon && go build -o "${STAGE_DIR}/usr/bin/snoozed" main.go)

# Build the CLI
echo "Building CLI..."
(cd ../../cli && go build -o "${STAGE_DIR}/usr/bin/snooze" main.go)

# Copy config
cp ../../config/snooze.json "${STAGE_DIR}/etc/snooze/"

# Copy systemd service file
cp ../../systemd/snoozed.service "${STAGE_DIR}/lib/systemd/system/"

# Copy docs
cp ../../README.md "${STAGE_DIR}/usr/share/doc/cloudsnooze/"
cp ../../docs/roadmap.md "${STAGE_DIR}/usr/share/doc/cloudsnooze/"
cp ../../man/snooze.1 "${STAGE_DIR}/usr/share/man/man1/" || echo "Warning: Man page not found"

# Create postinst script
cat > "${PACKAGE_DIR}/postinst" << 'EOF'
#!/bin/sh
set -e

# Enable and start the service
systemctl daemon-reload
systemctl enable snoozed.service
systemctl start snoozed.service || echo "Failed to start snoozed service"

exit 0
EOF
chmod 755 "${PACKAGE_DIR}/postinst"

# Create prerm script
cat > "${PACKAGE_DIR}/prerm" << 'EOF'
#!/bin/sh
set -e

# Stop and disable the service
systemctl stop snoozed.service || true
systemctl disable snoozed.service || true

exit 0
EOF
chmod 755 "${PACKAGE_DIR}/prerm"

# Build the package
echo "Creating DEB package..."
dpkg-deb --build "${STAGE_DIR}" "${BUILD_DIR}/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"

# Create latest symlink
ln -sf "${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb" "${BUILD_DIR}/${PKG_NAME}-latest_${PKG_ARCH}.deb"

echo "Package created: ${BUILD_DIR}/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"
echo "Latest symlink: ${BUILD_DIR}/${PKG_NAME}-latest_${PKG_ARCH}.deb"

# Cleanup 
rm -rf "${STAGE_DIR}"