#!/bin/bash
set -e

# CloudSnooze RPM package build script
# This script builds an RPM package for the CloudSnooze daemon and CLI

# Configuration
PKG_NAME="cloudsnooze"
PKG_VERSION=$(grep -oP 'const version = "\K[^"]+' ../../daemon/main.go)
PKG_RELEASE="1"
PKG_MAINTAINER="scott freedman <scott@freedman.io>"
PKG_DESCRIPTION="Automatically stop idle cloud instances to save costs"
PKG_LICENSE="Apache-2.0"
PKG_ARCH="x86_64"  # or aarch64 for ARM builds

# Create build directories
BUILD_DIR="./build"
mkdir -p "${BUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

echo "Building RPM package for CloudSnooze v${PKG_VERSION} (${PKG_ARCH})"

# Create spec file
cat > "${BUILD_DIR}/SPECS/${PKG_NAME}.spec" << EOF
Name:           ${PKG_NAME}
Version:        ${PKG_VERSION}
Release:        ${PKG_RELEASE}%{?dist}
Summary:        ${PKG_DESCRIPTION}

License:        ${PKG_LICENSE}
URL:            https://github.com/scttfrdmn/cloudsnooze
BuildArch:      ${PKG_ARCH}

Requires:       systemd

%description
CloudSnooze monitors CPU, memory, network, disk I/O, GPU, and input activity
to automatically stop idle cloud instances, saving costs.

%prep
# No source preparation needed for this build

%build
# No specific build instructions - binary is pre-built

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/etc/snooze
mkdir -p %{buildroot}/usr/lib/systemd/system
mkdir -p %{buildroot}/usr/share/doc/cloudsnooze

# Copy pre-built binaries
cp %{_builddir}/snoozed %{buildroot}/usr/bin/
cp %{_builddir}/snooze %{buildroot}/usr/bin/

# Copy config file
cp %{_builddir}/snooze.json %{buildroot}/etc/snooze/

# Copy systemd service file
cp %{_builddir}/snoozed.service %{buildroot}/usr/lib/systemd/system/

# Copy documentation
cp %{_builddir}/README.md %{buildroot}/usr/share/doc/cloudsnooze/
cp %{_builddir}/roadmap.md %{buildroot}/usr/share/doc/cloudsnooze/

%post
systemctl daemon-reload
systemctl enable snoozed.service
systemctl start snoozed.service || echo "Failed to start snoozed service"

%preun
systemctl stop snoozed.service || true
systemctl disable snoozed.service || true

%files
%defattr(-,root,root)
%{_bindir}/snoozed
%{_bindir}/snooze
%{_unitdir}/snoozed.service
%config(noreplace) %{_sysconfdir}/snooze/snooze.json
%{_datadir}/doc/cloudsnooze/*

%changelog
* $(date '+%a %b %d %Y') ${PKG_MAINTAINER} ${PKG_VERSION}-${PKG_RELEASE}
- Initial package release
EOF

# Build the daemon
echo "Building daemon..."
(cd ../../daemon && go build -o "${BUILD_DIR}/BUILD/snoozed" main.go)

# Build the CLI
echo "Building CLI..."
(cd ../../cli && go build -o "${BUILD_DIR}/BUILD/snooze" main.go)

# Copy files
cp ../../config/snooze.json "${BUILD_DIR}/BUILD/"
cp ../../systemd/snoozed.service "${BUILD_DIR}/BUILD/"
cp ../../README.md "${BUILD_DIR}/BUILD/"
cp ../../docs/roadmap.md "${BUILD_DIR}/BUILD/"

# Build the RPM package
echo "Creating RPM package..."
rpmbuild --define "_topdir $(pwd)/${BUILD_DIR}" -bb "${BUILD_DIR}/SPECS/${PKG_NAME}.spec"

# Move the resulting RPM and create latest symlink
mkdir -p "${BUILD_DIR}/dist"
mv "${BUILD_DIR}/RPMS/${PKG_ARCH}/${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}*.rpm" "${BUILD_DIR}/dist/${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}.${PKG_ARCH}.rpm"
ln -sf "${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}.${PKG_ARCH}.rpm" "${BUILD_DIR}/dist/${PKG_NAME}-latest.${PKG_ARCH}.rpm"

echo "Package created: ${BUILD_DIR}/dist/${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}.${PKG_ARCH}.rpm"
echo "Latest symlink: ${BUILD_DIR}/dist/${PKG_NAME}-latest.${PKG_ARCH}.rpm"