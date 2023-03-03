FROM nginx:alpine

RUN apk add --update nodejs npm

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm install

COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
