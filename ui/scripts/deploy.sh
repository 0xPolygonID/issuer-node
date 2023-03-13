#!/bin/sh

ENV_FILENAME=".env"

cd /app

# Create .env file
touch $ENV_FILENAME

# API env vars
echo "VITE_API=$API" >> $ENV_FILENAME

# Build app
npm run build

# Copy nginx config
cp deployment/nginx.conf /etc/nginx/conf.d/default.conf
echo $ISSUER_UI_AUTH_USERNAME
echo $ISSUER_UI_AUTH_PASSWORD
htpasswd -c -b /etc/nginx/.htpasswd $ISSUER_UI_AUTH_USERNAME $ISSUER_UI_AUTH_PASSWORD
cat /etc/nginx/.htpasswd
# Copy app dist
cp -r dist/. /usr/share/nginx/html

# Delete source code
rm -rf /app/*

# Run nginx
nginx -g 'daemon off;'
