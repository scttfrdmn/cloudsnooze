class Cloudsnooze < Formula
  desc "Automatically stop idle cloud instances and save costs"
  homepage "https://cloudsnooze.io"
  license "Apache-2.0"
  version "0.1.0"
  
  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ARM64_MAC_CHECKSUM"
    else
      url "https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_AMD64_MAC_CHECKSUM"
    end
  elsif OS.linux?
    if Hardware::CPU.arm?
      url "https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ARM64_LINUX_CHECKSUM"
    else
      url "https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_AMD64_LINUX_CHECKSUM"
    end
  end

  depends_on "go" => :build

  def install
    # Install binary executables
    bin.install "snooze"
    bin.install "snoozed"
    
    # Install systemd service file
    if OS.linux?
      (lib/"systemd/system").install "snoozed.service"
    end
    
    # Install standard configuration
    (etc/"cloudsnooze").install "config/snooze.json"
    
    # Create log and data directories
    (var/"log/cloudsnooze").mkpath
    (var/"lib/cloudsnooze").mkpath
    
    # Install man pages
    man1.install "man/snooze.1"
    man1.install "man/snoozed.1"
    
    # Install documentation
    doc.install "README.md", "LICENSE"
    doc.install Dir["docs/*"]
  end

  def post_install
    # Create configuration if it doesn't exist
    if !(etc/"cloudsnooze/snooze.json").exist?
      (etc/"cloudsnooze/snooze.json").write(<<~EOS)
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
            "log_file_path": "#{var}/log/cloudsnooze/cloudsnooze.log",
            "enable_syslog": false,
            "enable_cloudwatch": false,
            "cloudwatch_log_group": "CloudSnooze"
          },
          "monitoring_mode": "basic"
        }
      EOS
    end
  end

  def caveats
    <<~EOS
      Configuration was installed to:
        #{etc}/cloudsnooze/snooze.json

      Logs will be written to:
        #{var}/log/cloudsnooze
        
      To start CloudSnooze daemon at login on macOS:
        brew services start cloudsnooze
        
      To start the daemon on Linux manually:
        sudo systemctl enable snoozed.service
        sudo systemctl start snoozed.service
        
      Check your system status:
        snooze status
    EOS
  end

  service do
    run [opt_bin/"snoozed", "--config", etc/"cloudsnooze/snooze.json"]
    keep_alive true
    log_path var/"log/cloudsnooze/daemon.log"
    error_log_path var/"log/cloudsnooze/daemon.error.log"
    working_dir var/"lib/cloudsnooze"
  end

  test do
    system "#{bin}/snooze", "--version"
    system "#{bin}/snoozed", "--version"
  end
end