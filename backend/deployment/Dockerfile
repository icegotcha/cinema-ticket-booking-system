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
    -o /app/server \
    ./cmd/api

FROM scratch

WORKDIR /app

COPY --from=builder /app/server /server
COPY --from=builder /app/config /config

EXPOSE 8080
ENTRYPOINT ["/server"]
