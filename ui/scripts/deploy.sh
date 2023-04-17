#!/bin/sh
ENV_FILENAME="/app/.env"

# Create .env file
touch $ENV_FILENAME

# Mapping of issuer-level env variables to app-level
echo "VITE_API_URL=$ISSUER_API_UI_SERVER_URL" >> $ENV_FILENAME
echo "VITE_API_PASSWORD=$ISSUER_API_UI_AUTH_PASSWORD" >> $ENV_FILENAME
echo "VITE_API_USERNAME=$ISSUER_API_UI_AUTH_USER" >> $ENV_FILENAME
echo "VITE_ISSUER_DID=$ISSUER_API_UI_ISSUER_DID" >> $ENV_FILENAME
echo "VITE_ISSUER_LOGO=$ISSUER_API_UI_ISSUER_LOGO" >> $ENV_FILENAME
echo "VITE_ISSUER_NAME=$ISSUER_API_UI_ISSUER_NAME" >> $ENV_FILENAME
echo "VITE_BLOCK_EXPLORER_URL=$BLOCK_EXPLORER_URL" >> $ENV_FILENAME

# Build app
cd /app && npm run build

# Copy nginx config
cp deployment/nginx.conf /etc/nginx/conf.d/default.conf
echo $ISSUER_UI_AUTH_USERNAME
echo $ISSUER_UI_AUTH_PASSWORD
htpasswd -c -b /etc/nginx/.htpasswd $ISSUER_UI_AUTH_USERNAME $ISSUER_UI_AUTH_PASSWORD
cat /etc/nginx/.htpasswd

# Copy app dist
cp -r /app/dist/. /usr/share/nginx/html

# Delete source code
rm -rf /app

# Run nginx
nginx -g 'daemon off;'
