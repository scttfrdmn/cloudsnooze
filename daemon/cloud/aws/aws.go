// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/scttfrdmn/cloudsnooze/daemon/cloud"
	"github.com/scttfrdmn/cloudsnooze/daemon/monitor"
)

const (
	// MetadataBaseURL is the base URL for the AWS EC2 instance metadata service
	MetadataBaseURL = "http://169.254.169.254/latest/meta-data"
	// InstanceIdentityURL is the URL for the instance identity document
	InstanceIdentityURL = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	// IMDSv2 token TTL in seconds
	tokenTTL = "60"
)

// Config holds AWS-specific configuration
type Config struct {
	Region              string
	EnableTags          bool
	TaggingPrefix       string
	DetailedTags        bool  // Whether to add detailed tags
	TagPollingEnabled   bool  // Whether to poll for tags
	TagPollingInterval  int   // How often to poll for tags (in seconds)
	EnableCloudWatch    bool
	CloudWatchLogGroup  string
	EnableRestartFlag   bool  // Whether to add RestartAllowed flag
	AllowedRestarterIDs []string // List of service IDs allowed to restart instances
}

// Provider implements the cloud.Provider interface for AWS
type Provider struct {
	config Config
	instanceInfo *cloud.InstanceInfo
	// Add a channel for tag polling if needed in the future
	tagPollingDone chan bool
}

// NewProvider creates a new AWS cloud provider
func NewProvider(config Config) *Provider {
	provider := &Provider{
		config: config,
		tagPollingDone: make(chan bool),
	}
	
	// Start tag polling if enabled
	if config.TagPollingEnabled {
		go provider.startTagPolling()
	}
	
	return provider
}

// VerifyPermissions checks if the daemon has sufficient permissions to stop instances
func (p *Provider) VerifyPermissions() (bool, error) {
	// Get instance info to verify IMDS access
	_, err := p.GetInstanceInfo()
	if err != nil {
		return false, fmt.Errorf("failed to get instance info: %v", err)
	}

	// TODO: Make a dry-run call to StopInstances API to verify permissions
	// This would require implementing the AWS SDK
	
	// For now, just return true if we can access instance metadata
	return true, nil
}

// GetInstanceInfo retrieves information about the current instance
func (p *Provider) GetInstanceInfo() (*cloud.InstanceInfo, error) {
	// Return cached info if available
	if p.instanceInfo != nil {
		return p.instanceInfo, nil
	}

	// Get IMDSv2 token
	token, err := getIMDSv2Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get IMDSv2 token: %v", err)
	}

	// Get instance ID
	instanceID, err := getMetadata(token, "instance-id")
	if err != nil {
		return nil, fmt.Errorf("failed to get instance ID: %v", err)
	}

	// Get instance type
	instanceType, err := getMetadata(token, "instance-type")
	if err != nil {
		return nil, fmt.Errorf("failed to get instance type: %v", err)
	}

	// Use configured region or try to get it from IMDS
	region := p.config.Region
	if region == "" {
		// This field is available in the identity document, 
		// but we're using a simpler approach for this implementation
		availabilityZone, err := getMetadata(token, "placement/availability-zone")
		if err != nil {
			return nil, fmt.Errorf("failed to get availability zone: %v", err)
		}
		// Region is AZ minus the last character
		if len(availabilityZone) > 1 {
			region = availabilityZone[:len(availabilityZone)-1]
		}
	}

	// Create and cache instance info
	p.instanceInfo = &cloud.InstanceInfo{
		ID:       instanceID,
		Type:     instanceType,
		Region:   region,
		Provider: "aws",
		Tags:     make(map[string]string),
	}

	return p.instanceInfo, nil
}

