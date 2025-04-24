# AWS Integration Testing Setup Guide for CloudSnooze

This document provides comprehensive instructions for setting up AWS and GitHub for integration testing with CloudSnooze. It covers the entire process from AWS account setup, permission configuration, and GitHub repository configuration to ensure seamless integration testing.

## Overview

The CloudSnooze system requires integration testing with AWS to ensure its core functionality works as expected. This guide will help you set up:

1. A dedicated AWS account for testing
2. The necessary IAM permissions
3. GitHub repository secrets for secure CI/CD integration
4. GitHub Actions workflow configuration
5. Resource cleanup and cost controls

## AWS Account Setup

### Option 1: Create a Dedicated AWS Account (Recommended)

For optimal isolation and security, create a dedicated AWS account for testing:

1. **Create an AWS Organization**:
   - Sign in to your main AWS account
   - Navigate to AWS Organizations
   - Create a new member account specifically for CloudSnooze testing
   
2. **Configure the Account**: 
   - Set a descriptive name (e.g., "cloudsnooze-testing")
   - Create a root email (e.g., cloudsnooze-testing@yourdomain.com)
   - Set strong password and MFA for the root account

### Option 2: Use Existing AWS Account with Isolation

If creating a new account isn't feasible:

1. **Create an IAM User**: 
   - Name it specifically for CloudSnooze testing (e.g., "cloudsnooze-ci-user")
   - Do not assign console access - this is API-only

2. **Apply Restrictive Permissions**:
   - Create a specific IAM policy that only allows access to test resources
   - Use resource tagging to isolate test resources

## IAM Permission Configuration

### Create Testing IAM User

1. **Create a new IAM user**:
   ```bash
   aws iam create-user --user-name github-cloudsnooze-test
   ```

2. **Create policy document** (save as `cloudsnooze-test-policy.json`):
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

3. **Create and attach policy**:
   ```bash
   aws iam create-policy --policy-name CloudSnoozeTestPolicy --policy-document file://cloudsnooze-test-policy.json
   aws iam attach-user-policy --user-name github-cloudsnooze-test --policy-arn arn:aws:iam::<ACCOUNT_ID>:policy/CloudSnoozeTestPolicy
   ```

4. **Generate access keys**:
   ```bash
   aws iam create-access-key --user-name github-cloudsnooze-test
   ```
   
   Save the `AccessKeyId` and `SecretAccessKey` for use in GitHub secrets.

### Optional: Set Up OIDC Authentication (Enhanced Security)

For improved security, configure OIDC authentication between GitHub and AWS:

1. **Create an IAM OIDC provider**:
   ```bash
   aws iam create-open-id-connect-provider \
     --url https://token.actions.githubusercontent.com \
     --client-id-list sts.amazonaws.com \
     --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
   ```

2. **Create IAM role for GitHub Actions**:
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

3. **Attach the CloudSnooze test policy to this role**

## Cost Control Measures

To avoid unexpected charges, set up budget alerts and resource cleanup:

### Create a Budget

1. **Create a budget for the test account**:
   ```bash
   aws budgets create-budget --account-id <ACCOUNT_ID> --budget file://budget.json --notifications-with-subscribers file://notifications.json
   ```

2. **budget.json example**:
   ```json
   {
     "BudgetName": "CloudSnooze-Test-Budget",
     "BudgetLimit": {
       "Amount": "20",
       "Unit": "USD"
     },
     "BudgetType": "COST",
     "TimeUnit": "MONTHLY"
   }
   ```

3. **notifications.json example**:
   ```json
   [
     {
       "Notification": {
         "ComparisonOperator": "GREATER_THAN",
         "NotificationType": "ACTUAL",
         "Threshold": 80,
         "ThresholdType": "PERCENTAGE"
       },
       "Subscribers": [
         {
           "Address": "your-email@example.com",
           "SubscriptionType": "EMAIL"
         }
       ]
     }
   ]
   ```

### Set Up Automated Resource Cleanup

1. **Deploy the cleanup Lambda**:
   ```bash
   aws cloudformation deploy \
     --template-file scripts/cleanup_lambda.yaml \
     --stack-name cloudsnooze-test-cleanup \
     --capabilities CAPABILITY_IAM
   ```

2. **cleanup_lambda.yaml**: See the file in the CloudSnooze repository at `scripts/cleanup_lambda.yaml`

## GitHub Repository Configuration

### Add GitHub Secrets

Add the following secrets to your GitHub repository:

