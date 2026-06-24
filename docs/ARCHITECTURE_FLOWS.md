# Architecture Flows — Diagramas de Fluxo End-to-End

**Propósito:** Visualizar o fluxo completo de uma transação através dos componentes do sistema, incluindo caminhos de sucesso, erro e recuperação.

**Última atualização:** 2026-06-24

---

## 1. Fluxo de Transação Bem-Sucedida (Happy Path)

Este é o cenário esperado quando uma transação é processada sem erros.

```mermaid
sequenceDiagram
    participant Client
    participant GW as transaction-gw<br/>(Gateway)
    participant Kafka
    participant Core as ledger-core<br/>(Processador)
    participant DB as PostgreSQL
    participant Worker as OutboxWorker
    participant Notif as notification-service
    participant External as External Webhook

    Client->>GW: POST /transactions<br/>(source, target, amount,<br/>idempotency_key)
    
    GW->>GW: Validar payload
    GW->>GW: Gerar transaction_id
    
    GW->>DB: BEGIN transaction
    GW->>DB: INSERT transactions<br/>(status='PENDING')
    GW->>DB: INSERT outbox<br/>(event_type='transaction.created')
    GW->>DB: COMMIT transaction
    
    GW->>Kafka: PUBLISH transactions topic<br/>(transaction event)
    
    GW-->>Client: 202 Accepted<br/>(transaction_id, status='PENDING')
    
    Kafka->>Core: CONSUME transaction event
    
    Core->>Core: Validar regras de negócio
    Core->>Core: Calcular débito/crédito
    
    Core->>DB: BEGIN transaction
    Core->>DB: INSERT ledger_entries<br/>(DEBIT source account)<br/>(CREDIT target account)
    Core->>DB: UPDATE accounts_balance<br/>(com version check - OCC)
    Core->>DB: INSERT outbox<br/>(event_type='transaction.completed')
    Core->>DB: UPDATE transactions<br/>(status='COMPLETED')
    Core->>DB: COMMIT transaction
    
    Core->>Kafka: ACKNOWLEDGE message offset
    
    Worker->>DB: SELECT outbox<br/>WHERE status='PENDING'<br/>LIMIT 100
    
    Worker->>Kafka: PUBLISH notifications topic<br/>(notification event)
    
    Worker->>DB: UPDATE outbox<br/>SET status='PUBLISHED',<br/>published_at=NOW()
    
    Kafka->>Notif: CONSUME notification event
    
    Notif->>Notif: Parse notification<br/>(type, payload, recipient)
    
    Notif->>External: POST webhook<br/>(transaction data)
    External-->>Notif: 200 OK
    
    Notif->>Kafka: ACKNOWLEDGE message offset

    Note over Client,External: Transação concluída com sucesso!
```

**Tempo total esperado:** < 2 segundos (SLA)

**Verificação de sucesso:**
- ✅ `transactions.status` = `COMPLETED`
- ✅ `ledger_entries` contém 2 registros (débito + crédito balanceados)
- ✅ `accounts_balance` atualizado para ambas as contas
- ✅ `outbox.status` = `PUBLISHED`
- ✅ Cliente recebe webhook de notificação

---

## 2. Fluxo de Erro Transitório (Retry com Sucesso)

Quando um componente falha temporariamente mas se recupera após retry.

```mermaid
sequenceDiagram
    participant Client
    participant GW
    participant Kafka
    participant Core
    participant DB
    participant Worker
    participant External

    Client->>GW: POST /transactions
    GW->>DB: Inserir transaction + outbox
    GW->>Kafka: PUBLISH (1ª tentativa)
    GW-->>Client: 202 Accepted
    
    Note over Kafka: ⚠️ ERRO: Kafka cluster indisponível

    Worker->>DB: SELECT outbox (status='PENDING')
    Worker->>Kafka: PUBLISH notifications (1ª tentativa)
    Kafka-->>Worker: ❌ Connection timeout
    
    Worker->>DB: UPDATE outbox SET attempts=1,<br/>last_error='timeout'
    
    Worker->>Worker: Backoff: 500ms * 2^1 = 1s
    Note over Worker: Aguardando 1 segundo...
    
    Worker->>Kafka: PUBLISH notifications (2ª tentativa)
    Kafka-->>Worker: ✅ Success
    
    Worker->>DB: UPDATE outbox SET status='PUBLISHED'

    Note over Client,External: Recuperado após 1 retry!
```

