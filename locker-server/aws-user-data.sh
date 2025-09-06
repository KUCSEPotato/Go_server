#!/bin/bash
# AWS EC2 User Data Script
# This script runs when the EC2 instance first starts

# Update system
apt-get update -y
apt-get upgrade -y

# Install Docker
apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.21.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create docker-compose symlink
ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

# Add ubuntu user to docker group
usermod -aG docker ubuntu

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Install Git and other utilities
apt-get install -y git htop curl wget unzip

# Install AWS CLI v2
cd /tmp
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

# Create application directory
mkdir -p /home/ubuntu/locker-server
chown ubuntu:ubuntu /home/ubuntu/locker-server

# Install Node.js (for any future frontend needs)
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs

# Setup basic firewall
ufw allow ssh
ufw allow 3000
ufw allow 80
ufw allow 443
echo "y" | ufw enable

# Create a script for easy deployment
cat > /home/ubuntu/deploy.sh << 'EOF'
#!/bin/bash
echo "🚀 Deploying Locker Server..."

# Navigate to project directory
cd /home/ubuntu/locker-server

# Pull latest changes
git pull origin prod

# Build and start containers
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d

# Show status
echo "📊 Container Status:"
docker-compose -f docker-compose.prod.yml ps

echo "✅ Deployment completed!"
echo "🌐 Access your application at: http://$(curl -s ifconfig.me):3000"
EOF

chmod +x /home/ubuntu/deploy.sh
chown ubuntu:ubuntu /home/ubuntu/deploy.sh

# Create log file
touch /var/log/user-data.log
echo "$(date): User data script completed" >> /var/log/user-data.log
