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
	"github.com/scttfrdmn/cloudsnooze/daemon/plugin"
	cloudplugin "github.com/scttfrdmn/cloudsnooze/daemon/plugin/cloud"
	
	// Import all provider plugins to ensure they register themselves
	_ "github.com/scttfrdmn/cloudsnooze/daemon/plugin/cloud/aws"
)

var (
	configFile  = flag.String("config", "/etc/snooze/snooze.json", "Path to configuration file")
	socketPath  = flag.String("socket", api.DefaultSocketPath, "Path to Unix socket")
	showVersion = flag.Bool("version", false, "Show version and exit")
)

const version = "0.1.0"

// initializePlugins initializes and logs information about loaded plugins
func initializePlugins(config *Config) {
	// Built-in plugins are self-registered via their init() functions
	
	// Load external plugins if enabled
	if config != nil && config.PluginsEnabled && config.PluginsDir != "" {
		log.Printf("Loading external plugins from %s...", config.PluginsDir)
		if err := plugin.LoadExternalPlugins(config.PluginsDir); err != nil {
			log.Printf("Warning: Failed to load external plugins: %v", err)
		}
	}
	
	// List all available cloud provider plugins
	providers := cloudplugin.Registry.GetAllProviders()
	if len(providers) == 0 {
		log.Printf("Warning: No cloud provider plugins loaded")
	} else {
		log.Printf("Loaded %d cloud provider plugins:", len(providers))
		for _, p := range providers {
			info := p.Info()
			log.Printf("  - %s (%s) v%s", info.Name, info.ID, info.Version)
		}
	}
}

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
	
	// Initialize plugins with loaded config
	initializePlugins(&config)

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
	
	// Set up cloud provider
	var cloudProvider common.CloudProvider
	var providerType cloud.ProviderType
	
	// Determine provider type from config or auto-detect
	if config.ProviderType == "" {
		// Auto-detect provider
		log.Printf("No provider type specified, attempting auto-detection...")
		detectedType, detectErr := cloud.DetectProvider()
		if detectErr != nil {
			log.Printf("Warning: Failed to auto-detect cloud provider: %v", detectErr)
		} else {
			providerType = detectedType
			log.Printf("Detected cloud provider: %s", providerType)
		}
	} else {
		// Use configured provider
		providerType = cloud.ProviderType(config.ProviderType)
		log.Printf("Using configured cloud provider: %s", providerType)
	}
	
	// Create provider instance based on type
	if providerType != "" {
		switch providerType {
		case cloud.AWS:
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
			cloudProvider, err = cloud.CreateProvider(providerType, awsConfig)
			if err != nil {
				log.Printf("Warning: Failed to create AWS cloud provider: %v", err)
			}
		default:
			log.Printf("Warning: Unsupported cloud provider type: %s", providerType)
		}
	} else {
		log.Printf("No cloud provider available, running in local mode")
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
	
	// Stop all running plugins
	if config.PluginsEnabled {
		log.Println("Stopping all plugins...")
		providers := cloudplugin.Registry.GetAllProviders()
		for _, p := range providers {
			if p.IsRunning() {
				info := p.Info()
				log.Printf("Stopping plugin: %s (%s)", info.Name, info.ID)
				if err := p.Stop(); err != nil {
					log.Printf("Error stopping plugin %s: %v", info.ID, err)
				}
			}
		}
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
	
	// PLUGINS_LIST command
	server.RegisterHandler("PLUGINS_LIST", func(params map[string]interface{}) (interface{}, error) {
		providers := cloudplugin.Registry.GetAllProviders()
		
		var result []map[string]interface{}
		for _, p := range providers {
			info := p.Info()
			result = append(result, map[string]interface{}{
				"id":           info.ID,
				"name":         info.Name,
				"type":         info.Type,
				"version":      info.Version,
				"capabilities": info.Capabilities,
				"author":       info.Author,
				"website":      info.Website,
				"is_running":   p.IsRunning(),
			})
		}
		
		return result, nil
	})
}