**Parâmetros de retry:**
- `max_attempts` = 5
- `initial_backoff` = 500ms
- `backoff_multiplier` = 2
- `max_backoff` = 30s
- `jitter` = ±10% aleatório

**Quando isso acontece:**
- Kafka indisponível temporariamente
- Database connection pool esgotado
- Timeout de rede (network timeout)
- Rate limit temporário de webhook externo

---

## 3. Fluxo de Erro Permanente (Para DLQ)

Quando o erro persiste após max_attempts e a mensagem vai para Dead Letter Queue.

```mermaid
sequenceDiagram
    participant Core
    participant DB
    participant Worker
    participant Kafka
    participant DLQ as failed_events

    Core->>DB: INSERT invalid ledger_entries
    Core->>DB: INSERT outbox<br/>(event_type='transaction.invalid')
    
    Worker->>DB: SELECT outbox WHERE status='PENDING'
    
    loop 5 tentativas com backoff
        Worker->>Kafka: PUBLISH notifications
        Kafka-->>Worker: ❌ Schema validation failed
        Worker->>DB: UPDATE outbox SET attempts++
    end
    
    Worker->>Worker: attempts (5) >= max_attempts (5)?
    Note over Worker: SIM → Mover para DLQ
    
    Worker->>DB: UPDATE outbox SET status='FAILED'
    
    Worker->>DB: INSERT failed_events<br/>(id, source_topic, payload,<br/>error, attempts, metadata)
    
    Note over DLQ: Evento agora em DLQ<br/>Requer investigação manual

    Worker->>Kafka: Publicar alerta<br/>(LedgerDlqGrowth)

    Note over Worker,DLQ: SRE será notificado via Grafana
```

**Exemplo de erro permanente:**
- Schema mismatch (campo obrigatório faltando)
- Validação de negócio falha (saldo negativo)
- Conta de origem ou destino inválida
- Webhook externo retorna 410 Gone (cliente deletado)

**Operação manual necessária:**
- SRE investiga payload em `failed_events`
- Determina causa: bug no código vs dados ruins
- Se dados ruins: corrigir manualmente em metadata
- Se bug: fix, deploy, reprocessar

---

## 4. Fluxo de Conflito de Concorrência (OCC - Optimistic Concurrency Control)

Quando duas transações tentam atualizar a mesma conta simultaneamente.

```mermaid
sequenceDiagram
    participant Core1 as ledger-core<br/>(Transação A)
    participant Core2 as ledger-core<br/>(Transação B)
    participant DB

    par Processamento Paralelo
        Core1->>DB: SELECT accounts_balance<br/>WHERE account_id='ACC-001'<br/>(version=10, balance=1000)
        and
        Core2->>DB: SELECT accounts_balance<br/>WHERE account_id='ACC-001'<br/>(version=10, balance=1000)
    end
    
    Core1->>Core1: Calcular novo saldo: 1000 - 100 = 900
    Core2->>Core2: Calcular novo saldo: 1000 - 50 = 950
    
    par UPDATE com Version Check
        Core1->>DB: UPDATE accounts_balance<br/>SET balance=900, version=11<br/>WHERE account_id='ACC-001'<br/>AND version=10
        Note over DB: ✅ rows_affected=1
        and
        Core2->>DB: UPDATE accounts_balance<br/>SET balance=950, version=11<br/>WHERE account_id='ACC-001'<br/>AND version=10
        Note over DB: ❌ rows_affected=0<br/>(version mudou!)
    end
    
    Core1->>DB: COMMIT (sucesso)
    
    Core2->>Core2: ⚠️ Conflito detectado!
    Core2->>Core2: Incrementar backoff counter
    Core2->>DB: ROLLBACK
    
    Note over Core2: Retry com backoff exponencial
    
    Core2->>DB: SELECT accounts_balance<br/>(version=11, balance=900)
    Core2->>Core2: Recalcular: 900 - 50 = 850
    Core2->>DB: UPDATE com version=11
    Note over DB: ✅ rows_affected=1
    
    Core2->>DB: COMMIT (sucesso em retry)

    Note over Core1,Core2: Ambas transações processadas<br/>Saldo final: 850 (correto!)
```

