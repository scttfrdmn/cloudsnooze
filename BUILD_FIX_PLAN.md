# CloudSnooze Build Fix Plan

## Executive Summary

The CloudSnooze project is currently experiencing build failures in the GitHub Actions CI environment. This document outlines a comprehensive plan to address these issues, focusing on resolving Go module dependencies and fixing circular import problems. The plan includes a detailed timeline, resource allocation, and success criteria. We anticipate completing all fixes by April 29, 2025, with a total estimated effort of approximately 45 person-hours.

## Project Overview

This document outlines the issues with the current build process and provides a step-by-step plan to resolve them. The primary goals are to:

1. Fix all dependency issues causing build failures
2. Restructure the codebase to eliminate circular imports
3. Ensure CI/CD pipelines pass consistently
4. Update documentation to prevent future issues

## Current Issues

The GitHub Actions build and test workflows are failing due to the following issues:

1. **Missing Go Module Dependencies**
   - Missing required gopsutil packages for system monitoring:
     - `github.com/shirou/gopsutil/v3/cpu`
     - `github.com/shirou/gopsutil/v3/disk` 
     - `github.com/shirou/gopsutil/v3/mem`
     - `github.com/shirou/gopsutil/v3/net`

2. **Import Cycle Issues**
   - Circular dependencies between packages:
     - `daemon/monitor` imports `daemon/accelerator`
     - `daemon/accelerator` imports `daemon/monitor`
     - `daemon/cloud/aws` imports both `daemon/cloud` and `daemon/monitor`
     - `daemon/cloud` imports `daemon/cloud/aws`

## Fix Plan

### 1. Set Up Proper Go Modules

1. **Initialize Go Modules**
   - Ensure go.mod is properly set up at the appropriate level (daemon and cli)
   - Configure module paths correctly

2. **Add Required Dependencies**
   - Add missing gopsutil dependencies
   - Specify exact versions to ensure compatibility

### 2. Resolve Import Cycles

1. **Create Common Types Package**
   - Create a new package `daemon/types` or `daemon/common` to hold shared data structures
   - Move common types/interfaces to this package to break circular dependencies

2. **Restructure Monitor and Accelerator**
   - Refactor the monitor package to not depend on accelerator
   - Refactor the accelerator package to use types from the common package
   - Ensure monitor provides data to accelerator but doesn't import it

3. **Fix Cloud Provider Architecture**
   - Restructure cloud/aws to avoid direct imports of cloud package
   - Consider using interfaces defined in the common package
   - Implement factory pattern correctly to avoid circular references

### 3. Update GitHub Actions Workflow

1. **Improve Workflow Configuration**
   - Add proper Go dependency caching
   - Ensure the build process works with modules
   - Add a specific step to download dependencies before building

2. **Enhanced Error Reporting**
   - Add better error reporting to identify issues more quickly
   - Add GO111MODULE=on to ensure module mode is used

### 4. Specific Code Changes

#### 4.1 Go Modules Setup

```bash
# Set up go.mod in daemon directory
cd daemon
go mod init github.com/scttfrdmn/cloudsnooze/daemon
go mod tidy

# Add required dependencies
go get github.com/shirou/gopsutil/v3@v3.23.6
```

#### 4.2 Create Common Types Package

Create new file: `daemon/common/types.go`

```go
// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package common

// SystemMetrics contains all metrics collected from the system
type SystemMetrics struct {
    CPUUsage        float64
    MemoryUsage     float64
    DiskIORate      float64
    NetworkRate     float64
    IdleTime        int64
    GPUMetrics      []GPUMetrics
    LastInputTime   int64
    CollectionTime  int64
}

// GPUMetrics contains metrics specific to GPU devices
type GPUMetrics struct {
    ID              string
    Utilization     float64
    MemoryUsed      uint64
    MemoryTotal     uint64
    Temperature     float64
    Vendor          string
    Model           string
}

// CloudProvider defines the interface for cloud providers
type CloudProvider interface {
    GetInstanceID() (string, error)
    StopInstance(string) error
    IsInstanceRunning(string) (bool, error)
    TagInstance(string, map[string]string) error
    VerifyPermissions() error
}
```

#### 4.3 Update Monitor Package

Modify `daemon/monitor/types.go` to use the common package instead of directly importing accelerator.

#### 4.4 Update Accelerator Package

