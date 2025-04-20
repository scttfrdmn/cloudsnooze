#!/bin/bash
set -e

# CloudSnooze Release Script
# This script handles version tagging and package building for releases

# Usage information
function show_usage {
  echo "Usage: $0 [--major|--minor|--patch|--release]"
  echo ""
  echo "Options:"
  echo "  --major    Increment major version number (X.0.0)"
  echo "  --minor    Increment minor version number (x.X.0)"
  echo "  --patch    Increment patch version number (x.x.X)"
  echo "  --release  Use current version and build release packages"
  echo ""
  echo "Example: $0 --patch"
  exit 1
}

# Check if we have required tools
for cmd in git go rpmbuild dpkg-deb; do
  if ! command -v $cmd &> /dev/null; then
    echo "Error: $cmd is required but not installed"
    exit 1
  fi
done

# Get current version
CURRENT_VERSION=$(grep -oP 'const version = "\K[^"]+' ../daemon/main.go)
if [ -z "$CURRENT_VERSION" ]; then
  echo "Error: Could not determine current version"
  exit 1
fi

echo "Current version: $CURRENT_VERSION"
MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1)
MINOR=$(echo $CURRENT_VERSION | cut -d. -f2)
PATCH=$(echo $CURRENT_VERSION | cut -d. -f3)

# Process command line arguments
if [ $# -ne 1 ]; then
  show_usage
fi

case "$1" in
  --major)
    NEW_VERSION="$((MAJOR+1)).0.0"
    ;;
  --minor)
    NEW_VERSION="$MAJOR.$((MINOR+1)).0"
    ;;
  --patch)
    NEW_VERSION="$MAJOR.$MINOR.$((PATCH+1))"
    ;;
  --release)
    NEW_VERSION="$CURRENT_VERSION"
    ;;
  *)
    show_usage
    ;;
esac

echo "Release version: $NEW_VERSION"

# If we're updating the version, update source files
if [ "$NEW_VERSION" != "$CURRENT_VERSION" ]; then
  echo "Updating version in source files..."
  
  # Update daemon main.go
  sed -i.bak "s/const version = \"$CURRENT_VERSION\"/const version = \"$NEW_VERSION\"/" ../daemon/main.go
  rm -f ../daemon/main.go.bak
  
  # Update CLI main.go
  sed -i.bak "s/const version = \"$CURRENT_VERSION\"/const version = \"$NEW_VERSION\"/" ../cli/main.go
  rm -f ../cli/main.go.bak
  
  # Commit the version change
  git add ../daemon/main.go ../cli/main.go
  git commit -m "Bump version to $NEW_VERSION"
fi

# Create git tag
if [ "$NEW_VERSION" != "$CURRENT_VERSION" ]; then
  echo "Creating git tag v$NEW_VERSION..."
  git tag -a "v$NEW_VERSION" -m "Release version $NEW_VERSION"
fi

# Build architecture-specific packages
echo "Building packages..."

# Make build scripts executable
chmod +x ./deb/build.sh
chmod +x ./rpm/build.sh
chmod +x ./windows/build-msi.ps1

