// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin" // Go standard library plugin package
)

// ExternalPlugin provides access to dynamically loaded Go plugins
type ExternalPlugin struct {
	path        string
	goPlugin    *plugin.Plugin
	pluginInfo  PluginInfo
	pluginImpl  Plugin
	initialized bool
}

// LoadPluginFromFile loads a plugin from a Go plugin file (.so)
func LoadPluginFromFile(path string) (*ExternalPlugin, error) {
	// Load plugin from file
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin at %s: %v", path, err)
	}

	// Look up the Plugin symbol
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %s does not export 'Plugin' symbol: %v", path, err)
	}

	// Assert that the symbol is a Plugin
	pluginImpl, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s does not implement the Plugin interface", path)
	}

	return &ExternalPlugin{
		path:       path,
		goPlugin:   p,
		pluginImpl: pluginImpl,
		pluginInfo: pluginImpl.Info(),
	}, nil
}

// LoadPluginsFromDir loads all plugins from a directory
func LoadPluginsFromDir(dir string) ([]Plugin, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin directory %s does not exist", dir)
	}

	// Find all .so files in the directory
	matches, err := filepath.Glob(filepath.Join(dir, "*.so"))
	if err != nil {
		return nil, fmt.Errorf("error finding plugins in %s: %v", dir, err)
	}

	// Load each plugin
	var plugins []Plugin
	for _, match := range matches {
		plugin, err := LoadPluginFromFile(match)
		if err != nil {
			fmt.Printf("Warning: Failed to load plugin %s: %v\n", match, err)
			continue
		}

		plugins = append(plugins, plugin.pluginImpl)
	}

	return plugins, nil
}

// LoadPluginsFromManifest loads plugins based on manifest files
func LoadPluginsFromManifest(dir string) ([]Plugin, error) {
	// Find all manifest.json files
	manifests, err := filepath.Glob(filepath.Join(dir, "*/manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("error finding plugin manifests in %s: %v", dir, err)
	}

	var plugins []Plugin
	for _, manifestPath := range manifests {
		// Read and parse manifest
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			fmt.Printf("Warning: Failed to read manifest %s: %v\n", manifestPath, err)
			continue
		}

		var manifest PluginInfo
		if err := json.Unmarshal(data, &manifest); err != nil {
			fmt.Printf("Warning: Failed to parse manifest %s: %v\n", manifestPath, err)
			continue
		}

		// Find plugin binary in the same directory
		pluginDir := filepath.Dir(manifestPath)
		pluginPath := filepath.Join(pluginDir, manifest.ID+".so")
		
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Plugin binary not found for manifest %s\n", manifestPath)
			continue
		}

		// Load the plugin
		plugin, err := LoadPluginFromFile(pluginPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load plugin %s: %v\n", pluginPath, err)
			continue
		}

		plugins = append(plugins, plugin.pluginImpl)
	}

	return plugins, nil
}

// LoadExternalPlugins loads plugins from the specified directory and registers them
func LoadExternalPlugins(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory %s does not exist", dir)
	}

	// Try loading from manifests first
	plugins, err := LoadPluginsFromManifest(dir)
	if err != nil {
		fmt.Printf("Warning: Failed to load plugins from manifests: %v\n", err)
		// Fall back to direct .so loading
		plugins, err = LoadPluginsFromDir(dir)
		if err != nil {
			return fmt.Errorf("failed to load plugins from directory: %v", err)
		}
	}

	// Register loaded plugins
	for _, p := range plugins {
		if err := Registry.Register(p); err != nil {
			fmt.Printf("Warning: Failed to register plugin %s: %v\n", p.Info().ID, err)
		}
	}

	return nil
}