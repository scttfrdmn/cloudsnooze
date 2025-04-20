$ErrorActionPreference = 'Stop'

$packageName= 'cloudsnooze'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64      = 'https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_windows_amd64.zip'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  checksum64    = 'REPLACE_WITH_CHECKSUM' # Replace with actual checksum when available
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

# Install service
$serviceExePath = Join-Path $toolsDir 'snoozed.exe'
$serviceName = 'CloudSnooze'
$configDir = Join-Path $env:ProgramData 'CloudSnooze'
$configPath = Join-Path $configDir 'snooze.json'
$logDir = Join-Path $env:ProgramData 'CloudSnooze\logs'

# Create directories if they don't exist
if (!(Test-Path $configDir)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
}

if (!(Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
}

# Create default config if it doesn't exist
if (!(Test-Path $configPath)) {
    @'
{
  "check_interval_seconds": 60,
  "naptime_minutes": 30,
  "cpu_threshold_percent": 10.0,
  "memory_threshold_percent": 30.0,
  "network_threshold_kbps": 50.0,
  "disk_io_threshold_kbps": 100.0,
  "input_idle_threshold_secs": 900,
  "gpu_monitoring_enabled": true,
  "gpu_threshold_percent": 5.0,
  "aws_region": "",
  "enable_instance_tags": true,
  "tagging_prefix": "CloudSnooze",
  "detailed_instance_tags": true,
  "tag_polling_enabled": true,
  "tag_polling_interval_secs": 60,
  "logging": {
    "log_level": "info",
    "enable_file_logging": true,
    "log_file_path": "PROGRAM_DATA_PATH\\CloudSnooze\\logs\\cloudsnooze.log",
    "enable_syslog": false,
    "enable_cloudwatch": false,
    "cloudwatch_log_group": "CloudSnooze"
  },
  "monitoring_mode": "basic"
}
'@ -replace 'PROGRAM_DATA_PATH', $env:ProgramData.Replace('\', '\\') | Out-File -FilePath $configPath -Encoding UTF8
}

# Install the service if it doesn't exist
$serviceExists = Get-Service -Name $serviceName -ErrorAction SilentlyContinue

if (-not $serviceExists) {
    Write-Host "Installing CloudSnooze service..."
    & sc.exe create $serviceName binPath= "$serviceExePath --config $configPath" DisplayName= "CloudSnooze Daemon" start= auto
    Write-Host "CloudSnooze service installed!"
    
    # Start the service
    Start-Service -Name $serviceName
    Write-Host "CloudSnooze service started!"
} else {
    Write-Host "CloudSnooze service already exists. Stopping and starting service to apply changes..."
    Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
    Start-Service -Name $serviceName
}

# Add PATH environment variable for CLI
$binDir = $toolsDir
Install-ChocolateyPath -PathToInstall $binDir -PathType 'Machine'

Write-Host @"
CloudSnooze has been installed!

Configuration file:
  $configPath

Logs directory:
  $logDir

To check status:
  snooze status
"@