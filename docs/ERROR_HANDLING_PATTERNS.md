# Error Handling Patterns — Matriz, Estratégias e Exemplos Go

**Propósito:** Definir padrões claros de tratamento de erro para garantir consistência, resiliência e observabilidade em todo o sistema.

**Última atualização:** 2026-06-24

---

## 1. Matriz de Tipos de Erro

Nem todo erro é igual. Esta matriz classifica erros por características e define a ação apropriada.

```
┌─────────────────────────────────────────────────────────────────────────┐
│ MATRIZ DE CLASSIFICAÇÃO DE ERROS                                        │
├──────────────┬──────────────┬────────────────┬──────────┬──────────────┤
│ Tipo         │ Retentável?  │ Logar?         │ Alertar? │ Ação         │
├──────────────┼──────────────┼────────────────┼──────────┼──────────────┤
│ TRANSITÓRIO  │ Sim (5x)     │ WARN (limite:1)│ Não      │ Retry+Backoff│
│ - Timeout    │              │                │          │              │
│ - Network    │              │                │          │              │
│ - Rate limit │              │                │          │              │
├──────────────┼──────────────┼────────────────┼──────────┼──────────────┤
│ PERMANENTE   │ Não          │ ERROR (sempre) │ Sim      │ DLQ + Manual │
│ - Validação  │              │                │          │              │
│ - Schema     │              │                │          │              │
│ - Negócio    │              │                │          │              │
├──────────────┼──────────────┼────────────────┼──────────┼──────────────┤
│ DESCONHECIDO │ Sim (cauteloso) │ ERROR + WARN│ Sim (WARN)│ Retry1x+DLQ │
│ - Novo tipo  │              │                │          │              │
└──────────────┴──────────────┴────────────────┴──────────┴──────────────┘
```

---

## 2. Definições de Erro

### 2.1 Erro Transitório

**Características:**
- Relacionado a infraestrutura ou timing
- Provável sucesso em retry
- Não relacionado aos dados

**Exemplos:**
```
- "connection refused" (Kafka/DB offline)
- "i/o timeout" (rede lenta)
- "too many requests" (rate limit)
- "context deadline exceeded" (timeout)
```

**Ação:** Retry com backoff exponencial (500ms → 1s → 2s → 4s → 8s)

---

### 2.2 Erro Permanente

**Características:**
- Relacionado aos dados ou lógica
- Retry NÃO vai resolver
- Requer intervenção humana ou fix de código

**Exemplos:**
```
- "invalid account id" (uuid malformado)
- "insufficient balance" (saldo insuficiente)
- "duplicate idempotency_key" (duplicação detectada)
- "amount must be positive" (validação negócio)
- "schema validation failed" (contrato violado)
```

**Ação:** Log completo, move para DLQ, cria ticket de eng

---

### 2.3 Erro Desconhecido

**Características:**
- Tipo de erro não mapeado
- Pode ser transitório ou permanente
- Requer investigação

**Exemplos:**
```
- Erro customizado de terceira lib
- Erro com stack trace incompreensível
- Database error sem tipo específico
```

**Ação:** Log detalhado + 1 retry cauteloso + se persistir → DLQ

---

## 3. Padrões de Tratamento em Go

### 3.1 Pattern 1: Erro Simples com Log Estruturado

```go
package transaction

import (
    "context"
    "log/slog"
    "errors"
)

// Erro de domínio bem definido
var ErrInvalidAmount = errors.New("amount must be positive")

func ProcessTransaction(ctx context.Context, req *TransactionRequest) error {
    // Validação de erro PERMANENTE
    if req.Amount <= 0 {
        // Log estruturado com contexto
        slog.Error("Validação de transação falhou",
            "reason", "invalid_amount",
            "account_id", req.SourceAccountID,
            "amount", req.Amount,
            "idempotency_key", req.IdempotencyKey,
        )
        // Retorna erro permanente (não vai retentar)
        return ErrInvalidAmount
    }

    // Processamento...
    return nil
}
```

**Quando usar:** Erros de validação de negócio que NÃO devem retentar.

---

### 3.2 Pattern 2: Diferenciando Erro Transitório vs Permanente

