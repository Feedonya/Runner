FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o code-runner ./cmd/main.go

FROM golang:1.24

WORKDIR /app

COPY --from=builder /app/code-runner .

RUN chmod +x /app/code-runner

ENV REDIS_HOST=dragonfly:6379 \
    REDIS_PASSWORD= \
    REDIS_DB=0 \
    MINIO_ENDPOINT=minio:9000 \
    MINIO_ACCESS_KEY=minioadmin \
    MINIO_SECRET_KEY=minioadmin \
    USE_TMPFS=true

VOLUME /var/run/docker.sock

CMD ["./code-runner"]