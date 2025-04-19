
# CloudSnooze - Command Structure

This document outlines the command structure for the CloudSnooze CLI tool (`snooze`), including all available commands, options, and usage examples.

## Top-Level Command

```
snooze [command] [options]
```

Global flags:
- `--help, -h`: Display help for any command
- `--version, -v`: Display version information
- `--config, -c`: Specify an alternative configuration file path

## Primary Commands

### status

Display the current system status, including metrics and daemon information.

```
snooze status [options]
```

Options:
- `--watch, -w`: Continuously update the display (like top)
- `--interval=N, -i N`: Refresh interval in seconds when using watch mode

Examples:
```bash
# Show current status
snooze status

# Watch status with 5-second refresh
snooze status --watch --interval=5
```

Output example:
```
CloudSnooze Status
------------------
Status: Running
Monitoring since: 2025-04-19 14:30:45 UTC

Current metrics:
  - CPU: 5.2% (threshold: 10.0%)
  - Memory: 22.7% (threshold: 30.0%)
  - Network: 12.3 KB/s (threshold: 50.0 KB/s)
  - Disk I/O: 0.5 KB/s (threshold: 100.0 KB/s)
  - Input idle: 125s (threshold: 900s)
  - GPU [NVIDIA T4]: 0.0% (threshold: 5.0%)

System idle: No (Input activity detected)
Current naptime: 0 of 30 minutes
```

### config

View or modify configuration parameters.

```
snooze config [action] [parameter] [value]
```

Actions:
- `list`: Show all configuration parameters
- `get`: Get a specific parameter's value
- `set`: Set a parameter's value
- `reset`: Reset to default values
- `import`: Import configuration from a file
- `export`: Export configuration to a file
- `verify`: Verify configuration syntax

Examples:
```bash
# List all configuration options
snooze config list

# Get a specific parameter
snooze config get naptime

# Set a parameter
snooze config set cpu-threshold 15.0

# Reset to defaults
snooze config reset

# Import from file
snooze config import myconfig.json

# Export to file
snooze config export backup.json
```

### history

View the history of instance snooze events.

```
snooze history [options]
```

Options:
- `--limit=N, -l N`: Limit to N entries (default: 10)
- `--since=DATE, -s DATE`: Show entries since DATE
- `--format=FORMAT, -f FORMAT`: Output format (text, json, csv)
- `--output=FILE, -o FILE`: Write output to FILE

Examples:
```bash
# View last 10 snooze events
snooze history

# View last 20 events in JSON format
snooze history --limit=20 --format=json

# Export all events since last week to CSV
snooze history --since="1 week ago" --format=csv --output=history.csv
```

### start/stop/restart

Control the daemon service.

```
snooze start
snooze stop
snooze restart
```

Examples:
```bash
# Start the daemon
snooze start

# Stop the daemon
snooze stop

# Restart the daemon
snooze restart
```

### enable/disable

Control automatic startup of the daemon.

```
snooze enable
snooze disable
```

### simulate

Run a simulation based on specified metrics to see if snoozing would be triggered.

```
snooze simulate [options]
```

Options:
- `--cpu=PCT`: CPU usage percentage
- `--memory=PCT`: Memory usage percentage
- `--network=KBPS`: Network I/O in KB/s
- `--disk=KBPS`: Disk I/O in KB/s
- `--input=SECS`: Input idle time in seconds
- `--gpu=PCT`: GPU usage percentage
- `--duration=MINS`: Duration to simulate in minutes

Examples:
```bash
# Simulate with low resource usage for 40 minutes
snooze simulate --cpu=5 --memory=20 --network=10 --input=1800 --duration=40

# Simulate with high CPU but low everything else
snooze simulate --cpu=80 --memory=20 --network=5 --input=1800 --duration=35
```

## Advanced Commands

### tags

Manage instance tags (AWS-specific).

```
snooze tags [action] [instance-id]
```

Actions:
- `list`: List CloudSnooze tags for an instance
- `clear`: Clear CloudSnooze tags from an instance

Examples:
```bash
# List tags for current instance
snooze tags list

# List tags for a specific instance
snooze tags list i-1234567890abcdef0

# Clear tags from current instance
snooze tags clear
```

### logs

View and manage daemon logs.

```
snooze logs [options]
```

Options:
- `--tail=N, -n N`: Show the last N lines
- `--follow, -f`: Follow the log output (like tail -f)
- `--since=TIME`: Show logs since TIME

Examples:
```bash
# View last 50 log lines
snooze logs --tail=50

# Follow logs in real-time
snooze logs --follow

# View logs since yesterday
snooze logs --since="1 day ago"
```

### export-metrics

Export collected metrics for analysis.

```
snooze export-metrics [options]
```

Options:
- `--since=TIME`: Export metrics since TIME
- `--format=FORMAT, -f FORMAT`: Output format (json, csv)
- `--output=FILE, -o FILE`: Write to FILE instead of stdout

Examples:
```bash
# Export all metrics as JSON
snooze export-metrics --format=json --output=metrics.json

# Export last week's metrics as CSV
snooze export-metrics --since="1 week ago" --format=csv --output=metrics.csv
```

### test-permissions

Test cloud provider permissions required by CloudSnooze.

```
snooze test-permissions
```

Output example:
```
Testing cloud provider permissions...
✓ Can describe instances
✓ Can stop current instance
✓ Can create tags
✓ Can access CloudWatch (optional)
All required permissions are configured correctly.
```

## Exit Codes

The CLI tool uses the following exit codes:

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Permission error
- `4`: Communication error with daemon
- `5`: Command syntax error
