// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/scttfrdmn/cloudsnooze/daemon/accelerator"
	"github.com/scttfrdmn/cloudsnooze/daemon/api"
	"github.com/scttfrdmn/cloudsnooze/daemon/cloud"
	"github.com/scttfrdmn/cloudsnooze/daemon/cloud/aws"
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
	"github.com/scttfrdmn/cloudsnooze/daemon/monitor"
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
		config.GPUThresholdPercent,
		config.InputIdleThresholdSecs,
		config.NaptimeMinutes,
		config.CheckIntervalSeconds*1000,
		config.GPUMonitoringEnabled,
	)
	
	// Initialize GPU service and inject it into the system monitor
	if config.GPUMonitoringEnabled {
		// Use the factory function to create a GPU service
		gpuService := accelerator.CreateGPUService()
		// Initialize the service
		if err := gpuService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize GPU service: %v", err)
		}
		// Inject the service into the system monitor
		systemMonitor.SetGPUService(gpuService)
	}
	
	// Set up AWS cloud provider
	awsConfig := aws.Config{
		Region:             config.AWSRegion,
		EnableTags:         config.EnableInstanceTags,
		TaggingPrefix:      config.TaggingPrefix,
		DetailedTags:       config.DetailedInstanceTags,
		TagPollingEnabled:  config.TagPollingEnabled,
		TagPollingInterval: config.TagPollingIntervalSecs,
		EnableCloudWatch:   config.Logging.EnableCloudWatch,
		CloudWatchLogGroup: config.Logging.CloudWatchLogGroup,
	}
	cloudProvider, err := cloud.CreateProvider(cloud.AWS, awsConfig)
	if err != nil {
		log.Printf("Warning: Failed to create cloud provider: %v", err)
	}

	// Set up API socket server
	socketServer, err := api.NewSocketServer(*socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket server: %v", err)
	}

	// Register command handlers
	registerCommandHandlers(socketServer, systemMonitor, config, cloudProvider)

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
	go monitorLoop(systemMonitor, cloudProvider, config, done)

	// Wait for signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Stop the monitoring loop
	done <- true

	// Clean up
	if err := socketServer.Stop(); err != nil {
		log.Printf("Error stopping socket server: %v", err)
	}
	
	// Stop tag polling if the provider supports it
	// This is a type assertion to check if our provider is specifically an AWS provider
	if provider, ok := cloudProvider.(interface{ StopTagPolling() }); ok {
		provider.StopTagPolling()
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

func monitorLoop(systemMonitor *monitor.SystemMonitor, cloudProvider common.CloudProvider, config Config, done chan bool) {
	ticker := time.NewTicker(time.Duration(config.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Try to verify permissions at startup
	if cloudProvider != nil {
		log.Printf("Verifying cloud provider permissions...")
		if hasPerms, err := cloudProvider.VerifyPermissions(); err != nil {
			log.Printf("Warning: Failed to verify cloud provider permissions: %v", err)
		} else if !hasPerms {
			log.Printf("Warning: Insufficient permissions to stop instances")
		} else {
			log.Printf("Cloud provider permissions verified successfully")
		}
	}

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
				
				// Actually stop the instance via cloud provider
				if cloudProvider != nil {
					// Create a snooze event for logging
					event := &monitor.SnoozeEvent{
						Timestamp:   time.Now(),
						Reason:      reason,
						Metrics:     metrics,
						NaptimeMins: config.NaptimeMinutes,
					}
					
					// Get instance info if possible
					instanceInfo, err := cloudProvider.GetInstanceInfo()
					if err != nil {
						log.Printf("Warning: Failed to get instance info: %v", err)
					} else {
						event.InstanceID = instanceInfo.ID
						event.InstanceType = instanceInfo.Type
						event.Region = instanceInfo.Region
					}
					
					// Log the snooze event (ideally this would go to a proper logging system)
					eventJSON, _ := json.MarshalIndent(event, "", "  ")
					log.Printf("Snooze event: %s", string(eventJSON))
					
					// Stop the instance
					err = cloudProvider.StopInstance(reason, metrics)
					if err != nil {
						log.Printf("Failed to stop instance: %v", err)
					} else {
						log.Printf("Successfully initiated instance stop")
					}
				} else {
					log.Printf("No cloud provider available, would stop instance with reason: %s", reason)
				}
				
				// Reset idle state after stopping instance
				systemMonitor.ResetIdleState()
			}
		}
	}
}

func registerCommandHandlers(server *api.SocketServer, systemMonitor *monitor.SystemMonitor, config Config, cloudProvider common.CloudProvider) {
	
	// STATUS command
	server.RegisterHandler("STATUS", func(params map[string]interface{}) (interface{}, error) {
		metrics := systemMonitor.GetLastMetrics()
		
		var idleSinceStr string
		if idleSince := systemMonitor.GetIdleSince(); idleSince != nil {
			idleSinceStr = idleSince.Format(time.RFC3339)
		}
		
		shouldSnooze, reason := systemMonitor.ShouldSnooze()
		
		// Get instance info if available
		var instanceInfo *common.InstanceInfo
		if cloudProvider != nil {
			instanceInfo, _ = cloudProvider.GetInstanceInfo()
		}
		
		return map[string]interface{}{
			"metrics":       metrics,
			"idle_since":    idleSinceStr,
			"should_snooze": shouldSnooze,
			"snooze_reason": reason,
			"version":       version,
			"instance_info": instanceInfo,
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