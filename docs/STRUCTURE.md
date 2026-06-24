# Estrutura do Monorepo — Distributed Ledger Go

**Status:** Documentação Centralizada (2026-06-24)
**Proprietário:** Staff Engineering
**Objetivo:** Servir como referência única para arquitetura, layout e organização do monorepo

---

## Índice Rápido

- [Visão Geral](#visão-geral-da-arquitetura)
- [Layout do Monorepo](#layout-do-monorepo)
- [Microserviços](#microserviços-e-componentes)
- [Camadas & Padrões](#camadas--padrões-arquiteturais)
- [Convenções](#convenções-de-código-e-diretórios)
- [Como Expandir](#como-expandir-o-monorepo)

---

## Visão Geral da Arquitetura

O projeto **Distributed Ledger Go** é um ecossistema completo de ledger distribuído, orientado a eventos, implementado seguindo rigorosamente:

- **Arquitetura Hexagonal (Ports and Adapters):** lógica de negócio independente de frameworks, BDs, protocolos
- **SOLID Principles:** single responsibility, open/closed, Liskov substitution, interface segregation, dependency inversion
- **Clean Architecture:** separação clara entre camadas (domain, application, infrastructure)
- **Event-Driven Architecture:** comunicação assíncrona via Kafka, garantias de entrega eventual

**Stack Tecnológica:**
- **Linguagem:** Go 1.22+
- **Mensageria:** Apache Kafka (com KRaft)
- **Banco de Dados:** PostgreSQL 16
- **Cache/Rate Limit:** Redis 7.2
- **Containerização:** Docker & Docker Compose

---

## Layout do Monorepo

```
/distributed-ledger-go
├── README.md                          Root documentation (this should point to /docs)
├── docker-compose.yml                 Full stack setup (Kafka, Postgres, Redis, Apps)
├── .github/
│   └── workflows/
│       ├── ci.yml                     CI/CD pipeline (build, test, lint)
│       └── release.yml                Release automation
├── apps/                              Microservices layer
│   ├── rate-limiter/                  HTTP Rate Limiter (Project 2)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── README.md
│   ├── transaction-gw/                HTTP API Gateway (Project 1)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── README.md
│   ├── ledger-core/                   Kafka Consumer (Project 1)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── README.md
│   ├── notification-service/          Event Consumer (Project 3)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── README.md
│   ├── ledger-reconciler/             Batch Auditor (Project 4)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── README.md
│   └── ledger-backoffice/             Admin Dashboard (Project 5)
│       ├── cmd/
│       ├── internal/
│       ├── web/                       HTML templates & static assets
│       ├── go.mod
│       └── README.md
├── shared/                            Shared packages (events, DTOs, infrastructure)
│   ├── contracts/                     Event contracts & schemas
│   │   ├── events.go                  Domain events
│   │   └── transaction_v1.go          Transaction event versioning
│   ├── models/                        DTOs & value objects
│   │   ├── transaction.go
│   │   └── ledger_entry.go
│   ├── clients/                       Infrastructure clients
│   │   ├── kafkaClient.go
│   │   ├── postgresClient.go
│   │   └── redisClient.go
│   ├── middleware/                    HTTP middleware (auth, logging, etc)
│   │   ├── errorHandler.go
│   │   └── correlation.go
│   ├── go.mod
│   └── README.md
├── migrations/                        Database migrations
│   ├── 01_initial_schema.sql
│   ├── 02_outbox_table.sql
│   └── 03_outbox_failed_events.sql
├── schema/                            JSON schemas & validation
│   ├── json/
│   │   └── transactions/
│   │       ├── request.schema.json
│   │       └── response.schema.json
│   └── proto/                         Protocol buffers (future)
├── docs/                              Complete documentation
│   ├── README.md                      Documentation index
│   ├── QUICKSTART.md                  5-step onboarding
│   ├── ARCHITECTURE_FLOWS.md          8 Mermaid diagrams
│   ├── ERROR_HANDLING_PATTERNS.md     Error handling matrix + Go examples
│   ├── STRUCTURE.md                   This file
│   ├── system-design.md               Technical architecture
│   ├── business.md                    Business vision & SLAs
│   ├── dev-team.md                    Developer workflow
│   ├── reference/
│   │   ├── technical-contracts.md     API endpoints & events
│   │   ├── operational-compliance-policy.md  Unified policies
│   │   ├── idempotency-guide.md       Idempotency reference
│   │   ├── observability.md           Metrics & alerts
│   │   ├── faq.md                     Design decisions
│   │   └── shared-module.md           Shared package docs
│   └── playbooks/
│       ├── dlq-playbook.md            DLQ reprocessing runbook
│       └── operations-runbooks.md     Ops runbooks (lag, hot partitions, reconciliation)
└── .gitignore

```

---

## Microserviços e Componentes

### 1. Ledger Imutável Distribuído (`apps/ledger-core` + `apps/transaction-gw`)

**Responsabilidade:** Garantir consistência estrita dos saldos com double-entry bookkeeping

**Key Patterns:**
- **Append-Only:** Sem `UPDATE` ou `DELETE` em lançamentos — apenas `INSERT`
- **Integer Values:** Armazenamento em centavos (BIGINT) para evitar erros de ponto flutuante
- **Strict Ordering:** Chaveamento por `account_id` no Kafka para garantir sequencialidade por conta
- **Idempotency:** Uso de `idempotency_key` para garantir *At-Least-Once* delivery

**Fluxo de Transação:**
1. Client → `transaction-gw` (HTTP POST)
2. `transaction-gw` valida, gera `transaction_id`, escreve em `outbox` table
3. Outbox Worker publica em Kafka `transactions` topic
4. `ledger-core` consumer processa evento, aplica em `ledger_entries`
5. Se erro → Outbox marca como `FAILED`, DLQ captura para reprocessamento

**Banco de Dados:**
```sql
-- /migrations/01_initial_schema.sql
CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  idempotency_key VARCHAR(255) NOT NULL UNIQUE,
  status VARCHAR(20),  -- PENDING, COMPLETED, FAILED
  created_at TIMESTAMP NOT NULL,
  ...
);

CREATE TABLE ledger_entries (
  id UUID PRIMARY KEY,
  transaction_id UUID REFERENCES transactions(id),
  account_id UUID NOT NULL,
  amount BIGINT NOT NULL,  -- in cents
  created_at TIMESTAMP NOT NULL,
  ...
);

CREATE TABLE outbox (
  id UUID PRIMARY KEY,
  aggregate_type VARCHAR(50),
  aggregate_id UUID,
  event_type VARCHAR(50),
  payload JSONB,
  status VARCHAR(20),  -- PENDING, PUBLISHED, FAILED
  created_at TIMESTAMP NOT NULL,
  ...
);
```

---

### 2. Rate Limiter Distribuído Adaptativo (`apps/rate-limiter`)

**Responsabilidade:** Proteção contra exaustão de recursos com fallback local

**Key Patterns:**
- **Atomicidade:** Scripts Lua em Redis para evitar race conditions
- **Adaptive Limits:** Ajusta limites por conta baseado em histórico
- **Graceful Degradation:** Fallback para `sync.Map` em memória se Redis cair

**Integração:**
- Middleware HTTP em `transaction-gw`
- Protege ingress com tokens por account_id
- Retorna 429 (Too Many Requests) quando limite excedido

---

### 3. Notification Service (`apps/notification-service`)

**Responsabilidade:** Consumir eventos Kafka e enviar notificações (email, SMS, webhooks)

**Key Patterns:**
- **Worker Pool Pattern:** Pool fixo de goroutines para processar eventos
- **Exponential Backoff:** Retry com jitter para falhas transitórias
- **Circuit Breaker:** Proteção contra latência excessiva em provedores externos
- **At-Least-Once Semantics:** Commit offset após processamento bem-sucedido

**Tópicos Consumidos:**
- `transactions` — new transaction events
- `failed_events` — DLQ events para notificações de falha

---

### 4. Ledger Reconciliador (`apps/ledger-reconciler`)

**Responsabilidade:** Auditoria em batch para cura de eventual consistência

**Key Patterns:**
- **Cursor Pagination:** Varredura eficiente de tabelas massivas ($O(1)$ per batch)
- **Controlled Concurrency:** `errgroup` com limite máximo de workers
- **Idempotency:** Marcas de reconciliação para evitar reprocessamento

**Job Agendado:**
- Roda a cada hora (configurável)
- Verifica discrepâncias entre saldos esperados e observados
- Gera relatórios e alertas se inconsistências encontradas

---

### 5. Ledger Backoffice (`apps/ledger-backoffice`)

**Responsabilidade:** Dashboard administrativo e auditoria em tempo real

**Key Patterns:**
- **Server-Side Rendering:** Templates Go + Tailwind CSS
- **Audit Trail:** Inspeção imutável de transações por conta
- **System Health:** Monitoramento de DLQ e alertas de inconsistência

**Funcionalidades:**
- Visualizar saldo de qualquer conta
- Histórico completo de transações
- Status de DLQ e mensagens falhadas
- Reprocessamento manual de DLQ

---

## Camadas & Padrões Arquiteturais

### Hexagonal Architecture (Ports & Adapters)

Cada microserviço segue a estrutura abaixo dentro de `/internal`:

```
apps/ledger-core/internal/
├── domain/              ← Business logic (entities, value objects, use cases)
│   ├── ledger/
│   │   ├── ledger.go                (entity)
│   │   ├── entry.go                 (value object)
│   │   └── create_entry.go          (use case)
│   └── events/
│       └── transaction_created.go   (domain event)
├── application/         ← Application layer (services, DTOs)
│   ├── services/
│   │   └── transaction_service.go
│   └── dto/
│       └── transaction_dto.go
├── infrastructure/      ← Adapters (DB, Kafka, HTTP)
│   ├── persistence/
│   │   ├── postgres_repository.go
│   │   └── migrations.go
│   ├── messaging/
│   │   ├── kafka_consumer.go
│   │   └── kafka_publisher.go
│   └── http/
│       └── handler.go
└── shared/              ← Shared utilities
    ├── logger.go
    ├── errors.go
    └── middleware.go
```

### Key Principles

1. **Domain-Driven Design (DDD):**
   - Ubiquitous language
   - Bounded contexts (each app is a context)
   - Value objects vs entities

2. **SOLID:**
   - Single Responsibility: each file ~1 concern
   - Open/Closed: extensible without modification
   - Liskov Substitution: interfaces over implementations
   - Interface Segregation: small focused interfaces
   - Dependency Inversion: depend on abstractions

3. **Error Handling:**
   - Custom error types in `domain/` (e.g., `InvalidAmountError`)
   - Propagation via interface contracts
   - Log + alert + DLQ strategy (see `ERROR_HANDLING_PATTERNS.md`)

4. **Testing:**
   - Unit tests in `*_test.go` same package
   - Integration tests in `/test` folder
   - Mock interfaces for external dependencies

---

## Convenções de Código e Diretórios

### Naming Conventions

| Entity | Convention | Example |
|--------|-----------|---------|
| Go packages | lowercase, single word | `domain`, `infrastructure`, `ledger` |
| Go files | snake_case.go | `transaction_service.go` |
| Go interfaces | PascalCase + "er" suffix | `TransactionRepository` |
| Go structs | PascalCase | `Transaction`, `LedgerEntry` |
| Go functions | PascalCase | `CreateTransaction()` |
| Private functions | camelCase | `validateAmount()` |
| Constants | SCREAMING_SNAKE_CASE | `TRANSACTION_TIMEOUT_MS` |
| SQL tables | snake_case, plural | `transactions`, `ledger_entries` |
| SQL columns | snake_case | `transaction_id`, `created_at` |
| Kafka topics | kebab-case | `transactions`, `failed-events` |
| Docker containers | kebab-case | `ledger-core`, `transaction-gw` |
| Environment variables | SCREAMING_SNAKE_CASE | `DATABASE_URL`, `KAFKA_BROKERS` |

### Directory Conventions

- `/cmd/` — entry points (main.go for app, CLI tools)
- `/internal/` — private implementation (never imported by other apps)
- `/test/` — integration & contract tests
- `/web/` — static assets, HTML templates (for backoffice only)
- `/migrations/` — SQL migration scripts
- `/schema/` — JSON/Proto schemas

---

## Como Expandir o Monorepo

### Adicionar Novo Microserviço

**Step 1:** Criar estrutura
```bash
mkdir -p apps/my-service/internal/{domain,application,infrastructure}
mkdir -p apps/my-service/cmd
touch apps/my-service/go.mod
touch apps/my-service/go.sum
touch apps/my-service/cmd/main.go
```

**Step 2:** Inicializar módulo Go
```bash
cd apps/my-service
go mod init github.com/example/distributed-ledger-go/apps/my-service
go get github.com/example/distributed-ledger-go/shared
```

**Step 3:** Implementar domínio
```go
// apps/my-service/internal/domain/my_entity.go
package domain

type MyEntity struct {
  ID    string
  Value int
}

func (e *MyEntity) Validate() error {
  // validation logic
  return nil
}
```

**Step 4:** Implementar application service
```go
// apps/my-service/internal/application/my_service.go
package application

type MyService struct {
  repository Repository
}

func (s *MyService) DoSomething(ctx context.Context) error {
  // business logic
}
```

**Step 5:** Implementar adapters (HTTP, Kafka, DB)
```go
// apps/my-service/internal/infrastructure/http/handler.go
package http

func (h *Handler) POST(w http.ResponseWriter, r *http.Request) {
  // HTTP adapter
}
```

**Step 6:** Criar main.go
```go
// apps/my-service/cmd/main.go
package main

func main() {
  // initialize dependencies
  // start servers
}
```

**Step 7:** Adicionar ao docker-compose.yml
```yaml
services:
  my-service:
    build:
      context: .
      dockerfile: apps/my-service/Dockerfile
    environment:
      - DATABASE_URL=postgresql://...
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - kafka
```

---

## Padrões Comuns

### Error Handling

Veja [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) para detalhes.

```go
// Domain layer defines custom errors
type InvalidAmountError struct {
  Amount int64
}

func (e InvalidAmountError) Error() string {
  return fmt.Sprintf("invalid amount: %d", e.Amount)
}

// Application layer handles and propagates
func (s *Service) CreateTransaction(ctx context.Context, req *Request) error {
  if req.Amount <= 0 {
    return InvalidAmountError{Amount: req.Amount}
  }
  // ...
}

// Infrastructure layer logs, alerts, and sends to DLQ
if err != nil {
  log.Error("transaction failed", "error", err)
  if isTransient(err) {
    // retry
  } else {
    // send to DLQ
  }
}
```

### Idempotency

Veja [reference/idempotency-guide.md](reference/idempotency-guide.md) para detalhes.

```go
// Client provides idempotency_key
POST /transactions
X-Idempotency-Key: "client-generated-uuid-or-key"

// Service stores and checks
func (s *Service) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
  // Check if idempotency_key already exists
  existing, err := s.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
  if err == nil {
    return existing, nil  // idempotent response
  }
  
  // Otherwise create new
  tx := &Transaction{
    ID:             uuid.New().String(),
    IdempotencyKey: req.IdempotencyKey,
    Status:         StatusPending,
  }
  
  return s.repo.Create(ctx, tx)
}
```

### Kafka Consumer Pattern

```go
// Consumer group setup
consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
consumerGroup, err := sarama.NewConsumerGroup([]string{"kafka:9092"}, "ledger-core-group", nil)

// Handler
type Handler struct {
  service MyService
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
  for message := range claim.Messages() {
    var event TransactionEvent
    json.Unmarshal(message.Value, &event)
    
    if err := h.service.Handle(context.Background(), &event); err != nil {
      // log, alert, possibly send to DLQ
      return err
    }
    
    session.MarkMessage(message, "")
  }
  return nil
}

// Start consumer
go func() {
  for err := range consumerGroup.Errors() {
    log.Error("consumer error", "error", err)
  }
}()
```

---

## Links Relacionados

- **[QUICKSTART.md](QUICKSTART.md)** — Como começar rapidamente
- **[ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md)** — Diagramas de fluxo
- **[system-design.md](system-design.md)** — Design técnico profundo
- **[ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md)** — Padrões de erro
- **[dev-team.md](dev-team.md)** — Workflow do desenvolvedor
- **[reference/technical-contracts.md](reference/technical-contracts.md)** — APIs & eventos

---

**Documento criado:** 2026-06-24  
**Próxima revisão:** 2026-08-01  
**Proprietário:** Staff Engineering
