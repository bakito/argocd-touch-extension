FROM golang:1.26-alpine AS builder
WORKDIR /go/src/app

RUN apk update && apk add upx

ARG VERSION=main
ARG BUILD="N/A"

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

COPY . /go/src/app/

RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/argocd-touch-extension/pkg/version.Version=${VERSION} -X github.com/bakito/argocd-touch-extension/pkg/version.Build=${BUILD}" -o argocd-touch-extension . && \
    upx -q argocd-touch-extension

# application image
FROM scratch
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
ENTRYPOINT ["/opt/go/argocd-touch-extension"]
COPY --from=builder /go/src/app/argocd-touch-extension /opt/go/argocd-touch-extension
USER 999
