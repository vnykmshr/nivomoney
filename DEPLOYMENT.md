# Nivo Production Deployment Guide

## Security Model

**Standard Operating Procedure (SOP):** Use `deploy` user for all operations.
**Emergency/Maintenance:** Root access preserved (key-only).

| Operation | User | Notes |
|-----------|------|-------|
| Server provisioning | `root` | Initial setup |
| SSH login (SOP) | `deploy` | Standard operations |
| SSH login (emergency) | `root` | Key-only, preserved for emergencies |
| Application deployment | `deploy` | Docker via group membership |
| Docker commands | `deploy` | No sudo required |
| SSL certificates | `deploy` | Certbot in container |
| Database backups | `deploy` | Via docker compose |
| Viewing logs | `deploy` | Via docker compose |

## Prerequisites

1. Fresh Ubuntu 22.04+ or Debian 12+ server
2. SSH key configured for root access (temporarily)
3. Domain DNS configured (see below)

## DNS Configuration

Configure these DNS records BEFORE deployment:

| Record | Type | Value |
|--------|------|-------|
| nivomoney.com | A | `<server-ip>` |
| www.nivomoney.com | CNAME | nivomoney.com |
| admin.nivomoney.com | A | `<server-ip>` |
| api.nivomoney.com | A | `<server-ip>` |
| grafana.nivomoney.com | A | `<server-ip>` |
| docs.nivomoney.com | CNAME | vnykmshr.github.io |

## Step 1: Server Setup (Root - ONE TIME ONLY)

SSH as root (last time you'll use root):

```bash
ssh root@your-server-ip
```

Run the setup script:

```bash
curl -fsSL https://raw.githubusercontent.com/vnykmshr/nivo/main/scripts/setup-server.sh | bash
```

**What this does:**
- Creates `deploy` user with passwordless sudo
- Copies your SSH keys to deploy user
- Preserves root SSH access (key-only, for emergencies)
- Installs Docker, SOPS, age
- Configures UFW firewall (22, 80, 443 only)
- Configures fail2ban (24h ban after 3 failures)
- Enables automatic security updates

After setup, verify deploy user access:

```bash
ssh deploy@your-server-ip
```

Use `deploy` user for all standard operations (SOP).

## Step 2: Clone Repository (Deploy User)

```bash
ssh deploy@your-server-ip

git clone https://github.com/vnykmshr/nivo.git /opt/nivo
cd /opt/nivo
```

## Step 3: Generate Age Key (Deploy User)

```bash
# Generate new key pair
age-keygen -o ~/.config/sops/age/keys.txt

# View public key (needed for .sops.yaml)
cat ~/.config/sops/age/keys.txt
# Output: # public key: age1abc123...
#         AGE-SECRET-KEY-1...
```

**Save both keys securely!**
- Public key: Update `.sops.yaml` in the repo
- Private key: Keep in `~/.config/sops/age/keys.txt` on server

## Step 4: Create Encrypted Secrets (Local Machine)

On your local development machine:

```bash
# Update .sops.yaml with your public key
vim .sops.yaml  # Replace the age public key

# Create production env file
cp .env.template .env.prod

# Edit with real production values
vim .env.prod

# Encrypt
sops --encrypt --input-type dotenv --output-type dotenv --output .env.enc .env.prod

# Delete unencrypted file!
rm .env.prod

# Commit the encrypted file
git add .env.enc .sops.yaml
git commit -m "Add encrypted production secrets"
git push
```

## Step 5: Initial Deployment (Deploy User)

On your local machine:

```bash
# Create local deployment config
cat > .env.local << EOF
DEPLOY_HOST=your-server-ip
DEPLOY_USER=deploy
DEPLOY_DIR=/opt/nivo
EOF

# Deploy
make deploy
```

## Step 6: SSL Certificates (Deploy User)

After first deployment, obtain SSL certificates:

```bash
ssh deploy@your-server-ip
cd /opt/nivo

# Create certbot directories
mkdir -p certbot/conf certbot/www

# Get certificates (first time)
docker compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    -d nivomoney.com \
    -d www.nivomoney.com \
    -d admin.nivomoney.com \
    -d api.nivomoney.com \
    -d grafana.nivomoney.com \
    --email admin@nivomoney.com \
    --agree-tos \
    --no-eff-email

# Restart to pick up certificates
docker compose restart frontend
```

## Ongoing Operations

All operations run as `deploy` user:

### Deploy Updates
```bash
make deploy
```

### View Logs
```bash
ssh deploy@your-server-ip
cd /opt/nivo
docker compose logs -f
docker compose logs -f gateway  # Specific service
```

### Database Backup
```bash
ssh deploy@your-server-ip
cd /opt/nivo
make db-backup
```

### Restart Services
```bash
ssh deploy@your-server-ip
cd /opt/nivo
docker compose restart
```

### SSL Certificate Renewal
```bash
ssh deploy@your-server-ip
cd /opt/nivo
docker compose run --rm certbot renew
docker compose restart frontend
```

### Update Secrets
```bash
# On local machine
sops .env.enc  # Edit encrypted file directly
git add .env.enc
git commit -m "Update secrets"
make deploy
```

## Security Hardening Summary

### Network
- UFW firewall: Only ports 22, 80, 443
- All backend services on internal Docker network
- Only nginx exposed to internet

### SSH
- Root login: Key-only (preserved for emergencies)
- Deploy user: Key-only (for SOP)
- Strong ciphers only
- fail2ban: 24h ban after 3 failed attempts

### Docker
- All containers: `cap_drop: ALL`
- All containers: `no-new-privileges: true`
- Resource limits on all containers
- Non-root users in containers where possible
- Read-only filesystems where possible

### Secrets
- SOPS + age encryption
- Encrypted secrets committed to git
- Decrypted only at deploy time
- No plaintext secrets in repo

### Updates
- Automatic security updates enabled
- Log rotation configured
- Docker image cleanup on deploy

## Troubleshooting

### Can't SSH as deploy user
```bash
# If you still have root access somehow
sudo cat /var/log/auth.log | tail -50
```

### Services not starting
```bash
ssh deploy@your-server-ip
cd /opt/nivo
docker compose logs --tail=100
docker compose ps
```

### SSL certificate issues
```bash
# Check certificate status
docker compose run --rm certbot certificates

# Force renewal
docker compose run --rm certbot renew --force-renewal
docker compose restart frontend
```

### Database issues
```bash
# Connect to database
docker compose exec postgres psql -U nivo nivo

# Check database logs
docker compose logs postgres
```
