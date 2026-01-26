# =============================================================================
# Nivo Frontend - Multi-stage Docker Build
# =============================================================================
#
# IMPORTANT: Uses Vite aliases for @nivo/shared imports, NOT npm packages.
# This ensures single React instance (no hook context mismatch).
#

# =============================================================================
# Stage 1: Prepare shared source (no bundling, source-only)
# =============================================================================
FROM node:20-alpine AS shared-builder

WORKDIR /app/frontend

# Copy shared package source
COPY shared/ ./shared/

# Install shared dependencies only (needed for Vite to bundle correctly)
WORKDIR /app/frontend/shared
RUN npm install --silent --ignore-scripts

# =============================================================================
# Stage 2: Build user-app
# =============================================================================
FROM node:20-alpine AS user-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared source with node_modules
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy user-app package files
COPY user-app/package*.json ./user-app/

# Remove @nivo/shared from package.json (Vite aliases will handle it)
# Install remaining deps
WORKDIR /app/frontend/user-app
RUN sed -i '/"@nivo\/shared":/d' package.json && \
    npm install --silent

# Copy source and build
COPY user-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 3: Build admin-app
# =============================================================================
FROM node:20-alpine AS admin-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared source with node_modules
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy admin-app package files
COPY admin-app/package*.json ./admin-app/

# Remove @nivo/shared from package.json (Vite aliases will handle it)
# Install remaining deps
WORKDIR /app/frontend/admin-app
RUN sed -i '/"@nivo\/shared":/d' package.json && \
    npm install --silent

# Copy source and build
COPY admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 4: Build user-admin-app
# =============================================================================
FROM node:20-alpine AS user-admin-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared source with node_modules
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy user-admin-app package files
COPY user-admin-app/package*.json ./user-admin-app/

# Remove @nivo/shared from package.json (Vite aliases will handle it)
# Install remaining deps
WORKDIR /app/frontend/user-admin-app
RUN sed -i '/"@nivo\/shared":/d' package.json && \
    npm install --silent

# Copy source and build
COPY user-admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 5: Production nginx image
# =============================================================================
FROM nginx:1.25-alpine AS production

RUN rm /etc/nginx/conf.d/default.conf

RUN mkdir -p /usr/share/nginx/html/user-app \
    && mkdir -p /usr/share/nginx/html/admin-app \
    && mkdir -p /usr/share/nginx/html/user-admin-app \
    && mkdir -p /var/cache/nginx/client_temp \
    && mkdir -p /var/cache/nginx/proxy_temp \
    && mkdir -p /var/cache/nginx/fastcgi_temp \
    && mkdir -p /var/cache/nginx/uwsgi_temp \
    && mkdir -p /var/cache/nginx/scgi_temp \
    && mkdir -p /var/www/certbot

COPY --from=user-app-builder /app/frontend/user-app/dist /usr/share/nginx/html/user-app
COPY --from=admin-app-builder /app/frontend/admin-app/dist /usr/share/nginx/html/admin-app
COPY --from=user-admin-app-builder /app/frontend/user-admin-app/dist /usr/share/nginx/html/user-admin-app

RUN chown -R nginx:nginx /usr/share/nginx/html \
    && chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /var/www/certbot \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O - http://localhost/nginx-health || exit 1

CMD ["nginx", "-g", "daemon off;"]
