FROM golang:latest as builder

ARG VERSION

COPY . /service

WORKDIR /service
ENV GOBIN /service/bin
RUN go install -ldflags "-X main.build=${VERSION}" ./cmd/...

# Build an issuer image
FROM alpine:3.16.0

RUN apk add --no-cache libstdc++ gcompat libgomp

WORKDIR /service
COPY --from=builder /build/bin /service
COPY --from=builder "/go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so" \
"/go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so"