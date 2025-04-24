<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# AWS Integration Testing for CloudSnooze

This document explains how to run AWS integration tests for CloudSnooze.

## Overview

CloudSnooze includes integration tests that verify its AWS functionality works correctly in a real AWS environment. This document explains how to run these tests both locally and in CI/CD pipelines.

## Authentication Methods

We support two methods for authenticating with AWS:

1. **OIDC Authentication** (recommended): Uses GitHub's OIDC provider to get temporary AWS credentials. This is more secure and doesn't require storing long-lived access keys.

2. **Access Key Authentication**: Uses traditional IAM user access keys. This is simpler to set up but less secure.

For detailed setup instructions, see:
- [AWS Integration Setup with OIDC](aws_integration_setup_oidc.md)
- [AWS Integration Setup with Access Keys](aws_integration_setup_access_keys.md)

## Running Integration Tests

### In GitHub Actions

Integration tests will automatically run on:
- Pull requests with the `aws-test` label
- Branches with commit messages containing `[aws-test]`
- Manual workflow dispatches

To manually trigger the tests:
1. Go to the GitHub repository
2. Navigate to Actions → AWS Integration Tests
3. Click "Run workflow"
4. Select the branch to test
5. Click "Run workflow"

### Locally

To run integration tests on your local machine:

1. Set up AWS credentials:
   ```bash
   # For access key authentication
   export AWS_ACCESS_KEY_ID=your_access_key_id
   export AWS_SECRET_ACCESS_KEY=your_secret_access_key
   export AWS_REGION=us-west-2
   
   # OR for OIDC (requires additional setup)
   # See aws_integration_setup_oidc.md for details
   ```

2. Create a test instance:
   ```bash
   # Create a temporary EC2 instance for testing
   INSTANCE_ID=$(aws ec2 run-instances \
     --image-id ami-0c2b8ca1dad447f8a \
     --instance-type t2.micro \
     --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=cloudsnooze-local-test},{Key=Purpose,Value=Testing}]" \
     --query 'Instances[0].InstanceId' \
     --output text)
   
   # Export required environment variables
   export CLOUDSNOOZE_TEST_INSTANCE_ID=$INSTANCE_ID
   export CLOUDSNOOZE_TEST_REGION=us-west-2
   ```

3. Run the tests:
   ```bash
   cd daemon/cloud/aws
   go test -v -tags=integration ./...
   ```

4. Clean up:
   ```bash
   # Terminate the test instance
   aws ec2 terminate-instances --instance-ids $INSTANCE_ID
   ```

## Test Tags

Integration tests are tagged with `integration` using Go build tags:

```go
//go:build integration
// +build integration
```

This ensures they only run when explicitly requested with `-tags=integration`.

## Resource Cleanup

All test resources are tagged with `Purpose: Testing`. These resources are cleaned up in several ways:

1. **Immediate cleanup**: The GitHub Actions workflow automatically cleans up resources after tests complete (even if tests fail).

2. **Scheduled cleanup**: A Lambda function runs hourly to clean up any test resources older than 2 hours.

3. **Manual cleanup**: You can clean up all test resources manually:
   ```bash
   aws ec2 describe-instances --filters "Name=tag:Purpose,Values=Testing" --query "Reservations[].Instances[].InstanceId" --output text | xargs -n1 aws ec2 terminate-instances --instance-ids
   ```

## Cost Control

To prevent unexpected costs:

1. **Testing quotas**: The IAM policy is restricted to only allow t2.micro instances.

2. **Budget alerts**: Set up AWS budget alerts to notify you if costs exceed expected thresholds.

3. **Resource tagging**: All test resources are tagged for identification and automated cleanup.

## Troubleshooting

Common issues and solutions:

1. **Credentials Issues**:
   - For access keys: Check secrets are properly configured
   - For OIDC: Verify the trust relationship and role permissions

2. **Instance Creation Failures**:
   - Check the AMI exists in your region
   - Verify service quotas aren't exceeded
   - Use a different region if experiencing capacity issues

3. **Timeout Issues**:
   - AWS operations may sometimes be slow
   - Increase timeout values in test code if necessary

## Comparing Local vs CI Test Results

If you see tests passing locally but failing in CI (or vice versa):

1. **Check region differences**: Different regions may have different behavior
2. **Check credentials**: Ensure permissions are identical
3. **Examine logs**: Look for differences in instance properties or timing