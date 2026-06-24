# Contratos Técnicos — API, Eventos, Esquemas e DDL

Status: Esboço (Revisão pelo Staff)
Proprietário: Staff Engineering / API Guild
Última atualização: 2026-06-23

Objetivo: centralizar todos os contratos técnicos — formatos de API, eventos Kafka, DDL relevantes, versionamento e governança de contrato. Esta página é a fonte canônica para consumidores e produtores de eventos e para equipes que integram com o ledger.

Conteúdo
- Endpoints de API (gateway)
- Tópicos e esquemas de eventos
- Trechos principais de DDL (transactions, ledger_entries, outbox, failed_events)
- Guia de versionamento e registro de esquemas (schema registry)
- Governança de contratos e verificações de CI

---

## 1. API: Gateway de Transação (HTTP)

POST /transactions
- Corpo: TransactionRequest (JSON)
- Headers: `Idempotency-Key` (recomendado)
- Resposta: 202 Accepted (processando) ou 200 OK (concluído de forma síncrona)

TransactionRequest (JSON Schema)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "TransactionRequest",
  "type": "object",
  "required": ["source_account_id", "target_account_id", "amount", "idempotency_key"],
  "properties": {
    "source_account_id": {"type": "string", "format": "uuid"},
    "target_account_id": {"type": "string", "format": "uuid"},
    "amount": {"type": "integer", "minimum": 1},
    "idempotency_key": {"type": "string"},
    "description": {"type": "string"}
  }
}
```

Notas:
- `idempotency_key` deve ser fornecida para garantir a idempotência da transação.
- O gateway irá validar o request, verificar o rate limiting e publicar a transação no Kafka.

---

## 2. Eventos (Tópicos Kafka)

Regras gerais
- Use um registro de esquemas (Avro/Protobuf/JSON Schema) e imponha compatibilidade via CI.
- Adicione `schema_version` e `schema_name` nos headers das mensagens.
- Chaveie as mensagens por `source_account_id` para garantir a ordenação por conta.

Tópicos e exemplos de esquemas

1) Tópico: `transactions` (intenção)
- Objetivo: encaminhar a transação solicitada para o ledger-core para processamento
- Chave: `source_account_id`

Payload (JSON Schema):

```json
{
  "$id": "ledger.transactions.v1",
  "type": "object",
  "required": ["id", "source_account_id", "target_account_id", "amount", "idempotency_key", "created_at"],
  "properties": {
    "id": {"type": "string", "format": "uuid"},
    "source_account_id": {"type": "string", "format": "uuid"},
    "target_account_id": {"type": "string", "format": "uuid"},
    "amount": {"type": "integer", "minimum": 1},
    "idempotency_key": {"type": "string"},
    "description": {"type": "string"},
    "created_at": {"type": "string", "format": "date-time"}
  }
}
```

2) Tópico: `transaction_result` (resultado)
- Objetivo: publicar o status final para sistemas a jusante (downstream)
- Payload: idêntico ou similar, com status do processamento.

3) Tópico: `notifications`
- Objetivo: notificações enviadas ao notification-service.

```json
{
  "id": "not-123",
  "type": "transaction_processed",
  "target": "acc-A",
  "payload": "Transação de R$ 25.00 processada com sucesso."
}
```

Comportamento de DLQ
- Cada cliente de publicação deve implementar retries com backoff e mover para DLQ após `max_attempts`.
- Nomenclatura do tópico de DLQ: `<topic>-dlq` e a tabela `failed_events` deve ser populada para auditoria.

---

## 3. DDL Principal (referência)

Transações (Transactions)

```sql
CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY,
  idempotency_key VARCHAR(255) UNIQUE NOT NULL,
  status VARCHAR(20) NOT NULL,
  metadata JSONB,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

Lançamentos de Ledger (Ledger entries)

```sql
CREATE TABLE IF NOT EXISTS ledger_entries (
  id UUID PRIMARY KEY,
  transaction_id UUID NOT NULL REFERENCES transactions(id),
  account_id UUID NOT NULL,
  entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('DEBIT','CREDIT')),
  amount_in_cents BIGINT NOT NULL CHECK (amount_in_cents > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);
```

Outbox

```sql
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  aggregate_type VARCHAR(100) NOT NULL,
  aggregate_id UUID NULL,
  event_type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz NULL
);
```

Eventos com falha (DLQ)

```sql
CREATE TABLE IF NOT EXISTS failed_events (
  id UUID PRIMARY KEY,
  source_topic VARCHAR(255),
  payload JSONB NOT NULL,
  error TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  first_error_at timestamptz NOT NULL DEFAULT now(),
  last_error_at timestamptz NOT NULL DEFAULT now(),
  metadata JSONB NULL
);
```

Índices: garantir `transactions(idempotency_key)` e `outbox(status, created_at)`.

---

## 4. Versionamento e Compatibilidade

- Siga o versionamento semântico para esquemas: `v1`, `v2`, etc.
- Mudanças compatíveis com versões anteriores são permitidas (campos aditivos). Mudanças que quebram a compatibilidade (breaking changes) requerem implantação e migração coordenadas.
- Use o registro de esquemas para validar a compatibilidade e bloquear mudanças incompatíveis no CI.
- Adicione `schema_version` ao header da mensagem para auxiliar os consumidores na migração.

Política de compatibilidade
- Compatibilidade `BACKWARD` para produtores é preferível quando há muitos consumidores.
- Para breaking changes, abra uma RFC, agende um lançamento coordenado e mantenha um shim de compatibilidade, se necessário.

---

## 5. Governança de Contratos

- Todas as mudanças de contrato devem incluir:
  - diff do esquema
  - plano de migração
  - plano de atualização de consumidores
  - portão (gate) de CI que valida o novo esquema contra o registro
- Proprietários: cada tópico/esquema deve ter uma equipe proprietária listada no registro.
- Testes: adicione testes de contrato ao CI que garantam a compatibilidade entre produtor e consumidor (usando um mock do registro ou um harness de teste de contrato).

---

## 6. Exemplos e Snippets de Contrato

- Artefatos de JSON Schema devem residir em `/schema/json/transactions/` com versões.
- Considere Avro/Protobuf para esquemas binários e force via registro.
- Pseudocódigo de produtor de exemplo e validação de consumidor incluídos nos repositórios das equipes `transaction-gw` e `ledger-core`.

---

## 7. Onde encontrar mais detalhes
- Design do sistema: [../system-design.md](../system-design.md)
- Política operacional (idempotência/retry/DLQ): [./operational-compliance-policy.md](./operational-compliance-policy.md)
- Runbooks: [../playbooks/dlq-playbook.md](../playbooks/dlq-playbook.md) and [../playbooks/operations-runbooks.md](../playbooks/operations-runbooks.md)

---

Apêndice: checklist para alteração de contrato
- [ ] Criar diff do esquema
- [ ] Validar compatibilidade com o registro (registry)
- [ ] Adicionar testes de contrato no CI
- [ ] Notificar consumidores e programar janela de migração
- [ ] Atualizar documentação e changelog
