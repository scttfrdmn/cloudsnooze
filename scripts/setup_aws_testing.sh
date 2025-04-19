#!/bin/bash
# CloudSnooze AWS Testing Setup Script
# This script helps set up AWS testing infrastructure for CloudSnooze

set -e

echo "CloudSnooze AWS Testing Setup"
echo "============================"
echo

# Ensure prerequisites are installed
if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI not installed. Please install it first."
    exit 1
fi

if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI not installed. Please install it first."
    exit 1
fi

# Check AWS credentials
echo "Checking AWS authentication..."
if ! aws sts get-caller-identity &> /dev/null; then
    echo "Error: Not authenticated with AWS. Run 'aws configure' first."
    exit 1
fi

echo "Successfully authenticated with AWS."
ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
echo "AWS Account ID: $ACCOUNT_ID"

# Check GitHub authentication
echo "Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub. Run 'gh auth login' first."
    exit 1
fi

echo "Successfully authenticated with GitHub."
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo "GitHub Repository: $REPO"

# Get confirmation before proceeding
echo
echo "This script will:"
echo "1. Create an IAM user for GitHub Actions"
echo "2. Set up GitHub repository secrets"
echo "3. Create budget alerts for the AWS account"
echo "4. Set up test resource cleanup automation"
echo
read -p "Do you want to continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 1
fi

# Step 1: Create IAM user for GitHub Actions
echo
echo "Creating IAM user for GitHub Actions..."
USER_NAME="github-cloudsnooze-test"

# Check if user already exists
if aws iam get-user --user-name $USER_NAME &> /dev/null; then
    echo "User $USER_NAME already exists. Skipping creation."
