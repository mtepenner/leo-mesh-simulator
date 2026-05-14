FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o orbital-engine ./cmd/engine

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/orbital-engine .

EXPOSE 8080

CMD ["./orbital-engine", "-redis-addr", "redis:6379", "-stream-key", "topology-stream", "-snapshot-key", "topology-snapshot", "-num-sats", "500", "-verbose"]
