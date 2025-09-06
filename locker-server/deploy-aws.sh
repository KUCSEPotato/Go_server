#!/bin/bash

# AWS EC2 Deployment Script
# This script will create an EC2 instance and deploy your locker server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
INSTANCE_TYPE="t3.small"
KEY_NAME="locker-server-key"
SECURITY_GROUP_NAME="locker-server-sg"
IMAGE_ID="ami-0c7217cdde317cfec" # Ubuntu 22.04 LTS (update for your region)

print_status "🚀 Starting AWS EC2 Deployment..."

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    print_error "AWS CLI is not configured. Please run 'aws configure' first."
    exit 1
fi

print_success "✅ AWS CLI is configured"

# Get current region
REGION=$(aws configure get region)
if [ -z "$REGION" ]; then
    REGION="us-east-1"
    print_warning "No region set, using default: $REGION"
fi

print_status "📍 Using region: $REGION"

# Create or check key pair
print_status "🔑 Checking key pair..."
if ! aws ec2 describe-key-pairs --key-names $KEY_NAME --region $REGION > /dev/null 2>&1; then
    print_status "Creating key pair: $KEY_NAME"
    aws ec2 create-key-pair --key-name $KEY_NAME --region $REGION --query 'KeyMaterial' --output text > ${KEY_NAME}.pem
    chmod 400 ${KEY_NAME}.pem
    print_success "✅ Key pair created: ${KEY_NAME}.pem"
else
    print_success "✅ Key pair exists: $KEY_NAME"
fi

# Get default VPC
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --region $REGION --query 'Vpcs[0].VpcId' --output text)
if [ "$VPC_ID" = "None" ] || [ -z "$VPC_ID" ]; then
    print_error "No default VPC found. Please create one or specify a VPC ID."
    exit 1
fi

print_status "🌐 Using VPC: $VPC_ID"

# Create or check security group
print_status "🔒 Setting up security group..."
SG_ID=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=$SECURITY_GROUP_NAME" --region $REGION --query 'SecurityGroups[0].GroupId' --output text 2>/dev/null || echo "None")

if [ "$SG_ID" = "None" ] || [ -z "$SG_ID" ]; then
    print_status "Creating security group: $SECURITY_GROUP_NAME"
    SG_ID=$(aws ec2 create-security-group \
        --group-name $SECURITY_GROUP_NAME \
        --description "Security group for locker server" \
        --vpc-id $VPC_ID \
        --region $REGION \
        --query 'GroupId' \
        --output text)
    
    # Add rules
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_ID \
        --protocol tcp \
        --port 22 \
        --cidr 0.0.0.0/0 \
        --region $REGION
    
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_ID \
        --protocol tcp \
        --port 3000 \
        --cidr 0.0.0.0/0 \
        --region $REGION
    
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_ID \
        --protocol tcp \
        --port 80 \
        --cidr 0.0.0.0/0 \
        --region $REGION
    
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_ID \
        --protocol tcp \
        --port 443 \
        --cidr 0.0.0.0/0 \
        --region $REGION
    
    print_success "✅ Security group created: $SG_ID"
else
    print_success "✅ Security group exists: $SG_ID"
fi

# Get subnet ID
SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --region $REGION --query 'Subnets[0].SubnetId' --output text)

# Launch EC2 instance
print_status "🚀 Launching EC2 instance..."
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id $IMAGE_ID \
    --instance-type $INSTANCE_TYPE \
    --key-name $KEY_NAME \
    --security-group-ids $SG_ID \
    --subnet-id $SUBNET_ID \
    --user-data file://aws-user-data.sh \
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=locker-server},{Key=Environment,Value=production}]" \
    --region $REGION \
    --query 'Instances[0].InstanceId' \
    --output text)

print_success "✅ Instance launched: $INSTANCE_ID"

# Wait for instance to be running
print_status "⏳ Waiting for instance to be running..."
aws ec2 wait instance-running --instance-ids $INSTANCE_ID --region $REGION

# Get public IP
PUBLIC_IP=$(aws ec2 describe-instances \
    --instance-ids $INSTANCE_ID \
    --region $REGION \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text)

print_success "🎉 Instance is running!"
echo
print_status "📋 Instance Details:"
echo "   Instance ID: $INSTANCE_ID"
echo "   Public IP: $PUBLIC_IP"
echo "   Key Pair: ${KEY_NAME}.pem"
echo "   Security Group: $SG_ID"
echo
print_status "📝 Next Steps:"
echo "   1. Wait 3-5 minutes for user data script to complete"
echo "   2. SSH into your instance:"
echo "      ssh -i ${KEY_NAME}.pem ubuntu@${PUBLIC_IP}"
echo "   3. Clone your repository:"
echo "      git clone https://github.com/KUCSEPotato/Go_server.git locker-server"
echo "      cd locker-server"
echo "      git checkout prod"
echo "   4. Deploy your application:"
echo "      ./deploy.sh"
echo "   5. Access your application:"
echo "      http://${PUBLIC_IP}:3000"
echo
print_warning "💰 Remember to stop the instance when not in use to save costs!"
print_status "🛑 To stop: aws ec2 stop-instances --instance-ids $INSTANCE_ID --region $REGION"
