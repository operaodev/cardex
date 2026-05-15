# --- Stage 1: Build ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar dependencias primero para aprovechar cache de capas
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar binario estático
RUN CGO_ENABLED=0 GOOS=linux go build -o cardex ./cmd/api

# --- Stage 2: Runtime ---
FROM alpine:3.20

# Certificados para conexiones TLS externas (Yugipedia API, etc.)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/cardex .

EXPOSE 8080

CMD ["./cardex"]
