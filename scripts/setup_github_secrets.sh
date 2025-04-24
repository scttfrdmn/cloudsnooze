#!/bin/bash
# CloudSnooze GitHub Secrets Setup Script for AWS Integration Testing
# Copyright 2025 Scott Friedman and CloudSnooze Contributors
# SPDX-License-Identifier: Apache-2.0

set -e

echo "CloudSnooze GitHub Secrets Setup for AWS Integration Testing"
echo "==========================================================="
echo

# Check for required tools
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI not installed. Please install it first."
    echo "Visit: https://cli.github.com/manual/installation"
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

# Get AWS credentials
echo
echo "Please enter your AWS credentials for integration testing:"
echo "NOTE: These credentials will be stored as GitHub secrets"
echo

# Ask for credentials
read -p "AWS Access Key ID: " AWS_ACCESS_KEY_ID
if [[ -z "$AWS_ACCESS_KEY_ID" ]]; then
    echo "Error: AWS Access Key ID cannot be empty"
    exit 1
fi

read -p "AWS Secret Access Key: " -s AWS_SECRET_ACCESS_KEY
echo
if [[ -z "$AWS_SECRET_ACCESS_KEY" ]]; then
    echo "Error: AWS Secret Access Key cannot be empty"
    exit 1
fi

read -p "AWS Region [us-west-2]: " AWS_REGION
AWS_REGION=${AWS_REGION:-us-west-2}

# This script only configures access key authentication
# For OIDC setup, please refer to the aws_integration_setup_oidc.md guide

# Confirm before proceeding
echo
echo "The following GitHub secrets will be created or updated:"
echo "- AWS_ACCESS_KEY_ID"
echo "- AWS_SECRET_ACCESS_KEY"
echo "- AWS_REGION: $AWS_REGION"
echo

read -p "Do you want to continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 1
fi

# Set GitHub secrets
echo "Setting GitHub secrets..."

echo "Adding AWS_ACCESS_KEY_ID..."
echo "$AWS_ACCESS_KEY_ID" | gh secret set AWS_ACCESS_KEY_ID

echo "Adding AWS_SECRET_ACCESS_KEY..."
echo "$AWS_SECRET_ACCESS_KEY" | gh secret set AWS_SECRET_ACCESS_KEY

echo "Adding AWS_REGION..."
echo "$AWS_REGION" | gh secret set AWS_REGION


echo
echo "GitHub secrets have been successfully set up!"
echo
echo "Next steps:"
echo "1. Create AWS test resources using CloudFormation (see docs/testing/aws_integration_setup_access_keys.md)"
echo "2. Verify the GitHub Actions workflow runs successfully"
echo "3. Test locally with 'go test -tags=integration ./daemon/cloud/aws/...'"
echo
echo "For more information, see the docs/testing/aws_integration_setup_access_keys.md document."