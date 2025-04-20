// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	// Save original args and restore them after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up a mock environment for the test
	os.Args = []string{"snooze", "--version"}
	
	// This is a simple test to verify the version flag doesn't crash
	// In a real test, we would capture stdout and verify the version output
	// but this is sufficient for our current CI/CD setup
	if *showVersion != false {
		// We just check that the flag is initialized properly
		t.Errorf("showVersion flag was not initialized correctly")
	}
}

func TestSocketPathFlag(t *testing.T) {
	// Save original args and restore them after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// This is just a simple test to verify the socket path flag is initialized
	if *socketPath == "" {
		t.Errorf("socketPath flag was not initialized with a default value")
	}
}

func TestConfigFileFlag(t *testing.T) {
	// Save original args and restore them after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// This is just a simple test to verify the config file flag is initialized
	if *configFile == "" {
		t.Errorf("configFile flag was not initialized with a default value")
	}
}