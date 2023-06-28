FROM golang:1.20 as base
ARG VERSION
WORKDIR /service
ENV GOBIN /service/bin
COPY ./api ./api
COPY ./api_ui ./api_ui
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./go.mod ./
COPY ./go.sum ./
RUN go install -buildvcs=false -ldflags "-X main.build=${VERSION}" ./cmd/...

FROM alpine:latest
RUN apk add --no-cache libstdc++ gcompat libgomp
RUN ln -sfv ld-linux-x86-64.so.2 /lib/libresolv.so.2

RUN apk add doas; \
    adduser -S issuer -D -G wheel; \
    echo 'permit nopass :wheel as root' >> /etc/doas.d/doas.conf;
RUN chmod g+rx,o+rx /

COPY --from=base ./service/api ./api
COPY --from=base ./service/api_ui ./api_ui
COPY --from=base ./service/bin/* ./
COPY --from=base ./service/pkg/credentials ./pkg/credentials
COPY --from=base "/go/pkg/mod/github.com/iden3/wasmer-go@v0.0.1/wasmer/packaged/" \
 "/go/pkg/mod/github.com/iden3/wasmer-go@v0.0.1/wasmer/packaged/"