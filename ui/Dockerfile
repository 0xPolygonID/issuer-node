FROM nginx:alpine
# FROM nginx:1.29.3-alpine 

WORKDIR /app

RUN apk add --update nodejs npm
RUN apk add --update apache2-utils
COPY package.json package-lock.json ./

RUN npm install

COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