```go
package kafka

import (
    "context"
    "errors"
    "io"
    "log/slog"
    "net"
    "time"
)

// Classificar erro
func IsTransientError(err error) bool {
    // Timeout (transitório)
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }

    // Network error (transitório)
    var netErr net.Error
    if errors.As(err, &netErr) {
        return netErr.Timeout() || netErr.Temporary()
    }

    // EOF (transitório, pode reabrir conexão)
    if errors.Is(err, io.EOF) {
        return true
    }

    // Validation error (permanente)
    var valErr ValidationError
    if errors.As(err, &valErr) {
        return false
    }

    // Desconhecido: tratar como transitório com cautela
    return true
}

// Publicar com retry inteligente
func PublishWithRetry(ctx context.Context, broker *Broker, message []byte) error {
    var lastErr error
    maxAttempts := 5
    baseBackoff := 500 * time.Millisecond

    for attempt := 0; attempt < maxAttempts; attempt++ {
        err := broker.Publish(ctx, message)
        if err == nil {
            return nil // Sucesso!
        }

        lastErr = err
        isTransient := IsTransientError(err)

        // Log com nível apropriado
        if isTransient {
            slog.Warn("Erro transitório na publicação, retentando",
                "attempt", attempt+1,
                "max_attempts", maxAttempts,
                "error", err,
                "is_transient", true,
            )
        } else {
            // Erro permanente: não insistir
            slog.Error("Erro permanente na publicação, não vai retentar",
                "error", err,
                "is_transient", false,
            )
            return err
        }

        // Backoff exponencial com jitter
        if attempt < maxAttempts-1 {
            backoff := baseBackoff * (1 << uint(attempt)) // 500ms, 1s, 2s, 4s, 8s
            jitter := time.Duration(rand.Int63n(int64(backoff / 10)))
            sleep := backoff + jitter

            select {
            case <-time.After(sleep):
                // Continuar
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }

    // Excedeu tentativas: ir para DLQ
    slog.Error("Erro permanente ou transitório excedeu max_attempts",
        "attempts", maxAttempts,
        "last_error", lastErr,
    )
    return lastErr
}
```

**Quando usar:** Quando você precisa distinguir entre retry vs falha imediata.

---

### 3.3 Pattern 3: Erro com Fallback Local

```go
package cache

import (
    "context"
    "log/slog"
    "sync"
)

// Fallback local quando Redis falha
type CacheWithFallback struct {
    redis *RedisClient
    local sync.Map // Fallback em memória
}

func (c *CacheWithFallback) Get(ctx context.Context, key string) (string, error) {
    // Tentar Redis primeiro
    val, err := c.redis.Get(ctx, key)
    if err == nil {
        return val, nil // Sucesso
    }

    // Classificar erro
    if IsTransientError(err) {
        slog.Warn("Redis temporariamente indisponível, usando fallback local",
            "key", key,
            "error", err,
        )

        // Fallback: buscar do local
        if localVal, ok := c.local.Load(key); ok {
            slog.Debug("Cache local hit após falha de Redis")
            return localVal.(string), nil
        }

        // Local também vazio: erro
        slog.Error("Cache miss em Redis e local",
            "key", key,
        )
        return "", ErrCacheMiss
    }

    // Erro permanente (ex: Redis deletou a chave)
    slog.Error("Erro permanente no Redis, removendo do fallback",
        "key", key,
        "error", err,
    )
    c.local.Delete(key)
    return "", err
}

func (c *CacheWithFallback) Set(ctx context.Context, key string, value string) error {
    // Tentar Redis
    err := c.redis.Set(ctx, key, value)
    if err == nil {
        // Sucesso: atualizar local também
        c.local.Store(key, value)
        return nil
    }

    // Redis falhou, persistir localmente como fallback
    if IsTransientError(err) {
        slog.Warn("Redis indisponível, persistindo em fallback local",
            "key", key,
        )
        c.local.Store(key, value)
        // Não retornar erro ao caller: degradação graciosa
        return nil
    }

    // Erro permanente
    slog.Error("Erro ao set em Redis")
    return err
}
```

**Quando usar:** Serviços opcionais (cache, notificação) que têm fallback.

---

### 3.4 Pattern 4: Error Context com Dados Estruturados

