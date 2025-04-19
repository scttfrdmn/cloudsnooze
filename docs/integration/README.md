# CloudSnooze Integration Guide

This directory contains documentation for integrating external tools and services with CloudSnooze.

## Contents

- [API Reference](api-reference.md) - Detailed information about CloudSnooze's APIs
- [Restart Logic](restart-logic.md) - How to implement restart capabilities for stopped instances
- [External Tools](external-tools.md) - Guide for integrating specific external tools

## Key Integration Points

CloudSnooze provides several integration points for external tools and services:

1. **Socket API** - Local communication through a Unix socket
2. **Tag-based API** - Cloud provider tags for status and metadata
3. **Restart Capability** - Authorized restart of stopped instances

## Recent Updates

- Added support for explicit restart authorization through tags
- Added service-specific authorization for instance restarts
- Expanded documentation with more implementation examples