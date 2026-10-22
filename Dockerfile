
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

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/taskflow"]
