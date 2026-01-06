#!/bin/bash
# =============================================================================
# Nivo - Server Setup Script
# =============================================================================
#
# First-time setup for a fresh Ubuntu 22.04+ / Debian 12+ server.
#
# Usage:
#   # After cloning:
#   sudo ./scripts/setup-server.sh
#
#   # Or directly from GitHub:
#   curl -fsSL https://raw.githubusercontent.com/vnykmshr/nivo/main/scripts/setup-server.sh | sudo bash
#
# What this script does:
#   1. Installs Docker and Docker Compose
#   2. Installs SOPS and age for secrets management
#   3. Configures fail2ban for SSH protection
#   4. Enables unattended security upgrades
#   5. Configures Docker log rotation
#   6. Creates application directories
#
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (sudo)"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH_NAME="amd64" ;;
    aarch64) ARCH_NAME="arm64" ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

log_info "=========================================="
log_info "  Nivo Server Setup"
log_info "=========================================="
log_info "Architecture: $ARCH_NAME"
echo ""

# =============================================================================
# System Updates
# =============================================================================
log_step "Updating system packages..."
apt-get update
apt-get upgrade -y

# =============================================================================
# Install Prerequisites
# =============================================================================
log_step "Installing prerequisites..."
apt-get install -y \
    curl \
    wget \
    git \
    ca-certificates \
    gnupg \
    lsb-release \
    fail2ban \
    unattended-upgrades \
    apt-transport-https \
    htop \
    ncdu

# =============================================================================
# Install Docker
# =============================================================================
if ! command -v docker &> /dev/null; then
    log_step "Installing Docker..."

    # Add Docker's official GPG key
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    # Add the repository
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Enable and start Docker
    systemctl enable docker
    systemctl start docker

    log_info "Docker installed successfully"
else
    log_info "Docker already installed: $(docker --version)"
fi

# =============================================================================
# Install SOPS
# =============================================================================
SOPS_VERSION="3.8.1"
if ! command -v sops &> /dev/null; then
    log_step "Installing SOPS v${SOPS_VERSION}..."

    wget -q "https://github.com/getsops/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.linux.${ARCH_NAME}" \
        -O /usr/local/bin/sops
    chmod +x /usr/local/bin/sops

    log_info "SOPS installed successfully"
else
    log_info "SOPS already installed: $(sops --version)"
fi

# =============================================================================
# Install age
# =============================================================================
AGE_VERSION="1.1.1"
if ! command -v age &> /dev/null; then
    log_step "Installing age v${AGE_VERSION}..."

    wget -q "https://github.com/FiloSottile/age/releases/download/v${AGE_VERSION}/age-v${AGE_VERSION}-linux-${ARCH_NAME}.tar.gz" \
        -O /tmp/age.tar.gz
    tar -xzf /tmp/age.tar.gz -C /tmp
    mv /tmp/age/age /usr/local/bin/
    mv /tmp/age/age-keygen /usr/local/bin/
    rm -rf /tmp/age /tmp/age.tar.gz

    log_info "age installed successfully"
else
    log_info "age already installed: $(age --version)"
fi

# =============================================================================
# Configure fail2ban
# =============================================================================
log_step "Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 24h
findtime = 10m
maxretry = 3
ignoreip = 127.0.0.1/8 ::1

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
EOF

systemctl enable fail2ban
systemctl restart fail2ban

# =============================================================================
# Configure unattended upgrades
# =============================================================================
log_step "Configuring automatic security updates..."
cat > /etc/apt/apt.conf.d/20auto-upgrades << 'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

# =============================================================================
# Configure Docker log rotation
# =============================================================================
log_step "Configuring Docker log rotation..."
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker

# =============================================================================
# Create application directories
# =============================================================================
DEPLOY_DIR="${DEPLOY_DIR:-/opt/nivo}"
log_step "Creating application directories at ${DEPLOY_DIR}..."

mkdir -p "${DEPLOY_DIR}"
mkdir -p /var/backups/nivo
mkdir -p /var/log/nivo

# Set ownership if SUDO_USER is set
if [ -n "$SUDO_USER" ]; then
    chown -R "$SUDO_USER:$SUDO_USER" "${DEPLOY_DIR}"
    chown -R "$SUDO_USER:$SUDO_USER" /var/backups/nivo
    chown -R "$SUDO_USER:$SUDO_USER" /var/log/nivo

    # Add user to docker group
    usermod -aG docker "$SUDO_USER"
    log_info "Added $SUDO_USER to docker group (re-login required)"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
log_info "=========================================="
log_info "  Server setup complete!"
log_info "=========================================="
echo ""
echo "Installed:"
echo "  - Docker $(docker --version | cut -d' ' -f3 | tr -d ',')"
echo "  - Docker Compose $(docker compose version | cut -d' ' -f4)"
echo "  - SOPS $(sops --version 2>&1 | head -1)"
echo "  - age $(age --version)"
echo "  - fail2ban"
echo "  - unattended-upgrades"
echo ""
echo "Next steps:"
echo "  1. Clone the repository:"
echo "     git clone https://github.com/vnykmshr/nivo.git ${DEPLOY_DIR}"
echo ""
echo "  2. Set up age key for secrets decryption:"
echo "     mkdir -p ~/.config/sops/age"
echo "     # Copy your private key to ~/.config/sops/age/keys.txt"
echo ""
echo "  3. Create .env.local with deployment config:"
echo "     cd ${DEPLOY_DIR}"
echo "     cp .env.template .env.local"
echo "     # Edit with your settings"
echo ""
echo "  4. Run initial deployment:"
echo "     make deploy"
echo ""
if [ -n "$SUDO_USER" ]; then
    echo "  5. Re-login to apply docker group membership:"
    echo "     exit && ssh <server>"
fi
echo ""
log_info "=========================================="
