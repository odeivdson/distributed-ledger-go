# Idempotência Operacional — Guia de Implementação

**Status:** Referência Ativa (Consolidado)
**Proprietário:** Engenharia
**Última atualização:** 2026-06-24

👉 **Para política completa:** Consulte [operational-compliance-policy.md](./operational-compliance-policy.md) § 3 (fonte de verdade)

---

## Resumo Executivo

- `idempotency_key` é **obrigatória** para todas as requisições externas de transação
- Armazenar em coluna `VARCHAR(255) NOT NULL UNIQUE` na tabela `transactions`
- `transactions`, `ledger_entries` e `outbox` em **uma única transação de banco**
- TTL mínimo: **30 dias** após conclusão
- Estados: `PENDING`, `COMPLETED`, `FAILED`

---

### Fluxo de Implementação

1. **Cliente envia** transação com header `X-Idempotency-Key: <uuid>`
2. **Gateway valida** payload e tenta `INSERT` em `transactions` com `idempotency_key`
3. **Se conflito UNIQUE:** gateway consulta registro existente e retorna seu status
4. **Se sucesso:** gateway cria `ledger_entries` + `outbox` row na **mesma transação**
5. **Outbox worker** publica evento em Kafka e marca como `PUBLISHED`
6. **Duplicate request:** retorna resultado persistido **sem aplicar efeito novamente**

---

## Exemplo de SQL

### INSERT com Detecção de Duplicata

```sql
-- Tenta inserir; retorna nada se já existe
INSERT INTO transactions (id, idempotency_key, status, metadata)
VALUES ($1, $2, 'PENDING', $3)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING id, status, created_at;
```

Se não retornar linha (conflito detectado):

```sql
-- Consulta registro existente
SELECT id, status, created_at, updated_at
FROM transactions
WHERE idempotency_key = $1;
```

---

## UX & API Responses

- Requisição duplicada deve retornar `200 OK` ou `202 Accepted`.
- O corpo deve incluir:
  - `transaction_id`
  - `idempotency_key`
  - `status`
  - `created_at`
  - `updated_at`
  - `result` ou `reason` quando disponível.
- Para `PENDING`, retornar `202 Accepted` com `operation_id` ou `transaction_id`.
- Para `COMPLETED`, retornar o resultado final já calculado.
- Para `FAILED`, incluir informações de retry quando aplicável.

## 7. Erros e retry

### Métricas Recomendadas
### Métricas Recomendadas

- `idempotent_requests_total` — requisições com idempotency_key
- `duplicate_requests_total` — requisições duplicadas detectadas
- `idempotency_conflicts_total` — conflitos UNIQUE em inserts

Incluir `idempotency_key` em logs e propagar `trace_id` em todos os eventos.

---

## Referências Relacionadas

- [operational-compliance-policy.md](./operational-compliance-policy.md) § 3 — Política completa
- [technical-contracts.md](./technical-contracts.md) — Como usar header `X-Idempotency-Key`
- [system-design.md](../system-design.md) — DDL e arquitetura do banco

---

**Documento consolidado:** 2026-06-24  
**Próxima revisão:** 2026-08-01