# Alpine's nodejs package lags behind Node security releases; copy a pinned one.
FROM node:24.22.1-alpine3.24 AS node-runtime

# Floats for security rebuilds; its Alpine must be >= node-runtime's.
FROM nginx:alpine

WORKDIR /app

RUN apk add --no-cache libstdc++ apache2-utils
COPY --from=node-runtime /usr/local/bin/node /usr/local/bin/node
COPY --from=node-runtime /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/npm
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm
COPY package.json package-lock.json ./

RUN npm install

COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
