# Production Deployment Guide

## Overview
This `prod` branch contains production-ready deployment configurations for the locker management system.

## 🚀 Quick Production Deployment

### Prerequisites
- Docker and Docker Compose installed
- Git repository on `prod` branch
- Port 3000, 5432, 6379 available

### One-Command Deployment
```bash
./deploy-prod.sh
```

## 📋 Manual Deployment Steps

### 1. Switch to Production Branch
```bash
git checkout prod
```

### 2. Deploy with Docker Compose
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Verify Deployment
```bash
curl http://localhost:3000/api/v1/health
```

## 🏗️ Production Architecture

### Services
- **Application**: Go Fiber server (Port 3000)
- **Database**: PostgreSQL 16 (Port 5432)
- **Cache**: Redis 7 (Port 6379)

### Security Features
- Non-root container user
- Secure passwords
- Health checks
- Restart policies

## 🔧 Configuration

### Environment Variables (.env.prod)
- `DB_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis server address
- `JWT_SECRET`: JWT signing secret
- `GIN_MODE`: Set to "release" for production

### Docker Configuration
- Multi-stage build for optimized image size
- Health checks for all services
- Persistent data volumes
- Automatic restart policies

## 📊 Monitoring

### Health Checks
```bash
# Application health
curl http://localhost:3000/api/v1/health

# Database health
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U locker

# Redis health
docker-compose -f docker-compose.prod.yml exec redis redis-cli ping
```

### Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Application only
docker-compose -f docker-compose.prod.yml logs -f app

# Database only
docker-compose -f docker-compose.prod.yml logs -f postgres
```

## 🛑 Management Commands

### Stop Services
```bash
docker-compose -f docker-compose.prod.yml down
```

### Update Application
```bash
git pull origin prod
docker-compose -f docker-compose.prod.yml build app
docker-compose -f docker-compose.prod.yml up -d app
```

### Backup Database
```bash
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U locker locker > backup.sql
```

### Scale Services (if needed)
```bash
docker-compose -f docker-compose.prod.yml up -d --scale app=2
```

## 🔐 Security Considerations

1. **Change Default Passwords**: Update all passwords in `.env.prod`
2. **Network Security**: Configure firewall rules for production
3. **SSL/TLS**: Add HTTPS reverse proxy (nginx/traefik)
4. **Monitoring**: Set up monitoring and alerting
5. **Backups**: Implement regular database backups

## 🌍 External Access

For external network access, consider:

1. **Reverse Proxy**: Use nginx or traefik with SSL
2. **Cloud Deployment**: Deploy to AWS, GCP, or DigitalOcean
3. **VPN**: Set up VPN for secure remote access

## 📈 Performance Tuning

### PostgreSQL
- Adjust `shared_buffers` and `effective_cache_size`
- Configure connection pooling
- Set up read replicas if needed

### Redis
- Configure memory limits
- Set up persistence options
- Use Redis Cluster for high availability

### Application
- Adjust Go runtime parameters
- Configure connection pools
- Monitor memory usage

## 🚨 Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   sudo lsof -i :3000
   sudo lsof -i :5432
   sudo lsof -i :6379
   ```

2. **Permission Issues**
   ```bash
   sudo chown -R $USER:$USER .
   ```

3. **Docker Issues**
   ```bash
   docker system prune -a
   docker-compose -f docker-compose.prod.yml down
   docker-compose -f docker-compose.prod.yml up -d
   ```

## 📞 Support

- Check logs: `docker-compose -f docker-compose.prod.yml logs -f`
- Health status: `docker-compose -f docker-compose.prod.yml ps`
- System resources: `docker stats`
