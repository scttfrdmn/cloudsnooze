// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package aws

import (
	"testing"
	"time"
)

// Test NewProvider function
func TestNewProvider(t *testing.T) {
	config := Config{
		Region:             "us-west-2",
		EnableTags:         true,
		TaggingPrefix:      "cloudsnooze",
		DetailedTags:       true,
		TagPollingEnabled:  true,
		TagPollingInterval: 60,
	}

	provider := NewProvider(config)

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.config.Region != "us-west-2" {
		t.Errorf("Expected Region to be us-west-2, got %s", provider.config.Region)
	}

	if provider.config.EnableTags != true {
		t.Errorf("Expected EnableTags to be true")
	}

	if provider.config.TaggingPrefix != "cloudsnooze" {
		t.Errorf("Expected TaggingPrefix to be cloudsnooze, got %s", provider.config.TaggingPrefix)
	}

	if provider.stopTagPoll == nil {
		t.Fatal("Expected stopTagPoll channel to be initialized")
	}
}

// Test StopTagPolling function
func TestStopTagPolling(t *testing.T) {
	provider := NewProvider(Config{
		Region:             "us-west-2",
		TagPollingEnabled:  true,
		TagPollingInterval: 1,
	})

	// Set up a ticker to simulate polling
	provider.tagPoller = time.NewTicker(1 * time.Second)

	// Call StopTagPolling - this should not block or panic
	provider.StopTagPolling()

	// Verify the ticker was stopped
	if provider.tagPoller != nil {
		t.Errorf("Expected tagPoller to be nil after stopping")
	}
}