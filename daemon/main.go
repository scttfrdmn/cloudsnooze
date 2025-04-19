package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/cloudsnooze/daemon/api"
	"github.com/yourusername/cloudsnooze/daemon/monitor"
)

var (
	configFile  = flag.String("config", "/etc/snooze/snooze.json", "Path to configuration file")
	socketPath  = flag.String("socket", api.DefaultSocketPath, "Path to Unix socket")
	showVersion = flag.Bool("version", false, "Show version and exit")
)

const version = "0.1.0"

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("CloudSnooze daemon v%s\n", version)
		return
	}

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up system monitor
	systemMonitor := monitor.NewSystemMonitor(
		config.CPUThresholdPercent,
		config.MemoryThresholdPercent,
		config.NetworkThresholdKBps,
		config.DiskIOThresholdKBps,
		config.InputIdleThresholdSecs,
		config.NaptimeMinutes,
		config.CheckIntervalSeconds*1000,
	)

	// Set up API socket server
	socketServer, err := api.NewSocketServer(*socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket server: %v", err)
	}

	// Register command handlers
	registerCommandHandlers(socketServer, systemMonitor, config)

	// Start socket server in a goroutine
	go func() {
		if err := socketServer.Start(); err != nil {
			log.Fatalf("Socket server error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring loop
	done := make(chan bool)
	go monitorLoop(systemMonitor, config, done)

	// Wait for signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Stop the monitoring loop
	done <- true

	// Clean up
	if err := socketServer.Stop(); err != nil {
		log.Printf("Error stopping socket server: %v", err)
	}
}

func loadConfig(path string) (Config, error) {
	// Start with default config
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create config directory if it doesn't exist
		configDir := "/etc/snooze"
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return config, fmt.Errorf("failed to create config directory: %v", err)
		}

		// Write default config if file doesn't exist
		defaultConfig, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, fmt.Errorf("failed to serialize default config: %v", err)
		}

		if err := os.WriteFile(path, defaultConfig, 0644); err != nil {
			return config, fmt.Errorf("failed to write default config: %v", err)
		}

		log.Printf("Created default configuration at %s", path)
		return config, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func monitorLoop(systemMonitor *monitor.SystemMonitor, config Config, done chan bool) {
	ticker := time.NewTicker(time.Duration(config.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			metrics, err := systemMonitor.CollectMetrics()
			if err != nil {
				log.Printf("Error collecting metrics: %v", err)
				continue
			}

			shouldSnooze, reason := systemMonitor.ShouldSnooze()
			if shouldSnooze {
				log.Printf("Instance should be snoozed: %s", reason)
				// TODO: Implement actual instance stopping via cloud provider API
				
				// For now, we just log that we would stop the instance
				log.Printf("Would stop instance with reason: %s", reason)
				
				// Reset idle state after "stopping" instance
				systemMonitor.ResetIdleState()
			}
		}
	}
}

func registerCommandHandlers(server *api.SocketServer, systemMonitor *monitor.SystemMonitor, config Config) {
	// STATUS command
	server.RegisterHandler("STATUS", func(params map[string]interface{}) (interface{}, error) {
		metrics := systemMonitor.GetLastMetrics()
		
		var idleSinceStr string
		if idleSince := systemMonitor.GetIdleSince(); idleSince != nil {
			idleSinceStr = idleSince.Format(time.RFC3339)
		}
		
		shouldSnooze, reason := systemMonitor.ShouldSnooze()
		
		return map[string]interface{}{
			"metrics":      metrics,
			"idle_since":   idleSinceStr,
			"should_snooze": shouldSnooze,
			"snooze_reason": reason,
			"version":      version,
		}, nil
	})
	
	// CONFIG_GET command
	server.RegisterHandler("CONFIG_GET", func(params map[string]interface{}) (interface{}, error) {
		return config, nil
	})
	
	// CONFIG_SET command - placeholder
	server.RegisterHandler("CONFIG_SET", func(params map[string]interface{}) (interface{}, error) {
		// TODO: Implement configuration updates
		return map[string]interface{}{"updated": false, "message": "Not implemented yet"}, nil
	})
	
	// HISTORY command - placeholder
	server.RegisterHandler("HISTORY", func(params map[string]interface{}) (interface{}, error) {
		// TODO: Implement history retrieval
		return []interface{}{}, nil
	})
}