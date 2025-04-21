// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"errors"
	"sync"
)

// Plugin types
const (
	TypeCloudProvider = "cloud-provider"
	// Add more plugin types as needed
)

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	ID           string            // Unique identifier
	Name         string            // Human-readable name
	Type         string            // Plugin type (e.g., "cloud-provider")
	Version      string            // Version string
	Capabilities map[string]bool   // Capabilities this plugin supports
	Author       string            // Plugin author
	Website      string            // Plugin website or repository
	Dependencies []string          // IDs of plugins this plugin depends on
}

// Plugin defines the base interface all plugins must implement
type Plugin interface {
	// Info returns plugin metadata
	Info() PluginInfo
	
	// Init initializes the plugin with configuration
	Init(config interface{}) error
	
	// Start starts the plugin
	Start() error
	
	// Stop gracefully stops the plugin
	Stop() error
	
	// IsRunning returns true if the plugin is running
	IsRunning() bool
}

// PluginRegistry is the global registry of plugins
type PluginRegistry struct {
	plugins map[string]Plugin
	lock    sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry
func (r *PluginRegistry) Register(p Plugin) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	
	info := p.Info()
	if _, exists := r.plugins[info.ID]; exists {
		return errors.New("plugin already registered")
	}
	
	r.plugins[info.ID] = p
	return nil
}

// Get returns a plugin by ID
func (r *PluginRegistry) Get(id string) (Plugin, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	
	p, exists := r.plugins[id]
	return p, exists
}

// GetByType returns all plugins of a specific type
func (r *PluginRegistry) GetByType(pluginType string) []Plugin {
	r.lock.RLock()
	defer r.lock.RUnlock()
	
	var result []Plugin
	for _, p := range r.plugins {
		if p.Info().Type == pluginType {
			result = append(result, p)
		}
	}
	
	return result
}

// Global registry instance
var Registry = NewPluginRegistry()