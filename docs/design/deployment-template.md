# CloudSnooze - Deployment Template

This document provides templates and instructions for deploying CloudSnooze in various cloud environments.

## AWS Deployment

### IAM Role Configuration

The following IAM policy allows CloudSnooze to stop instances and tag them:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:StopInstances",
            "Resource": "arn:aws:ec2:*:*:instance/*",
            "Condition": {
                "StringEquals": {"ec2:ResourceID": "${ec2:InstanceID}"}
            }
        },
        {
            "Effect": "Allow",
            "Action": "ec2:DescribeInstances",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ec2:CreateTags",
                "ec2:DeleteTags"
            ],
            "Resource": "arn:aws:ec2:*:*:instance/*",
            "Condition": {
                "StringEquals": {"ec2:ResourceID": "${ec2:InstanceID}"}
            }
        }
    ]
}
```

### CloudFormation Template

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'CloudSnooze IAM Role and Instance Profile'

Resources:
  CloudSnoozeRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: ec2.amazonaws.com
            Action: sts:AssumeRole
      Path: /
      ManagedPolicyArns:
        - !Ref CloudSnoozePolicy

  CloudSnoozePolicy:
    Type: AWS::IAM::ManagedPolicy
    Properties:
      Description: Policy for CloudSnooze to stop instances and manage tags
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Action: ec2:StopInstances
            Resource: !Sub arn:aws:ec2:*:${AWS::AccountId}:instance/*
            Condition:
              StringEquals:
                ec2:ResourceID: "${ec2:InstanceID}"
          - Effect: Allow
            Action: ec2:DescribeInstances
            Resource: "*"
          - Effect: Allow
            Action:
              - ec2:CreateTags
              - ec2:DeleteTags
            Resource: !Sub arn:aws:ec2:*:${AWS::AccountId}:instance/*
            Condition:
              StringEquals:
                ec2:ResourceID: "${ec2:InstanceID}"
          - Effect: Allow
            Action:
              - logs:CreateLogGroup
              - logs:CreateLogStream
              - logs:PutLogEvents
            Resource: !Sub arn:aws:logs:*:${AWS::AccountId}:log-group:CloudSnooze:*
            Condition:
              StringEquals:
                aws:SourceAccount: "${aws:PrincipalAccount}"

  CloudSnoozeInstanceProfile:
    Type: AWS::IAM::InstanceProfile
    Properties:
      Path: /
      Roles:
        - !Ref CloudSnoozeRole

Outputs:
  InstanceProfile:
    Description: Instance profile for EC2 instances to use CloudSnooze
    Value: !Ref CloudSnoozeInstanceProfile
    Export:
      Name: !Sub "${AWS::StackName}-InstanceProfile"
```

### Terraform Template

```hcl
# CloudSnooze IAM Configuration

resource "aws_iam_policy" "cloudsnooze_policy" {
  name        = "CloudSnoozePolicy"
  description = "Policy for CloudSnooze to stop instances and manage tags"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "ec2:StopInstances"
        Resource = "arn:aws:ec2:*:*:instance/*"
        Condition = {
          StringEquals = {
            "ec2:ResourceID" = "${ec2:InstanceID}"
          }
        }
      },
      {
        Effect   = "Allow"
        Action   = "ec2:DescribeInstances"
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ec2:CreateTags",
          "ec2:DeleteTags"
        ]
        Resource = "arn:aws:ec2:*:*:instance/*"
        Condition = {
          StringEquals = {
            "ec2:ResourceID" = "${ec2:InstanceID}"
          }
        }
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:log-group:CloudSnooze:*"
      }
    ]
  })
}

resource "aws_iam_role" "cloudsnooze_role" {
  name = "CloudSnoozeRole"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cloudsnooze_attachment" {
  role       = aws_iam_role.cloudsnooze_role.name
  policy_arn = aws_iam_policy.cloudsnooze_policy.arn
}

resource "aws_iam_instance_profile" "cloudsnooze_profile" {
  name = "CloudSnoozeProfile"
  role = aws_iam_role.cloudsnooze_role.name
}

# Example EC2 instance with CloudSnooze
resource "aws_instance" "example" {
  ami           = "ami-12345678"  # Replace with appropriate AMI
  instance_type = "t3.micro"
  
  iam_instance_profile = aws_iam_instance_profile.cloudsnooze_profile.name
  
  user_data = <<-EOF
#!/bin/bash
# Install CloudSnooze
wget https://github.com/scttfrdmn/cloudsnooze/releases/download/v1.0.0/cloudsnooze_1.0.0_amd64.deb
dpkg -i cloudsnooze_1.0.0_amd64.deb

# Configure and start the service
systemctl enable snoozed
systemctl start snoozed
  EOF
  
  tags = {
    Name = "CloudSnooze-Enabled-Instance"
  }
}
```

### AWS CDK (TypeScript)

