// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
)

const (
	// How often to refresh the token
	tokenTTL = "300"
)

// Config holds the AWS provider configuration
type Config struct {
	Region             string
	EnableTags         bool
	TaggingPrefix      string
	DetailedTags       bool
	TagPollingEnabled  bool
	TagPollingInterval int
	EnableCloudWatch   bool
	CloudWatchLogGroup string
}

// AWSProvider is an implementation of CloudProvider for AWS
type AWSProvider struct {
	config     Config
	client     *ec2.Client
	tagPoller  *time.Ticker
	stopTagPoll chan struct{}
	instanceID string
	region     string
	instanceType string
	lock       sync.RWMutex
}

// NewProvider creates a new AWS provider instance
func NewProvider(config Config) *AWSProvider {
	return &AWSProvider{
		config:     config,
		stopTagPoll: make(chan struct{}),
	}
}

// Initialize sets up the AWS provider
func (p *AWSProvider) Initialize() error {
	// Load default AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.config.Region))
	if err != nil {
		return fmt.Errorf("error loading AWS config: %v", err)
	}

	// Create EC2 client
	p.client = ec2.NewFromConfig(cfg)

	// Get instance ID and region info
	if err := p.loadInstanceInfo(); err != nil {
		return fmt.Errorf("error loading instance info: %v", err)
	}

	// Start tag polling if enabled
	if p.config.TagPollingEnabled && p.config.TagPollingInterval > 0 {
		interval := time.Duration(p.config.TagPollingInterval) * time.Second
		p.tagPoller = time.NewTicker(interval)
		go p.pollTags()
	}

	return nil
}

// StopInstance stops the EC2 instance
func (p *AWSProvider) StopInstance(reason string, metrics common.SystemMetrics) error {
	// Get the instance ID
	instanceID, err := p.getInstanceID()
	if err != nil {
		return fmt.Errorf("error getting instance ID: %v", err)
	}

	// Apply tags if enabled
	if p.config.EnableTags {
		// Create basic tags
		tags := []types.Tag{
			{
				Key:   aws.String(fmt.Sprintf("%s:stopped_at", p.config.TaggingPrefix)),
				Value: aws.String(time.Now().Format(time.RFC3339)),
			},
			{
				Key:   aws.String(fmt.Sprintf("%s:reason", p.config.TaggingPrefix)),
				Value: aws.String(reason),
			},
		}

		// Add detailed metrics tags if enabled
		if p.config.DetailedTags {
			tags = append(tags, 
				types.Tag{
					Key:   aws.String(fmt.Sprintf("%s:cpu_percent", p.config.TaggingPrefix)),
					Value: aws.String(fmt.Sprintf("%.2f", metrics.CPUUsage)),
				},
				types.Tag{
					Key:   aws.String(fmt.Sprintf("%s:memory_percent", p.config.TaggingPrefix)),
					Value: aws.String(fmt.Sprintf("%.2f", metrics.MemoryUsage)),
				},
				types.Tag{
					Key:   aws.String(fmt.Sprintf("%s:idle_time_mins", p.config.TaggingPrefix)),
					Value: aws.String(fmt.Sprintf("%.1f", float64(metrics.IdleTime)/60.0)), // Convert from seconds to minutes
				},
			)
		}

		// Apply the tags
		_, err = p.client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
			Resources: []string{instanceID},
			Tags:      tags,
		})
		if err != nil {
			// Log the error but don't fail
			fmt.Printf("Warning: Failed to apply tags: %v\n", err)
		}
	}

	// Stop the instance
	_, err = p.client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	return err
}

// VerifyPermissions checks if the current AWS credentials have the required permissions
func (p *AWSProvider) VerifyPermissions() (bool, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.config.Region))
	if err != nil {
		return false, fmt.Errorf("error loading AWS config: %v", err)
	}

	// Create EC2 client
	client := ec2.NewFromConfig(cfg)

	// Check if we can describe instances
	_, err = client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		MaxResults: aws.Int32(5),
	})
	if err != nil {
		return false, fmt.Errorf("error checking EC2 permissions: %v", err)
	}

	// If tags are enabled, verify tag permissions
	if p.config.EnableTags {
		instanceID, err := p.getInstanceID()
		if err != nil {
			return false, fmt.Errorf("error getting instance ID: %v", err)
		}

		// Try to add a test tag
		_, err = client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
			Resources: []string{instanceID},
			Tags: []types.Tag{
				{
					Key:   aws.String("cloudsnooze:test"),
					Value: aws.String("permission-check"),
				},
			},
		})
		if err != nil {
			return false, fmt.Errorf("error checking tag permissions: %v", err)
		}

		// Try to remove the test tag
		_, err = client.DeleteTags(context.TODO(), &ec2.DeleteTagsInput{
			Resources: []string{instanceID},
			Tags: []types.Tag{
				{
					Key:   aws.String("cloudsnooze:test"),
					Value: aws.String("permission-check"),
				},
			},
		})
		if err != nil {
			return false, fmt.Errorf("error checking tag delete permissions: %v", err)
		}
	}

	return true, nil
}

