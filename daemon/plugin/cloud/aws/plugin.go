// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/scttfrdmn/cloudsnooze/daemon/cloud/aws"
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
	"github.com/scttfrdmn/cloudsnooze/daemon/plugin"
	cloudplugin "github.com/scttfrdmn/cloudsnooze/daemon/plugin/cloud"
)

// AWSPlugin implements the CloudProviderPlugin interface for AWS
type AWSPlugin struct {
	running bool
	config  interface{}
}

// Ensure AWSPlugin implements required interfaces
var _ cloudplugin.CloudProviderPlugin = &AWSPlugin{}
var _ plugin.Plugin = &AWSPlugin{}

// NewAWSPlugin creates a new AWS plugin
func NewAWSPlugin() *AWSPlugin {
	return &AWSPlugin{}
}

// Info returns plugin metadata
func (p *AWSPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:          "aws",
		Name:        "AWS Cloud Provider",
		Type:        plugin.TypeCloudProvider,
		Version:     "1.0.0",
		Capabilities: map[string]bool{
			"tagging":     true,
			"tag-polling": true,
			"restart":     true,
		},
		Author:   "CloudSnooze Contributors",
		Website:  "https://github.com/scttfrdmn/cloudsnooze",
	}
}

// Init initializes the plugin
func (p *AWSPlugin) Init(config interface{}) error {
	p.config = config
	return nil
}

// Start starts the plugin
func (p *AWSPlugin) Start() error {
	p.running = true
	return nil
}

// Stop stops the plugin
func (p *AWSPlugin) Stop() error {
	p.running = false
	return nil
}

// IsRunning returns true if the plugin is running
func (p *AWSPlugin) IsRunning() bool {
	return p.running
}

// CreateProvider creates a new AWS provider instance
func (p *AWSPlugin) CreateProvider(config interface{}) (common.CloudProvider, error) {
	awsConfig, ok := config.(aws.Config)
	if !ok {
		return nil, errors.New("invalid AWS configuration")
	}
	
	return aws.NewProvider(awsConfig), nil
}

// CanDetect returns true as AWS can be detected
func (p *AWSPlugin) CanDetect() bool {
	return true
}

// Detect tries to detect if running on AWS
func (p *AWSPlugin) Detect() (bool, error) {
	// Check if we're in a CI environment
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		// Skip actual detection in CI environments to avoid failures
		log.Println("AWS detection skipped in CI environment")
		return false, nil
	}

	// Check for AWS instance metadata service
	if _, err := os.Stat("/sys/devices/virtual/dmi/id/product_uuid"); err == nil {
		// Check if we can access the instance metadata service
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://169.254.169.254/latest/meta-data")
		if err == nil {
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					log.Printf("Error closing response body: %v", closeErr)
				}
			}()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// Register the plugin
func init() {
	err := plugin.Registry.Register(NewAWSPlugin())
	if err != nil {
		// Don't panic, just log it (in a production environment we'd use a proper logger)
		println("Failed to register AWS plugin:", err.Error())
	}
}