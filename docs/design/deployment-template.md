<!--
Copyright 2025 Scott Friedman and CloudSnooze Contributors
SPDX-License-Identifier: Apache-2.0
-->

# CloudSnooze Deployment Templates

This document provides templates and examples for deploying CloudSnooze in various cloud environments.

## AWS Deployment

### CloudFormation Template

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'CloudSnooze Deployment'

Parameters:
  InstanceType:
    Type: String
    Default: t3.medium
    Description: Instance type for deployment
  KeyName:
    Type: AWS::EC2::KeyPair::KeyName
    Description: SSH key for instance access
  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC for deployment
  SubnetId:
    Type: AWS::EC2::Subnet::Id
    Description: Subnet for deployment
  CloudSnoozeVersion:
    Type: String
    Default: 0.1.0
    Description: CloudSnooze version to deploy

Resources:
  CloudSnoozeInstanceRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: ec2.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
      Policies:
        - PolicyName: CloudSnoozeInstancePolicy
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - ec2:StopInstances
                  - ec2:DescribeInstances
                  - ec2:CreateTags
                  - ec2:DescribeTags
                Resource: '*'
              - Effect: Allow
                Action:
                  - logs:CreateLogGroup
                  - logs:CreateLogStream
                  - logs:PutLogEvents
                Resource: 'arn:aws:logs:*:*:log-group:/aws/cloudsnooze/*'

  CloudSnoozeInstanceProfile:
    Type: AWS::IAM::InstanceProfile
    Properties:
      Roles:
        - !Ref CloudSnoozeInstanceRole

  CloudSnoozeSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for CloudSnooze instance
      VpcId: !Ref VpcId
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 22
          ToPort: 22
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0

  CloudSnoozeInstance:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: !Ref InstanceType
      KeyName: !Ref KeyName
      IamInstanceProfile: !Ref CloudSnoozeInstanceProfile
      SubnetId: !Ref SubnetId
      SecurityGroupIds:
        - !GetAtt CloudSnoozeSecurityGroup.GroupId
      ImageId: !FindInMap [RegionMap, !Ref 'AWS::Region', AMI]
      UserData:
        Fn::Base64: !Sub |
          #!/bin/bash -xe
          # Install CloudSnooze
          mkdir -p /opt/cloudsnooze
          cd /opt/cloudsnooze
          
          # Install dependencies
          apt-get update
          apt-get install -y curl systemd
          
          # Download and install CloudSnooze
          curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/v${CloudSnoozeVersion}/cloudsnooze_${CloudSnoozeVersion}_amd64.deb
          dpkg -i cloudsnooze_${CloudSnoozeVersion}_amd64.deb
          
          # Configure CloudSnooze
          cat > /etc/snooze/snooze.json << 'EOF'
          {
            "CPUThresholdPercent": 5.0,
            "MemoryThresholdPercent": 10.0,
            "NetworkThresholdKBps": 5.0,
            "DiskIOThresholdKBps": 10.0,
            "GPUThresholdPercent": 5.0,
            "InputIdleThresholdSecs": 1800,
            "NaptimeMinutes": 60,
            "CheckIntervalSeconds": 60,
            "EnableInstanceTags": true,
            "DetailedInstanceTags": true,
            "TaggingPrefix": "CloudSnooze",
            "GPUMonitoringEnabled": true,
            "AWSRegion": "${AWS::Region}",
            "Logging": {
              "EnableCloudWatch": true,
              "CloudWatchLogGroup": "/aws/cloudsnooze/instance-logs"
            }
          }
          EOF
          
          # Start the service
          systemctl enable snoozed
          systemctl start snoozed

Mappings:
  RegionMap:
    us-east-1:
      AMI: ami-0c55b159cbfafe1f0
    us-east-2:
      AMI: ami-05d72852800cbf29e
    us-west-1:
      AMI: ami-0f2176987ee50226e
    us-west-2:
      AMI: ami-01fee56b22f308154

Outputs:
  InstanceId:
    Description: Instance ID of the CloudSnooze instance
    Value: !Ref CloudSnoozeInstance
  PublicDNS:
    Description: Public DNS of the CloudSnooze instance
    Value: !GetAtt CloudSnoozeInstance.PublicDnsName
```

### Terraform Template

```hcl
provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  default = "us-east-1"
}

variable "instance_type" {
  default = "t3.medium"
}

variable "key_name" {
  description = "SSH key for instance access"
}

variable "vpc_id" {
  description = "VPC for deployment"
}

variable "subnet_id" {
  description = "Subnet for deployment"
}

variable "cloudsnooze_version" {
  default = "0.1.0"
  description = "CloudSnooze version to deploy"
}

# IAM Role and Policy
resource "aws_iam_role" "cloudsnooze_role" {
  name = "cloudsnooze-instance-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "cloudsnooze_policy" {
  name = "cloudsnooze-instance-policy"
  role = aws_iam_role.cloudsnooze_role.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:StopInstances",
          "ec2:DescribeInstances",
          "ec2:CreateTags",
          "ec2:DescribeTags"
        ]
        Effect = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Effect = "Allow"
        Resource = "arn:aws:logs:*:*:log-group:/aws/cloudsnooze/*"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "cloudsnooze_profile" {
  name = "cloudsnooze-instance-profile"
  role = aws_iam_role.cloudsnooze_role.name
}

