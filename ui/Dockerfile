FROM nginx:alpine

WORKDIR /app

RUN apk add --update nodejs npm
RUN apk add --update apache2-utils
RUN apk add --update libwebp=1.3.2-r0
#RUN apk add --update libx11=1.8.7-r0
RUN apk add --update openssl>3.1.4-r1
COPY package.json package-lock.json ./

RUN npm install


COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