```go
package domain

import (
    "fmt"
)

// Erro de domínio com contexto
type TransactionError struct {
    Code       string                 // "INVALID_AMOUNT", "INSUFFICIENT_BALANCE"
    Message    string
    Transient  bool
    Metadata   map[string]interface{} // Dados para debug
    OriginalErr error                  // Stack original
}

func (e *TransactionError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *TransactionError) Unwrap() error {
    return e.OriginalErr
}

// Constructor
func NewTransactionError(code, message string, transient bool, metadata map[string]interface{}, originalErr error) *TransactionError {
    return &TransactionError{
        Code:        code,
        Message:     message,
        Transient:   transient,
        Metadata:    metadata,
        OriginalErr: originalErr,
    }
}

// Exemplos de uso
func ValidateTransaction(req *TransactionRequest) error {
    if req.Amount <= 0 {
        return NewTransactionError(
            "INVALID_AMOUNT",
            "Amount must be positive",
            false, // Não transitório
            map[string]interface{}{
                "received_amount": req.Amount,
                "source_account": req.SourceAccountID,
            },
            nil,
        )
    }

    if req.SourceAccountID == req.TargetAccountID {
        return NewTransactionError(
            "SAME_ACCOUNT",
            "Source and target must be different",
            false,
            map[string]interface{}{
                "account_id": req.SourceAccountID,
            },
            nil,
        )
    }

    return nil
}

// No caller: tratar diferentemente
func ProcessRequest(req *TransactionRequest) error {
    err := ValidateTransaction(req)
    if err != nil {
        var txnErr *TransactionError
        if errors.As(err, &txnErr) {
            if txnErr.Transient {
                // Retry
                return RetryWithBackoff(func() error {
                    return ValidateTransaction(req)
                })
            } else {
                // Log e retornar erro permanente
                slog.Error("Erro de transação permanente",
                    "code", txnErr.Code,
                    "message", txnErr.Message,
                    "metadata", txnErr.Metadata,
                )
                return txnErr
            }
        }
    }
    return nil
}
```

**Quando usar:** Erros de domínio que precisam de contexto estruturado.

---

### 3.5 Pattern 5: Middleware de Error Handling HTTP

```go
package http

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "log/slog"
)

type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    TraceID string `json:"trace_id,omitempty"`
}

// Middleware que captura e formata erros
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := r.Header.Get("X-Trace-ID")
        ctx := context.WithValue(r.Context(), "trace_id", traceID)

        // Detectar panic
        defer func() {
            if rec := recover(); rec != nil {
                slog.Error("Panic recuperado",
                    "trace_id", traceID,
                    "panic", rec,
                )
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(ErrorResponse{
                    Code:    "INTERNAL_ERROR",
                    Message: "Internal server error",
                    TraceID: traceID,
                })
            }
        }()

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler com tratamento de erro
func HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
    traceID := r.Context().Value("trace_id").(string)

    var req TransactionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        slog.Error("Erro ao decodificar request",
            "trace_id", traceID,
            "error", err,
        )
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{
            Code:    "INVALID_REQUEST",
            Message: "Invalid request body",
            TraceID: traceID,
        })
        return
    }

    // Validar
    if err := ValidateTransaction(&req); err != nil {
        var txnErr *TransactionError
        if errors.As(err, &txnErr) {
            // Error de domínio: tratado
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(ErrorResponse{
                Code:    txnErr.Code,
                Message: txnErr.Message,
                TraceID: traceID,
            })
            slog.Warn("Validação falhou",
                "trace_id", traceID,
                "code", txnErr.Code,
            )
            return
        }
    }

    // Processar (pode falhar transitoriamente)
    result, err := Process(r.Context(), &req)
    if err != nil {
        // Transitório: retry (aplicação)
        var txnErr *TransactionError
        if errors.As(err, &txnErr) && txnErr.Transient {
            // Retornar 202 Accepted (em processamento)
            w.WriteHeader(http.StatusAccepted)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "PENDING",
                "trace_id": traceID,
            })
            slog.Warn("Processamento pendente",
                "trace_id", traceID,
                "error", err,
            )
            return
        }

        // Permanente: erro
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{
            Code:    "PROCESSING_ERROR",
            Message: "Failed to process transaction",
            TraceID: traceID,
        })
        slog.Error("Processamento falhou",
            "trace_id", traceID,
            "error", err,
        )
        return
    }

    // Sucesso
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)
}
```

