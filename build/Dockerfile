FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pull-requester ./cmd/pull-requester

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/pull-requester .
COPY --from=builder /app/internal/infrastructure/storage/migrations ./migrations
COPY --from=builder /app/config ./config

RUN echo "=== Checking config files ===" && \
    ls -la /root/config/ && \
    echo "=== Testing file access ===" && \
    cat /root/config/config.docker.yaml

RUN adduser -D -s /bin/sh appuser && \
    chown -R appuser:appuser /root/ && \
    chmod -R 755 /root/config

USER appuser

#для проверки
RUN echo "=== Checking as appuser ===" && \
    whoami && \
    ls -la /root/config/ && \
    cat /root/config/config.docker.yaml

EXPOSE 8080

CMD ["./pull-requester"]