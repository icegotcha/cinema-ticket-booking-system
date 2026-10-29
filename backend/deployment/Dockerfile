FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api


RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/worker \
    ./cmd/worker

FROM alpine:3.23 AS api

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/api /server

EXPOSE 8080
ENTRYPOINT ["/server"]

FROM alpine:3.23 AS worker

WORKDIR /app

COPY --from=builder /out/worker /worker

ENTRYPOINT ["/worker"]