FROM golang:1.19 as builder

ARG VERSION

COPY . /service

WORKDIR /service
ENV GOBIN /service/bin
RUN go install -ldflags "-X main.build=${VERSION}" ./cmd/...

# Build an issuer image
FROM alpine:3.16.0

RUN apk add --no-cache libstdc++ gcompat libgomp

WORKDIR /service
COPY --from=builder /service/bin/migrate /service
COPY --from=builder /service/bin/platform /service
COPY --from=builder /service/bin/pending_publisher /service
COPY --from=builder /service/bin/admin /service
COPY --from=builder /service/config.toml /service/config.toml
COPY --from=builder /service/api/spec.html /service/api/spec.html
COPY --from=builder /service/api/api.yaml /service/api/api.yaml
COPY --from=builder "/go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so" \
"/go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so"
COPY --from=builder "/service/pkg/credentials" \
"/service/pkg/credentials"