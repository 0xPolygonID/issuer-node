FROM nginx:alpine

ARG UI_BASIC_USERNAME
ARG UI_BASIC_PASSWORD

ENV ISSUER_UI_BASIC_USERNAME=$UI_BASIC_USERNAME
ENV ISSUER_UI_BASIC_PASSWORD=$UI_BASIC_PASSWORD

WORKDIR /app
RUN apk add --update nodejs npm
RUN apk add --update apache2-utils
COPY package.json package-lock.json ./

RUN npm install

COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
