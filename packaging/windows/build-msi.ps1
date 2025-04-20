# PowerShell script for building a Windows MSI installer with WiX Toolset
# Requires WiX Toolset v3.11 or higher to be installed

param (
    [string]$Version = "0.1.0",
    [string]$BuildDir = "..\..\dist\windows"
)

# Ensure WiX is installed
function Check-WiX {
    try {
        $candle = Get-Command "candle.exe" -ErrorAction Stop
        $light = Get-Command "light.exe" -ErrorAction Stop
        Write-Host "WiX Toolset found at: $($candle.Source)" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "WiX Toolset not found in PATH. Please install WiX Toolset v3.11 or higher." -ForegroundColor Red
        return $false
    }
}

# Ensure the build directory exists
function Ensure-BuildDir {
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
        Write-Host "Created build directory: $BuildDir" -ForegroundColor Green
    }
}

# Create WiX source files
function Create-WixSource {
    $wxsPath = Join-Path $BuildDir "cloudsnooze.wxs"
    
    @"
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
    <Product Id="*" Name="CloudSnooze" Language="1033" Version="$Version" Manufacturer="CloudSnooze Team" UpgradeCode="81F9D371-F257-4F14-B4B7-58F64557CFB7">
        <Package InstallerVersion="200" Compressed="yes" InstallScope="perMachine" Comments="Windows Installer Package for CloudSnooze"/>
        <MajorUpgrade DowngradeErrorMessage="A newer version of CloudSnooze is already installed." />
        <MediaTemplate EmbedCab="yes" />
        <Feature Id="ProductFeature" Title="CloudSnooze" Level="1">
            <ComponentGroupRef Id="ProductComponents" />
            <ComponentRef Id="ApplicationShortcut" />
        </Feature>
        <UIRef Id="WixUI_InstallDir" />
        <Property Id="WIXUI_INSTALLDIR" Value="INSTALLFOLDER" />
        <WixVariable Id="WixUILicenseRtf" Value="license.rtf" />
        <Icon Id="icon.ico" SourceFile="..\..\resources\icon.ico"/>
        <Property Id="ARPPRODUCTICON" Value="icon.ico" />
    </Product>

    <Fragment>
        <Directory Id="TARGETDIR" Name="SourceDir">
            <Directory Id="ProgramFiles64Folder">
                <Directory Id="INSTALLFOLDER" Name="CloudSnooze">
                    <Directory Id="BinDir" Name="bin" />
                    <Directory Id="ConfigDir" Name="config" />
                    <Directory Id="DocsDir" Name="docs" />
                </Directory>
            </Directory>
            <Directory Id="ProgramMenuFolder">
                <Directory Id="ApplicationProgramsFolder" Name="CloudSnooze"/>
            </Directory>
            <Directory Id="CommonAppDataFolder">
                <Directory Id="AppDataCloudSnooze" Name="CloudSnooze">
                    <Directory Id="AppDataLogs" Name="logs" />
                </Directory>
            </Directory>
        </Directory>
    </Fragment>

    <Fragment>
        <ComponentGroup Id="ProductComponents" Directory="INSTALLFOLDER">
            <Component Id="ProductComponent" Guid="*">
                <File Id="README" Name="README.md" Source="..\..\README.md" KeyPath="yes" />
                <ServiceInstall
                    Id="ServiceInstaller"
                    Type="ownProcess"
                    Name="CloudSnooze"
                    DisplayName="CloudSnooze Daemon"
                    Description="Monitors system resources and automatically stops idle cloud instances"
                    Start="auto"
                    ErrorControl="normal"
                    Arguments="--config &quot;[CommonAppDataFolder]CloudSnooze\\snooze.json&quot;"
                    Account="LocalSystem" />
                <ServiceControl 
                    Id="StartService" 
                    Start="install" 
                    Stop="both" 
                    Remove="uninstall" 
                    Name="CloudSnooze" 
                    Wait="yes" />
                <Environment 
                    Id="PATH" 
                    Name="PATH" 
                    Value="[BinDir]" 
                    Permanent="no" 
                    Part="last" 
                    Action="set" 
                    System="yes" />
            </Component>
        </ComponentGroup>

        <ComponentGroup Id="BinComponents" Directory="BinDir">
            <Component Id="Daemon" Guid="*">
                <File Id="SnoozedEXE" Name="snoozed.exe" Source="..\..\dist\snoozed_windows_amd64.exe" KeyPath="yes" />
            </Component>
            <Component Id="CLI" Guid="*">
                <File Id="SnoozeEXE" Name="snooze.exe" Source="..\..\dist\snooze_windows_amd64.exe" KeyPath="yes" />
            </Component>
        </ComponentGroup>

        <ComponentGroup Id="ConfigComponents" Directory="ConfigDir">
            <Component Id="DefaultConfig" Guid="*">
                <File Id="ConfigJSON" Name="snooze.json" Source="..\..\config\snooze.json" KeyPath="yes" />
                <CopyFile Id="CopyConfig" SourceProperty="ConfigJSON" DestinationDirectory="AppDataCloudSnooze" DestinationName="snooze.json" />
            </Component>
        </ComponentGroup>

        <Component Id="ApplicationShortcut" Guid="*" Directory="ApplicationProgramsFolder">
            <Shortcut Id="ApplicationStartMenuShortcut" 
                      Name="CloudSnooze Status" 
                      Description="Check CloudSnooze status"
                      Target="[BinDir]snooze.exe" 
                      Arguments="status"
                      WorkingDirectory="INSTALLFOLDER">
                <Icon Id="ApplicationIcon" SourceFile="..\..\resources\icon.ico" />
            </Shortcut>
            <RemoveFolder Id="CleanUpShortCut" Directory="ApplicationProgramsFolder" On="uninstall"/>
            <RegistryValue Root="HKCU" Key="Software\CloudSnooze" Name="installed" Type="integer" Value="1" KeyPath="yes"/>
        </Component>
    </Fragment>
</Wix>
"@ | Out-File -FilePath $wxsPath -Encoding UTF8
    
    # Create a simple RTF license file
    $licenseRtf = Join-Path $BuildDir "license.rtf"
    @"
{\rtf1\ansi\ansicpg1252\deff0\nouicompat\deflang1033{\fonttbl{\f0\fnil\fcharset0 Calibri;}}
{\*\generator Riched20 10.0.19041}\viewkind4\uc1 
\pard\sa200\sl276\slmult1\f0\fs22\lang9 CloudSnooze License\par
\par
Copyright 2025 Scott Friedman and CloudSnooze Contributors\par
\par
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this software except in compliance with the License. You may obtain a copy of the License at:\par
\par
http://www.apache.org/licenses/LICENSE-2.0\par
\par
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.\par
}
"@ | Out-File -FilePath $licenseRtf -Encoding UTF8
}

# Build the MSI
function Build-MSI {
    $wxsPath = Join-Path $BuildDir "cloudsnooze.wxs"
    $objPath = Join-Path $BuildDir "cloudsnooze.wixobj"
    $msiPath = Join-Path $BuildDir "cloudsnooze-$Version-windows-amd64.msi"
    
    # Compile the WiX source
    & candle.exe -ext WixUtilExtension -ext WixUIExtension -out $objPath $wxsPath
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: WiX candle compilation failed" -ForegroundColor Red
        exit 1
    }
    
    # Link the object file to create the MSI
    & light.exe -ext WixUtilExtension -ext WixUIExtension -out $msiPath $objPath
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: WiX light linking failed" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "MSI installer created successfully: $msiPath" -ForegroundColor Green
}

# Main execution
if (-not (Check-WiX)) {
    exit 1
}

Ensure-BuildDir
Create-WixSource
Build-MSI