# Create output directories
RELEASE_DIR="./release"
BUILD_DIR="../dist/build"
mkdir -p "$RELEASE_DIR"
mkdir -p "$BUILD_DIR"
rm -rf "$RELEASE_DIR"/*

# Build binaries for all platforms
echo "Building binaries for all platforms..."
(
  cd ..
  
  # Linux (amd64)
  echo "Building for Linux (amd64)..."
  GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/snoozed_linux_amd64" ./daemon
  GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/snooze_linux_amd64" ./cli
  
  # Linux (arm64)
  echo "Building for Linux (arm64)..."
  GOOS=linux GOARCH=arm64 go build -o "$BUILD_DIR/snoozed_linux_arm64" ./daemon
  GOOS=linux GOARCH=arm64 go build -o "$BUILD_DIR/snooze_linux_arm64" ./cli
  
  # macOS (amd64)
  echo "Building for macOS (amd64)..."
  GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/snoozed_darwin_amd64" ./daemon
  GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/snooze_darwin_amd64" ./cli
  
  # macOS (arm64)
  echo "Building for macOS (arm64)..."
  GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/snoozed_darwin_arm64" ./daemon
  GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/snooze_darwin_arm64" ./cli
  
  # Windows (amd64)
  echo "Building for Windows (amd64)..."
  GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/snoozed_windows_amd64.exe" ./daemon
  GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/snooze_windows_amd64.exe" ./cli
)

# Create tarballs and zip archives
echo "Creating archives..."
(
  cd "$BUILD_DIR"
  
  # Linux amd64 tarball
  echo "Creating Linux (amd64) tarball..."
  mkdir -p linux_amd64/{bin,config,man}
  cp snoozed_linux_amd64 linux_amd64/bin/snoozed
  cp snooze_linux_amd64 linux_amd64/bin/snooze
  cp "../../config/snooze.json" linux_amd64/config/
  cp "../../man/"* linux_amd64/man/ 2>/dev/null || true
  cp "../../README.md" linux_amd64/
  cp "../../LICENSE" linux_amd64/
  tar -czf "$RELEASE_DIR/cloudsnooze_${NEW_VERSION}_linux_amd64.tar.gz" linux_amd64
  
  # Linux arm64 tarball
  echo "Creating Linux (arm64) tarball..."
  mkdir -p linux_arm64/{bin,config,man}
  cp snoozed_linux_arm64 linux_arm64/bin/snoozed
  cp snooze_linux_arm64 linux_arm64/bin/snooze
  cp "../../config/snooze.json" linux_arm64/config/
  cp "../../man/"* linux_arm64/man/ 2>/dev/null || true
  cp "../../README.md" linux_arm64/
  cp "../../LICENSE" linux_arm64/
  tar -czf "$RELEASE_DIR/cloudsnooze_${NEW_VERSION}_linux_arm64.tar.gz" linux_arm64
  
  # macOS amd64 tarball
  echo "Creating macOS (amd64) tarball..."
  mkdir -p darwin_amd64/{bin,config,man}
  cp snoozed_darwin_amd64 darwin_amd64/bin/snoozed
  cp snooze_darwin_amd64 darwin_amd64/bin/snooze
  cp "../../config/snooze.json" darwin_amd64/config/
  cp "../../man/"* darwin_amd64/man/ 2>/dev/null || true
  cp "../../README.md" darwin_amd64/
  cp "../../LICENSE" darwin_amd64/
  tar -czf "$RELEASE_DIR/cloudsnooze_${NEW_VERSION}_darwin_amd64.tar.gz" darwin_amd64
  
  # macOS arm64 tarball
  echo "Creating macOS (arm64) tarball..."
  mkdir -p darwin_arm64/{bin,config,man}
  cp snoozed_darwin_arm64 darwin_arm64/bin/snoozed
  cp snooze_darwin_arm64 darwin_arm64/bin/snooze
  cp "../../config/snooze.json" darwin_arm64/config/
  cp "../../man/"* darwin_arm64/man/ 2>/dev/null || true
  cp "../../README.md" darwin_arm64/
  cp "../../LICENSE" darwin_arm64/
  tar -czf "$RELEASE_DIR/cloudsnooze_${NEW_VERSION}_darwin_arm64.tar.gz" darwin_arm64
  
  # Windows amd64 zip
  echo "Creating Windows (amd64) zip..."
  mkdir -p windows_amd64/{bin,config}
  cp snoozed_windows_amd64.exe windows_amd64/bin/snoozed.exe
  cp snooze_windows_amd64.exe windows_amd64/bin/snooze.exe
  cp "../../config/snooze.json" windows_amd64/config/
  cp "../../README.md" windows_amd64/
  cp "../../LICENSE" windows_amd64/
  
  # Check if zip command exists
  if command -v zip >/dev/null 2>&1; then
    zip -r "$RELEASE_DIR/cloudsnooze_${NEW_VERSION}_windows_amd64.zip" windows_amd64
  else
    echo "Warning: zip command not found, skipping Windows zip creation"
    # Create a directory for Windows files
    mkdir -p "$RELEASE_DIR/windows_amd64"
    cp -r windows_amd64/* "$RELEASE_DIR/windows_amd64/"
  fi
)

# Build Linux packages (x86_64)
echo "Building Linux (x86_64) packages..."
(cd ./deb && ./build.sh "$NEW_VERSION" "amd64")
cp ./deb/build/cloudsnooze_${NEW_VERSION}_amd64.deb "$RELEASE_DIR/"
cp ./deb/build/cloudsnooze-latest_amd64.deb "$RELEASE_DIR/"

(cd ./rpm && PKG_ARCH=x86_64 ./build.sh "$NEW_VERSION")
cp ./rpm/build/dist/cloudsnooze-${NEW_VERSION}-1.x86_64.rpm "$RELEASE_DIR/"
cp ./rpm/build/dist/cloudsnooze-latest.x86_64.rpm "$RELEASE_DIR/"

# Build Linux packages (arm64)
echo "Building Linux (arm64) packages..."
(cd ./deb && ./build.sh "$NEW_VERSION" "arm64")
cp ./deb/build/cloudsnooze_${NEW_VERSION}_arm64.deb "$RELEASE_DIR/"
cp ./deb/build/cloudsnooze-latest_arm64.deb "$RELEASE_DIR/"

(cd ./rpm && PKG_ARCH=aarch64 ./build.sh "$NEW_VERSION")
cp ./rpm/build/dist/cloudsnooze-${NEW_VERSION}-1.aarch64.rpm "$RELEASE_DIR/"
cp ./rpm/build/dist/cloudsnooze-latest.aarch64.rpm "$RELEASE_DIR/"

# Build Windows MSI installer (if WiX toolset available)
if command -v candle.exe >/dev/null 2>&1 && command -v light.exe >/dev/null 2>&1; then
  echo "Building Windows MSI installer..."
  (cd ./windows && powershell -Command "./build-msi.ps1 -Version $NEW_VERSION -BuildDir '$BUILD_DIR/windows'")
  if [ -f "$BUILD_DIR/windows/cloudsnooze-$NEW_VERSION-windows-amd64.msi" ]; then
    cp "$BUILD_DIR/windows/cloudsnooze-$NEW_VERSION-windows-amd64.msi" "$RELEASE_DIR/"
  fi
else
  echo "WiX toolset not found, skipping MSI build"
fi

# Update Homebrew formula
echo "Updating Homebrew formula..."
(
  cd "$RELEASE_DIR"
  # Calculate checksums
  DARWIN_AMD64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_darwin_amd64.tar.gz" | awk '{print $1}')
  DARWIN_ARM64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_darwin_arm64.tar.gz" | awk '{print $1}')
  LINUX_AMD64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_linux_amd64.tar.gz" | awk '{print $1}')
  LINUX_ARM64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_linux_arm64.tar.gz" | awk '{print $1}')
  
  cd ../homebrew
  # Update version in formula
  sed -i.bak "s/version \"[0-9.]*\"/version \"$NEW_VERSION\"/" cloudsnooze.rb
  # Update checksums
  sed -i.bak "s/REPLACE_WITH_AMD64_MAC_CHECKSUM/$DARWIN_AMD64_CHECKSUM/g" cloudsnooze.rb
  sed -i.bak "s/REPLACE_WITH_ARM64_MAC_CHECKSUM/$DARWIN_ARM64_CHECKSUM/g" cloudsnooze.rb
  sed -i.bak "s/REPLACE_WITH_AMD64_LINUX_CHECKSUM/$LINUX_AMD64_CHECKSUM/g" cloudsnooze.rb
  sed -i.bak "s/REPLACE_WITH_ARM64_LINUX_CHECKSUM/$LINUX_ARM64_CHECKSUM/g" cloudsnooze.rb
  rm -f cloudsnooze.rb.bak
  
  # Copy updated formula to release dir
  cp cloudsnooze.rb "$RELEASE_DIR/"
)

# Update Chocolatey package
echo "Updating Chocolatey package..."
(
  cd "$RELEASE_DIR"
  # Calculate Windows zip checksum
  if [ -f "cloudsnooze_${NEW_VERSION}_windows_amd64.zip" ]; then
    WINDOWS_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_windows_amd64.zip" | awk '{print $1}')
    
    cd ../windows
    # Update version in nuspec
    sed -i.bak "s/<version>[0-9.]*<\/version>/<version>$NEW_VERSION<\/version>/" cloudsnooze.nuspec
    # Update checksum in install script
    sed -i.bak "s/REPLACE_WITH_CHECKSUM/$WINDOWS_CHECKSUM/g" tools/chocolateyinstall.ps1
    rm -f cloudsnooze.nuspec.bak tools/chocolateyinstall.ps1.bak
    
    # Copy files to release dir
    mkdir -p "$RELEASE_DIR/chocolatey"
    cp -r . "$RELEASE_DIR/chocolatey/"
    
    # Create Chocolatey package if choco is installed
    if command -v choco >/dev/null 2>&1; then
      echo "Building Chocolatey package..."
      cd "$RELEASE_DIR/chocolatey"
      choco pack
      mv cloudsnooze.*.nupkg ../ 2>/dev/null || true
    else
      echo "Chocolatey not installed, skipping package creation"
    fi
  else
    echo "Windows zip file not found, skipping Chocolatey package update"
  fi
)

# Create release notes
echo "Creating release notes..."
(
  cd "$RELEASE_DIR"
  
  # Calculate checksums for release notes
  DARWIN_AMD64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_darwin_amd64.tar.gz" 2>/dev/null | awk '{print $1}' || echo "N/A")
  DARWIN_ARM64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_darwin_arm64.tar.gz" 2>/dev/null | awk '{print $1}' || echo "N/A")
  LINUX_AMD64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_linux_amd64.tar.gz" 2>/dev/null | awk '{print $1}' || echo "N/A")
  LINUX_ARM64_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_linux_arm64.tar.gz" 2>/dev/null | awk '{print $1}' || echo "N/A")
  WINDOWS_CHECKSUM=$(shasum -a 256 "cloudsnooze_${NEW_VERSION}_windows_amd64.zip" 2>/dev/null | awk '{print $1}' || echo "N/A")
  
  cat > "RELEASE_NOTES.md" << EOF
# CloudSnooze v$NEW_VERSION Release Notes

## Overview
This release provides a new version of CloudSnooze, a tool for automatically stopping idle cloud instances to save costs.

## Features
- Comprehensive monitoring of CPU, memory, network, disk I/O, user input, and GPU activity
- Support for AWS cloud instances
- Cross-platform support (Linux, macOS, Windows)
- Low-resource footprint daemon
- Command-line interface for configuration and status
- Configurable thresholds and idle detection

## Installation

### Linux (Debian/Ubuntu)
\`\`\`bash
sudo dpkg -i cloudsnooze_${NEW_VERSION}_*.deb
\`\`\`

### Linux (RHEL/Fedora/Amazon Linux)
\`\`\`bash
sudo rpm -i cloudsnooze-${NEW_VERSION}-*.rpm
\`\`\`

### macOS (Homebrew)
\`\`\`bash
brew tap scttfrdmn/cloudsnooze
brew install cloudsnooze
\`\`\`

### Windows (Chocolatey)
\`\`\`powershell
choco install cloudsnooze
\`\`\`

### Windows (MSI)
Download and run the MSI installer.

## SHA-256 Checksums
\`\`\`
$DARWIN_AMD64_CHECKSUM  cloudsnooze_${NEW_VERSION}_darwin_amd64.tar.gz
$DARWIN_ARM64_CHECKSUM  cloudsnooze_${NEW_VERSION}_darwin_arm64.tar.gz
$LINUX_AMD64_CHECKSUM  cloudsnooze_${NEW_VERSION}_linux_amd64.tar.gz
$LINUX_ARM64_CHECKSUM  cloudsnooze_${NEW_VERSION}_linux_arm64.tar.gz
$WINDOWS_CHECKSUM  cloudsnooze_${NEW_VERSION}_windows_amd64.zip
\`\`\`

## Known Issues
- This is an alpha release and may have stability issues
- Limited cloud provider support (AWS only)
- Advanced idle detection capabilities are planned for future releases

## Next Steps
See our roadmap for upcoming features: https://github.com/scttfrdmn/cloudsnooze/blob/main/docs/roadmap.md
EOF
)

echo "Packages built successfully:"
ls -la "$RELEASE_DIR"

# Instructions for GitHub release
echo ""
echo "===== Next steps for GitHub release ====="
echo "1. Push the tag: git push origin v$NEW_VERSION"
echo "2. Create a new release on GitHub with tag v$NEW_VERSION"
echo "3. Upload the package files from ./release directory"
echo "4. Add release notes describing the changes"
echo "4. Use RELEASE_NOTES.md as the release description"
echo "5. For Homebrew tap, push the updated formula to:"
echo "   https://github.com/scttfrdmn/homebrew-cloudsnooze"
echo "6. For Chocolatey, submit the package to:"
echo "   https://chocolatey.org/packages/upload"
echo "==========================================="