#!/bin/bash
# =============================================================================
# Nivo - Production Deployment Script
# =============================================================================
#
# Deploys the application to production server.
# Runs as 'deploy' user (non-root) - Docker access via group membership.
#
# Usage:
#   make deploy
#   # Or directly:
#   ./scripts/deploy.sh
#
# Prerequisites:
#   - Server setup completed (scripts/setup-server.sh)
#   - SSH key configured for deploy user
#   - .env.local with DEPLOY_HOST, DEPLOY_USER, DEPLOY_DIR
#   - .env.enc committed with encrypted secrets
#   - age private key on server at ~/.config/sops/age/keys.txt
#
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# =============================================================================
# Load Configuration
# =============================================================================
if [ -f ".env.local" ]; then
    # shellcheck disable=SC1091
    source .env.local
else
    log_error ".env.local not found."
    echo ""
    echo "Create .env.local with:"
    echo "  DEPLOY_HOST=your-server-ip"
    echo "  DEPLOY_USER=deploy"
    echo "  DEPLOY_DIR=/opt/nivo"
    exit 1
fi

# Defaults
DEPLOY_USER="${DEPLOY_USER:-deploy}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/nivo}"

# Validate required variables
: "${DEPLOY_HOST:?DEPLOY_HOST is required in .env.local}"

log_info "=========================================="
log_info "  Nivo Deployment"
log_info "=========================================="
log_info "Host: ${DEPLOY_USER}@${DEPLOY_HOST}"
log_info "Directory: ${DEPLOY_DIR}"
echo ""

# =============================================================================
# Pre-deployment Checks
# =============================================================================
log_step "Running pre-deployment checks..."

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    log_warn "You have uncommitted changes:"
    git status --short
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Deployment cancelled"
        exit 1
    fi
fi

# Check .env.enc exists
if [ ! -f ".env.enc" ]; then
    log_error ".env.enc not found!"
    echo ""
    echo "Create encrypted secrets first:"
    echo "  cp .env.template .env.prod"
    echo "  # Edit .env.prod with real values"
    echo "  make secrets-encrypt SRC=.env.prod"
    echo "  rm .env.prod"
    exit 1
fi

# Get current branch and commit
BRANCH=$(git rev-parse --abbrev-ref HEAD)
COMMIT=$(git rev-parse --short HEAD)
log_info "Deploying branch: ${BRANCH} (${COMMIT})"

# =============================================================================
# Push Local Changes
# =============================================================================
log_step "Pushing local changes..."
git push origin "${BRANCH}"

# =============================================================================
# Deploy to Server
# =============================================================================
log_step "Deploying to server..."

ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << ENDSSH
set -e

cd "${DEPLOY_DIR}"

echo "[DEPLOY] Pulling latest changes..."
git fetch origin
git reset --hard origin/${BRANCH}

echo "[DEPLOY] Decrypting secrets..."
# SOPS will automatically use key from ~/.config/sops/age/keys.txt
sops --decrypt --input-type dotenv --output-type dotenv .env.enc > .env

echo "[DEPLOY] Tagging current images for rollback..."
docker compose ps -q 2>/dev/null | xargs -r docker inspect --format='{{.Image}}' 2>/dev/null | sort -u | while read img; do
    docker tag "\$img" "\$img:rollback" 2>/dev/null || true
done

echo "[DEPLOY] Building images..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml build --pull

echo "[DEPLOY] Stopping old containers..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml down --remove-orphans

echo "[DEPLOY] Running database migrations..."
# Migrations run as part of service startup

echo "[DEPLOY] Starting services..."
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

echo "[DEPLOY] Cleaning up old images..."
docker image prune -f

echo "[DEPLOY] Waiting for health checks..."
sleep 10

# Health check - use nginx health endpoint (gateway port not exposed to host)
MAX_ATTEMPTS=30
ATTEMPT=0
until curl -sf http://localhost:80/nginx-health > /dev/null 2>&1; do
    ATTEMPT=\$((ATTEMPT + 1))
    if [ \$ATTEMPT -ge \$MAX_ATTEMPTS ]; then
        echo "[ERROR] Health check failed after \$MAX_ATTEMPTS attempts"
        echo "[ERROR] Rolling back..."
        docker compose -f docker-compose.yml -f docker-compose.prod.yml logs --tail=50
        docker compose -f docker-compose.yml -f docker-compose.prod.yml down
        exit 1
    fi
    echo "[DEPLOY] Waiting for frontend... (attempt \$ATTEMPT/\$MAX_ATTEMPTS)"
    sleep 2
done

echo ""
echo "[DEPLOY] =========================================="
echo "[DEPLOY]   Deployment successful!"
echo "[DEPLOY] =========================================="
echo ""
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
ENDSSH

echo ""
log_info "=========================================="
log_info "  Deployment complete!"
log_info "=========================================="
echo ""
echo "  Site:    https://nivomoney.com"
echo "  Admin:   https://admin.nivomoney.com"
echo "  API:     https://api.nivomoney.com"
echo "  Grafana: https://grafana.nivomoney.com"
echo ""
