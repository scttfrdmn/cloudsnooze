# AWS Integration Testing Setup Guide for CloudSnooze (OIDC Method)

This document provides instructions for setting up AWS integration testing with CloudSnooze using GitHub's OIDC provider for authentication. This method is more secure than using long-lived access keys.

## Overview

The OIDC (OpenID Connect) authentication method creates a trust relationship between GitHub Actions and AWS. This approach:

- Eliminates the need to store long-lived AWS credentials in GitHub
- Provides temporary, short-lived credentials for each workflow run
- Reduces the risk of credential exposure or compromise
- Allows for fine-grained permission control

## Prerequisites

Before you begin, you need:

- Administrative access to an AWS account dedicated for testing
- Administrative access to the GitHub repository
- AWS CLI and GitHub CLI installed locally

## Step 1: Create an IAM OIDC Provider in AWS

First, create an IAM OIDC identity provider to establish trust with GitHub:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

## Step 2: Create an IAM Role for GitHub Actions

1. Create a policy document for trust relationships (save as `trust-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:scttfrdmn/cloudsnooze:*"
        }
      }
    }
  ]
}
```

2. Create a policy document for permissions (save as `permissions-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:StartInstances",
        "ec2:StopInstances",
        "ec2:CreateTags",
        "ec2:RunInstances",
        "ec2:TerminateInstances"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "aws:RequestTag/Purpose": "Testing"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeImages",
        "ec2:DescribeInstanceTypes"
      ],
      "Resource": "*"
    }
  ]
}
```

3. Create the IAM role:

```bash
# Create the role with the trust policy
aws iam create-role \
  --role-name CloudSnoozeGitHubActionsRole \
  --assume-role-policy-document file://trust-policy.json

# Create the permissions policy
aws iam create-policy \
  --policy-name CloudSnoozeTestPolicy \
  --policy-document file://permissions-policy.json

# Attach the policy to the role
aws iam attach-role-policy \
  --role-name CloudSnoozeGitHubActionsRole \
  --policy-arn arn:aws:iam::<ACCOUNT_ID>:policy/CloudSnoozeTestPolicy
```

4. Get the role ARN:

```bash
aws iam get-role --role-name CloudSnoozeGitHubActionsRole --query Role.Arn --output text
```

## Step 3: Add the Role ARN to GitHub Secrets

Add the role ARN to your GitHub repository secrets:

1. **Using GitHub UI**:
   - Go to your repository → Settings → Secrets and variables → Actions
   - Add a new repository secret named `AWS_ROLE_ARN` with the value of your role ARN
   - Add a new repository secret named `AWS_REGION` with your preferred AWS region (e.g., `us-west-2`)

2. **Using GitHub CLI**:
   ```bash
   gh secret set AWS_ROLE_ARN --body "arn:aws:iam::<ACCOUNT_ID>:role/CloudSnoozeGitHubActionsRole"
   gh secret set AWS_REGION --body "us-west-2"
   ```

## Step 4: Update GitHub Actions Workflow

The repository contains a workflow file `.github/workflows/aws-tests.yml` that already supports OIDC authentication. Ensure it contains:

```yaml
jobs:
  test:
    # ... other configuration ...
    permissions:
      contents: read
      id-token: write # Required for OIDC
    
    steps:
      # ... other steps ...
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ secrets.AWS_REGION || 'us-west-2' }}
      
      # ... remaining steps ...
```

## Step 5: Set Up Cost Control

### Create a Budget

Create a budget using the AWS Management Console:

1. Go to AWS Billing Dashboard → Budgets → Create budget
2. Choose "Cost budget"
3. Set a monthly budget amount (e.g., $20)
4. Set up alerts at 50%, 80%, and 100% of the budget
5. Add email notifications to appropriate team members

### Deploy Resource Cleanup Lambda

To automatically clean up test resources:

```bash
# Deploy the cleanup Lambda using CloudFormation
aws cloudformation deploy \
  --template-file scripts/aws_cleanup_lambda.yaml \
  --stack-name cloudsnooze-test-cleanup \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides MaxAgeHours=2 TagKey=Purpose TagValue=Testing
```

This Lambda function will run every hour and clean up any test resources tagged with `Purpose: Testing` that are older than 2 hours.

## Step 6: Test the Integration

1. **Add the AWS test label to a PR**:
   Add the `aws-test` label to a pull request to trigger the integration tests.

2. **Manually trigger the workflow**:
   Go to Actions → AWS Integration Tests → Run workflow → Select a branch and run.

3. **Monitor the workflow**:
   Check the Actions tab to verify that the workflow can:
   - Successfully assume the IAM role
   - Create test resources in AWS
   - Run the integration tests
   - Clean up resources after completion

## Troubleshooting

### Common OIDC Issues

1. **Thumbprint Mismatch**:
   - If the OIDC provider's thumbprint is incorrect, authentication will fail
   - The thumbprint `6938fd4d98bab03faadb97b34396831e3780aea1` is correct for GitHub Actions as of 2025
   - If updated, check GitHub's documentation for the current thumbprint

2. **Incorrect Repository Pattern**:
   - The `StringLike` condition in the trust policy must match your repository
   - Ensure the pattern is `repo:scttfrdmn/cloudsnooze:*` (or your actual repository path)

3. **Missing Permissions**:
   - Ensure the GitHub Actions workflow has the `id-token: write` permission
   - Verify the IAM role has the correct permissions for EC2 operations

## Local Testing with OIDC Role

To test locally using the same role:

1. Install the AWS Session Token Plugin:
   ```bash
   pip install aws-session-token-plugin
   ```

2. Configure temporary credentials using web identity:
   ```bash
   aws configure set plugins.session-token aws-session-token
   aws configure set web_identity_token_file /path/to/jwt_token.txt
   aws configure set role_arn arn:aws:iam::<ACCOUNT_ID>:role/CloudSnoozeGitHubActionsRole
   aws configure set role_session_name local-testing
   ```

3. Run integration tests locally:
   ```bash
   cd daemon/cloud/aws
   go test -v -tags=integration ./...
   ```

## Additional Resources

- [AWS Security Blog: Use OIDC with GitHub Actions](https://aws.amazon.com/blogs/security/use-iam-roles-to-connect-github-actions-to-actions-in-aws/)
- [GitHub Docs: Configuring OpenID Connect in AWS](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [AWS IAM Roles Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html)