#!/bin/bash
# =============================================================================
# Nivo - Production Deployment Script
# =============================================================================
#
# Deploys the application to production server.
#
# Usage:
#   make deploy
#   # Or directly:
#   ./scripts/deploy.sh
#
# Prerequisites:
#   - .env.local with DEPLOY_HOST, DEPLOY_DIR, DEPLOY_AGE_KEY
#   - SSH key configured for DEPLOY_HOST
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
    echo "  DEPLOY_DIR=/opt/nivo"
    echo "  DEPLOY_AGE_KEY=AGE-SECRET-KEY-xxx"
    exit 1
fi

# Validate required variables
: "${DEPLOY_HOST:?DEPLOY_HOST is required in .env.local}"
: "${DEPLOY_DIR:?DEPLOY_DIR is required in .env.local}"
: "${DEPLOY_AGE_KEY:?DEPLOY_AGE_KEY is required in .env.local}"

log_info "=========================================="
log_info "  Nivo Deployment"
log_info "=========================================="
log_info "Host: ${DEPLOY_HOST}"
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

ssh "${DEPLOY_HOST}" << ENDSSH
set -e

cd "${DEPLOY_DIR}"

echo "[DEPLOY] Pulling latest changes..."
git fetch origin
git reset --hard origin/${BRANCH}

echo "[DEPLOY] Decrypting secrets..."
export SOPS_AGE_KEY="${DEPLOY_AGE_KEY}"
sops --decrypt --input-type dotenv --output-type dotenv .env.enc > .env

echo "[DEPLOY] Tagging current images for rollback..."
docker compose ps -q 2>/dev/null | xargs -r docker inspect --format='{{.Image}}' 2>/dev/null | sort -u | while read img; do
    docker tag "\$img" "\$img:rollback" 2>/dev/null || true
done

echo "[DEPLOY] Building images..."
docker compose build --pull

echo "[DEPLOY] Stopping old containers..."
docker compose down --remove-orphans

echo "[DEPLOY] Starting services..."
docker compose up -d

echo "[DEPLOY] Cleaning up old images..."
docker image prune -f

echo "[DEPLOY] Waiting for health checks..."
sleep 10

# Health check
MAX_ATTEMPTS=30
ATTEMPT=0
until curl -sf http://localhost:8000/health > /dev/null 2>&1; do
    ATTEMPT=\$((ATTEMPT + 1))
    if [ \$ATTEMPT -ge \$MAX_ATTEMPTS ]; then
        echo "[ERROR] Health check failed after \$MAX_ATTEMPTS attempts"
        echo "[DEPLOY] Rolling back..."
        docker compose down
        # Could add rollback logic here
        exit 1
    fi
    echo "[DEPLOY] Waiting for gateway... (attempt \$ATTEMPT/\$MAX_ATTEMPTS)"
    sleep 2
done

echo ""
echo "[DEPLOY] =========================================="
echo "[DEPLOY]   Deployment successful!"
echo "[DEPLOY] =========================================="
echo ""
docker compose ps
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
