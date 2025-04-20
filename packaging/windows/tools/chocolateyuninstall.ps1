$ErrorActionPreference = 'Stop'

$packageName= 'cloudsnooze'
$serviceName = 'CloudSnooze'

# Stop and remove the service if it exists
$serviceExists = Get-Service -Name $serviceName -ErrorAction SilentlyContinue

if ($serviceExists) {
    Write-Host "Stopping and removing CloudSnooze service..."
    Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
    & sc.exe delete $serviceName
    Write-Host "CloudSnooze service removed!"
}

# Configuration and logs are kept by default unless user specifies to remove them
# This is a common practice for service uninstallations
$configDir = Join-Path $env:ProgramData 'CloudSnooze'

Write-Host @"
CloudSnooze has been uninstalled!

Note: Configuration files and logs have been preserved at:
  $configDir

If you want to completely remove all data, delete this directory manually.
"@