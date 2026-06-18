#!/bin/sh
# Runtime entrypoint for EKS — the app was already built at Docker build time.
# Only nginx configuration is handled here.

if [ "${ISSUER_UI_INSECURE}" = "true" ]; then
  cp /app/deployment/nginx_insecure.conf /etc/nginx/conf.d/default.conf
else
  cp /app/deployment/nginx.conf /etc/nginx/conf.d/default.conf
  htpasswd -c -b /etc/nginx/.htpasswd "$ISSUER_UI_AUTH_USERNAME" "$ISSUER_UI_AUTH_PASSWORD"
fi

nginx -g 'daemon off;'
