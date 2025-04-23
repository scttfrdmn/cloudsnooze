// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// IssueType represents the type of GitHub issue to create
type IssueType string

const (
	BugReport       IssueType = "bug_report"
	FeatureRequest  IssueType = "feature_request"
	IntegrationIssue IssueType = "integration_issue"
	Documentation   IssueType = "documentation"
)

// IssueData holds information about a GitHub issue
type IssueData struct {
	Type        IssueType
	Title       string
	Description string
	Environment map[string]string
	LogData     string
	ExtraInfo   string
}

// CreateIssue creates a new GitHub issue or opens the issue creation page in a browser
func CreateIssue(issueType IssueType, title, description string, browser bool) error {
	// Get environment information
	env, err := collectEnvironmentInfo()
	if err != nil {
		fmt.Printf("Warning: Could not collect all environment information: %v\n", err)
	}

	// Get log data
	logs, err := collectLogData()
	if err != nil {
		fmt.Printf("Warning: Could not collect log data: %v\n", err)
	}

	issueData := IssueData{
		Type:        issueType,
		Title:       title,
		Description: description,
		Environment: env,
		LogData:     logs,
	}

	// Format the issue based on the template
	issueBody, err := formatIssueBody(issueData)
	if err != nil {
		return fmt.Errorf("error formatting issue: %v", err)
	}

	// If browser flag is true, open GitHub issue page in browser
	if browser {
		return openBrowserWithIssue(issueType, title, issueBody)
	}

	// Otherwise, submit issue directly using GitHub API
	// Note: This would require a GitHub token which is more complex
	// For simplicity, we'll default to browser method
	return openBrowserWithIssue(issueType, title, issueBody)
}

// collectEnvironmentInfo gathers system and CloudSnooze information
func collectEnvironmentInfo() (map[string]string, error) {
	env := make(map[string]string)

	// Get CloudSnooze version
	cmd := exec.Command("snooze", "--version")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		if strings.Contains(version, "CloudSnooze CLI v") {
			version = strings.TrimPrefix(version, "CloudSnooze CLI v")
		}
		env["CloudSnooze Version"] = version
	} else {
		env["CloudSnooze Version"] = "Unknown"
	}

	// Get OS information
	env["OS"] = runtime.GOOS
	if runtime.GOOS == "linux" {
		// Try to get Linux distribution
		cmd := exec.Command("lsb_release", "-d")
		output, err := cmd.Output()
		if err == nil {
			distro := strings.TrimSpace(string(output))
			if strings.Contains(distro, "Description:") {
				distro = strings.TrimPrefix(distro, "Description:")
				distro = strings.TrimSpace(distro)
				env["OS"] = distro
			}
		}
	} else if runtime.GOOS == "darwin" {
		// Get macOS version
		cmd := exec.Command("sw_vers", "-productVersion")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			env["OS"] = "macOS " + version
		}
	} else if runtime.GOOS == "windows" {
		// Get Windows version
		cmd := exec.Command("powershell", "-Command", "(Get-WmiObject -class Win32_OperatingSystem).Caption")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			env["OS"] = version
		}
	}

	// Get architecture
	env["Architecture"] = runtime.GOARCH

	// Attempt to determine installation method
	switch runtime.GOOS {
	case "linux":
		// Check if it was installed via DEB or RPM
		cmd := exec.Command("dpkg", "-s", "cloudsnooze")
		if err := cmd.Run(); err == nil {
			env["Installation Method"] = "DEB package"
		} else {
			cmd := exec.Command("rpm", "-q", "cloudsnooze")
			if err := cmd.Run(); err == nil {
				env["Installation Method"] = "RPM package"
			} else {
				env["Installation Method"] = "Unknown"
			}
		}
	case "darwin":
		// Check if it was installed via Homebrew
		cmd := exec.Command("brew", "list", "cloudsnooze")
		if err := cmd.Run(); err == nil {
			env["Installation Method"] = "Homebrew"
		} else {
			env["Installation Method"] = "Unknown"
		}
	case "windows":
		// Check if it was installed via Chocolatey
		cmd := exec.Command("powershell", "-Command", "choco list --local-only cloudsnooze")
		if err := cmd.Run(); err == nil {
			env["Installation Method"] = "Chocolatey"
		} else {
			env["Installation Method"] = "MSI or Unknown"
		}
	}

	// Try to determine cloud provider by checking AWS metadata
	awsMetadata := checkAwsMetadata()
	if awsMetadata {
		env["Cloud Provider"] = "AWS"
		// Try to get instance type
		instanceType, err := getAwsInstanceType()
		if err == nil {
			env["Instance Type"] = instanceType
		}
	} else {
		env["Cloud Provider"] = "None (local)"
	}

	return env, nil
}

