# Política Operacional e de Conformidade — Distributed Ledger

Este documento consolida as políticas operacionais e de conformidade (compliance) do ledger, evitando dispersão em múltiplos guias.
Use-o como referência principal para idempotência, retry/DLQ, observabilidade, auditabilidade, retenção e governança.

## 1. Objetivo

Definir políticas unificadas que garantam:
- operação segura e previsível do ledger;
- conformidade com requisitos de auditoria e retenção;
- resposta clara a incidentes e falhas;
- preservação da integridade financeira e da idempotência.

## 2. Escopo

Este documento cobre:
- políticas de idempotência e entrega garantida;
- retry, outbox, DLQ e reprocessamento;
- observabilidade e alertas operacionais;
- segurança, compliance de dados e retenção;
- papéis e responsabilidades.

## 3. Idempotência

### 3.1 Regra básica

- `idempotency_key` é obrigatória para todas as requisições externas que iniciam transações.
- Se o cliente não fornecer a chave, o gateway deve gerar uma por operação lógica e retorná-la.
- A chave deve ser estável para a mesma intenção de negócio e única por intenção de transação.

### 3.2 Persistência

- A tabela `transactions` deve incluir `idempotency_key VARCHAR(255) NOT NULL UNIQUE`.
- Gravar `transactions`, `ledger_entries` e `outbox` na mesma transação de banco de dados.
- O registro deve ser criado inicialmente com `status = 'PENDING'`.

### 3.3 Estados e comportamentos

- `PENDING`: transação aceita e em processamento.
- `COMPLETED`: transação processada com sucesso.
- `FAILED`: transação processada e falhou.

Regras de duplicata:
- se existir registro `COMPLETED`, retornar o resultado persistido;
- se existir registro `PENDING`, informar que a operação ainda está em processamento;
- se existir registro `FAILED`, permitir retry seguro com a mesma `idempotency_key` caso o erro seja transitório.

### 3.4 TTL e retenção

- A `idempotency_key` deve ser mantida por pelo menos **30 dias** após a conclusão da transação.
- O TTL pode ser ajustado conforme compliance, mas não deve ser inferior ao período de replay esperado.
- A limpeza de chaves expiradas não deve excluir histórico de transações nem auditoria.

### 3.5 Controles duplos

- Nível 1: gateway valida e registra `idempotency_key` no ponto de entrada.
- Nível 2: `ledger-core` verifica `idempotency_key` novamente antes de aplicar efeitos.
- Isso protege contra replays síncronos e assíncronos.

### 3.6 UX

- Duplicatas devem retornar `200 OK` ou `202 Accepted`.
- O corpo deve incluir:
  - `transaction_id`
  - `idempotency_key`
  - `status`
  - `created_at`
  - `updated_at`
  - `result` ou `reason` quando disponível.
- Para `PENDING`, retornar `202 Accepted` com `operation_id` ou `transaction_id`.

## 4. Retry, Outbox e DLQ

### 4.1 Padrões

- Usar outbox local para publicar eventos downstream após commit da transação.
- Outbox states: `PENDING`, `PUBLISHED`, `FAILED`.
- Persistir outbox em mesma transação que `transactions` e `ledger_entries`.

### 4.2 Parâmetros de retry

- `max_attempts = 5`
- `initial_backoff = 500ms`
- `backoff_multiplier = 2`
- `max_backoff = 30s`
- jitter randômico para evitar sincronização de retries.

### 4.3 DLQ

- Após `max_attempts`, mover mensagem para DLQ e registrar em `failed_events`.
- `failed_events` deve reter dados por no mínimo **90 dias**.
- O DLQ deve ser acompanhado por runbook de reprocessamento ou investigação.

### 4.4 Reprocessamento

- Reprocessar em small-batches e validar contadores de sucesso.
- Marcar `reprocessed_by` e `reprocessed_at` no metadata ao re-executar.
- Evitar reprocessar massas sem análise prévia.

## 5. Observabilidade e alertas

### 5.1 Métricas essenciais

- `transactions_processed_total`
- `transactions_processed_errors_total`
- `transaction_processing_duration_seconds`
- `consumer_lag{topic,partition}`
- `outbox_pending`
- `dlq_messages_total`
- `reconciler_discrepancies_total`
- `hot_partition_lag`
- `transaction_rate_by_account`

### 5.2 Alertas básicos

- `LedgerConsumerHighPartitionLag`: lag > 1000 em 5m
- `LedgerDlqGrowth`: aumento em DLQ por 10m
- `LedgerOutboxPendingHigh`: outbox pendente > 100 por 5m
- `LedgerReconciliationDiscrepancy`: discrepância detectada
- `LedgerTransactionErrorRate`: erro > 1% em 10m

### 5.3 Dashboards

- Ingestão: taxa, latência, erros
- Consumers: lag, throughput, tempo de processamento
- Resiliência: retries, DLQ, outbox
- Hot partitions: lag por partição e por conta

## 6. Segurança, compliance e retenção

### 6.1 Auditabilidade

- `ledger_entries` deve ser append-only.
- Correções somente por transações de estorno ou compensação.
- Registrar metadata de auditoria em todas as transações.

### 6.2 Retenção de dados

- `transactions`, `ledger_entries`, `accounts_balance`: retenção conforme regra contábil vigente.
- `failed_events`: mínimo **90 dias**.
- `outbox`: mínimo **7 dias**.
- `idempotency_key`: mínimo **30 dias**.
 
> Confirmação pendente: o staff recomenda TTL de `idempotency_key` = **30 dias** e retenção de `failed_events` = **90 dias**. Produto/Compliance devem confirmar ou propor ajustes. Data de confirmação proposta: 2026-07-07.
- Logs de auditoria: mínimo **1 ano** ou conforme regulamentação.

### 6.3 Proteção de dados sensíveis

- Classificar dados em financeiro, PII e operacional.
- Mascarar/anonymizar PII em logs e dashboards.
- Controlar acesso aos dados sensíveis.

### 6.4 Controle de acesso

- Acesso a produção deve ser auditado.
- Segregar roles para leitura de auditoria e reprocessamento.

### 6.5 Governança de schema

- Usar schema registry ou validação de contrato para eventos e payloads.
- Versionar todo schema antes de promover mudanças.
- Confirmar compliance e auditoria para alterações de dados financeiros.

## 7. Papéis e responsabilidades

- **Product:** valida SLAs, retenção e UX de idempotência.
- **Engineering:** implementa idempotência, outbox, observabilidade e políticas.
- **SRE/Observability:** monitora métricas, alertas e incidentes.
- **Compliance:** valida retenção, PII e auditoria.

## 8. Referência

- `./idempotency-guide.md`: detalhes de idempotência operacionais e exemplos de SQL.
- `./observability.md`: métricas e alertas de monitoramento.
- `../playbooks/dlq-playbook.md`: runbook de reprocessamento.
- `../system-design.md`: arquitetura de alto nível e padrões técnicos.
