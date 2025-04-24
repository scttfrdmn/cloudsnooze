#!/bin/bash
# CloudSnooze GitHub OIDC Setup Script for AWS Integration Testing
# Copyright 2025 Scott Friedman and CloudSnooze Contributors
# SPDX-License-Identifier: Apache-2.0

set -e

echo "CloudSnooze GitHub OIDC Setup for AWS Integration Testing"
echo "========================================================"
echo

# Check for required tools
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI not installed. Please install it first."
    echo "Visit: https://cli.github.com/manual/installation"
    exit 1
fi

if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI not installed. Please install it first."
    echo "Visit: https://aws.amazon.com/cli/"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq not installed. Please install it first."
    echo "Visit: https://stedolan.github.io/jq/download/"
    exit 1
fi

# Check GitHub authentication
echo "Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub. Run 'gh auth login' first."
    exit 1
fi

echo "Successfully authenticated with GitHub."
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo "GitHub Repository: $REPO"

# Check AWS authentication
echo "Checking AWS authentication..."
if ! aws sts get-caller-identity &> /dev/null; then
    echo "Error: Not authenticated with AWS. Run 'aws configure' first."
    exit 1
fi

echo "Successfully authenticated with AWS."
ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
echo "AWS Account ID: $ACCOUNT_ID"

echo
echo "This script will:"
echo "1. Create an IAM OIDC provider for GitHub Actions"
echo "2. Create an IAM role with necessary permissions"
echo "3. Set up GitHub repository secrets for AWS integration testing"
echo

read -p "Do you want to continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 1
fi

# Get AWS region
read -p "AWS Region [us-west-2]: " AWS_REGION
AWS_REGION=${AWS_REGION:-us-west-2}

# Step 1: Create IAM OIDC provider
echo
echo "Creating IAM OIDC provider for GitHub Actions..."

# Check if provider already exists
if aws iam list-open-id-connect-providers | grep -q "token.actions.githubusercontent.com"; then
    echo "OIDC provider for GitHub Actions already exists."
else
    aws iam create-open-id-connect-provider \
        --url https://token.actions.githubusercontent.com \
        --client-id-list sts.amazonaws.com \
        --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
    echo "Created OIDC provider for GitHub Actions."
fi

# Step 2: Create trust policy
echo
echo "Creating IAM role trust policy..."

# Extract repository owner and name
REPO_OWNER=$(echo $REPO | cut -d '/' -f 1)
REPO_NAME=$(echo $REPO | cut -d '/' -f 2)

cat > /tmp/trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:${REPO_OWNER}/${REPO_NAME}:*"
        }
      }
    }
  ]
}
EOF

# Step 3: Create permissions policy
echo "Creating IAM policy for CloudSnooze testing..."

cat > /tmp/permissions-policy.json << EOF
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

# Step 4: Create or update role
ROLE_NAME="CloudSnoozeGitHubActionsRole"
echo "Creating IAM role: $ROLE_NAME..."

# Check if role exists
ROLE_EXISTS=false
if aws iam get-role --role-name $ROLE_NAME &> /dev/null; then
    ROLE_EXISTS=true
    echo "Role $ROLE_NAME already exists. Updating trust policy..."
    aws iam update-assume-role-policy --role-name $ROLE_NAME --policy-document file:///tmp/trust-policy.json
else
    echo "Creating new role $ROLE_NAME..."
    aws iam create-role --role-name $ROLE_NAME --assume-role-policy-document file:///tmp/trust-policy.json
fi

# Step 5: Create and attach policy
POLICY_NAME="CloudSnoozeTestPolicy"
if ! $ROLE_EXISTS || ! aws iam list-attached-role-policies --role-name $ROLE_NAME --query "AttachedPolicies[?PolicyName=='$POLICY_NAME'].PolicyArn" --output text | grep -q "arn:aws"; then
    echo "Creating and attaching policy $POLICY_NAME..."
    
    # Check if policy exists
    POLICY_ARN=$(aws iam list-policies --query "Policies[?PolicyName=='$POLICY_NAME'].Arn" --output text)
    
    if [[ -z "$POLICY_ARN" || "$POLICY_ARN" == "None" ]]; then
        POLICY_ARN=$(aws iam create-policy --policy-name $POLICY_NAME --policy-document file:///tmp/permissions-policy.json --query "Policy.Arn" --output text)
        echo "Created policy: $POLICY_ARN"
    else
        echo "Policy $POLICY_NAME already exists with ARN: $POLICY_ARN"
    fi
    
    # Attach policy to role
    aws iam attach-role-policy --role-name $ROLE_NAME --policy-arn $POLICY_ARN
    echo "Attached policy to role $ROLE_NAME"
else
    echo "Policy $POLICY_NAME is already attached to role $ROLE_NAME"
fi

# Step 6: Get role ARN
ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query "Role.Arn" --output text)
echo "Role ARN: $ROLE_ARN"

# Step 7: Set GitHub secrets
echo
echo "Setting up GitHub repository secrets..."

echo "Adding AWS_ROLE_ARN..."
echo "$ROLE_ARN" | gh secret set AWS_ROLE_ARN

echo "Adding AWS_REGION..."
echo "$AWS_REGION" | gh secret set AWS_REGION

echo
echo "GitHub OIDC setup complete!"
echo
echo "Next steps:"
echo "1. Create AWS test resources using CloudFormation (see docs/testing/aws_integration_setup_oidc.md)"
echo "2. Verify the GitHub Actions workflow runs successfully"
echo "3. Test the OIDC authentication by manually triggering the workflow"
echo
echo "For more information, see the docs/testing/aws_integration_setup_oidc.md document."