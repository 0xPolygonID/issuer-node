FROM golang:1.22 AS base
ARG VERSION
WORKDIR /service
ENV GOBIN=/service/bin
COPY ./api ./api
COPY ./cmd ./cmd
COPY ./tools/vault-migrator ./tools/vault-migrator
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./go.mod ./
COPY ./go.sum ./

# uncoment if you want to use resolvers_settings.yaml file in the build
# COPY ./resolvers_settings.* ./

RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./cmd/...
RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./tools/...

FROM node:alpine AS ui-builder
WORKDIR /app
COPY ./ui ./
RUN npm install
RUN npm run build

FROM alpine:latest
RUN apk add --no-cache libstdc++ gcompat libgomp
RUN apk add --update busybox>1.3.1-r0
RUN apk add --update openssl>3.1.4-r1
RUN apk add --update nodejs npm
RUN apk add --update apache2-utils
RUN apk add --update libwebp=1.3.2-r0
RUN apk add --update openssl>3.1.4-r1
RUN apk add curl

COPY --from=base ./service/api ./api
COPY --from=base ./service/bin/* ./
COPY --from=base ./service/pkg/credentials ./pkg/credentials
COPY --from=base ./service/resolvers_settings.* ./
COPY --from=ui-builder ./app/dist ./ui/dist

