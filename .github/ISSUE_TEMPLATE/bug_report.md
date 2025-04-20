---
name: Bug report
about: Create a report to help us improve CloudSnooze
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description
<!-- A clear and concise description of what the bug is -->

## Environment
<!-- Please complete the following information -->
- CloudSnooze Version: <!-- e.g. v0.1.0 -->
- OS: <!-- e.g. Ubuntu 22.04, macOS 13.4, Windows 11 -->
- Architecture: <!-- e.g. x86_64, arm64 -->
- Installation Method: <!-- e.g. DEB, RPM, Homebrew, Chocolatey, MSI -->
- Cloud Provider: <!-- e.g. AWS EC2, AWS Lightsail, None (local) -->
- Instance Type (if applicable): <!-- e.g. t3.micro -->

## Steps To Reproduce
<!-- Steps to reproduce the behavior -->
1. 
2. 
3. 

## Expected Behavior
<!-- A clear and concise description of what you expected to happen -->

## Actual Behavior
<!-- What actually happened -->

## Debug Output
<!-- Please provide the output of the following commands -->
```
snooze status --debug
```

```
systemctl status snoozed  # Linux
# OR
brew services info cloudsnooze  # macOS
# OR
sc query CloudSnooze  # Windows
```

## Log Output
<details>
<summary>CloudSnooze logs</summary>

```
<!-- Paste relevant logs here - Linux: /var/log/cloudsnooze.log, macOS: /usr/local/var/log/cloudsnooze, Windows: C:\ProgramData\CloudSnooze\logs\cloudsnooze.log -->
```
</details>

## Screenshots
<!-- If applicable, add screenshots to help explain your problem -->

## Additional Context
<!-- Add any other context about the problem here -->