package accelerator

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/scttfrdmn/cloudsnooze/daemon/monitor"
)

// GPUMonitor is the interface for GPU monitoring
type GPUMonitor interface {
	// GetMetrics returns metrics for all detected GPUs
	GetMetrics() ([]monitor.GPUMetric, error)
	
	// IsAvailable returns true if this GPU type is available
	IsAvailable() bool
}

// NvidiaMonitor monitors NVIDIA GPUs
type NvidiaMonitor struct{}

// NewNvidiaMonitor creates a new NVIDIA GPU monitor
func NewNvidiaMonitor() *NvidiaMonitor {
	return &NvidiaMonitor{}
}

// IsAvailable checks if NVIDIA GPUs are available
func (m *NvidiaMonitor) IsAvailable() bool {
	_, err := exec.LookPath("nvidia-smi")
	return err == nil
}

// GetMetrics returns metrics for all NVIDIA GPUs
func (m *NvidiaMonitor) GetMetrics() ([]monitor.GPUMetric, error) {
	if !m.IsAvailable() {
		return nil, fmt.Errorf("nvidia-smi not available")
	}

	// Run nvidia-smi to get GPU info
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %v", err)
	}

	// Parse output
	var metrics []monitor.GPUMetric
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ", ")
		if len(parts) < 6 {
			continue
		}

		index, _ := strconv.Atoi(parts[0])
		utilization, _ := strconv.ParseFloat(parts[2], 64)
		memoryUsed, _ := strconv.ParseUint(parts[3], 10, 64)
		memoryTotal, _ := strconv.ParseUint(parts[4], 10, 64)
		temperature, _ := strconv.ParseFloat(parts[5], 64)

		metrics = append(metrics, monitor.GPUMetric{
			Type:        "NVIDIA",
			ID:          index,
			Name:        parts[1],
			Utilization: utilization,
			MemoryUsed:  memoryUsed,
			MemoryTotal: memoryTotal,
			Temperature: temperature,
		})
	}

	return metrics, nil
}

// AMDMonitor monitors AMD GPUs
type AMDMonitor struct{}

// NewAMDMonitor creates a new AMD GPU monitor
func NewAMDMonitor() *AMDMonitor {
	return &AMDMonitor{}
}

// IsAvailable checks if AMD GPUs are available
func (m *AMDMonitor) IsAvailable() bool {
	_, err := exec.LookPath("rocm-smi")
	return err == nil
}

// GetMetrics returns metrics for all AMD GPUs
func (m *AMDMonitor) GetMetrics() ([]monitor.GPUMetric, error) {
	if !m.IsAvailable() {
		return nil, fmt.Errorf("rocm-smi not available")
	}

	// Run rocm-smi to get GPU info
	cmd := exec.Command("rocm-smi", "--showuse", "--showmemuse", "--showtemp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run rocm-smi: %v", err)
	}

	// Parse output
	var metrics []monitor.GPUMetric
	
	// AMD GPUs don't have a clean CSV output like NVIDIA
	// This is a simplified parser
	lines := strings.Split(string(output), "\n")
	gpuRegex := regexp.MustCompile(`GPU\[(\d+)\]`)
	utilizationRegex := regexp.MustCompile(`GPU use\s+:\s+(\d+)%`)
	memoryUsedRegex := regexp.MustCompile(`GPU memory use\s+:\s+(\d+)MiB / (\d+)MiB`)
	tempRegex := regexp.MustCompile(`Temperature\s+:\s+(\d+\.\d+)c`)
	
	var currentID int
	var currentGPU monitor.GPUMetric
	
	for _, line := range lines {
		if match := gpuRegex.FindStringSubmatch(line); match != nil {
			// Save previous GPU if we're processing a new one
			if currentGPU.Type != "" {
				metrics = append(metrics, currentGPU)
			}
			
			// Start new GPU
			id, _ := strconv.Atoi(match[1])
			currentID = id
			currentGPU = monitor.GPUMetric{
				Type: "AMD",
				ID:   id,
				Name: fmt.Sprintf("AMD GPU %d", id),
			}
		} else if match := utilizationRegex.FindStringSubmatch(line); match != nil {
			utilization, _ := strconv.ParseFloat(match[1], 64)
			currentGPU.Utilization = utilization
		} else if match := memoryUsedRegex.FindStringSubmatch(line); match != nil {
			usedMiB, _ := strconv.ParseUint(match[1], 10, 64)
			totalMiB, _ := strconv.ParseUint(match[2], 10, 64)
			currentGPU.MemoryUsed = usedMiB * 1024 * 1024  // Convert to bytes
			currentGPU.MemoryTotal = totalMiB * 1024 * 1024 // Convert to bytes
		} else if match := tempRegex.FindStringSubmatch(line); match != nil {
			temp, _ := strconv.ParseFloat(match[1], 64)
			currentGPU.Temperature = temp
		}
	}
	
	// Add the last GPU if we have one
	if currentGPU.Type != "" {
		metrics = append(metrics, currentGPU)
	}

	return metrics, nil
}

// GPUService coordinates monitoring of multiple GPU types
type GPUService struct {
	monitors []GPUMonitor
}

// NewGPUService creates a new GPU monitoring service
func NewGPUService() *GPUService {
	service := &GPUService{
		monitors: []GPUMonitor{
			NewNvidiaMonitor(),
			NewAMDMonitor(),
			// Could add Intel GPU monitoring here
		},
	}
	return service
}

// GetAllMetrics returns metrics from all available GPU types
func (s *GPUService) GetAllMetrics() ([]monitor.GPUMetric, error) {
	var allMetrics []monitor.GPUMetric
	var errs []string

	for _, monitor := range s.monitors {
		if !monitor.IsAvailable() {
			continue
		}

		metrics, err := monitor.GetMetrics()
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		allMetrics = append(allMetrics, metrics...)
	}

	if len(allMetrics) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("failed to get GPU metrics: %s", strings.Join(errs, "; "))
	}

	return allMetrics, nil
}