else
    aws iam create-user --user-name $USER_NAME
    echo "Created IAM user: $USER_NAME"
    
    # Create policy document for test permissions
    cat > /tmp/test-policy.json << EOF
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

    # Create policy
    POLICY_ARN=$(aws iam create-policy --policy-name CloudSnoozeTestPolicy --policy-document file:///tmp/test-policy.json --query "Policy.Arn" --output text)
    echo "Created IAM policy: $POLICY_ARN"
    
    # Attach policy to user
    aws iam attach-user-policy --user-name $USER_NAME --policy-arn $POLICY_ARN
    echo "Attached policy to user"
    
    # Create access key
    ACCESS_KEY=$(aws iam create-access-key --user-name $USER_NAME)
    ACCESS_KEY_ID=$(echo $ACCESS_KEY | jq -r .AccessKey.AccessKeyId)
    SECRET_ACCESS_KEY=$(echo $ACCESS_KEY | jq -r .AccessKey.SecretAccessKey)
    
    echo "Created access key for user"
fi

# Step 2: Set up GitHub repository secrets
echo
echo "Setting up GitHub repository secrets..."

if [[ -n "$ACCESS_KEY_ID" && -n "$SECRET_ACCESS_KEY" ]]; then
    echo "Adding AWS_ACCESS_KEY_ID secret to GitHub repository..."
    echo "$ACCESS_KEY_ID" | gh secret set AWS_ACCESS_KEY_ID
    
    echo "Adding AWS_SECRET_ACCESS_KEY secret to GitHub repository..."
    echo "$SECRET_ACCESS_KEY" | gh secret set AWS_SECRET_ACCESS_KEY
    
    echo "GitHub secrets created successfully."
else
    echo "Skipping GitHub secret creation. No new access keys were generated."
    echo "If you need to update the secrets, you'll need to create new access keys manually."
fi

# Step 3: Create budget alerts
echo
echo "Setting up AWS budget alerts..."

# Create a budget with alerts at 50%, 80%, and 100%
cat > /tmp/budget.json << EOF
{
    "BudgetName": "CloudSnoozeTestBudget",
    "BudgetType": "COST",
    "BudgetLimit": {
        "Amount": "20",
        "Unit": "USD"
    },
    "CostFilters": {},
    "CostTypes": {
        "IncludeTax": true,
        "IncludeSubscription": true,
        "UseBlended": false,
        "IncludeRefund": false,
        "IncludeCredit": false,
        "IncludeUpfront": true,
        "IncludeRecurring": true,
        "IncludeOtherSubscription": true,
        "IncludeDiscount": true,
        "UseAmortized": false
    },
    "TimeUnit": "MONTHLY",
    "TimePeriod": {
        "Start": $(date +%s),
        "End": 2584186236
    },
    "NotificationsWithSubscribers": [
        {
            "Notification": {
                "NotificationType": "ACTUAL",
                "ComparisonOperator": "GREATER_THAN",
                "Threshold": 50,
                "ThresholdType": "PERCENTAGE",
                "NotificationState": "ALARM"
            },
            "Subscribers": [
                {
                    "SubscriptionType": "EMAIL",
                    "Address": "$(aws iam get-user --query User.UserName --output text)@example.com"
                }
            ]
        },
        {
            "Notification": {
                "NotificationType": "ACTUAL",
                "ComparisonOperator": "GREATER_THAN",
                "Threshold": 80,
                "ThresholdType": "PERCENTAGE",
                "NotificationState": "ALARM"
            },
            "Subscribers": [
                {
                    "SubscriptionType": "EMAIL",
                    "Address": "$(aws iam get-user --query User.UserName --output text)@example.com"
                }
            ]
        },
        {
            "Notification": {
                "NotificationType": "ACTUAL",
                "ComparisonOperator": "GREATER_THAN",
                "Threshold": 100,
                "ThresholdType": "PERCENTAGE",
                "NotificationState": "ALARM"
            },
            "Subscribers": [
                {
                    "SubscriptionType": "EMAIL",
                    "Address": "$(aws iam get-user --query User.UserName --output text)@example.com"
                }
            ]
        }
    ]
}
EOF

# Note: Budget API requires special permissions not commonly granted
echo "To create the budget, please go to the AWS Billing console:"
echo "https://console.aws.amazon.com/billing/home#/budgets/create"
echo "Use the configuration from scripts/cloudsnooze_test_budget.json"

# Create the JSON file for manual upload
cp /tmp/budget.json scripts/cloudsnooze_test_budget.json
echo "Budget configuration saved to scripts/cloudsnooze_test_budget.json"

# Step 4: Set up test resource cleanup automation
echo
echo "Setting up test resource cleanup automation..."

# Create Lambda function for cleanup
cat > /tmp/cleanup.py << EOF
import boto3
import datetime

def lambda_handler(event, context):
    """
    Lambda function to clean up test resources
    """
    ec2 = boto3.resource('ec2')
    
    # Find all instances with the Testing tag
    instances = ec2.instances.filter(
        Filters=[{'Name': 'tag:Purpose', 'Values': ['Testing']}]
    )
    
    # Terminate any test instances older than 2 hours
    current_time = datetime.datetime.now(datetime.timezone.utc)
    terminated = []
    
    for instance in instances:
        # Get launch time
        launch_time = instance.launch_time
        if (current_time - launch_time).total_seconds() > 7200:  # 2 hours in seconds
            instance.terminate()
            terminated.append(instance.id)
    
    return {
        'terminatedInstances': terminated
    }
EOF

# Create CloudFormation template for the Lambda function
cat > scripts/cleanup_lambda.yaml << EOF
AWSTemplateFormatVersion: '2010-09-09'
Description: 'CloudSnooze Test Resource Cleanup'

Resources:
  CleanupLambdaRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      Policies:
        - PolicyName: EC2CleanupPolicy
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - ec2:DescribeInstances
                  - ec2:TerminateInstances
                Resource: '*'
                Condition:
                  StringEquals:
                    'ec2:ResourceTag/Purpose': 'Testing'

  CleanupLambda:
    Type: AWS::Lambda::Function
    Properties:
      Handler: index.lambda_handler
      Role: !GetAtt CleanupLambdaRole.Arn
      Runtime: python3.9
      Timeout: 60
      Code:
        ZipFile: |
          import boto3
          import datetime
          
          def lambda_handler(event, context):
              """
              Lambda function to clean up test resources
              """
              ec2 = boto3.resource('ec2')
              
              # Find all instances with the Testing tag
              instances = ec2.instances.filter(
                  Filters=[{'Name': 'tag:Purpose', 'Values': ['Testing']}]
              )
              
              # Terminate any test instances older than 2 hours
              current_time = datetime.datetime.now(datetime.timezone.utc)
              terminated = []
              
              for instance in instances:
                  # Get launch time
                  launch_time = instance.launch_time
                  if (current_time - launch_time).total_seconds() > 7200:  # 2 hours in seconds
                      instance.terminate()
                      terminated.append(instance.id)
              
              return {
                  'terminatedInstances': terminated
              }

  CleanupSchedule:
    Type: AWS::Events::Rule
    Properties:
      ScheduleExpression: rate(1 hour)
      State: ENABLED
      Targets:
        - Arn: !GetAtt CleanupLambda.Arn
          Id: CleanupScheduleTarget

  LambdaPermission:
    Type: AWS::Lambda::Permission
    Properties:
      Action: lambda:InvokeFunction
      FunctionName: !GetAtt CleanupLambda.Arn
      Principal: events.amazonaws.com
      SourceArn: !GetAtt CleanupSchedule.Arn

Outputs:
  CleanupLambdaArn:
    Description: ARN of the cleanup Lambda function
    Value: !GetAtt CleanupLambda.Arn
EOF

echo "Created CloudFormation template for cleanup automation"
echo "To deploy: aws cloudformation deploy --template-file scripts/cleanup_lambda.yaml --stack-name cloudsnooze-test-cleanup --capabilities CAPABILITY_IAM"

echo
echo "AWS Testing Setup Complete!"
echo "============================="
echo
echo "Next steps:"
echo "1. Create a budget in the AWS Billing console using the provided template"
echo "2. Deploy the cleanup Lambda function with CloudFormation"
echo "3. Run a test GitHub Actions workflow"
echo
echo "For more details, see the README.md file."