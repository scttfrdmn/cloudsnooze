package cloud

import (
	"github.com/scttfrdmn/cloudsnooze/daemon/monitor"
)

// Provider defines the interface for cloud provider operations
type Provider interface {
	// VerifyPermissions checks if the daemon has sufficient permissions
	VerifyPermissions() (bool, error)
	
	// GetInstanceInfo retrieves information about the current instance
	GetInstanceInfo() (*InstanceInfo, error)
	
	// StopInstance stops the current instance
	StopInstance(reason string, metrics monitor.SystemMetrics) error
	
	// TagInstance adds tags to the current instance
	TagInstance(tags map[string]string) error
	
	// GetExternalTags checks for tags from external systems that might control this instance
	GetExternalTags() (map[string]string, error)
}

// InstanceInfo contains information about the current cloud instance
type InstanceInfo struct {
	ID         string
	Type       string
	Region     string
	Provider   string
	LaunchTime string
	Tags       map[string]string
}

// ProviderFactory creates and returns a cloud provider
type ProviderFactory interface {
	// CreateProvider creates a provider with the given config
	CreateProvider(config interface{}) (Provider, error)
}