# Security Group
resource "aws_security_group" "cloudsnooze_sg" {
  name        = "cloudsnooze-sg"
  description = "Security group for CloudSnooze instance"
  vpc_id      = var.vpc_id
  
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# EC2 Instance
resource "aws_instance" "cloudsnooze_instance" {
  ami                    = "ami-0c55b159cbfafe1f0" # Ubuntu 20.04 LTS
  instance_type          = var.instance_type
  key_name               = var.key_name
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [aws_security_group.cloudsnooze_sg.id]
  iam_instance_profile   = aws_iam_instance_profile.cloudsnooze_profile.name
  
  user_data = <<-EOF
    #!/bin/bash -xe
    # Install CloudSnooze
    mkdir -p /opt/cloudsnooze
    cd /opt/cloudsnooze
    
    # Install dependencies
    apt-get update
    apt-get install -y curl systemd
    
    # Download and install CloudSnooze
    curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/v${var.cloudsnooze_version}/cloudsnooze_${var.cloudsnooze_version}_amd64.deb
    dpkg -i cloudsnooze_${var.cloudsnooze_version}_amd64.deb
    
    # Configure CloudSnooze
    cat > /etc/snooze/snooze.json << 'EOF'
    {
      "CPUThresholdPercent": 5.0,
      "MemoryThresholdPercent": 10.0,
      "NetworkThresholdKBps": 5.0,
      "DiskIOThresholdKBps": 10.0,
      "GPUThresholdPercent": 5.0,
      "InputIdleThresholdSecs": 1800,
      "NaptimeMinutes": 60,
      "CheckIntervalSeconds": 60,
      "EnableInstanceTags": true,
      "DetailedInstanceTags": true,
      "TaggingPrefix": "CloudSnooze",
      "GPUMonitoringEnabled": true,
      "AWSRegion": "${var.aws_region}",
      "Logging": {
        "EnableCloudWatch": true,
        "CloudWatchLogGroup": "/aws/cloudsnooze/instance-logs"
      }
    }
    EOF
    
    # Start the service
    systemctl enable snoozed
    systemctl start snoozed
  EOF
  
  tags = {
    Name = "CloudSnooze-Instance"
  }
}

output "instance_id" {
  value = aws_instance.cloudsnooze_instance.id
}

output "public_dns" {
  value = aws_instance.cloudsnooze_instance.public_dns
}
```

## User Data Scripts

### Amazon Linux 2

```bash
#!/bin/bash
# CloudSnooze installation for Amazon Linux 2

# Install dependencies
yum update -y
yum install -y curl tar

# Download and install CloudSnooze
curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze-0.1.0-1.x86_64.rpm
# Use latest version instead:
# curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/latest/cloudsnooze-latest.x86_64.rpm
rpm -i cloudsnooze-0.1.0-1.x86_64.rpm

# Configure CloudSnooze
cat > /etc/snooze/snooze.json << 'EOF'
{
  "CPUThresholdPercent": 5.0,
  "MemoryThresholdPercent": 10.0,
  "NetworkThresholdKBps": 5.0,
  "DiskIOThresholdKBps": 10.0,
  "GPUThresholdPercent": 5.0,
  "InputIdleThresholdSecs": 1800,
  "NaptimeMinutes": 60,
  "CheckIntervalSeconds": 60,
  "EnableInstanceTags": true,
  "DetailedInstanceTags": true,
  "TaggingPrefix": "CloudSnooze",
  "GPUMonitoringEnabled": true,
  "AWSRegion": "us-east-1", 
  "Logging": {
    "EnableCloudWatch": true,
    "CloudWatchLogGroup": "/aws/cloudsnooze/instance-logs"
  }
}
EOF

# Start the service
systemctl enable snoozed
systemctl start snoozed
```

### Ubuntu/Debian

```bash
#!/bin/bash
# CloudSnooze installation for Ubuntu/Debian

# Install dependencies
apt-get update
apt-get install -y curl systemd

# Download and install CloudSnooze
curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_amd64.deb
# Use latest version instead:
# curl -LO https://github.com/scttfrdmn/cloudsnooze/releases/download/latest/cloudsnooze-latest_amd64.deb
dpkg -i cloudsnooze_0.1.0_amd64.deb

# Configure CloudSnooze
cat > /etc/snooze/snooze.json << 'EOF'
{
  "CPUThresholdPercent": 5.0,
  "MemoryThresholdPercent": 10.0,
  "NetworkThresholdKBps": 5.0,
  "DiskIOThresholdKBps": 10.0,
  "GPUThresholdPercent": 5.0,
  "InputIdleThresholdSecs": 1800,
  "NaptimeMinutes": 60,
  "CheckIntervalSeconds": 60,
  "EnableInstanceTags": true,
  "DetailedInstanceTags": true,
  "TaggingPrefix": "CloudSnooze",
  "GPUMonitoringEnabled": true,
  "AWSRegion": "us-east-1", 
  "Logging": {
    "EnableCloudWatch": true,
    "CloudWatchLogGroup": "/aws/cloudsnooze/instance-logs"
  }
}
EOF

# Start the service
systemctl enable snoozed
systemctl start snoozed
```

## Version Information

This deployment template is compatible with CloudSnooze v0.1.0 and uses the versioning scheme documented in the packaging system.

For latest package links, use:
- DEB: `https://github.com/scttfrdmn/cloudsnooze/releases/download/latest/cloudsnooze-latest_amd64.deb`
- RPM: `https://github.com/scttfrdmn/cloudsnooze/releases/download/latest/cloudsnooze-latest.x86_64.rpm`

For specific version links, use:
- DEB: `https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze_0.1.0_amd64.deb`
- RPM: `https://github.com/scttfrdmn/cloudsnooze/releases/download/v0.1.0/cloudsnooze-0.1.0-1.x86_64.rpm`