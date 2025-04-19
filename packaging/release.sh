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

# Create output directories
mkdir -p ./release
rm -rf ./release/*

# Build x86_64 packages
echo "Building x86_64 packages..."
(cd ./deb && ./build.sh)
cp ./deb/build/cloudsnooze_${NEW_VERSION}_amd64.deb ./release/
cp ./deb/build/cloudsnooze-latest_amd64.deb ./release/

(cd ./rpm && PKG_ARCH=x86_64 ./build.sh)
cp ./rpm/build/dist/cloudsnooze-${NEW_VERSION}-1.x86_64.rpm ./release/
cp ./rpm/build/dist/cloudsnooze-latest.x86_64.rpm ./release/

echo "Packages built successfully:"
ls -la ./release/

# Instructions for GitHub release
echo ""
echo "===== Next steps for GitHub release ====="
echo "1. Push the tag: git push origin v$NEW_VERSION"
echo "2. Create a new release on GitHub with tag v$NEW_VERSION"
echo "3. Upload the package files from ./release directory"
echo "4. Add release notes describing the changes"
echo "=========================================="