1. Go to your repository → Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `AWS_ACCESS_KEY_ID`: The IAM user's access key
   - `AWS_SECRET_ACCESS_KEY`: The IAM user's secret key
   - `AWS_REGION`: The AWS region to use (e.g., "us-west-2")
   - `AWS_ROLE_ARN` (if using OIDC): The ARN of the IAM role

### Using GitHub CLI

Alternatively, use the GitHub CLI to add secrets:

```bash
gh secret set AWS_ACCESS_KEY_ID --body "<access-key>"
gh secret set AWS_SECRET_ACCESS_KEY --body "<secret-key>"
gh secret set AWS_REGION --body "us-west-2"
```

## GitHub Actions Workflow Configuration

The repository contains a workflow file `.github/workflows/aws-tests.yml` that runs the integration tests.

Key aspects of this workflow:

1. **Trigger Conditions**:
   - Pull requests to the main branch
   - Manual workflow dispatch
   - PRs with the "aws-test" label

2. **Authentication**:
   - Uses GitHub's OIDC provider (preferred) or access key authentication

3. **Resource Creation and Cleanup**:
   - Creates temporary EC2 instances for testing
   - Tags all resources with "Purpose: Testing"
   - Ensures cleanup even if tests fail

## Running Integration Tests Locally

To run the integration tests locally:

1. **Configure AWS credentials**:
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-west-2
   ```

2. **Create a test EC2 instance**:
   ```bash
   INSTANCE_ID=$(aws ec2 run-instances \
     --image-id ami-0c55b159cbfafe1f0 \
     --instance-type t2.micro \
     --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=cloudsnooze-local-test},{Key=Purpose,Value=Testing}]" \
     --query 'Instances[0].InstanceId' \
     --output text)
   
   echo "Created test instance: $INSTANCE_ID"
   export CLOUDSNOOZE_TEST_INSTANCE_ID=$INSTANCE_ID
   export CLOUDSNOOZE_TEST_REGION=us-west-2
   ```

3. **Run the tests**:
   ```bash
   cd daemon/cloud/aws
   go test -v -tags=integration ./...
   ```

4. **Clean up**:
   ```bash
   aws ec2 terminate-instances --instance-ids $INSTANCE_ID
   ```

## Troubleshooting

### Common Issues

1. **Permission Errors**:
   - Verify IAM policy has the correct permissions
   - Check GitHub secrets are set correctly
   - Test the IAM user's credentials locally

2. **Resources Not Found**:
   - Verify the AWS region settings
   - Check AMI availability in the selected region
   - Ensure the test instance is properly tagged

3. **Tests Timing Out**:
   - EC2 instance startup may take longer than expected
   - Increase timeout values in test code
   - Check AWS service health

### Logs and Monitoring

1. **CloudTrail Logs**:
   - Enable CloudTrail for API call logging
   - Filter for the IAM user's actions

2. **CloudWatch Logs**:
   - Configure the Lambda function to log to CloudWatch
   - Monitor EC2 instance state changes

## Best Practices

1. **Minimize Test Resources**:
   - Use the smallest viable instance types (t2.micro)
   - Keep test duration short
   - Ensure all resources are tagged for cleanup

2. **Ensure Cleanup**:
   - Always use the "if: always()" condition for cleanup steps
   - Implement redundant cleanup mechanisms
   - Set up an automated cleanup Lambda function

3. **Security**:
   - Rotate access keys regularly
   - Use OIDC authentication when possible
   - Limit permissions to the absolute minimum required

## AMI Selection for Testing

For reliable testing, use Amazon Linux 2 AMIs available in all regions:

| Region       | AMI ID                | Architecture |
|--------------|------------------------|--------------|
| us-west-2    | ami-0c55b159cbfafe1f0 | x86_64       |
| us-east-1    | ami-0c2b8ca1dad447f8a | x86_64       |
| us-west-1    | ami-0d9858aa3c6322f73 | x86_64       |
| eu-west-1    | ami-062a49a8152e4c413 | x86_64       |

For ARM64 testing:

| Region       | AMI ID                | Architecture |
|--------------|------------------------|--------------|
| us-west-2    | ami-0a4a377a7617e3fbd | ARM64        |
| us-east-1    | ami-0ae74ae55fc4f48a5 | ARM64        |

## Related Documents

For more information, refer to:
- [AWS Integration Testing Guide](aws_integration_testing.md)
- [CloudSnooze Plugin Architecture](../design/plugin-architecture.md)
- [AWS Plugin Documentation](../plugins/aws.md)