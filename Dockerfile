FROM golang:1.19 as builder

ARG VERSION
WORKDIR /service

COPY ./api ./api
COPY ./api_admin ./api_admin
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./config.toml ./
COPY ./go.mod ./
COPY ./go.sum ./

ENV GOBIN /service/bin

RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./cmd/...
RUN mv /service/bin/* /service/

RUN rm -R /service/bin
