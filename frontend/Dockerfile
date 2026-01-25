# =============================================================================
# Nivo Frontend - Multi-stage Docker Build
# =============================================================================
#
# Builds user-app, admin-app, and user-admin-app, serves via nginx.
#
# Build args:
#   VITE_API_URL - API URL for frontend (default: https://api.nivomoney.com)
#
# =============================================================================

# =============================================================================
# Stage 1: Build shared package (dependency for all apps)
# =============================================================================
FROM node:20-alpine AS shared-builder

WORKDIR /app

# Copy shared package
COPY shared/ ./shared/

# Build shared package
WORKDIR /app/shared
RUN npm install --silent && npm run build

# =============================================================================
# Stage 2: Build user-app
# =============================================================================
FROM node:20-alpine AS user-app-builder

WORKDIR /app

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared build artifacts
COPY --from=shared-builder /app/shared ./shared

# Copy package files
COPY user-app/package*.json ./user-app/

# Update @nivo/shared to use local file path
WORKDIR /app/user-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json

# Install dependencies
RUN npm install --silent

# Copy source code and build
COPY user-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 3: Build admin-app
# =============================================================================
FROM node:20-alpine AS admin-app-builder

WORKDIR /app

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared build artifacts
COPY --from=shared-builder /app/shared ./shared

# Copy package files
COPY admin-app/package*.json ./admin-app/

# Update @nivo/shared to use local file path
WORKDIR /app/admin-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json

# Install dependencies
RUN npm install --silent

# Copy source code and build
COPY admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 4: Build user-admin-app (verification portal)
# =============================================================================
FROM node:20-alpine AS user-admin-app-builder

WORKDIR /app

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared build artifacts
COPY --from=shared-builder /app/shared ./shared

# Copy package files
COPY user-admin-app/package*.json ./user-admin-app/

# Update @nivo/shared to use local file path
WORKDIR /app/user-admin-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json

# Install dependencies
RUN npm install --silent

# Copy source code and build
COPY user-admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 5: Production nginx image
# =============================================================================
FROM nginx:1.25-alpine AS production

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Create directories for all apps and nginx cache
RUN mkdir -p /usr/share/nginx/html/user-app \
    && mkdir -p /usr/share/nginx/html/admin-app \
    && mkdir -p /usr/share/nginx/html/user-admin-app \
    && mkdir -p /var/cache/nginx/client_temp \
    && mkdir -p /var/cache/nginx/proxy_temp \
    && mkdir -p /var/cache/nginx/fastcgi_temp \
    && mkdir -p /var/cache/nginx/uwsgi_temp \
    && mkdir -p /var/cache/nginx/scgi_temp \
    && mkdir -p /var/www/certbot

# Copy built apps from builder stages
COPY --from=user-app-builder /app/user-app/dist /usr/share/nginx/html/user-app
COPY --from=admin-app-builder /app/admin-app/dist /usr/share/nginx/html/admin-app
COPY --from=user-admin-app-builder /app/user-admin-app/dist /usr/share/nginx/html/user-admin-app

# Set permissions for nginx user
RUN chown -R nginx:nginx /usr/share/nginx/html \
    && chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /var/www/certbot \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

# Expose ports
EXPOSE 80 443

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O - http://localhost/nginx-health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
