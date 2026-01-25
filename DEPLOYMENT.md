# Nivo Deployment Guide

## Server Information

- **IP**: 157.245.96.200
- **OS**: Debian 13 (Trixie)
- **Deploy User**: nivo
- **Deploy Directory**: /opt/nivo

## Domains

| Domain | Purpose | Status |
|--------|---------|--------|
| nivomoney.com | User app | Active |
| www.nivomoney.com | Redirect to apex | Active |
| admin.nivomoney.com | Admin dashboard | Active |
| app.nivomoney.com | User admin app | Active |
| api.nivomoney.com | API Gateway | Pending backend (returns 503) |
| verify.nivomoney.com | Verification portal | Pending DNS |
| grafana.nivomoney.com | Monitoring | Pending DNS |

### SSL Certificate Coverage (Let's Encrypt)
```
Certificate: /etc/letsencrypt/live/nivomoney.com/fullchain.pem
Domains: nivomoney.com, www.nivomoney.com, admin.nivomoney.com,
         api.nivomoney.com, app.nivomoney.com
Auto-renewal: certbot.timer (runs twice daily)
```

To add new domains to certificate:
```bash
# Stop nginx, expand cert, restart
docker stop nivo-frontend
sudo certbot certonly --standalone --expand \
  -d nivomoney.com -d www.nivomoney.com \
  -d admin.nivomoney.com -d api.nivomoney.com \
  -d app.nivomoney.com -d NEW_DOMAIN.nivomoney.com
docker start nivo-frontend
```

## Quick Deploy Commands

### Frontend Only
```bash
# SSH to server
ssh nivo@157.245.96.200

# Pull latest code
cd /opt/nivo && git pull origin main

# Build frontend Docker image
cd frontend && docker build --target production -t nivo-frontend:latest .

# Deploy frontend (without backend)
docker stop nivo-frontend 2>/dev/null; docker rm nivo-frontend 2>/dev/null
docker run -d \
  --name nivo-frontend \
  -p 80:80 -p 443:443 \
  -v /opt/nivo/frontend/nginx-frontend-only.conf:/etc/nginx/nginx.conf:ro \
  -v /etc/letsencrypt:/etc/letsencrypt:ro \
  -v /var/www/certbot:/var/www/certbot:ro \
  --restart unless-stopped \
  nivo-frontend:latest

# Verify
docker ps --filter name=nivo-frontend
curl -I https://admin.nivomoney.com
```

### Full Stack (with backend)
```bash
# Use docker-compose for full stack
cd /opt/nivo
docker-compose -f docker-compose.prod.yml up -d

# Or use nginx-prod.conf which includes gateway upstream
docker run -d \
  --name nivo-frontend \
  --network nivo-network \
  -p 80:80 -p 443:443 \
  -v /opt/nivo/frontend/nginx-prod.conf:/etc/nginx/nginx.conf:ro \
  -v /etc/letsencrypt:/etc/letsencrypt:ro \
  -v /var/www/certbot:/var/www/certbot:ro \
  --restart unless-stopped \
  nivo-frontend:latest
```

## SSL Certificates

Certificates managed by Let's Encrypt (certbot).

### Renewal (automatic via cron)
```bash
# Manual renewal test
sudo certbot renew --dry-run

# Force renewal
sudo certbot renew --force-renewal
```

### Add new domains to certificate
```bash
# Stop nginx first
docker stop nivo-frontend

# Add domain (requires DNS to be configured first)
sudo certbot certonly --standalone -d newdomain.nivomoney.com

# Or expand existing certificate
sudo certbot certonly --standalone --expand \
  -d nivomoney.com -d www.nivomoney.com \
  -d admin.nivomoney.com -d api.nivomoney.com \
  -d newdomain.nivomoney.com

# Restart nginx
docker start nivo-frontend
```

## Nginx Configurations

| File | Purpose |
|------|---------|
| nginx-frontend-only.conf | Frontend apps only (no backend) |
| nginx-prod.conf | Full production with API gateway |
| nginx.conf | Original full config with all domains |
| nginx-local.conf | Local development |

## Docker Images

