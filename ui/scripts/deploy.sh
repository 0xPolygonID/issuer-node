#!/bin/sh
ENV_FILENAME="/app/.env"

# Create .env file
touch $ENV_FILENAME

# API env vars
echo "VITE_API=$API" >> $ENV_FILENAME

# Build app
cd /app && npm run build

# Copy nginx config
cp /app/deployment/nginx.conf /etc/nginx/conf.d/default.conf

# Copy app dist
cp -r /app/dist/. /usr/share/nginx/html

# Delete source code
rm -rf /app/*

# Run nginx
nginx -g 'daemon off;'
