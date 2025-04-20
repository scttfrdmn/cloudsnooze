// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/scttfrdmn/cloudsnooze/daemon/cloud/aws"
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
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
func DetectProvider() (ProviderType, error) {
	// Check for AWS
	if _, err := os.Stat("/sys/devices/virtual/dmi/id/product_uuid"); err == nil {
		// Check if we can access the instance metadata service
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://169.254.169.254/latest/meta-data")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return AWS, nil
			}
		}
	}

	// Could add detection for other providers here
	
	return "", errors.New("unable to detect cloud provider")
}

// CreateProvider creates a new provider of the specified type with the given config
func CreateProvider(providerType ProviderType, config interface{}) (common.CloudProvider, error) {
	switch providerType {
	case AWS:
		awsConfig, ok := config.(aws.Config)
		if !ok {
			return nil, errors.New("invalid AWS configuration type")
		}
		return aws.NewProvider(awsConfig), nil
	// case GCP:
	//    return nil, errors.New("GCP provider not implemented")
	// case Azure:
	//    return nil, errors.New("Azure provider not implemented")
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}