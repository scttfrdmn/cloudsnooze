<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze CLI Reference

This document provides detailed information on the CloudSnooze CLI commands and their usage.

## Global Options

The following options apply to all commands:

- `--version`: Display version information and exit
- `--socket=PATH`: Path to the Unix socket for communicating with the daemon
- `--config=PATH`: Path to the configuration file
- `--help`: Display help information about the specified command

## Commands

### `status`

Display the current system status, including metrics and daemon information.

```
snooze status [options]
```

Options:
- `--watch`, `-w`: Continuously update the display
- `--interval=N`, `-i N`: Refresh interval in seconds when using watch mode (default: 5)
- `--json`, `-j`: Output in JSON format
- `--debug`, `-d`: Include additional debug information

Examples:
```bash
snooze status
snooze status --watch
snooze status --watch --interval=10
snooze status --json
```

### `config`

View or modify configuration settings.

```
snooze config <subcommand> [options]
```

Subcommands:
- `list`: Display all configuration settings
- `get <name>`: Display a specific configuration setting
- `set <name> <value>`: Set a configuration setting
- `reset`: Reset configuration to defaults
- `import <file>`: Import configuration from a file
- `export <file>`: Export configuration to a file

Examples:
```bash
snooze config list
snooze config get cpu-threshold
snooze config set naptime 20
snooze config export my-config.json
```

### `history`

View snooze history and events.

```
snooze history [options]
```

Options:
- `--limit=N`: Limit to N entries (default: 10)
- `--since=DATE`: Show entries since DATE
- `--format=FORMAT`: Output format (text, json, csv) (default: text)
- `--output=FILE`: Write output to FILE

Examples:
```bash
snooze history
snooze history --limit=20
snooze history --since="2025-01-01" --format=json
```

### `issue`

Report issues to the CloudSnooze GitHub repository.

```
snooze issue [options]
```

Options:
- `--type=TYPE`: Issue type (bug, feature, integration, docs) (default: bug)
- `--title=TITLE`: Issue title
- `--description=DESC`: Issue description (if not provided, will prompt for input)
- `--browser`: Open in browser instead of submitting via API (default: true)

Examples:
```bash
snooze issue --type=bug --title="Memory leak in daemon" --description="Observed high memory usage"
snooze issue --type=feature --title="Add support for GCP"
```

### `debug`

Generate debug information for troubleshooting.

```
snooze debug [options]
```

Options:
- `--output=FILE`: Output file (if not specified, outputs to stdout)

Examples:
```bash
snooze debug
snooze debug --output=debug.json
```

### Service Control Commands

#### `start`

Start the CloudSnooze daemon.

```
snooze start
```

#### `stop`

Stop the CloudSnooze daemon.

```
snooze stop
```

#### `restart`

Restart the CloudSnooze daemon.

```
snooze restart
```

## Configuration Parameters

The following configuration parameters can be viewed and modified using the `config` command:

| Parameter | Description | Default | Type |
|-----------|-------------|---------|------|
| `check_interval_seconds` | How frequently to check system metrics | 60 | Integer |
| `naptime_minutes` | How long the system must be idle before stopping | 30 | Integer |
| `cpu_threshold_percent` | CPU usage threshold for idle detection | 10.0 | Float |
| `memory_threshold_percent` | Memory usage threshold for idle detection | 30.0 | Float |
| `network_threshold_kbps` | Network traffic threshold for idle detection | 50.0 | Float |
| `disk_io_threshold_kbps` | Disk I/O threshold for idle detection | 100.0 | Float |
| `input_idle_threshold_secs` | User input idle time threshold | 900 | Integer |
| `gpu_monitoring_enabled` | Whether to monitor GPU usage | true | Boolean |
| `gpu_threshold_percent` | GPU usage threshold for idle detection | 5.0 | Float |
| `aws_region` | AWS region to use | "" (auto-detect) | String |
| `enable_instance_tags` | Whether to tag instances when stopping | true | Boolean |
| `tagging_prefix` | Prefix for instance tags | "CloudSnooze" | String |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Command syntax error |
| 3 | Connection error with daemon |
| 4 | Permission error |
| 5 | Configuration error |