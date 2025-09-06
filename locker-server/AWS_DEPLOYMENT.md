# AWS EC2 Deployment Guide

## 🚀 Quick AWS Deployment

### Prerequisites
```bash
# 1. Install AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# 2. Configure AWS CLI
aws configure
# Enter your Access Key ID, Secret Access Key, Region, and Output format
```

### One-Command Deployment
```bash
# Deploy to AWS EC2
./deploy-aws.sh
```

## 📋 Manual Deployment Steps

### 1. Push Code to GitHub
```bash
# Commit all changes
git add .
git commit -m "feat: AWS deployment configuration"

# Push to GitHub
git push origin prod
```

### 2. Run AWS Deployment Script
```bash
./deploy-aws.sh
```

The script will:
- ✅ Create EC2 instance (t3.small)
- ✅ Create security group with ports 22, 80, 443, 3000
- ✅ Create key pair for SSH access
- ✅ Install Docker and Docker Compose
- ✅ Provide connection details

### 3. Connect to Your Instance
```bash
# SSH into your EC2 instance (replace with your actual IP and key)
ssh -i locker-server-key.pem ubuntu@YOUR_PUBLIC_IP
```

### 4. Deploy Your Application
```bash
# On the EC2 instance
git clone https://github.com/KUCSEPotato/Go_server.git locker-server
cd locker-server
git checkout prod

# Deploy
./deploy.sh
```

## 🌐 Access Your Application

Your application will be available at:
- **Application**: `http://YOUR_PUBLIC_IP:3000`
- **Health Check**: `http://YOUR_PUBLIC_IP:3000/api/v1/health`

## � Architecture

```
Internet
    ↓
EC2 Instance (t3.small)
    ├── Docker Container: Go Application (Port 3000)
    ├── Docker Container: PostgreSQL (Port 5432)
    └── Docker Container: Redis (Port 6379)
```

## 🔧 Management Commands

### Check Application Status
```bash
# SSH to instance
ssh -i locker-server-key.pem ubuntu@YOUR_PUBLIC_IP

# Check containers
docker-compose -f docker-compose.prod.yml ps

# Check logs
docker-compose -f docker-compose.prod.yml logs -f app
```

### Update Application
```bash
# On EC2 instance
cd /home/ubuntu/locker-server
git pull origin prod
./deploy.sh
```

### Stop/Start Instance
```bash
# Stop instance (saves money)
aws ec2 stop-instances --instance-ids YOUR_INSTANCE_ID

# Start instance
aws ec2 start-instances --instance-ids YOUR_INSTANCE_ID
```

## 💰 Cost Estimation

### Monthly Costs (us-east-1):
- **EC2 t3.small**: ~$15/month
- **Data Transfer**: ~$1-5/month
- **EBS Storage**: ~$1/month
- **Total**: ~$17-21/month

### Cost Optimization:
- Stop instance when not in use
- Use t3.micro for development ($6/month)
- Monitor usage with AWS Cost Explorer

## 🔒 Security Best Practices

### 1. Update Security Group (Optional)
```bash
# Restrict SSH access to your IP only
aws ec2 authorize-security-group-ingress \
  --group-id YOUR_SECURITY_GROUP_ID \
  --protocol tcp \
  --port 22 \
  --cidr YOUR_IP/32
```

### 2. Enable HTTPS (Production)
```bash
# Install Certbot on EC2
sudo apt install certbot

# Get SSL certificate (requires domain)
sudo certbot certonly --standalone -d your-domain.com
```

### 3. Setup Firewall
```bash
# Already configured in user-data script
sudo ufw status
```

## 🚨 Troubleshooting

### Common Issues

1. **Can't connect to instance**
   ```bash
   # Check security group allows SSH (port 22)
   # Verify key pair permissions
   chmod 400 locker-server-key.pem
   ```

2. **Application not accessible**
   ```bash
   # Check if containers are running
   docker-compose -f docker-compose.prod.yml ps
   
   # Check application logs
   docker-compose -f docker-compose.prod.yml logs app
   ```

3. **Out of memory**
   ```bash
   # Upgrade to larger instance
   aws ec2 modify-instance-attribute \
     --instance-id YOUR_INSTANCE_ID \
     --instance-type Value=t3.medium
   ```

## 🔄 Advanced: Auto-scaling Setup

For production with high traffic, consider:
- Application Load Balancer
- Auto Scaling Group
- RDS for database
- ElastiCache for Redis
- CloudWatch monitoring

## 📞 Support

- **AWS Documentation**: https://docs.aws.amazon.com/
- **EC2 Pricing**: https://aws.amazon.com/ec2/pricing/
- **Free Tier**: 750 hours/month of t2.micro for 12 months
