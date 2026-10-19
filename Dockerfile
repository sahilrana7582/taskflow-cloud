
# Build stage
FROM golang:1.26.6 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/taskflow ./cmd/main.go

# Minimal production image
FROM scratch

WORKDIR /app

COPY --from=builder /out/taskflow /app/taskflow

# TLS certificates for secure database connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

# Goose reads the SQL migrations from disk at startup. Config is not copied:
# production reads only TASKFLOW_* environment variables, and config/ is
# excluded by .dockerignore.
COPY --from=builder /src/internal/db/migrations/ /app/internal/db/migrations/

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/taskflow"]
