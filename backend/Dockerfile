# Build stage
FROM golang:1.24-alpine AS builder

# Permitir download automático de toolchain
ENV GOTOOLCHAIN=auto

# Instalar certificados SSL e outras dependências
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build do binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Final stage
FROM alpine:latest

# Instalar certificados SSL
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copiar binário do builder
COPY --from=builder --chown=appuser:appuser /app/server .

# Copiar diretório config (firebase credentials serão adicionadas durante build no CI/CD)
COPY --from=builder --chown=appuser:appuser /app/config ./config

# Usar usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Executar aplicação
CMD ["./server"]
