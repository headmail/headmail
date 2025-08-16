# Multi-stage Dockerfile to build and serve the frontend.
# Builds the frontend app (pnpm) and serves static files with nginx.
# Place at repo root as "frontend.Dockerfile".
# Build example:
#   docker build -f frontend.Dockerfile -t ghcr.io/headmail/headmail-frontend:0.0.1 .
# Then push to registry and set values.frontend.image.repository/tag in the Helm chart.

FROM node:20-alpine AS builder

ENV NODE_ENV=production
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git python3 make g++ && \
    npm install -g pnpm

# Copy package & lockfiles first for better caching
COPY ./package.json ./pnpm-lock.yaml ./pnpm-workspace.yaml ./
# If the project uses a workspace, the workspace file is copied too.
RUN pnpm install --frozen-lockfile

# Copy the rest of the frontend source and build
COPY ./ .
RUN pnpm build

# Serve stage
FROM nginx:stable-alpine AS runner

# Remove default nginx content
RUN rm -rf /usr/share/nginx/html/*

# Copy built static files from builder
COPY --from=builder /app/dist/ /usr/share/nginx/html/

# Copy custom nginx config to enable SPA fallback and listen on port 3000
# This file should implement try_files $uri $uri/ /index.html;
COPY ./docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

STOPSIGNAL SIGTERM

CMD ["nginx", "-g", "daemon off;"]
