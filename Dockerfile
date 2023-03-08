FROM ubuntu:22.04 as builder

ARG VERSION

RUN apt-get update && apt-get install -y wget build-essential ca-certificates

# Configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /usr/local/go/bin:/go/bin:$PATH
ENV GOBIN /service/bin

WORKDIR /usr/local
RUN wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
RUN tar -xvf go1.19.linux-amd64.tar.gz
RUN rm ./go1.19.linux-amd64.tar.gz

WORKDIR /service
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./cmd/...
COPY ./api ./api
COPY ./api_admin ./api_admin
RUN mv /service/bin/* /service/
RUN rm -R /usr/local/go
RUN rm -R /service/bin

COPY ./config.toml ./
