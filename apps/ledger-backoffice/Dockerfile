FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work go.work.sum ./
COPY apps/ledger-backoffice/go.mod apps/ledger-backoffice/go.sum ./apps/ledger-backoffice/
COPY apps/ledger-core/go.mod apps/ledger-core/go.sum ./apps/ledger-core/
COPY apps/ledger-reconciler/go.mod apps/ledger-reconciler/go.sum ./apps/ledger-reconciler/
COPY apps/notification-service/go.mod apps/notification-service/go.sum ./apps/notification-service/
COPY apps/rate-limiter/go.mod apps/rate-limiter/go.sum ./apps/rate-limiter/
COPY apps/transaction-gw/go.mod apps/transaction-gw/go.sum ./apps/transaction-gw/
COPY shared/go.mod shared/go.sum ./shared/

RUN go mod download

COPY apps/ledger-backoffice ./apps/ledger-backoffice
COPY shared ./shared

RUN go build -o ledger-backoffice ./apps/ledger-backoffice/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/ledger-backoffice .
# Copiar templates para o container
COPY --from=builder /app/apps/ledger-backoffice/web ./apps/ledger-backoffice/web

EXPOSE 8081

CMD ["./ledger-backoffice"]
