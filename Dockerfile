FROM ubuntu:22.04 as builder

ARG VERSION
WORKDIR /service
COPY ./api ./api
COPY api_ui ./api_admin
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./go.mod ./
COPY ./go.sum ./

RUN apt-get update
RUN apt-get install -y wget build-essential ca-certificates
RUN wget https://go.dev/dl/go1.19.linux-amd64.tar.gz

# Configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /usr/local/go/bin:/go/bin:$PATH
ENV GOBIN /service/bin

RUN tar -xvf go1.19.linux-amd64.tar.gz
RUN mv go /usr/local

RUN go mod download
RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./cmd/...
RUN mv /service/bin/* /service/
RUN rm -R /usr/local/go
RUN rm -R /service/bin
RUN rm ./go1.19.linux-amd64.tar.gz