// GetInstanceInfo returns information about the current instance
func (p *AWSProvider) GetInstanceInfo() (*common.InstanceInfo, error) {
	instanceID, err := p.getInstanceID()
	if err != nil {
		return nil, fmt.Errorf("error getting instance ID: %v", err)
	}

	// Check if we already have the instance type
	p.lock.RLock()
	if p.instanceType != "" {
		info := &common.InstanceInfo{
			ID:       instanceID,
			Type:     p.instanceType,
			Region:   p.region,
			Provider: "aws",
		}
		p.lock.RUnlock()
		return info, nil
	}
	p.lock.RUnlock()

	// Get the instance type from the metadata service
	instanceType, err := getMetadata("instance-type")
	if err != nil {
		return nil, fmt.Errorf("error getting instance type: %v", err)
	}

	// Get region from the metadata service if not already set
	region := p.config.Region
	if region == "" {
		// Try to get region from instance metadata
		az, err := getMetadata("placement/availability-zone")
		if err == nil {
			// Convert AZ to region by removing the last character (e.g., us-west-2a -> us-west-2)
			if len(az) > 1 {
				region = az[:len(az)-1]
			}
		}
	}

	// Store the values
	p.lock.Lock()
	p.instanceType = instanceType
	p.region = region
	p.lock.Unlock()

	return &common.InstanceInfo{
		ID:       instanceID,
		Type:     instanceType,
		Region:   region,
		Provider: "aws",
	}, nil
}

// getInstanceID returns the EC2 instance ID, caching the result
func (p *AWSProvider) getInstanceID() (string, error) {
	// Check if we already have the instance ID
	p.lock.RLock()
	if p.instanceID != "" {
		id := p.instanceID
		p.lock.RUnlock()
		return id, nil
	}
	p.lock.RUnlock()

	// Get instance ID from metadata service
	instanceID, err := getMetadata("instance-id")
	if err != nil {
		return "", fmt.Errorf("error getting instance ID: %v", err)
	}

	// Store the instance ID
	p.lock.Lock()
	p.instanceID = instanceID
	p.lock.Unlock()

	return instanceID, nil
}

// loadInstanceInfo loads instance information from the AWS metadata service
func (p *AWSProvider) loadInstanceInfo() error {
	// Get instance ID
	instanceID, err := getMetadata("instance-id")
	if err != nil {
		return fmt.Errorf("error getting instance ID: %v", err)
	}

	// Get instance type
	instanceType, err := getMetadata("instance-type")
	if err != nil {
		return fmt.Errorf("error getting instance type: %v", err)
	}

	// Get availability zone and derive region
	az, err := getMetadata("placement/availability-zone")
	if err != nil {
		return fmt.Errorf("error getting availability zone: %v", err)
	}

	// Convert AZ to region by removing the last character (e.g., us-west-2a -> us-west-2)
	var region string
	if len(az) > 1 {
		region = az[:len(az)-1]
	}

	// Store the values
	p.lock.Lock()
	p.instanceID = instanceID
	p.instanceType = instanceType
	p.region = region
	p.lock.Unlock()

	return nil
}

// getIMDSToken gets a token for IMDSv2
func getIMDSToken() (string, error) {
	// Create a request to get the token
	req, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", tokenTTL)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get IMDSv2 token, status: %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

// getMetadata gets a value from the EC2 instance metadata service
func getMetadata(path string) (string, error) {
	// Get token for IMDSv2
	token, err := getIMDSToken()
	if err != nil {
		return "", err
	}

	// Create a request with the token
	req, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/"+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metadata at path %s, status: %d", path, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// pollTags periodically checks for tags that might control the behavior of the daemon
func (p *AWSProvider) pollTags() {
	for {
		select {
		case <-p.tagPoller.C:
			// Get instance ID
			instanceID, err := p.getInstanceID()
			if err != nil {
				fmt.Printf("Error in tag polling: %v\n", err)
				continue
			}

			// Filter for the tags we're interested in
			tagFilter := fmt.Sprintf("%s:*", p.config.TaggingPrefix)

			// Get the instance tags
			result, err := p.client.DescribeTags(context.TODO(), &ec2.DescribeTagsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("resource-id"),
						Values: []string{instanceID},
					},
					{
						Name:   aws.String("key"),
						Values: []string{tagFilter},
					},
				},
			})
			if err != nil {
				fmt.Printf("Error getting tags: %v\n", err)
				continue
			}

			// Process tags - this is a placeholder, add real tag handling logic here
			for _, tag := range result.Tags {
				if tag.Key != nil && tag.Value != nil {
					fmt.Printf("Found tag: %s = %s\n", *tag.Key, *tag.Value)
					// TODO: Implement actual tag handling logic
					// For example, if there's a tag like "cloudsnooze:disable", pause monitoring
				}
			}

		case <-p.stopTagPoll:
			// Stop was requested
			if p.tagPoller != nil {
				p.tagPoller.Stop()
				p.tagPoller = nil
			}
			return
		}
	}
}

// StopTagPolling stops the tag polling goroutine
func (p *AWSProvider) StopTagPolling() {
	if p.tagPoller != nil {
		p.stopTagPoll <- struct{}{}
	}
}

// TagInstance adds tags to the current instance
func (p *AWSProvider) TagInstance(tags map[string]string) error {
	instanceID, err := p.getInstanceID()
	if err != nil {
		return fmt.Errorf("error getting instance ID: %v", err)
	}
	
	// Convert map to EC2 tag format
	var ec2Tags []types.Tag
	for k, v := range tags {
		ec2Tags = append(ec2Tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	
	// Apply the tags
	_, err = p.client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: []string{instanceID},
		Tags:      ec2Tags,
	})
	return err
}

// GetExternalTags checks for tags from external systems that might control this instance
func (p *AWSProvider) GetExternalTags() (map[string]string, error) {
	instanceID, err := p.getInstanceID()
	if err != nil {
		return nil, fmt.Errorf("error getting instance ID: %v", err)
	}
	
	// Get all tags for the instance
	result, err := p.client.DescribeTags(context.TODO(), &ec2.DescribeTagsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{instanceID},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %v", err)
	}
	
	// Convert to map
	tags := make(map[string]string)
	for _, tag := range result.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}
	
	return tags, nil
}