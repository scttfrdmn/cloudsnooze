// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"errors"

	"github.com/scttfrdmn/cloudsnooze/daemon/common"
	"github.com/scttfrdmn/cloudsnooze/daemon/plugin"
)

// CloudProviderPlugin extends the base Plugin interface for cloud providers
type CloudProviderPlugin interface {
	plugin.Plugin
	
	// CreateProvider creates a new provider instance with the given configuration
	CreateProvider(config interface{}) (common.CloudProvider, error)
	
	// CanDetect returns true if this plugin can detect if it's running on this cloud provider
	CanDetect() bool
	
	// Detect tries to detect if the current environment is running on this cloud provider
	Detect() (bool, error)
}

// ProviderRegistry provides access to cloud provider plugins
type ProviderRegistry struct {
	registry *plugin.PluginRegistry
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(registry *plugin.PluginRegistry) *ProviderRegistry {
	return &ProviderRegistry{
		registry: registry,
	}
}

// GetProvider gets a cloud provider plugin by ID
func (r *ProviderRegistry) GetProvider(id string) (CloudProviderPlugin, bool) {
	p, exists := r.registry.Get(id)
	if !exists {
		return nil, false
	}
	
	cp, ok := p.(CloudProviderPlugin)
	return cp, ok
}

// GetAllProviders gets all registered cloud provider plugins
func (r *ProviderRegistry) GetAllProviders() []CloudProviderPlugin {
	plugins := r.registry.GetByType(plugin.TypeCloudProvider)
	result := make([]CloudProviderPlugin, 0, len(plugins))
	
	for _, p := range plugins {
		if cp, ok := p.(CloudProviderPlugin); ok {
			result = append(result, cp)
		}
	}
	
	return result
}

// DetectProvider tries to detect which cloud provider the system is running on
func (r *ProviderRegistry) DetectProvider() (CloudProviderPlugin, error) {
	providers := r.GetAllProviders()
	
	for _, p := range providers {
		if !p.CanDetect() {
			continue
		}
		
		isRunningOn, err := p.Detect()
		if err != nil {
			continue
		}
		
		if isRunningOn {
			return p, nil
		}
	}
	
	return nil, errors.New("unable to detect cloud provider")
}

// Global provider registry instance
var Registry = NewProviderRegistry(plugin.Registry)