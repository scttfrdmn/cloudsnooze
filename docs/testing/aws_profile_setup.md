# AWS Profile Setup for CloudSnooze Testing

This document explains how to set up a dedicated AWS CLI profile for CloudSnooze integration testing. This approach allows you to:

1. Maintain separation between your normal AWS credentials and testing credentials
2. Avoid accidental use of your default credentials for tests
3. Easily switch between multiple AWS accounts

## Overview

When working with a dedicated AWS testing account, you should create a separate AWS CLI profile to store its credentials. This profile can then be used explicitly when running CloudSnooze integration tests.

## Prerequisites

- [AWS CLI installed](https://aws.amazon.com/cli/)
- Access to your CloudSnooze AWS testing account credentials

## Setting Up a Named Profile

### 1. Create a Profile Using the AWS CLI

```bash
aws configure --profile cloudsnooze-testing
```

When prompted, enter:
- AWS Access Key ID for your testing account
- AWS Secret Access Key for your testing account
- Default region (e.g., us-west-2)
- Default output format (json is recommended)

### 2. Verify the Profile Setup

```bash
aws sts get-caller-identity --profile cloudsnooze-testing
```

This should return information about the identity associated with the credentials:

```json
{
    "UserId": "AIDA...",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/github-cloudsnooze-test"
}
```

Confirm that this shows the correct account and user for your testing account.

## Using the Profile with CloudSnooze Scripts

All CloudSnooze testing scripts that interact with AWS should be used with the `--profile` flag:

### Manual AWS Commands

```bash
# General pattern
aws [service] [command] --profile cloudsnooze-testing

# Examples
aws ec2 describe-instances --profile cloudsnooze-testing
aws cloudformation deploy --template-file template.yaml --stack-name test-stack --profile cloudsnooze-testing
```

### CloudSnooze Setup Scripts

Our setup scripts now support the `--profile` parameter:

```bash
# For access key authentication setup
./scripts/setup_github_secrets.sh --profile cloudsnooze-testing

# For OIDC authentication setup
./scripts/setup_github_oidc.sh --profile cloudsnooze-testing
```

### Running Integration Tests Locally

When running tests locally, specify the profile:

```bash
# First, create a test instance using the profile
INSTANCE_ID=$(aws ec2 run-instances \
  --image-id ami-0c2b8ca1dad447f8a \
  --instance-type t2.micro \
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=cloudsnooze-local-test},{Key=Purpose,Value=Testing}]" \
  --query 'Instances[0].InstanceId' \
  --output text \
  --profile cloudsnooze-testing)

# Export required environment variables
export CLOUDSNOOZE_TEST_INSTANCE_ID=$INSTANCE_ID
export CLOUDSNOOZE_TEST_REGION=us-west-2
export AWS_PROFILE=cloudsnooze-testing  # This tells the AWS SDK to use this profile

# Run the tests
cd daemon/cloud/aws
go test -v -tags=integration ./...

# Clean up using the profile
aws ec2 terminate-instances --instance-ids $INSTANCE_ID --profile cloudsnooze-testing
```

## Setting the Profile in Your Shell

For convenience, you can set the AWS_PROFILE environment variable to make the profile the default for your current shell session:

```bash
# For Bash/Zsh
export AWS_PROFILE=cloudsnooze-testing

# For PowerShell
$env:AWS_PROFILE = "cloudsnooze-testing"

# Now you can use AWS CLI commands without specifying --profile
aws sts get-caller-identity  # Will use cloudsnooze-testing profile
```

## Profile Configuration Files

AWS CLI profiles are stored in:

- `~/.aws/credentials` - Contains access keys
- `~/.aws/config` - Contains region and output format settings

You can also edit these files directly. The profile section will look something like:

```ini
# In ~/.aws/credentials
[cloudsnooze-testing]
aws_access_key_id = AKIA...
aws_secret_access_key = your-secret-key

# In ~/.aws/config
[profile cloudsnooze-testing]
region = us-west-2
output = json
```

## Using Multiple Profiles for Different Environments

You can create multiple profiles for different environments:

```bash
# Set up a development testing profile
aws configure --profile cloudsnooze-dev-testing

# Set up a staging testing profile
aws configure --profile cloudsnooze-staging-testing

# Set up a production testing profile
aws configure --profile cloudsnooze-prod-testing
```

Then use the appropriate profile based on your current testing needs.

## Using AWS Profiles with GitHub Actions

GitHub Actions doesn't use AWS profiles directly, but you can set up multiple repositories or repository environments with different secrets to achieve a similar separation of credentials.

## Troubleshooting Profile Issues

### Profile Not Found

If you get "profile not found" errors:

1. Check that the profile name is spelled correctly
2. Verify the profile exists in your AWS credentials file
3. Try using the full path to your credentials file with the AWS_SHARED_CREDENTIALS_FILE environment variable

### Permissions Issues

If you have permission errors when using a profile:

1. Verify the profile is using the correct access keys
2. Check that the IAM user has the necessary permissions
3. Use `aws sts get-caller-identity --profile cloudsnooze-testing` to confirm you're using the correct account

### Conflicting Default Credentials

If AWS commands are using your default credentials instead of the profile:

1. Always explicitly use `--profile cloudsnooze-testing` with commands
2. Or set the AWS_PROFILE environment variable
3. Check that you don't have conflicting AWS environment variables set (e.g., AWS_ACCESS_KEY_ID)