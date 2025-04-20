// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package integration_test

import (
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// TestInstanceStop tests the ability to stop an EC2 instance
func TestInstanceStop(t *testing.T) {
	// Skip if not running in a test environment
	instanceID := os.Getenv("CLOUDSNOOZE_TEST_INSTANCE_ID")
	if instanceID == "" {
		t.Skip("Skipping integration test: No test instance ID provided")
	}

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(
		config.WithRegion(os.Getenv("CLOUDSNOOZE_TEST_REGION")),
	)
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create EC2 client
	client := ec2.NewFromConfig(cfg)

	// Check initial state (should be running)
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeResult, err := client.DescribeInstances(nil, describeInput)
	if err != nil {
		t.Fatalf("Failed to describe instance: %v", err)
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		t.Fatalf("Test instance not found: %s", instanceID)
	}

	instance := describeResult.Reservations[0].Instances[0]
	initialState := instance.State.Name

	// Should be running initially
	if initialState != "running" {
		t.Logf("Warning: Test instance is not in 'running' state. Current state: %s", initialState)
	}

	// Add a tag that we're about to stop it via CloudSnooze
	_, err = client.CreateTags(nil, &ec2.CreateTagsInput{
		Resources: []string{instanceID},
		Tags: []ec2.Tag{
			{
				Key:   aws.String("StoppedBy"),
				Value: aws.String("CloudSnooze-Test"),
			},
			{
				Key:   aws.String("StopReason"),
				Value: aws.String("Integration Test"),
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to tag instance: %v", err)
	}

	// Stop the instance
	stopInput := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err = client.StopInstances(nil, stopInput)
	if err != nil {
		t.Fatalf("Failed to stop instance: %v", err)
	}

	// Wait for the instance to stop (with timeout)
	t.Log("Waiting for instance to stop...")
	waited := 0
	maxWait := 120 // seconds
	stopped := false

	for waited < maxWait {
		describeResult, err := client.DescribeInstances(nil, describeInput)
		if err != nil {
			t.Fatalf("Failed to describe instance after stop attempt: %v", err)
		}

		currentState := describeResult.Reservations[0].Instances[0].State.Name
		if currentState == "stopped" {
			t.Logf("Instance successfully stopped after %d seconds", waited)
			stopped = true
			break
		}

		time.Sleep(5 * time.Second)
		waited += 5
	}

	if !stopped {
		t.Fatalf("Timed out waiting for instance to stop after %d seconds", maxWait)
	}

	// Start the instance again to leave it in the original state
	_, err = client.StartInstances(nil, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		t.Logf("Warning: Failed to restart instance after test: %v", err)
	} else {
		t.Log("Successfully restarted instance after test")
	}
}
