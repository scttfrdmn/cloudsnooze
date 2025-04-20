// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/cloudsnooze/daemon/api"
)

var (
	socketPath  = flag.String("socket", api.DefaultSocketPath, "Path to Unix socket")
	showVersion = flag.Bool("version", false, "Show version and exit")
	configFile  = flag.String("config", "/etc/snooze/snooze.json", "Path to configuration file")
)

const version = "0.1.0"

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("CloudSnooze CLI v%s\n", version)
		return
	}

	// Check if enough arguments are provided
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Create socket client
	client := api.NewSocketClient(*socketPath)

	// Process command
	command := args[0]
	switch command {
	case "status":
		showStatus(client, args[1:])
	case "config":
		handleConfig(client, args[1:])
	case "history":
		showHistory(client, args[1:])
	case "start", "stop", "restart":
		controlDaemon(client, command)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: snooze [options] command [args]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nCommands:")
	fmt.Println("  status       Show current system status")
	fmt.Println("  config       View or modify configuration")
	fmt.Println("  history      View snooze history")
	fmt.Println("  start        Start the daemon")
	fmt.Println("  stop         Stop the daemon")
	fmt.Println("  restart      Restart the daemon")
	fmt.Println("  help         Show this help message")
	fmt.Println("\nRun 'snooze help command' for more information on a command")
}

func showStatus(client *api.SocketClient, args []string) {
	result, err := client.SendCommand("STATUS", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Convert result to a map
	data, ok := result.(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unexpected response format\n")
		os.Exit(1)
	}

	// Extract metrics
	metrics, ok := data["metrics"].(map[string]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: metrics not found in response\n")
		os.Exit(1)
	}

	// Display status
	fmt.Println("CloudSnooze Status")
	fmt.Println("------------------")
	fmt.Printf("Version: %s\n", data["version"])
	
	// Display idle status
	if idleSince, ok := data["idle_since"].(string); ok && idleSince != "" {
		t, err := time.Parse(time.RFC3339, idleSince)
		if err == nil {
			fmt.Printf("Idle since: %s (%s ago)\n", 
				t.Format("2006-01-02 15:04:05"),
				time.Since(t).Round(time.Second))
		} else {
			fmt.Printf("Idle since: %s\n", idleSince)
		}
	} else {
		fmt.Println("System is active")
	}
	
	// Display should snooze
	if shouldSnooze, ok := data["should_snooze"].(bool); ok {
		if shouldSnooze {
			fmt.Printf("Status: WILL SNOOZE - %s\n", data["snooze_reason"])
		} else {
			fmt.Printf("Status: %s\n", data["snooze_reason"])
		}
	}
	
	fmt.Println("\nCurrent metrics:")
	fmt.Printf("  - CPU: %.1f%%\n", metrics["cpu_percent"])
	fmt.Printf("  - Memory: %.1f%%\n", metrics["memory_percent"])
	fmt.Printf("  - Network: %.1f KB/s\n", metrics["network_kbps"])
	fmt.Printf("  - Disk I/O: %.1f KB/s\n", metrics["disk_io_kbps"])
	fmt.Printf("  - Input idle: %ds\n", int(metrics["input_idle_secs"].(float64)))
	
	// Display GPU metrics if available
	if gpuMetrics, ok := metrics["gpu_metrics"].([]interface{}); ok && len(gpuMetrics) > 0 {
		fmt.Println("\nGPU Metrics:")
		for i, gpu := range gpuMetrics {
			gpuData := gpu.(map[string]interface{})
			fmt.Printf("  - GPU %d [%s %s]: %.1f%% utilized, %.1f MB / %.1f MB memory\n",
				i+1, 
				gpuData["type"], 
				gpuData["name"],
				gpuData["utilization"],
				float64(gpuData["memory_used"].(float64))/1024/1024,
				float64(gpuData["memory_total"].(float64))/1024/1024)
		}
	}
}

func handleConfig(client *api.SocketClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: snooze config [list|get|set|reset|import|export]")
		os.Exit(1)
	}

	action := args[0]
	switch action {
	case "list":
		// Get all configuration
		result, err := client.SendCommand("CONFIG_GET", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Pretty print configuration
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting config: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println(string(jsonData))
		
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: snooze config get <parameter>")
			os.Exit(1)
		}
		
		paramName := args[1]
		
		// Get all configuration
		result, err := client.SendCommand("CONFIG_GET", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Extract the requested parameter
		config, ok := result.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: unexpected response format\n")
			os.Exit(1)
		}
		
		// Try to find the parameter
		value, found := config[paramName]
		if !found {
			fmt.Fprintf(os.Stderr, "Error: parameter '%s' not found\n", paramName)
			os.Exit(1)
		}
		
		fmt.Printf("%v\n", value)
		
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: snooze config set <parameter> <value>")
			os.Exit(1)
		}
		
		paramName := args[1]
		paramValue := args[2]
		
		params := map[string]interface{}{
			"name":  paramName,
			"value": paramValue,
		}
		
		_, err := client.SendCommand("CONFIG_SET", params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Parameter '%s' updated to '%s'\n", paramName, paramValue)
		
	default:
		fmt.Fprintf(os.Stderr, "Unknown config action: %s\n", action)
		fmt.Println("Usage: snooze config [list|get|set|reset|import|export]")
		os.Exit(1)
	}
}

func showHistory(client *api.SocketClient, args []string) {
	// Parse flags for history command
	historyCmd := flag.NewFlagSet("history", flag.ExitOnError)
	limit := historyCmd.Int("limit", 10, "Limit to N entries")
	since := historyCmd.String("since", "", "Show entries since DATE")
	format := historyCmd.String("format", "text", "Output format (text, json, csv)")
	output := historyCmd.String("output", "", "Write output to FILE")
	
	if err := historyCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}
	
	params := map[string]interface{}{
		"limit": *limit,
	}
	
	if *since != "" {
		params["since"] = *since
	}
	
	// Send request
	result, err := client.SendCommand("HISTORY", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Process results
	events, ok := result.([]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unexpected response format\n")
		os.Exit(1)
	}
	
	// Output results
	var output_data []byte
	var output_err error
	
	switch *format {
	case "json":
		output_data, output_err = json.MarshalIndent(events, "", "  ")
	case "csv":
		// TODO: Implement CSV output
		fmt.Fprintf(os.Stderr, "CSV output not implemented yet\n")
		os.Exit(1)
	case "text":
		fallthrough
	default:
		fmt.Printf("Snooze History (last %d events)\n", *limit)
		fmt.Println("-------------------------------")
		
		if len(events) == 0 {
			fmt.Println("No snooze events found")
		} else {
			for i, event := range events {
				e, ok := event.(map[string]interface{})
				if !ok {
					continue
				}
				
				timestamp := e["timestamp"].(string)
				reason := e["reason"].(string)
				
				t, err := time.Parse(time.RFC3339, timestamp)
				if err != nil {
					t = time.Time{}
				}
				
				fmt.Printf("%d. %s - %s\n", i+1, t.Format("2006-01-02 15:04:05"), reason)
			}
		}
	}
	
	if output_err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", output_err)
		os.Exit(1)
	}
	
	// Write to file if specified
	if *output != "" && *format != "text" {
		outputDir := filepath.Dir(*output)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
		
		if err := os.WriteFile(*output, output_data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Output written to %s\n", *output)
	} else if *format != "text" {
		fmt.Println(string(output_data))
	}
}

func controlDaemon(client *api.SocketClient, command string) {
	// TODO: Implement daemon control
	fmt.Printf("Command '%s' not implemented yet\n", command)
}