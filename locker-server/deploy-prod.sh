#!/bin/bash

# Production Deployment Script
set -e

echo "🚀 Starting Production Deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if we're on prod branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "prod" ]; then
    print_error "You must be on 'prod' branch to deploy to production!"
    print_status "Current branch: $CURRENT_BRANCH"
    print_status "Run: git checkout prod"
    exit 1
fi

print_success "✅ On production branch"

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    print_warning "You have uncommitted changes. Commit them before deploying."
    git status --short
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Stop existing containers
print_status "🛑 Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down || true

# Clean up old images (optional)
print_status "🧹 Cleaning up old images..."
docker image prune -f || true

# Build and start production containers
print_status "🔨 Building production containers..."
docker-compose -f docker-compose.prod.yml build --no-cache

print_status "🚀 Starting production services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
print_status "⏳ Waiting for services to be healthy..."
sleep 10

# Check service health
print_status "🏥 Checking service health..."

# Check PostgreSQL
if docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -U locker -d locker > /dev/null 2>&1; then
    print_success "✅ PostgreSQL is healthy"
else
    print_error "❌ PostgreSQL is not healthy"
    exit 1
fi

# Check Redis
if docker-compose -f docker-compose.prod.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
    print_success "✅ Redis is healthy"
else
    print_error "❌ Redis is not healthy"
    exit 1
fi

# Check Application
sleep 5
if curl -f http://localhost:3000/api/v1/health > /dev/null 2>&1; then
    print_success "✅ Application is healthy"
else
    print_error "❌ Application is not healthy"
    print_status "Checking application logs..."
    docker-compose -f docker-compose.prod.yml logs app
    exit 1
fi

# Show running containers
print_status "📋 Production containers status:"
docker-compose -f docker-compose.prod.yml ps

# Show application info
print_success "🎉 Production deployment completed successfully!"
echo
print_status "📍 Service URLs:"
echo "   🌐 Application: http://localhost:3000"
echo "   🗄️  PostgreSQL: localhost:5432"
echo "   🔴 Redis: localhost:6379"
echo
print_status "📊 Health check: curl http://localhost:3000/api/v1/health"
print_status "📝 Logs: docker-compose -f docker-compose.prod.yml logs -f"
print_status "🛑 Stop: docker-compose -f docker-compose.prod.yml down"

echo
print_success "🚀 Production server is now running!"
