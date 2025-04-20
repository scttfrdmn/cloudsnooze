// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"github.com/scttfrdmn/cloudsnooze/daemon/common"
)

// The Provider interface has been moved to common.CloudProvider
// This file is kept for backward compatibility during transition

// ProviderFactory creates and returns a cloud provider
type ProviderFactory interface {
	// CreateProvider creates a provider with the given config
	CreateProvider(config interface{}) (common.CloudProvider, error)
}