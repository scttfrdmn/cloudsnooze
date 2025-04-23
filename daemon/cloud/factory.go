// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"

	"github.com/scttfrdmn/cloudsnooze/daemon/common"
	cloudplugin "github.com/scttfrdmn/cloudsnooze/daemon/plugin/cloud"
)

// ProviderType represents a cloud provider type
type ProviderType string

const (
	// AWS is the Amazon Web Services provider
	AWS ProviderType = "aws"
	// GCP is the Google Cloud Platform provider
	GCP ProviderType = "gcp"
	// Azure is the Microsoft Azure provider
	Azure ProviderType = "azure"
)

// DetectProvider attempts to detect which cloud provider we're running on
// This is now a wrapper around the plugin-based detection for backward compatibility
func DetectProvider() (ProviderType, error) {
	plugin, err := cloudplugin.Registry.DetectProvider()
	if err != nil {
		return ProviderType(""), err
	}
	
	// Convert plugin ID to ProviderType
	return ProviderType(plugin.Info().ID), nil
}

// CreateProvider creates a new provider of the specified type with the given config
// This is now a wrapper around the plugin-based provider creation for backward compatibility
func CreateProvider(providerType ProviderType, config interface{}) (common.CloudProvider, error) {
	// Get the provider plugin
	plugin, exists := cloudplugin.Registry.GetProvider(string(providerType))
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
	
	// Initialize the plugin if not already initialized
	if !plugin.IsRunning() {
		if err := plugin.Init(nil); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin: %v", err)
		}
		if err := plugin.Start(); err != nil {
			return nil, fmt.Errorf("failed to start plugin: %v", err)
		}
	}
	
	// Create a provider instance
	return plugin.CreateProvider(config)
}