Modify `daemon/accelerator/gpu.go` to use types from the common package and not import monitor.

#### 4.5 Fix Cloud Package

Refactor the cloud and AWS packages to use the common interfaces and avoid circular imports.

## Implementation Timeline

1. **Setup Phase (April 21-22, 2025)**
   - **April 21, 2025**
     - Set up proper Go modules for daemon directory (2 hours)
     - Set up proper Go modules for cli directory (1 hour)
     - Add required gopsutil dependencies with fixed versions (2 hours)
   - **April 22, 2025**
     - Create common types package to break dependency cycles (3 hours)
     - Initial test of build with new module structure (1 hour)

2. **Refactoring Phase (April 23-25, 2025)**
   - **April 23, 2025**
     - Refactor monitor package to use common types (4 hours)
     - Update all monitor subpackages to follow new structure (3 hours)
   - **April 24, 2025**
     - Refactor accelerator package to use common types (3 hours)
     - Fix GPU monitoring dependency issues (2 hours)
   - **April 25, 2025**
     - Update cloud provider implementation to avoid circular imports (4 hours)
     - Refactor AWS-specific code to use interfaces properly (3 hours)

3. **Testing Phase (April 26-27, 2025)**
   - **April 26, 2025**
     - Local build testing across all packages (2 hours)
     - Run unit tests and fix any test-specific issues (3 hours)
     - Address any remaining import cycle issues (2 hours)
   - **April 27, 2025**
     - CI/CD workflow testing with updated GitHub Actions (3 hours)
     - Cross-platform testing (Linux/macOS) (2 hours)
     - Fix any platform-specific issues (2 hours)

4. **Documentation Phase (April 28-29, 2025)**
   - **April 28, 2025**
     - Update developer documentation to reflect new architecture (3 hours)
     - Create package dependency diagram for future reference (2 hours)
   - **April 29, 2025**
     - Create contributing guidelines for package structure (2 hours)
     - Final review and submission of all changes (2 hours)
     - Verification of build badges showing passing status (1 hour)

## Success Criteria

- All GitHub Actions workflows pass successfully by April 29, 2025
- No import cycle errors reported by the Go compiler
- All required dependencies are properly included and versioned
- Code builds and tests pass on all target platforms (Linux/macOS, x86_64/ARM64)
- Documentation is updated to reflect architectural changes
- Build status badge shows "passing" on the GitHub repository

## Project Management

### Key Milestones
- **April 22, 2025**: Complete module setup and initial common package
- **April 25, 2025**: Complete all code refactoring to resolve import cycles
- **April 27, 2025**: All tests passing locally and in CI environment
- **April 29, 2025**: Documentation updated and project fully restored to working state

### Responsible Team Members
- **Lead Developer**: Will oversee the overall implementation plan
- **Backend Developer 1**: Responsible for module setup and dependency management
- **Backend Developer 2**: Responsible for refactoring monitor and accelerator packages
- **DevOps Engineer**: Responsible for GitHub Actions workflow updates
- **Technical Writer**: Responsible for documentation updates

### Status Reporting
- Daily stand-up meetings to report progress and blockers
- GitHub issues created for each specific task
- Progress tracked via project board with Kanban methodology

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Additional undiscovered import cycles | Medium | High | Conduct thorough static analysis of imports before starting implementation |
| Go module versioning conflicts | Medium | Medium | Pin all dependencies to specific versions; create lockfile |
| Platform-specific build issues | Low | Medium | Test on both Linux and macOS throughout development, not just at the end |
| Timeline extension due to complexity | Medium | Medium | Build in 20% buffer time; prioritize critical path items first |
| Knowledge gaps in Go modules | Low | Medium | Schedule training session on Go modules before implementation begins |

## Cost and Resource Estimation

- **Total Developer Hours**: ~45 hours
- **Testing Hours**: ~10 hours
- **Documentation Hours**: ~8 hours
- **Project Management**: ~5 hours
- **Total Estimated Cost**: 68 person-hours

## Future Considerations

- Set up a development container for consistent environments
- Implement stronger test coverage to catch issues earlier
- Consider using a monorepo tool like Bazel for better dependency management
- Document internal package dependencies to prevent future cycles
- Implement automated import cycle detection in pre-commit hooks
- Consider modularizing the codebase further to reduce tight coupling