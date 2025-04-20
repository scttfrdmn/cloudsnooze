// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// InputMonitor tracks user input activity
type InputMonitor struct {
	lastActivity time.Time
	platform     string
}

// NewInputMonitor creates a new input activity monitor
func NewInputMonitor() *InputMonitor {
	return &InputMonitor{
		lastActivity: time.Now(),
		platform:     runtime.GOOS,
	}
}

// GetIdleSeconds returns the number of seconds since the last input activity
func (m *InputMonitor) GetIdleSeconds() (int, error) {
	var idleSeconds int
	var err error

	switch m.platform {
	case "linux":
		idleSeconds, err = m.getLinuxIdleTime()
	case "darwin":
		idleSeconds, err = m.getMacIdleTime()
	default:
		return 0, fmt.Errorf("unsupported platform: %s", m.platform)
	}

	if err != nil {
		return 0, err
	}

	return idleSeconds, nil
}

// getLinuxIdleTime gets idle time on Linux systems using xprintidle
func (m *InputMonitor) getLinuxIdleTime() (int, error) {
	// Check if X11 is running
	if _, err := exec.LookPath("xprintidle"); err != nil {
		return 0, fmt.Errorf("xprintidle not found, install it for input monitoring")
	}

	// Get idle time in milliseconds
	cmd := exec.Command("xprintidle")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run xprintidle: %v", err)
	}

	// Parse output
	idleMs, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse idle time: %v", err)
	}

	return int(idleMs / 1000), nil
}

// getMacIdleTime gets idle time on macOS using ioreg
func (m *InputMonitor) getMacIdleTime() (int, error) {
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run ioreg: %v", err)
	}

	// Parse output to find HIDIdleTime
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "HIDIdleTime") {
			parts := strings.Split(line, " = ")
			if len(parts) != 2 {
				continue
			}

			// Value is in nanoseconds
			idleNs, err := strconv.ParseInt(strings.Trim(parts[1], " "), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse idle time: %v", err)
			}

			// Convert to seconds
			return int(idleNs / 1000000000), nil
		}
	}

	return 0, fmt.Errorf("HIDIdleTime not found in ioreg output")
}