# AWS Integration Testing for CloudSnooze

This document explains how to set up and run AWS integration tests for CloudSnooze.

## Setup Overview

CloudSnooze uses AWS integration tests to verify that cloud provider functionality works correctly. The setup process creates:

1. An IAM user with limited permissions for testing
2. Budget alerts to prevent unexpected charges
3. GitHub repository secrets for secure credential storage
4. Automated cleanup of test resources

## Prerequisites

Before running the setup script, ensure you have:

- AWS CLI installed and configured with administrative permissions
- GitHub CLI installed and authenticated to your repository
- jq installed for JSON processing

## Automated Setup

We provide a script to automate most of the setup process:

```bash
# Run the setup script
./scripts/setup_aws_testing.sh
```

The script will:
1. Check prerequisites
2. Create an IAM user for GitHub Actions
3. Set up GitHub repository secrets
4. Provide templates for budget alerts
5. Create a CloudFormation template for test resource cleanup

## Manual Steps

Some steps require manual intervention:

### 1. Creating an AWS Organization (Optional)

For better isolation, consider using AWS Organizations to create a dedicated test account:

1. Sign in to your main AWS account
2. Navigate to AWS Organizations
3. Create a new member account for testing
4. Use "Switch Role" to access the test account

### 2. Setting Up Budget Alerts

The setup script creates a template but you need to manually create the budget:

1. Go to AWS Billing console: https://console.aws.amazon.com/billing/home#/budgets/create
2. Upload the configuration from `scripts/cloudsnooze_test_budget.json`
3. Verify notification emails are correct

### 3. Deploying the Cleanup Lambda

Deploy the CloudFormation stack for automated cleanup:

```bash
aws cloudformation deploy \
  --template-file scripts/cleanup_lambda.yaml \
  --stack-name cloudsnooze-test-cleanup \
  --capabilities CAPABILITY_IAM
```

## Running Integration Tests

Integration tests are tagged with `integration` and excluded from normal test runs.

### Running Locally

To run integration tests locally:

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-west-2

# Run integration tests
cd daemon/cloud/aws
go test -v -tags=integration ./...
```

### GitHub Actions Workflow

The integration tests will run automatically on pull requests with the `aws-test` label or when manually triggered via the GitHub Actions UI.

## Security Considerations

- **Resource Limits**: All test resources are tagged with `Purpose: Testing` and automatically cleaned up
- **Permission Boundaries**: The IAM user has minimal permissions required for testing
- **Cost Control**: Budget alerts prevent unexpected charges
- **Secret Management**: Credentials are stored securely in GitHub Secrets

## Troubleshooting

If you encounter issues with AWS integration tests:

1. **Permission Errors**: Verify the IAM user has the correct permissions
2. **Timeout Errors**: Check if AWS resource creation is taking too long
3. **Cleanup Failures**: Manually verify and terminate any lingering test resources
4. **GitHub Secrets**: Ensure secrets are correctly configured in the repository

For more help, see the AWS CloudShell Guide in the CloudSnooze documentation.