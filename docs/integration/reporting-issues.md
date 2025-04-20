<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# Reporting Issues with CloudSnooze

This document explains the various ways to report issues, request features, or provide feedback for CloudSnooze.

## GitHub Issue Templates

CloudSnooze provides structured issue templates in our GitHub repository to help you provide all the necessary information when reporting issues or requesting features.

### Available Issue Templates

1. **Bug Report**: For any unexpected behavior, errors, or crashes.
2. **Feature Request**: For suggesting new functionality or improvements.
3. **Integration Issue**: For problems related to integrating CloudSnooze with external systems.
4. **Documentation Request**: For suggesting improvements to the documentation.

### Using GitHub Issues Directly

To report an issue directly from GitHub:

1. Go to [CloudSnooze Issues](https://github.com/scttfrdmn/cloudsnooze/issues)
2. Click on "New Issue"
3. Select the appropriate template
4. Fill in the required information
5. Submit the issue

The templates will prompt you for specific details needed to address your issue effectively.

## CLI Issue Reporting

CloudSnooze's CLI includes built-in functionality to report issues directly from the command line. This feature automatically collects relevant system information to help with troubleshooting.

### Basic Usage

```bash
snooze issue -type <bug|feature|integration|docs> -title "Your issue title"
```

This will open your default web browser with a pre-filled GitHub issue form.

### Options

- `-type`: Type of issue (bug, feature, integration, docs)
- `-title`: Issue title
- `-description`: Issue description (if not provided, will prompt for input)
- `-browser`: Open in browser (default: true)

### Examples

```bash
# Report a bug
snooze issue -type bug -title "Memory leak in daemon" -description "Observed high memory usage"

# Request a feature
snooze issue -type feature -title "Add support for GCP"

# Report an integration issue
snooze issue -type integration -title "Slack notification failure"

# Request documentation improvements
snooze issue -type docs -title "Clarify installation instructions"
```

### What information is collected?

When you report an issue via the CLI, the following information is automatically collected:

- CloudSnooze version
- Operating system and version
- System architecture
- Installation method
- Cloud provider (if running on a cloud instance)
- Instance type (if running on a cloud instance)
- Recent log entries (truncated to protect privacy)

You can review this information before submitting the issue.

## Generating Debug Information

For troubleshooting complex issues, you can generate comprehensive debug information:

```bash
snooze debug -output debug.json
```

This command collects detailed information about your CloudSnooze installation, configuration, and system environment.

### Options

- `-output`: Path to save the debug information (if not specified, outputs to console)

### Example Output

```json
{
  "environment": {
    "CloudSnooze Version": "0.1.0",
    "OS": "Ubuntu 22.04.2 LTS",
    "Architecture": "x86_64",
    "Installation Method": "DEB package",
    "Cloud Provider": "AWS",
    "Instance Type": "t3.medium"
  },
  "status": {
    "version": "0.1.0",
    "idle_since": null,
    "should_snooze": false,
    "snooze_reason": "System is active"
  },
  "config": {
    "check_interval_seconds": 60,
    "naptime_minutes": 30,
    "cpu_threshold_percent": 10.0,
    "memory_threshold_percent": 30.0,
    "network_threshold_kbps": 50.0,
    "disk_io_threshold_kbps": 100.0,
    "input_idle_threshold_secs": 900,
    "gpu_monitoring_enabled": true
  },
  "logs": "..."
}
```

## Best Practices for Issue Reporting

1. **Be specific**: Clearly describe what you were doing when the issue occurred
2. **Include steps to reproduce**: List the exact steps needed to reproduce the issue
3. **Provide expected vs. actual behavior**: Explain what you expected to happen and what actually happened
4. **Share error messages**: Include any error messages or unexpected output
5. **Attach debug information**: When possible, include the output from `snooze debug`

## Security Issues

For security-related issues, please DO NOT report them publicly on the issue tracker.

Instead, please send an email to security@cloudsnooze.io with details of the vulnerability. We will address security issues with the highest priority.