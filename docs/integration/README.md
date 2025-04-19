# CloudSnooze Integration Guide

## Overview

This directory contains documentation for integrating external tools with CloudSnooze. CloudSnooze is designed to automatically stop idle cloud instances to save costs, and these guides explain how other systems can interact with it.

## Contents

- [External Tools Integration Guide](external-tools.md) - Overview of integration methods and best practices
- [Restart Logic](restart-logic.md) - Patterns for implementing instance restart functionality
- [API Reference](api-reference.md) - Detailed API documentation for both socket and tag-based APIs

## Quick Start

If you're building a tool that needs to interact with CloudSnooze, start with the [External Tools Integration Guide](external-tools.md). This provides an overview of the available integration methods and will help you choose the right approach.

For more specific needs:

- If you need to restart instances that CloudSnooze has stopped, see [Restart Logic](restart-logic.md)
- If you need detailed API information, see [API Reference](api-reference.md)

## Integration Patterns

CloudSnooze supports two main integration patterns:

1. **Tag-Based Integration**: Uses instance tags to communicate state and metadata
   - Ideal for external tools that manage many instances
   - Works across cloud boundaries
   - Simple polling mechanism

2. **Socket API**: Direct communication with the CloudSnooze daemon
   - Ideal for tools running on the same instance
   - Provides real-time status information
   - Allows configuration queries

## Security Considerations

When integrating with CloudSnooze, keep these security considerations in mind:

1. **Avoid Tag Manipulation**: External tools should not modify CloudSnooze tags directly, except when restarting instances

2. **Socket API Security**: The Unix socket is protected by filesystem permissions. Ensure your application has appropriate access.

3. **IAM Permissions**: Make sure your external tools have the necessary permissions to:
   - Read instance tags
   - Start instances (if implementing restart logic)

## Getting Help

If you encounter issues or have questions about integrating with CloudSnooze, please:

1. Check the documentation in this directory
2. Review the main [README.md](../../README.md) for general information
3. File an issue on the project repository