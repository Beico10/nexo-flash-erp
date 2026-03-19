# =============================================================================
# Nexo Flash — Dockerfile multi-stage
# Imagem final: ~10MB (scratch), zero shell, zero vulnerabilidades de SO
# =============================================================================

# Estágio 1: build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always)" \
    -trimpath \
    -o nexo-flash \
    ./cmd/api

# Estágio 2: imagem mínima
# scratch = sem SO, sem shell, sem vulnerabilidades conhecidas
FROM scratch

# Certificados TLS e timezone (copiados do builder)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Usuário não-root (UID 65534 = nobody em scratch)
USER 65534:65534

COPY --from=builder /build/nexo-flash /nexo-flash

EXPOSE 8080

ENTRYPOINT ["/nexo-flash"]