// checkAwsMetadata checks if we're running on AWS by attempting to access the metadata service
func checkAwsMetadata() bool {
	client := &http.Client{
		Timeout: 1 * time.Second, // Short timeout
	}
	_, err := client.Get("http://169.254.169.254/latest/meta-data")
	return err == nil
}

// getAwsInstanceType retrieves the instance type from AWS metadata
func getAwsInstanceType() (string, error) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get("http://169.254.169.254/latest/meta-data/instance-type")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// collectLogData retrieves CloudSnooze logs
func collectLogData() (string, error) {
	var logPath string
	var readCmd *exec.Cmd

	// Determine log path based on OS
	switch runtime.GOOS {
	case "linux":
		logPath = "/var/log/cloudsnooze.log"
		readCmd = exec.Command("tail", "-n", "100", logPath)
	case "darwin":
		logPath = "/usr/local/var/log/cloudsnooze/cloudsnooze.log"
		readCmd = exec.Command("tail", "-n", "100", logPath)
	case "windows":
		logPath = "C:\\ProgramData\\CloudSnooze\\logs\\cloudsnooze.log"
		// PowerShell command to get last 100 lines
		readCmd = exec.Command("powershell", "-Command", 
			fmt.Sprintf("Get-Content -Tail 100 -Path '%s'", logPath))
	default:
		return "Log collection not supported on this OS", fmt.Errorf("unsupported OS")
	}

	output, err := readCmd.Output()
	if err != nil {
		return fmt.Sprintf("Could not read log file %s: %v", logPath, err), err
	}

	return string(output), nil
}

// formatIssueBody creates the issue body based on the template
func formatIssueBody(data IssueData) (string, error) {
	var body strings.Builder

	switch data.Type {
	case BugReport:
		body.WriteString("## Bug Description\n")
		body.WriteString(data.Description)
		body.WriteString("\n\n## Environment\n")
		for k, v := range data.Environment {
			body.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
		body.WriteString("\n## Steps To Reproduce\n1. \n2. \n3. \n\n")
		body.WriteString("## Expected Behavior\n\n\n")
		body.WriteString("## Actual Behavior\n\n\n")
		body.WriteString("## Log Output\n<details>\n<summary>CloudSnooze logs</summary>\n\n```\n")
		body.WriteString(data.LogData)
		body.WriteString("\n```\n</details>\n\n")
		body.WriteString("## Additional Context\n")
		if data.ExtraInfo != "" {
			body.WriteString(data.ExtraInfo)
		}

	case FeatureRequest:
		body.WriteString("## Problem Statement\n")
		body.WriteString(data.Description)
		body.WriteString("\n\n## Proposed Solution\n\n\n")
		body.WriteString("## Alternative Solutions\n\n\n")
		body.WriteString("## Cloud Providers Affected\n")
		body.WriteString("- [ ] AWS\n")
		body.WriteString("- [ ] Future GCP Support\n")
		body.WriteString("- [ ] Future Azure Support\n")
		body.WriteString("- [ ] Local development machines\n")
		body.WriteString("- [ ] Other (please specify)\n\n")
		body.WriteString("## Additional Context\n")
		if data.ExtraInfo != "" {
			body.WriteString(data.ExtraInfo)
		}

	case IntegrationIssue:
		body.WriteString("## Integration Issue Description\n")
		body.WriteString(data.Description)
		body.WriteString("\n\n## Environment\n")
		for k, v := range data.Environment {
			body.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
		body.WriteString("\n## Steps To Reproduce\n1. \n2. \n3. \n\n")
		body.WriteString("## Expected Behavior\n\n\n")
		body.WriteString("## Actual Behavior\n\n\n")
		body.WriteString("## Log Output\n<details>\n<summary>CloudSnooze logs</summary>\n\n```\n")
		body.WriteString(data.LogData)
		body.WriteString("\n```\n</details>\n\n")
		body.WriteString("## Additional Context\n")
		if data.ExtraInfo != "" {
			body.WriteString(data.ExtraInfo)
		}

	case Documentation:
		body.WriteString("## Documentation Issue/Request\n")
		body.WriteString(data.Description)
		body.WriteString("\n\n## Current Documentation Location\n")
		body.WriteString("- URL: \n")
		body.WriteString("- Section: \n\n")
		body.WriteString("## Proposed Changes\n\n\n")
		body.WriteString("## Additional Information\n")
		if data.ExtraInfo != "" {
			body.WriteString(data.ExtraInfo)
		}

	default:
		return "", fmt.Errorf("unknown issue type: %s", data.Type)
	}

	return body.String(), nil
}

// openBrowserWithIssue opens the browser with a pre-filled GitHub issue
func openBrowserWithIssue(issueType IssueType, title, body string) error {
	// GitHub new issue URL
	baseURL := "https://github.com/scttfrdmn/cloudsnooze/issues/new"

	// URL-encode the title and body
	queryParams := fmt.Sprintf("?template=%s.md&title=%s&body=%s", 
		issueType,
		encodeURIComponent(title),
		encodeURIComponent(body))

	url := baseURL + queryParams

	// Open the URL in the default browser
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		return fmt.Errorf("error opening browser: %v", err)
	}

	return nil
}

// encodeURIComponent is a simple implementation of JavaScript's encodeURIComponent
func encodeURIComponent(str string) string {
	// This is a simplified version - a production version would need more complete encoding
	replacer := strings.NewReplacer(
		" ", "%20",
		"\"", "%22",
		"<", "%3C",
		">", "%3E",
		"#", "%23",
		"%", "%25",
		"{", "%7B",
		"}", "%7D",
		"|", "%7C",
		"\\", "%5C",
		"^", "%5E",
		"~", "%7E",
		"[", "%5B",
		"]", "%5D",
		"`", "%60",
		";", "%3B",
		"/", "%2F",
		"?", "%3F",
		":", "%3A",
		"@", "%40",
		"=", "%3D",
		"&", "%26",
		"$", "%24",
	)
	return replacer.Replace(str)
}

// ReportIssue handles the report-issue command
func ReportIssue(issueType, title, description string, browser bool) error {
	var reportType IssueType

	// Validate issue type
	switch strings.ToLower(issueType) {
	case "bug":
		reportType = BugReport
	case "feature":
		reportType = FeatureRequest
	case "integration":
		reportType = IntegrationIssue
	case "docs", "documentation":
		reportType = Documentation
	default:
		return fmt.Errorf("unknown issue type: %s (valid types: bug, feature, integration, docs)", issueType)
	}

	// Validate title
	if title == "" {
		return fmt.Errorf("issue title cannot be empty")
	}

	// If description is empty, prompt from stdin
	if description == "" {
		fmt.Print("Enter issue description (end with Ctrl+D on a new line):\n")
		descBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading description: %v", err)
		}
		description = string(descBytes)
	}

	return CreateIssue(reportType, title, description, browser)
}

