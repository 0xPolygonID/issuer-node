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

FROM alpine:latest
RUN apk add --no-cache libstdc++ gcompat libgomp
RUN apk add --update busybox>1.3.1-r0
RUN apk add --update openssl>3.1.4-r1


RUN apk add doas; \
    adduser -S issuer -D -G wheel; \
    echo 'permit nopass :wheel as root' >> /etc/doas.d/doas.conf;
RUN chmod g+rx,o+rx /

RUN apk add curl
COPY --from=base ./service/api ./api
COPY --from=base ./service/bin/* ./
COPY --from=base ./service/pkg/credentials ./pkg/credentials
COPY --from=base ./service/resolvers_settings.* ./