**Garantia:** A version check (OCC) garante que não há "lost updates" — cada transação vê a versão correta da conta.

**Quando isso acontece:**
- Múltiplas transações na mesma conta
- Particularmente com hot accounts (contas muito ativas)
- Exemplo: Conta compartilhada recebendo múltiplos pagamentos

---

## 5. Fluxo de Reprocessamento de DLQ (Operacional)

Como um SRE/Ops reprocessa eventos que falharam.

```mermaid
sequenceDiagram
    participant SRE
    participant DB as PostgreSQL
    participant Kafka
    participant Worker
    participant Monitoring

    SRE->>Monitoring: Detectar LedgerDlqGrowth alert<br/>no Grafana

    SRE->>DB: SELECT FROM failed_events<br/>ORDER BY first_error_at DESC LIMIT 50

    SRE->>SRE: Analisar payloads e erros
    Note over SRE: Classificar:<br/>- Transitório (infra)<br/>- Permanente (validação)

    SRE->>SRE: Caso 1: Transitório<br/>(ex: timeout Kafka)
    SRE->>DB: UPDATE outbox<br/>SET status='PENDING', attempts=0<br/>WHERE id IN (select_ids)

    Worker->>DB: SELECT outbox (status='PENDING')
    Worker->>Kafka: PUBLISH

    alt Sucesso
        Worker->>DB: UPDATE status='PUBLISHED'
        Monitoring->>SRE: Métrica dlq_messages_total reduz
    else Falha novamente
        Worker->>DB: UPDATE attempts++
    end

    SRE->>SRE: Caso 2: Permanente<br/>(ex: schema mismatch)

    SRE->>DB: UPDATE failed_events<br/>SET metadata={'reprocessed_by': 'ops-alice',<br/>'reprocessed_at': now(),<br/>'reason': 'fixed invalid amount'}<br/>WHERE id='...'

    Note over SRE: Criar ticket de engenharia<br/>para fix do código

    SRE->>Monitoring: Validar: dlq_messages_total<br/>não crescendo mais
```

**Runbook associado:** [`playbooks/dlq-playbook.md`](playbooks/dlq-playbook.md)

---

## 6. Fluxo de Hot Partition (Estrangulamento)

Como o sistema detecta e mitiga uma "hot partition" (uma partição Kafka recebendo muito tráfego).

```mermaid
sequenceDiagram
    participant Client
    participant GW
    participant Monitoring
    participant RateLimiter
    participant Kafka
    participant Core

    Client->>Client: 🔥 Super Account<br/>recebendo 1000s trans/sec

    loop Cada transação
        Client->>GW: POST /transactions
        GW->>RateLimiter: Check rate<br/>(account_id='SUPER-ACC')
        alt Dentro do limite
            RateLimiter->>GW: ✅ Allow
            GW->>Kafka: PUBLISH (partition K=SUPER-ACC)
        else Excedeu limite
            RateLimiter->>GW: ❌ Reject (429 Too Many Requests)
            GW-->>Client: 429 Too Many Requests
        end
    end

    Monitoring->>Monitoring: Detectar<br/>consumer_lag{partition=N} > 1000<br/>por 5 minutos

    Monitoring->>Monitoring: Dispara alerta:<br/>LedgerConsumerHighPartitionLag

    Core->>Kafka: Consumer lag continua crescendo

    Note over GW: Ativar mitigation:

    GW->>GW: Aumentar backoff global<br/>por 2x (adaptive throttle)

    Client->>Client: Receber rejeições<br/>Retry com backoff

    GW->>Kafka: Lag diminui gradualmente

    Monitoring->>Monitoring: Alerta resolvido<br/>após lag < 100 por 5m
```

**Quando ativar:**
- Consumer lag > 1000 por > 5m
- Taxa de erro > 1% durante > 10m
- `transaction_processing_duration_seconds` p95 > 5s

**Estratégias de mitigação:**
1. **Rate Limiting** — Throttle cliente (429 Too Many Requests)
2. **Sharding** — Splittar super account em subcontas
3. **Escalar consumers** — Adicionar mais replicas (se stateless)
4. **Reroute** — Enviar para partição alternativa (se suportado)

---

## 7. Fluxo de Reconciliação (Auditoria em Lote)