// SubmitDebugInfo collects and submits debug information to assist with troubleshooting
func SubmitDebugInfo(outputFile string) error {
	debugInfo := make(map[string]interface{})

	// Get environment information
	env, err := collectEnvironmentInfo()
	if err != nil {
		fmt.Printf("Warning: Could not collect all environment information: %v\n", err)
	}
	debugInfo["environment"] = env

	// Get status output
	statusCmd := exec.Command("snooze", "status", "--json")
	statusOutput, err := statusCmd.Output()
	if err == nil {
		var statusData interface{}
		if err := json.Unmarshal(statusOutput, &statusData); err == nil {
			debugInfo["status"] = statusData
		} else {
			debugInfo["status"] = string(statusOutput)
		}
	} else {
		debugInfo["status"] = "Error retrieving status"
		debugInfo["status_error"] = err.Error()
	}

	// Get configuration
	configCmd := exec.Command("snooze", "config", "list", "--json")
	configOutput, err := configCmd.Output()
	if err == nil {
		var configData interface{}
		if err := json.Unmarshal(configOutput, &configData); err == nil {
			debugInfo["config"] = configData
		} else {
			debugInfo["config"] = string(configOutput)
		}
	} else {
		debugInfo["config"] = "Error retrieving configuration"
		debugInfo["config_error"] = err.Error()
	}

	// Get logs
	logs, err := collectLogData()
	if err != nil {
		fmt.Printf("Warning: Could not collect log data: %v\n", err)
	}
	debugInfo["logs"] = logs

	// Get service status
	var serviceStatusCmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		serviceStatusCmd = exec.Command("systemctl", "status", "snoozed", "--no-pager")
	case "darwin":
		serviceStatusCmd = exec.Command("brew", "services", "info", "cloudsnooze")
	case "windows":
		serviceStatusCmd = exec.Command("sc", "query", "CloudSnooze")
	}

	if serviceStatusCmd != nil {
		serviceOutput, err := serviceStatusCmd.Output()
		if err == nil {
			debugInfo["service_status"] = string(serviceOutput)
		} else {
			debugInfo["service_status"] = "Error retrieving service status"
			debugInfo["service_status_error"] = err.Error()
		}
	}

	// Get system info
	debugInfo["system"] = map[string]string{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"time":       time.Now().Format(time.RFC3339),
	}

	// Serialize the debug info
	jsonData, err := json.MarshalIndent(debugInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing debug info: %v", err)
	}

	// If outputFile is provided, write to that file
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			return fmt.Errorf("error writing debug info to file: %v", err)
		}
		fmt.Printf("Debug information written to %s\n", outputFile)
		return nil
	}

	// Otherwise, output to stdout
	fmt.Println(string(jsonData))
	return nil
}