```typescript
import * as cdk from 'aws-cdk-lib';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

export class CloudSnoozeStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // CloudSnooze IAM Policy
    const cloudSnoozePolicy = new iam.ManagedPolicy(this, 'CloudSnoozePolicy', {
      statements: [
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          actions: ['ec2:StopInstances'],
          resources: ['arn:aws:ec2:*:*:instance/*'],
          conditions: {
            StringEquals: {
              'ec2:ResourceID': '${ec2:InstanceID}'
            }
          }
        }),
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          actions: ['ec2:DescribeInstances'],
          resources: ['*']
        }),
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          actions: ['ec2:CreateTags', 'ec2:DeleteTags'],
          resources: ['arn:aws:ec2:*:*:instance/*'],
          conditions: {
            StringEquals: {
              'ec2:ResourceID': '${ec2:InstanceID}'
            }
          }
        }),
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          actions: [
            'logs:CreateLogGroup',
            'logs:CreateLogStream',
            'logs:PutLogEvents'
          ],
          resources: [`arn:aws:logs:*:${this.account}:log-group:CloudSnooze:*`]
        })
      ]
    });

    // CloudSnooze Role
    const role = new iam.Role(this, 'CloudSnoozeRole', {
      assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
      managedPolicies: [cloudSnoozePolicy]
    });

    // CloudSnooze Instance Profile
    const instanceProfile = new iam.CfnInstanceProfile(this, 'CloudSnoozeInstanceProfile', {
      roles: [role.roleName]
    });

    // Output the instance profile name
    new cdk.CfnOutput(this, 'InstanceProfileName', {
      value: instanceProfile.ref,
      description: 'Name of the instance profile for EC2 instances'
    });
  }
}
```

### AWS CLI Commands

```bash
# Create IAM policy
aws iam create-policy \
  --policy-name CloudSnoozePolicy \
  --policy-document file://cloudsnooze-policy.json

# Create IAM role
aws iam create-role \
  --role-name CloudSnoozeRole \
  --assume-role-policy-document file://ec2-trust-policy.json

# Attach policy to role
aws iam attach-role-policy \
  --role-name CloudSnoozeRole \
  --policy-arn arn:aws:iam::ACCOUNT_ID:policy/CloudSnoozePolicy

# Create instance profile
aws iam create-instance-profile \
  --instance-profile-name CloudSnoozeProfile

# Add role to instance profile
aws iam add-role-to-instance-profile \
  --instance-profile-name CloudSnoozeProfile \
  --role-name CloudSnoozeRole

# Associate instance profile with existing instance
aws ec2 associate-iam-instance-profile \
  --instance-id i-1234567890abcdef0 \
  --iam-instance-profile Name=CloudSnoozeProfile
```

## User Data Scripts

### Amazon Linux 2

```bash
#!/bin/bash
# Install CloudSnooze on Amazon Linux 2

# Install dependencies
yum update -y
yum install -y wget

# Download and install CloudSnooze
wget https://github.com/scttfrdmn/cloudsnooze/releases/download/v1.0.0/cloudsnooze-1.0.0-1.x86_64.rpm
rpm -i cloudsnooze-1.0.0-1.x86_64.rpm

# Configure CloudSnooze
cat > /etc/snooze/snooze.json << 'EOF'
{
  "check_interval_seconds": 60,
  "naptime_minutes": 30,
  "cpu_threshold_percent": 10.0,
  "memory_threshold_percent": 30.0,
  "network_threshold_kbps": 50.0,
  "disk_io_threshold_kbps": 100.0,
  "input_idle_threshold_secs": 900,
  "gpu_monitoring_enabled": false,
  "aws_region": "us-east-1",
  "enable_instance_tags": true,
  "tagging_prefix": "CloudSnooze",
  "logging": {
    "log_level": "info",
    "enable_file_logging": true,
    "log_file_path": "/var/log/cloudsnooze.log",
    "enable_syslog": false,
    "enable_cloudwatch": false
  },
  "monitoring_mode": "basic"
}
EOF

# Enable and start the service
systemctl enable snoozed
systemctl start snoozed
```

### Ubuntu/Debian

```bash
#!/bin/bash
# Install CloudSnooze on Ubuntu/Debian

# Update and install dependencies
apt-get update
apt-get install -y wget

# Download and install CloudSnooze
wget https://github.com/scttfrdmn/cloudsnooze/releases/download/v1.0.0/cloudsnooze_1.0.0_amd64.deb
dpkg -i cloudsnooze_1.0.0_amd64.deb

# Configure CloudSnooze
cat > /etc/snooze/snooze.json << 'EOF'
{
  "check_interval_seconds": 60,
  "naptime_minutes": 30,
  "cpu_threshold_percent": 10.0,
  "memory_threshold_percent": 30.0,
  "network_threshold_kbps": 50.0,
  "disk_io_threshold_kbps": 100.0,
  "input_idle_threshold_secs": 900,
  "gpu_monitoring_enabled": false,
  "aws_region": "us-east-1",
  "enable_instance_tags": true,
  "tagging_prefix": "CloudSnooze",
  "logging": {
    "log_level": "info",
    "enable_file_logging": true,
    "log_file_path": "/var/log/cloudsnooze.log",
    "enable_syslog": false,
    "enable_cloudwatch": false
  },
  "monitoring_mode": "basic"
}
EOF

# Enable and start the service
systemctl enable snoozed
systemctl start snoozed
```
