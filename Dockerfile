# ============================================================
# NEXO ONE ERP — Dockerfile Multi-Stage (Backend Go)
# ============================================================
# DIRETRIZES: Máxima segurança, Mínimo custo
#
# Segurança:
#   - Build em imagem separada (secrets não vazam)
#   - Binário estático (zero dependências runtime)
#   - Imagem final distroless (sem shell, sem package manager)
#   - Usuário non-root (UID 65532)
#   - Sem capabilities extras
#
# Custo:
#   - Imagem final ~15MB (vs ~1GB com golang:alpine)
#   - Build cache otimizado (go mod download separado)
#   - Multi-stage elimina ferramentas de build
# ============================================================

# ═══════════════════════════════════════════════════════════
# STAGE 1: Builder — Compila o binário Go
# ═══════════════════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

# Segurança: certificados para HTTPS + timezone
RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /build

# Cache de dependências (só rebuilda se go.mod/sum mudar)
COPY go.mod go.sum* ./
RUN go mod download

# Copia código fonte
COPY . .

# Atualiza go.sum se necessário e compila
RUN go mod tidy

# Compila binário estático
# CGO_ENABLED=0: sem dependência de libc (portável)
# -ldflags: remove símbolos de debug (-s -w = -30% tamanho)
# -trimpath: remove paths locais do binário (segurança)
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o /build/nexo-one \
    ./cmd/api

# Verifica se compilou
RUN test -f /build/nexo-one && echo "✅ Build OK: $(ls -lh /build/nexo-one)"

# ═══════════════════════════════════════════════════════════
# STAGE 2: Runtime — Imagem mínima de produção
# ═══════════════════════════════════════════════════════════
FROM gcr.io/distroless/static-debian12:nonroot

# Metadata
LABEL org.opencontainers.image.title="Nexo One ERP"
LABEL org.opencontainers.image.description="ERP Multi-Tenant Brasil 2026 - Do TOTVS ao cafezinho"
LABEL org.opencontainers.image.vendor="Nexo One"
LABEL org.opencontainers.image.source="https://github.com/Beico10/nexo-one-erp"

# Copia certificados e timezone
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copia binário
COPY --from=builder /build/nexo-one /app/nexo-one

# Usuário non-root (distroless já vem com nonroot:65532)
USER nonroot:nonroot

# Variáveis de ambiente padrão
ENV TZ=America/Sao_Paulo
ENV APP_ENV=production
ENV PORT=8080

# Health check endpoint
EXPOSE 8080

# Ponto de entrada
ENTRYPOINT ["/app/nexo-one"]

# ============================================================
# COMO USAR:
#
# Build:
#   docker build -t nexo-one:latest \
#     --build-arg VERSION=$(git describe --tags --always) \
#     --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#     .
#
# Run:
#   docker run -d --name nexo-one \
#     -p 8080:8080 \
#     -e DATABASE_URL="postgres://..." \
#     -e REDIS_URL="redis://..." \
#     -e NATS_URL="nats://..." \
#     -e JWT_SECRET="seu-segredo-forte" \
#     --read-only \
#     --cap-drop=ALL \
#     nexo-one:latest
#
# Tamanho final: ~15-20MB
# ============================================================