| Image | Purpose |
|-------|---------|
| nivo-frontend:latest | Nginx serving all frontend apps |
| nivo-gateway:latest | API Gateway (Go) |
| nivo-identity:latest | Identity service |
| ... | Other microservices |

## Troubleshooting

### Frontend container not starting
```bash
# Check logs
docker logs nivo-frontend

# Common issue: gateway upstream not available
# Solution: Use nginx-frontend-only.conf instead of nginx-prod.conf
```

### SSL certificate issues
```bash
# Check certificate
sudo certbot certificates

# Check expiration
openssl s_client -connect nivomoney.com:443 -servername nivomoney.com </dev/null 2>/dev/null | openssl x509 -noout -dates
```

### DNS not resolving
```bash
# Check from server
dig +short admin.nivomoney.com

# Check A record points to 157.245.96.200
```

## Security Hardening Applied

- SSH key-only authentication
- UFW firewall (ports 22, 80, 443 only)
- Fail2ban for brute-force protection
- Docker containers run as non-root
- Security headers (HSTS, CSP, X-Frame-Options, etc.)
- Rate limiting on API endpoints
- TLS 1.2/1.3 only

## Monitoring

### Health checks
```bash
# Nginx health
curl http://localhost/nginx-health

# Container health
docker inspect --format='{{.State.Health.Status}}' nivo-frontend
```

### Logs
```bash
# Nginx access logs
docker exec nivo-frontend tail -f /var/log/nginx/access.log

# Nginx error logs
docker exec nivo-frontend tail -f /var/log/nginx/error.log
```

## Docker Network Security

### External Exposure
Only the frontend container exposes ports externally:
- **Port 80** → HTTP (redirects to HTTPS)
- **Port 443** → HTTPS

All backend services communicate via internal Docker network only:
- postgres (5432) - Internal only
- redis (6379) - Internal only
- identity-service (8080) - Internal only
- ledger-service (8081) - Internal only
- rbac-service (8082) - Internal only
- wallet-service (8083) - Internal only
- transaction-service (8084) - Internal only
- risk-service (8085) - Internal only
- simulation-service (8086) - Internal only
- notification-service (8087) - Internal only
- gateway (8000) - Internal only (nginx proxies to it)

### Verification
```bash
# Verify only frontend exposes ports
docker ps --format "{{.Names}}: {{.Ports}}" | grep -v "0.0.0.0"

# Confirm no backend services are accessible externally
curl http://157.245.96.200:8080/health  # Should fail (connection refused)
curl http://157.245.96.200:5432         # Should fail (connection refused)

# Only these should work
curl http://157.245.96.200              # Redirect to HTTPS
curl -I https://nivomoney.com           # 200 OK
```

### Container Security Hardening
All containers implement:
- `cap_drop: ALL` - Remove all Linux capabilities
- `no-new-privileges: true` - Prevent privilege escalation
- Resource limits (memory, pids) - Prevent resource exhaustion

---

## Security Configuration Details

### SSH Hardening (/etc/ssh/sshd_config)
- PermitRootLogin: prohibit-password (key-only)
- PasswordAuthentication: no
- MaxAuthTries: 3
- ClientAliveInterval: 300
- ClientAliveCountMax: 2

### Firewall (UFW)
```
Port 22/tcp  - SSH
Port 80/tcp  - HTTP (redirects to HTTPS)
Port 443/tcp - HTTPS
```

### Fail2ban
- SSH jail active
- Auto-bans after failed attempts

### Kernel Security (/etc/sysctl.d/99-security.conf)
- IP spoofing protection
- ICMP redirect disabled
- Source routing disabled
- SYN flood protection
- Martian packet logging

### Automatic Updates
- unattended-upgrades enabled for security patches

### Docker Container Security
- Non-root nginx user
- Read-only config mounts
- Restart policy: unless-stopped
- Health checks enabled

### Web Security Headers
All HTTPS responses include:
- Strict-Transport-Security (HSTS)
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Referrer-Policy
- Content-Security-Policy (user app)

### Rate Limiting
- Auth endpoints: 5 requests/minute
- API endpoints: 10 requests/second
- General: 30 requests/second