// StopInstance stops the current instance
func (p *Provider) StopInstance(reason string, metrics monitor.SystemMetrics) error {
	instanceInfo, err := p.GetInstanceInfo()
	if err != nil {
		return fmt.Errorf("failed to get instance info: %v", err)
	}

	// Add tags if enabled
	if p.config.EnableTags {
		// Base tags that are always added
		tags := map[string]string{
			fmt.Sprintf("%s:StopTimestamp", p.config.TaggingPrefix): time.Now().Format(time.RFC3339),
			fmt.Sprintf("%s:StopReason", p.config.TaggingPrefix): reason,
			fmt.Sprintf("%s:Status", p.config.TaggingPrefix): "Stopped",
		}
		
		// Add restart capability if enabled
		if p.config.EnableRestartFlag {
			tags[fmt.Sprintf("%s:RestartAllowed", p.config.TaggingPrefix)] = "true"
			
			// Add allowed restarter IDs if configured
			if len(p.config.AllowedRestarterIDs) > 0 {
				tags[fmt.Sprintf("%s:AllowedRestarters", p.config.TaggingPrefix)] = strings.Join(p.config.AllowedRestarterIDs, ",")
			}
		}
		
		// Add detailed tags if enabled
		if p.config.DetailedTags {
			// Add system metrics
			tags[fmt.Sprintf("%s:CPUPercent", p.config.TaggingPrefix)] = fmt.Sprintf("%.2f", metrics.CPUPercent)
			tags[fmt.Sprintf("%s:MemoryPercent", p.config.TaggingPrefix)] = fmt.Sprintf("%.2f", metrics.MemoryPercent)
			tags[fmt.Sprintf("%s:NetworkKBps", p.config.TaggingPrefix)] = fmt.Sprintf("%.2f", metrics.NetworkKBps)
			tags[fmt.Sprintf("%s:DiskIOKBps", p.config.TaggingPrefix)] = fmt.Sprintf("%.2f", metrics.DiskIOKBps)
			tags[fmt.Sprintf("%s:InputIdleSecs", p.config.TaggingPrefix)] = fmt.Sprintf("%d", metrics.InputIdleSecs)
			
			// Add GPU metrics if available
			if len(metrics.GPUMetrics) > 0 {
				// Summarize GPU metrics - average utilization of all GPUs
				var totalGPUUtil float64
				for _, gpu := range metrics.GPUMetrics {
					totalGPUUtil += gpu.Utilization
				}
				avgGPUUtil := totalGPUUtil / float64(len(metrics.GPUMetrics))
				tags[fmt.Sprintf("%s:GPUPercent", p.config.TaggingPrefix)] = fmt.Sprintf("%.2f", avgGPUUtil)
				tags[fmt.Sprintf("%s:GPUCount", p.config.TaggingPrefix)] = fmt.Sprintf("%d", len(metrics.GPUMetrics))
			}
			
			// System information
			tags[fmt.Sprintf("%s:NaptimeMinutes", p.config.TaggingPrefix)] = fmt.Sprintf("%d", p.config.TagPollingInterval/60)
			// This is a placeholder since we don't have this info in the Provider struct
			// In a real implementation, this would come from the config
			tags[fmt.Sprintf("%s:InstanceType", p.config.TaggingPrefix)] = instanceInfo.Type
			tags[fmt.Sprintf("%s:Region", p.config.TaggingPrefix)] = instanceInfo.Region
		}
		
		// Tag the instance before stopping it
		if err := p.TagInstance(tags); err != nil {
			// Log but continue with stopping
			fmt.Printf("Warning: Failed to tag instance: %v\n", err)
		}
	}

	// TODO: Implement actual instance stopping using AWS SDK
	// For now, log what we would do
	fmt.Printf("Would stop AWS EC2 instance %s (type: %s) in region %s with reason: %s\n", 
		instanceInfo.ID, instanceInfo.Type, instanceInfo.Region, reason)

	return nil
}

// TagInstance adds tags to the current instance
func (p *Provider) TagInstance(tags map[string]string) error {
	instanceInfo, err := p.GetInstanceInfo()
	if err != nil {
		return fmt.Errorf("failed to get instance info: %v", err)
	}

	// TODO: Implement actual tagging using AWS SDK
	// For now, log what we would do
	fmt.Printf("Would add the following tags to instance %s:\n", instanceInfo.ID)
	for key, value := range tags {
		fmt.Printf("  %s: %s\n", key, value)
	}

	return nil
}

// GetExternalTags checks for tags from external systems that might control this instance
func (p *Provider) GetExternalTags() (map[string]string, error) {
	instanceInfo, err := p.GetInstanceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance info: %v", err)
	}

	// TODO: Implement actual tag retrieval using AWS SDK
	// For now, return empty tags
	fmt.Printf("Would check for external tags on instance %s\n", instanceInfo.ID)
	
	// In a real implementation, this would call DescribeTags API
	// and filter for tags with external tool prefixes
	
	return make(map[string]string), nil
}

// startTagPolling starts a goroutine to poll for external tags
func (p *Provider) startTagPolling() {
	ticker := time.NewTicker(time.Duration(p.config.TagPollingInterval) * time.Second)
	defer ticker.Stop()
	
	fmt.Printf("Starting tag polling every %d seconds\n", p.config.TagPollingInterval)
	
	for {
		select {
		case <-p.tagPollingDone:
			fmt.Println("Stopping tag polling")
			return
		case <-ticker.C:
			tags, err := p.GetExternalTags()
			if err != nil {
				fmt.Printf("Error polling for tags: %v\n", err)
				continue
			}
			
			// Process external tags
			if len(tags) > 0 {
				fmt.Printf("Found %d external tags\n", len(tags))
				// Handle commands from external systems
				// This could include wake-up commands, status changes, etc.
			}
		}
	}
}

// StopTagPolling stops the tag polling goroutine
func (p *Provider) StopTagPolling() {
	if p.config.TagPollingEnabled {
		p.tagPollingDone <- true
	}
}

// getIMDSv2Token gets a token for IMDSv2 requests
func getIMDSv2Token() (string, error) {
	req, err := http.NewRequest(http.MethodPut, "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", tokenTTL)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get IMDSv2 token, status: %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

// getMetadata gets metadata from the instance metadata service
func getMetadata(token, path string) (string, error) {
	url := fmt.Sprintf("%s/%s", MetadataBaseURL, path)
	
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metadata at path %s, status: %d", path, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// AWSFactory implements the cloud.ProviderFactory interface
type AWSFactory struct{}

// CreateProvider creates an AWS provider with the given config
func (f *AWSFactory) CreateProvider(config interface{}) (cloud.Provider, error) {
	awsConfig, ok := config.(Config)
	if !ok {
		return nil, errors.New("invalid AWS configuration")
	}
	
	return NewProvider(awsConfig), nil
}