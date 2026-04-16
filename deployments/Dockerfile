# Stage 1 — Builder
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/server

# Stage 2 — Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app/server /server
COPY --from=builder /app/migrations /migrations
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8000/api/v1/health || exit 1
ENTRYPOINT ["/server"]
