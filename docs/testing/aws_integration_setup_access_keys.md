# AWS Integration Testing Setup Guide for CloudSnooze (Access Keys Method)

This document provides instructions for setting up AWS integration testing with CloudSnooze using IAM user access keys for authentication. While not as secure as OIDC, this method is simpler to set up and works in all environments.

## Overview

The access keys method creates an IAM user with limited permissions and uses its access keys for GitHub Actions authentication. This approach:

- Works in all environments without special configuration
- Is simpler to set up than OIDC
- Uses long-lived credentials stored in GitHub secrets

## Prerequisites

Before you begin, you need:

- Administrative access to an AWS account dedicated for testing
- Administrative access to the GitHub repository
- AWS CLI and GitHub CLI installed locally

### Setting Up an AWS Profile (Recommended)

When working with a dedicated AWS testing account, it's recommended to set up a separate AWS CLI profile to avoid confusion with your default credentials:

```bash
# Create a new AWS profile for CloudSnooze testing
aws configure --profile cloudsnooze-testing
```

Enter the access key, secret key, and region for your testing account when prompted. For more detailed instructions on AWS profile setup, see the [AWS Profile Setup Guide](aws_profile_setup.md).

All commands in this guide can be run with the `--profile cloudsnooze-testing` flag to ensure you're using the correct AWS account.

## Step 1: Create an IAM User

First, create a dedicated IAM user for GitHub Actions:

```bash
# If using an AWS profile
aws --profile cloudsnooze-testing iam create-user --user-name github-cloudsnooze-test

# OR, if using default credentials
aws iam create-user --user-name github-cloudsnooze-test

# Create policy document (save as cloudsnooze-test-policy.json)
cat > cloudsnooze-test-policy.json << EOF
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
EOF

# Create the policy (with profile)
POLICY_ARN=$(aws --profile cloudsnooze-testing iam create-policy \
  --policy-name CloudSnoozeTestPolicy \
  --policy-document file://cloudsnooze-test-policy.json \
  --query 'Policy.Arn' --output text)

# Attach policy to user
aws --profile cloudsnooze-testing iam attach-user-policy \
  --user-name github-cloudsnooze-test \
  --policy-arn $POLICY_ARN

# OR, without profile:
# POLICY_ARN=$(aws iam create-policy \
#   --policy-name CloudSnoozeTestPolicy \
#   --policy-document file://cloudsnooze-test-policy.json \
#   --query 'Policy.Arn' --output text)
# 
# aws iam attach-user-policy \
#   --user-name github-cloudsnooze-test \
#   --policy-arn $POLICY_ARN
```

## Step 2: Generate Access Keys

Create access keys for the IAM user:

```bash
# Generate access keys (with profile)
ACCESS_KEY=$(aws --profile cloudsnooze-testing iam create-access-key --user-name github-cloudsnooze-test)

# OR, without profile:
# ACCESS_KEY=$(aws iam create-access-key --user-name github-cloudsnooze-test)

# Extract access key ID and secret key
ACCESS_KEY_ID=$(echo $ACCESS_KEY | jq -r '.AccessKey.AccessKeyId')
SECRET_ACCESS_KEY=$(echo $ACCESS_KEY | jq -r '.AccessKey.SecretAccessKey')

# Display the keys (save these securely!)
echo "Access Key ID: $ACCESS_KEY_ID"
echo "Secret Access Key: $SECRET_ACCESS_KEY"
```

**IMPORTANT**: Save these keys securely. After this point, you won't be able to retrieve the secret access key again.

## Step 3: Add Access Keys to GitHub Secrets

Add the access keys to your GitHub repository secrets:

1. **Using GitHub UI**:
   - Go to your repository → Settings → Secrets and variables → Actions
   - Add a new repository secret named `AWS_ACCESS_KEY_ID` with the value of your access key ID
   - Add a new repository secret named `AWS_SECRET_ACCESS_KEY` with the value of your secret access key
   - Add a new repository secret named `AWS_REGION` with your preferred AWS region (e.g., `us-west-2`)

2. **Using GitHub CLI**:
   ```bash
   gh secret set AWS_ACCESS_KEY_ID --body "$ACCESS_KEY_ID"
   gh secret set AWS_SECRET_ACCESS_KEY --body "$SECRET_ACCESS_KEY"
   gh secret set AWS_REGION --body "us-west-2"
   ```

3. **Using the helper script**:
   ```bash
   # Using profile
   ./scripts/setup_github_secrets.sh --profile cloudsnooze-testing
   
   # OR, using default credentials
   ./scripts/setup_github_secrets.sh
   ```

## Step 4: Update GitHub Actions Workflow

The repository contains a workflow file `.github/workflows/aws-tests.yml` that already supports access key authentication. Ensure it contains:

```yaml
jobs:
  test:
    # ... other configuration ...
    
    steps:
      # ... other steps ...
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
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
# Deploy the cleanup Lambda using CloudFormation (with profile)
aws --profile cloudsnooze-testing cloudformation deploy \
  --template-file scripts/aws_cleanup_lambda.yaml \
  --stack-name cloudsnooze-test-cleanup \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides MaxAgeHours=2 TagKey=Purpose TagValue=Testing

# OR, without profile:
# aws cloudformation deploy \
#   --template-file scripts/aws_cleanup_lambda.yaml \
#   --stack-name cloudsnooze-test-cleanup \
#   --capabilities CAPABILITY_IAM \
#   --parameter-overrides MaxAgeHours=2 TagKey=Purpose TagValue=Testing
```

This Lambda function will run every hour and clean up any test resources tagged with `Purpose: Testing` that are older than 2 hours.

## Step 6: Test the Integration

1. **Add the AWS test label to a PR**:
   Add the `aws-test` label to a pull request to trigger the integration tests.

2. **Manually trigger the workflow**:
   Go to Actions → AWS Integration Tests → Run workflow → Select a branch and run.

3. **Monitor the workflow**:
   Check the Actions tab to verify that the workflow can:
   - Successfully authenticate using the access keys
   - Create test resources in AWS
   - Run the integration tests
   - Clean up resources after completion

## Running Integration Tests Locally

To run the integration tests locally:

1. **Configure AWS credentials**:
   ```bash
   # Option 1: Set environment variables directly
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-west-2
   
   # Option 2: Use an AWS profile (recommended)
   export AWS_PROFILE=cloudsnooze-testing
   ```

2. **Create a test EC2 instance**:
   ```bash
   # If using AWS_PROFILE
   INSTANCE_ID=$(aws ec2 run-instances \
     --image-id ami-0c2b8ca1dad447f8a \
     --instance-type t2.micro \
     --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=cloudsnooze-local-test},{Key=Purpose,Value=Testing}]" \
     --query 'Instances[0].InstanceId' \
     --output text)

   # OR, if using an explicit profile
   # INSTANCE_ID=$(aws --profile cloudsnooze-testing ec2 run-instances \
   #   --image-id ami-0c2b8ca1dad447f8a \
   #   --instance-type t2.micro \
   #   --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=cloudsnooze-local-test},{Key=Purpose,Value=Testing}]" \
   #   --query 'Instances[0].InstanceId' \
   #   --output text)
   
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
   # If using AWS_PROFILE
   aws ec2 terminate-instances --instance-ids $INSTANCE_ID
   
   # OR, if using an explicit profile
   # aws --profile cloudsnooze-testing ec2 terminate-instances --instance-ids $INSTANCE_ID
   ```

## Best Practices for Access Key Security

1. **Limit access key permissions**: 
   - Use the principle of least privilege
   - Restrict to only the necessary actions and resources
   - Include conditions in IAM policies to further restrict usage

2. **Regular key rotation**:
   - Rotate access keys periodically (e.g., every 90 days)
   - To rotate keys:
     ```bash
     # Create new access key (with profile)
     NEW_ACCESS_KEY=$(aws --profile cloudsnooze-testing iam create-access-key --user-name github-cloudsnooze-test)
     
     # OR, without profile
     # NEW_ACCESS_KEY=$(aws iam create-access-key --user-name github-cloudsnooze-test)
     
     # Update GitHub secrets with the new key
     NEW_ACCESS_KEY_ID=$(echo $NEW_ACCESS_KEY | jq -r '.AccessKey.AccessKeyId')
     NEW_SECRET_ACCESS_KEY=$(echo $NEW_ACCESS_KEY | jq -r '.AccessKey.SecretAccessKey')
     gh secret set AWS_ACCESS_KEY_ID --body "$NEW_ACCESS_KEY_ID"
     gh secret set AWS_SECRET_ACCESS_KEY --body "$NEW_SECRET_ACCESS_KEY"
     
     # Delete old access key (with profile)
     aws --profile cloudsnooze-testing iam delete-access-key --user-name github-cloudsnooze-test --access-key-id OLD_ACCESS_KEY_ID
     
     # OR, without profile
     # aws iam delete-access-key --user-name github-cloudsnooze-test --access-key-id OLD_ACCESS_KEY_ID
     ```

3. **Monitor for key usage**:
   - Enable AWS CloudTrail to log all API calls
   - Set up alerts for suspicious activity
   - Regularly review the logs for unexpected usage

4. **Additional protection**:
   - Enable MFA for the IAM user if it has console access
   - Use IP restrictions in IAM policies if GitHub Actions IPs are predictable

## Troubleshooting

### Common Access Key Issues

1. **Insufficient Permissions**:
   - Check the IAM policy attached to the user
   - Verify the policy includes all required EC2 actions
   - Check CloudTrail logs for specific permission denials

2. **Invalid Credentials**:
   - Verify the GitHub secrets contain the correct values
   - Check if the access key has been deactivated or deleted
   - Make sure the region configuration is correct

3. **Resource Creation Failures**:
   - Check for service quotas that might be limiting resource creation
   - Verify the AMI ID exists in the specified region
   - Check that instance type is available in the specified region

## Additional Resources

- [AWS IAM User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html)
- [AWS Security Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [GitHub Encrypted Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)