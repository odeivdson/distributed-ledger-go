FROM golang:1.26-alpine AS builder

WORKDIR /app

# Instalar dependências básicas
RUN apk add --no-cache git

# Copiar go.work e todos os go.mod para aproveitar o cache de camadas
COPY go.work go.work.sum ./
COPY apps/ledger-backoffice/go.mod apps/ledger-backoffice/go.sum ./apps/ledger-backoffice/
COPY apps/ledger-core/go.mod apps/ledger-core/go.sum ./apps/ledger-core/
COPY apps/ledger-reconciler/go.mod apps/ledger-reconciler/go.sum ./apps/ledger-reconciler/
COPY apps/notification-service/go.mod apps/notification-service/go.sum ./apps/notification-service/
COPY apps/rate-limiter/go.mod apps/rate-limiter/go.sum ./apps/rate-limiter/
COPY apps/transaction-gw/go.mod apps/transaction-gw/go.sum ./apps/transaction-gw/
COPY shared/go.mod shared/go.sum ./shared/

RUN go mod download

# Copiar o código fonte
COPY apps/transaction-gw ./apps/transaction-gw
COPY apps/rate-limiter ./apps/rate-limiter
COPY shared ./shared

# Build do binário
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./apps/transaction-gw/cmd/app/main.go

# Estágio final
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/app .
EXPOSE 8080
CMD ["./app"]
