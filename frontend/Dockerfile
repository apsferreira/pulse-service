# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

ARG BUILD_SHA=dev
ENV BUILD_SHA=$BUILD_SHA

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Production stage — serve com nginx
# Compatível com readOnlyRootFilesystem: k8s deve montar emptyDir em:
#   /var/cache/nginx, /var/run, /tmp
FROM nginx:alpine

# Remove default config
RUN rm /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Patch nginx.conf principal para não criar pid em /var/run
RUN sed -i 's|pid\s*/var/run/nginx.pid;|pid /tmp/nginx.pid;|g' /etc/nginx/nginx.conf \
 && sed -i '/^user /d' /etc/nginx/nginx.conf

# Cria diretórios writáveis e seta permissões para nginx user (uid 101)
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp \
 && chown -R nginx:nginx /var/cache/nginx /usr/share/nginx/html \
 && chmod -R 755 /var/cache/nginx

EXPOSE 80

USER nginx

CMD ["nginx", "-g", "daemon off;"]
