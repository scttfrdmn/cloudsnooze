// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/cloudsnooze/daemon/api"
)

// StatusCommand handler for the 'status' CLI command
type StatusCommand struct {
	Watch    bool
	Interval int
	Json     bool
	Debug    bool
}

// NewStatusCommand creates a new status command
func NewStatusCommand() *StatusCommand {
	return &StatusCommand{
		Watch:    false,
		Interval: 5, // Default 5-second refresh
		Json:     false,
		Debug:    false,
	}
}

// Execute runs the status command
func (c *StatusCommand) Execute(client *api.SocketClient) error {
	// If watch mode is enabled, run in a loop
	if c.Watch {
		ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)
		defer ticker.Stop()

		// Clear screen and show status
		fmt.Print("\033[H\033[2J") // ANSI escape codes to clear screen
		if err := c.showStatus(client); err != nil {
			return err
		}

		for {
			select {
			case <-ticker.C:
				// Clear screen and show status
				fmt.Print("\033[H\033[2J")
				if err := c.showStatus(client); err != nil {
					return err
				}
			}
		}
	} else {
		// Single display
		return c.showStatus(client)
	}
}

// showStatus displays the current system status
func (c *StatusCommand) showStatus(client *api.SocketClient) error {
	if c.Json {
		jsonData, err := GetStatusJson(client)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonData))
		return nil
	}
	
	formatted, err := FormatStatusOutput(client)
	if err != nil {
		return err
	}
	
	fmt.Println(formatted)
	
	if c.Watch {
		fmt.Printf("\nWatch mode: refreshing every %d seconds (press Ctrl+C to exit)\n", c.Interval)
	}
	
	return nil
}

// Help returns the help text for the status command
func (c *StatusCommand) Help() string {
	return `Usage: snooze status [options]

Display the current system status, including metrics and daemon information.

Options:
  --watch, -w        Continuously update the display
  --interval=N, -i N Refresh interval in seconds when using watch mode (default: 5)
  --json, -j         Output in JSON format
  --debug, -d        Include additional debug information

Examples:
  snooze status
  snooze status --watch
  snooze status --watch --interval=10
  snooze status --json
  snooze status --debug`
}

// GetStatusJson retrieves the status and returns it as JSON
func GetStatusJson(client *api.SocketClient) ([]byte, error) {
	result, err := client.SendCommand("STATUS", nil)
	if err != nil {
		return nil, err
	}
	
	return json.MarshalIndent(result, "", "  ")
}

// FormatStatusOutput formats the status output for human-readable display
func FormatStatusOutput(client *api.SocketClient) (string, error) {
	result, err := client.SendCommand("STATUS", nil)
	if err != nil {
		return "", err
	}
	
	// Convert result to a map
	data, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	
	// Extract metrics
	metrics, ok := data["metrics"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("metrics not found in response")
	}
	
	// Build formatted output
	var output string
	output += "CloudSnooze Status\n"
	output += "------------------\n"
	output += fmt.Sprintf("Version: %s\n", data["version"])
	
	// Display idle status
	if idleSince, ok := data["idle_since"].(string); ok && idleSince != "" {
		t, err := time.Parse(time.RFC3339, idleSince)
		if err == nil {
			output += fmt.Sprintf("Idle since: %s (%s ago)\n", 
				t.Format("2006-01-02 15:04:05"),
				time.Since(t).Round(time.Second))
		} else {
			output += fmt.Sprintf("Idle since: %s\n", idleSince)
		}
	} else {
		output += "System is active\n"
	}
	
	// Display should snooze
	if shouldSnooze, ok := data["should_snooze"].(bool); ok {
		if shouldSnooze {
			output += fmt.Sprintf("Status: WILL SNOOZE - %s\n", data["snooze_reason"])
		} else {
			output += fmt.Sprintf("Status: %s\n", data["snooze_reason"])
		}
	}
	
	output += "\nCurrent metrics:\n"
	output += fmt.Sprintf("  - CPU: %.1f%%\n", metrics["cpu_percent"])
	output += fmt.Sprintf("  - Memory: %.1f%%\n", metrics["memory_percent"])
	output += fmt.Sprintf("  - Network: %.1f KB/s\n", metrics["network_kbps"])
	output += fmt.Sprintf("  - Disk I/O: %.1f KB/s\n", metrics["disk_io_kbps"])
	output += fmt.Sprintf("  - Input idle: %ds\n", int(metrics["input_idle_secs"].(float64)))
	
	// Display GPU metrics if available
	if gpuMetrics, ok := metrics["gpu_metrics"].([]interface{}); ok && len(gpuMetrics) > 0 {
		output += "\nGPU Metrics:\n"
		for i, gpu := range gpuMetrics {
			gpuData := gpu.(map[string]interface{})
			output += fmt.Sprintf("  - GPU %d [%s %s]: %.1f%% utilized, %.1f MB / %.1f MB memory\n",
				i+1, 
				gpuData["type"], 
				gpuData["name"],
				gpuData["utilization"],
				float64(gpuData["memory_used"].(float64))/1024/1024,
				float64(gpuData["memory_total"].(float64))/1024/1024)
		}
	}
	
	// Display instance info if available
	if instanceInfo, ok := data["instance_info"].(map[string]interface{}); ok {
		output += "\nInstance Information:\n"
		output += fmt.Sprintf("  - ID: %s\n", instanceInfo["ID"])
		output += fmt.Sprintf("  - Type: %s\n", instanceInfo["Type"])
		output += fmt.Sprintf("  - Region: %s\n", instanceInfo["Region"])
		output += fmt.Sprintf("  - Provider: %s\n", instanceInfo["Provider"])
	}
	
	return output, nil
}