Como o reconciliador detecta e classifica discrepâncias.

```mermaid
graph TD
    A["Reconciliador inicia<br/>a cada 6 horas"] -->|Cron job| B["Buscar todas as contas<br/>com Cursor Pagination"]
    B -->|Para cada lote| C["Calcular saldo a partir<br/>de ledger_entries"]
    C -->|SUM CASE WHEN DEBIT THEN+ ELSE-| D["Comparar com<br/>accounts_balance"]
    
    D -->|Saldo igual| E["✅ Conta OK"]
    D -->|Saldo diferente| F{"Classificar<br/>discrepância"}
    
    F -->|Transitória| G["Em voo:<br/>outbox pendente"]
    G -->|Aguardar próximo ciclo| H["Reprocessar outbox"]
    H --> E
    
    F -->|Operacional| I["Falha de publicação:<br/>evento perdido"]
    I --> J["DLQ: Reprocessar<br/>com runbook"]
    J --> E
    
    F -->|Bug| K["Persistente:<br/>lógica incorreta"]
    K --> L["🚨 Alerta crítico<br/>reconciler_discrepancies_total++"]
    L --> M["Criar ticket<br/>de Engenharia"]
    M --> N["Fix + Deploy +<br/>Transação corretiva"]
    N --> E
    
    E -->|Fim do ciclo| O["Publicar métricas<br/>reconciler_success_count"]
    O -->|Próximo ciclo| B

    style E fill:#90EE90
    style L fill:#FF6B6B
    style M fill:#FFD93D
```

**Tempos:**
- Cada ciclo: ~30 min (depende do volume)
- Detecção de discrepância: < 1h após acontecer
- Classificação: manual via SRE/Ops
- Resolução: 24h (SLA de negócio)

**Métricas:**
- `reconciler_cycles_total`
- `reconciler_discrepancies_total`
- `reconciler_cycle_duration_seconds`

---

## 8. Referência Rápida de Estados

```mermaid
graph LR
    A["PENDING<br/>(Aceito)"] -->|Processado com sucesso| B["COMPLETED<br/>(Finalizado)"]
    A -->|Validação falha| C["FAILED<br/>(Erro)"]
    C -->|Erro transitório| A
    C -->|Erro permanente| D["DLQ<br/>(Manual)"]
    D -->|SRE fixa| A
    D -->|SRE analisa| E["Ticket<br/>de Eng"]
```

**Estados por tabela:**

| Tabela | Status | Significado |
|--------|--------|-------------|
| `transactions` | PENDING | Em processamento |
| `transactions` | COMPLETED | Processada com sucesso |
| `transactions` | FAILED | Validação ou erro permanente |
| `outbox` | PENDING | Aguardando publicação no Kafka |
| `outbox` | PUBLISHED | Publicada com sucesso |
| `outbox` | FAILED | Excedeu max_attempts, movida para DLQ |
| `failed_events` | N/A | Evento permanentemente falhado |

---

## 9. Links para Documentação Detalhada

| Fluxo | Documentação |
|-------|--------------|
| Happy Path | `system-design.md` § 1) Arquitetura |
| Erro & DLQ | `playbooks/dlq-playbook.md` |
| OCC & Concorrência | `reference/faq.md` § 1) Consistência |
| Hot Partition | `playbooks/operations-runbooks.md` § 1) |
| Reconciliação | `reference/faq.md` § 3) Escalabilidade |
| Idempotência | `reference/operational-compliance-policy.md` § 3) |

---

## 10. Troubleshooting por Sintoma

| Sintoma | Provável Causa | Fluxo | Ação |
|---------|----------------|-------|------|
| Transação fica em PENDING | Processador não está rodando | 1 | Reiniciar ledger-core |
| Muitos eventos em DLQ | Bug no código ou schema inválido | 3 | Investigar failed_events |
| Consumer lag alto | Hot partition ou processador lento | 6 | Escalar consumers ou mitigar throttle |
| Discrepância de saldo | Outbox event perdido ou bug | 7 | Executar reconciliador |
| OCC conflict persistente | Muita contention na mesma conta | 4 | Considerar sharding de conta |

---

**Versão:** 1.0  
**Última atualização:** 2026-06-24  
**Próxima revisão:** 2026-08-01  
**Proprietário:** Staff Engineering / Architecture Guild
