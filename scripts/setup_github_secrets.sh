#!/bin/bash
# CloudSnooze GitHub Secrets Setup Script for AWS Integration Testing
# Copyright 2025 Scott Friedman and CloudSnooze Contributors
# SPDX-License-Identifier: Apache-2.0

set -e

PROFILE_OPTION=""

# Process command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --profile)
      PROFILE_OPTION="--profile $2"
      echo "Using AWS profile: $2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--profile PROFILE_NAME]"
      exit 1
      ;;
  esac
done

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
echo "Please choose how to obtain AWS credentials for integration testing:"
echo "1. Enter credentials manually"
echo "2. Use current AWS profile credentials"
echo "NOTE: These credentials will be stored as GitHub secrets"
echo

read -p "Enter your choice (1/2): " CRED_CHOICE

case $CRED_CHOICE in
  1)
    # Manual credential entry
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
    ;;
  2)
    # Use AWS profile credentials
    echo "Retrieving credentials from AWS profile..."
    if [[ -z "$PROFILE_OPTION" ]]; then
      echo "No profile specified. Using default profile."
    fi
    
    # Get credentials from AWS CLI
    CREDENTIALS=$(aws $PROFILE_OPTION configure get aws_access_key_id aws_secret_access_key 2>/dev/null)
    if [[ $? -ne 0 ]]; then
      echo "Error retrieving credentials from AWS profile."
      echo "Make sure your AWS profile is properly configured."
      exit 1
    fi
    
    # Parse the credentials
    AWS_ACCESS_KEY_ID=$(echo "$CREDENTIALS" | grep aws_access_key_id | awk '{print $2}')
    AWS_SECRET_ACCESS_KEY=$(echo "$CREDENTIALS" | grep aws_secret_access_key | awk '{print $2}')
    
    if [[ -z "$AWS_ACCESS_KEY_ID" || -z "$AWS_SECRET_ACCESS_KEY" ]]; then
      echo "Error: Could not retrieve valid credentials from AWS profile."
      exit 1
    fi
    
    echo "Successfully retrieved credentials from AWS profile."
    ;;
  *)
    echo "Invalid choice. Exiting."
    exit 1
    ;;
esac

# Get AWS region
read -p "AWS Region [us-west-2]: " AWS_REGION
if [[ -z "$AWS_REGION" ]]; then
  # Try to get region from profile if available
  if [[ -n "$PROFILE_OPTION" ]]; then
    PROFILE_REGION=$(aws $PROFILE_OPTION configure get region 2>/dev/null)
    if [[ -n "$PROFILE_REGION" ]]; then
      AWS_REGION=$PROFILE_REGION
      echo "Using region from AWS profile: $AWS_REGION"
    else
      AWS_REGION="us-west-2"
    fi
  else
    AWS_REGION="us-west-2"
  fi
fi

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