**Quando usar:** Handler HTTP que precisa traduzir erros internos para HTTP status codes.

---

## 4. Checklist de Implementação por Tipo

### Para Todo Novo Componente:

- [ ] Definir erros de domínio específicos (ex: `ErrInvalidAmount`)
- [ ] Classificar erro como `Transient` vs `Permanent`
- [ ] Log estruturado com `trace_id` e `idempotency_key`
- [ ] Backoff exponencial com jitter (se retentável)
- [ ] Fallback local (se aplicável)
- [ ] Métrica de erro incrementada
- [ ] Alerta dispara (se crítico)

---

## 5. Fluxo de Decisão (Decision Tree)

```
                    ┌─────────────────────┐
                    │   Erro Ocorreu?     │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │ Tipo de Erro Conhecido?  │
                    └──────┬──────────────┘
                           │
         ┌─────────────────┴─────────────────┐
         │                                   │
    ┌────▼─────┐                      ┌─────▼────┐
    │SIM       │                      │NÃO       │
    └────┬─────┘                      └─────┬────┘
         │                                  │
    ┌────▼────────────┐            ┌────────▼─────┐
    │ É Transitório?  │            │ Assumir      │
    └────┬────────────┘            │ Transitório? │
         │                         └────────┬─────┘
    ┌────┴─────┐                        │
    │SIM  │NÃO │                        │
    │    └────▼──────┐                  │
    │         │      │                  │
    │         │  ┌───▼────┐             │
    │         │  │Log      │             │
    │         │  │Error    │             │
    │         │  │Send DLQ │             │
    │         │  └─────────┘             │
    │         │                          │
┌───▼──┐   ┌──▼────┐        ┌──────────▼──┐
│Retry │   │Return │        │Retry 1x +   │
│+Boff │   │Error  │        │Then DLQ     │
└──────┘   └───────┘        └─────────────┘
```

---

## 6. Exemplos de Erro Completos

### Exemplo 1: Validação de Amount

```go
// Domain
var ErrNegativeAmount = errors.New("amount must be positive")

// Handler
func ValidateAmount(amount int64) error {
    if amount <= 0 {
        slog.Error("Validação falhou",
            "reason", "negative_amount",
            "amount", amount,
        )
        return ErrNegativeAmount // Permanente
    }
    return nil
}

// Middleware responde
resp := ErrorResponse{Code: "INVALID_AMOUNT", Message: "Amount must be positive"}
w.WriteHeader(http.StatusBadRequest) // 400
```

### Exemplo 2: Timeout de Kafka

```go
// Publicação falha
err := broker.Publish(message)
if errors.Is(err, context.DeadlineExceeded) {
    slog.Warn("Timeout ao publicar, retentando",
        "attempt", 1,
        "backoff_ms", 500,
    )
    // Retry com backoff
    return RetryWithBackoff(...)
}
```

### Exemplo 3: Schema Validation

```go
// Kafka valida schema
err := ValidateSchema(payload, "transaction.v1")
if err != nil {
    slog.Error("Schema inválido, movendo para DLQ",
        "schema_version", "transaction.v1",
        "error", err,
    )
    // Permanente: não vai melhorar
    return SaveToFailedEvents(payload, err)
}
```

---

## 7. Métricas Recomendadas

Para cada erro, incrementar métricas:

```go
// Erro transitório → retry
errorTransientCounter.WithLabelValues("kafka_timeout").Inc()
retryCounter.WithLabelValues("kafka_timeout").Inc()

// Erro permanente → DLQ
errorPermanentCounter.WithLabelValues("validation_failed").Inc()
dlqCounter.WithLabelValues("validation_failed").Inc()

// Alerta
alertCounter.WithLabelValues("dlq_growth").Inc()
```

---

## 8. Referências

- `playbooks/dlq-playbook.md` — Reprocessamento de DLQ
- `reference/operational-compliance-policy.md` § 4) Retry/Outbox/DLQ
- `system-design.md` — Arquitetura e padrões

---

**Versão:** 1.0  
**Última atualização:** 2026-06-24  
**Proprietário:** Engenharia / Architecture Guild
