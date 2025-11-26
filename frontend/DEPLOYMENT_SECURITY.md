# Frontend Deployment Security Guide

This guide covers security configuration for deploying Nivo Money frontend applications to production.

## Table of Contents

1. [Nginx Configuration](#nginx-configuration)
2. [IP Whitelisting (Admin App)](#ip-whitelisting-admin-app)
3. [SSL/TLS Configuration](#ssltls-configuration)
4. [Content Security Policy](#content-security-policy)
5. [Production Checklist](#production-checklist)

---

## Nginx Configuration

### User App Configuration (`app.nivomoney.com`)

```nginx
server {
    listen 443 ssl http2;
    server_name app.nivomoney.com;

    # SSL Configuration
    ssl_certificate /etc/ssl/certs/nivomoney.com.crt;
    ssl_certificate_key /etc/ssl/private/nivomoney.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Root directory
    root /var/www/nivo/user-app/dist;
    index index.html;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;

    # Content Security Policy (adjust domains as needed)
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https://api.nivomoney.com" always;

    # Serve static files
    location / {
        try_files $uri $uri/ /index.html;

        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # Disable access to sensitive files
    location ~ /\. {
        deny all;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name app.nivomoney.com;
    return 301 https://$server_name$request_uri;
}
```

### Admin App Configuration (`admin.nivomoney.com`)

**IMPORTANT**: Admin app requires stricter security with IP whitelisting.

```nginx
server {
    listen 443 ssl http2;
    server_name admin.nivomoney.com;

    # SSL Configuration
    ssl_certificate /etc/ssl/certs/nivomoney.com.crt;
    ssl_certificate_key /etc/ssl/private/nivomoney.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Root directory
    root /var/www/nivo/admin-app/dist;
    index index.html;

    # IP Whitelisting - ONLY ALLOW TRUSTED IPS
    # Add your office/VPN IPs here
    allow 203.0.113.0/24;     # Example: Office network
    allow 198.51.100.50;      # Example: VPN gateway IP
    allow 192.0.2.100;        # Example: Admin home IP
    deny all;                 # Deny everyone else

    # Security Headers (stricter than user app)
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), payment=()" always;

    # Content Security Policy (stricter)
    add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' https://api.nivomoney.com; frame-ancestors 'none'" always;

    # Remove server version info
    server_tokens off;

    # Serve static files
    location / {
        try_files $uri $uri/ /index.html;

        # Cache static assets (shorter cache for admin updates)
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 7d;
            add_header Cache-Control "public";
        }
    }

    # Disable access to sensitive files
    location ~ /\. {
        deny all;
    }

    # Rate limiting (optional but recommended)
    limit_req_zone $binary_remote_addr zone=admin:10m rate=10r/s;
    limit_req zone=admin burst=20;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name admin.nivomoney.com;
    return 301 https://$server_name$request_uri;
}
```

---

## IP Whitelisting (Admin App)

### Why IP Whitelisting?

The admin app handles sensitive operations (KYC approval, user management, etc.) and should only be accessible from trusted networks.

### Configuration Steps

1. **Identify Trusted IPs**:
   - Office network IP range
   - VPN gateway IP addresses
   - Admin home/remote IPs (if applicable)

2. **Update Nginx Config**:
   ```nginx
   # In the admin.nivomoney.com server block
   allow 203.0.113.0/24;     # Office network
   allow 198.51.100.50;      # VPN gateway
   allow 192.0.2.100;        # Specific admin IP
   deny all;
   ```

3. **Testing**:
   ```bash
   # Test from allowed IP
   curl -I https://admin.nivomoney.com
   # Should return 200 OK

   # Test from blocked IP (use VPN or different network)
   curl -I https://admin.nivomoney.com
   # Should return 403 Forbidden
   ```

4. **Monitoring**:
   - Set up alerts for 403 errors on admin app
   - Review access logs regularly
   - Update IP whitelist when admins change networks

### Alternative: VPN-Only Access

For maximum security, consider requiring VPN connection:

```nginx
# Only allow VPN subnet
allow 10.0.0.0/8;          # VPN subnet
deny all;
```

This ensures all admin access goes through your secure VPN tunnel.

---

## SSL/TLS Configuration

### Certificate Setup

1. **Obtain SSL Certificate**:
   ```bash
   # Using Let's Encrypt (recommended)
   certbot certonly --nginx -d app.nivomoney.com -d admin.nivomoney.com
   ```

2. **Auto-Renewal**:
   ```bash
   # Add to crontab
   0 0 1 * * certbot renew --quiet && systemctl reload nginx
   ```

3. **Test SSL Configuration**:
   - Use [SSL Labs](https://www.ssllabs.com/ssltest/) to verify A+ rating
   - Ensure TLS 1.2+ only
   - Check certificate chain

---

## Content Security Policy

### Understanding CSP

Content Security Policy prevents XSS attacks by controlling which resources can be loaded.

### User App CSP (Less Restrictive)

```
Content-Security-Policy: default-src 'self';
  script-src 'self' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: https:;
  font-src 'self' data:;
  connect-src 'self' https://api.nivomoney.com
```

**Note**: `'unsafe-inline'` is needed for Vite's build output. Consider removing in future with nonce-based CSP.

### Admin App CSP (Stricter)

```
Content-Security-Policy: default-src 'self';
  script-src 'self';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data:;
  font-src 'self' data:;
  connect-src 'self' https://api.nivomoney.com;
  frame-ancestors 'none'
```

### CSP Reporting

Add CSP reporting to monitor violations:

```nginx
add_header Content-Security-Policy "default-src 'self'; ...; report-uri https://api.nivomoney.com/csp-report" always;
```

---

## Production Checklist

### Pre-Deployment

- [ ] SSL certificates installed and valid
- [ ] DNS records configured (A/CNAME)
- [ ] Firewall rules configured (allow 80, 443)
- [ ] Nginx installed and configured
- [ ] IP whitelist configured for admin app
- [ ] Environment variables set correctly

### Security Verification

- [ ] HTTPS enforced (HTTP redirects to HTTPS)
- [ ] Security headers present (check with browser dev tools)
- [ ] CSP configured and not blocking resources
- [ ] IP whitelist working (test from blocked IP)
- [ ] Session timeout working (2h for admin, 24h for user)
- [ ] Audit logging enabled and working
- [ ] CSRF protection enabled (TODO: implement)

### Testing

- [ ] User app accessible from public internet
- [ ] Admin app blocked from non-whitelisted IPs
- [ ] Admin app accessible from whitelisted IPs
- [ ] Login/logout flows working
- [ ] KYC submission and approval working
- [ ] All API endpoints responding correctly
- [ ] Error handling working (network errors, auth errors)

### Monitoring

- [ ] Set up uptime monitoring (UptimeRobot, Pingdom, etc.)
- [ ] Configure log aggregation (ELK, Datadog, etc.)
- [ ] Set up alerts for:
  - HTTP 5xx errors
  - HTTP 403 errors on admin app (potential attacks)
  - SSL certificate expiration
  - High response times
- [ ] Review audit logs regularly

### Ongoing Maintenance

- [ ] Review access logs weekly
- [ ] Update IP whitelist as needed
- [ ] Rotate SSL certificates before expiry
- [ ] Update dependencies regularly
- [ ] Review and update CSP as app evolves
- [ ] Perform security audits quarterly

---

## Additional Security Measures

### 1. Rate Limiting

Prevent brute force attacks:

```nginx
# In http block
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;

# In location block for login
location /api/v1/auth/login {
    limit_req zone=login burst=2;
    proxy_pass http://backend;
}
```

### 2. DDoS Protection

Consider using Cloudflare or AWS Shield for DDoS protection.

### 3. Web Application Firewall (WAF)

Use ModSecurity or cloud-based WAF:
- Cloudflare WAF
- AWS WAF
- Imperva

### 4. Regular Security Scanning

- Use OWASP ZAP for penetration testing
- Run Lighthouse security audits
- Use Snyk or Dependabot for dependency scanning

### 5. Backup Strategy

- Daily backups of configurations
- Version control all config files
- Document all manual steps
- Test restore procedures

---

## Emergency Procedures

### Security Breach

1. **Immediately**: Block access by removing from IP whitelist
2. Revoke all active sessions
3. Force password reset for affected users
4. Review audit logs for suspicious activity
5. Notify affected users if data compromised

### SSL Certificate Issues

1. Have backup certificates ready
2. Know manual renewal process
3. Monitor expiration dates (60 days before)

### IP Whitelist Update

```bash
# Edit nginx config
sudo nano /etc/nginx/sites-available/admin.nivomoney.com

# Add/remove IPs
allow 198.51.100.51;  # New IP

# Test config
sudo nginx -t

# Reload
sudo systemctl reload nginx
```

---

## Support and Updates

For questions or updates to this guide:
- Check shared package security configuration: `frontend/shared/src/lib/security.ts`
- Review application security stores:
  - User app: `frontend/user-app/src/stores/authStore.ts`
  - Admin app: `frontend/admin-app/src/stores/adminAuthStore.ts`
- Audit logging: `frontend/admin-app/src/lib/auditLogger.ts`

**Last Updated**: Phase 7 - Security Hardening
