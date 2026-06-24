# Dev Team — Guia de Implementação, Runbooks e CI

Status: Ativo
Proprietário: Engineering Tech Lead
Última atualização: 2026-06-23

Objetivo

Este documento é o manual do desenvolvedor e operador: como implementar, testar, executar e operar os componentes do ledger.

1. Pré-requisitos e Ambiente

Para desenvolver e executar o projeto localmente, você precisará de:

- **Go 1.26.4** ou superior.
- **Docker e Docker Compose** para subir a infraestrutura (Postgres, Kafka, Redis).
- **Git** para controle de versão.

### Subindo a Infraestrutura Local

```bash
docker-compose up -d
```
Isso iniciará as dependências necessárias listadas no `docker-compose.yaml`.

2. Fluxo de trabalho do desenvolvedor

- Use os módulos do monorepo: `transaction-gw`, `ledger-core`, `notification-service`, `ledger-reconciler`, `ledger-backoffice`.
- Gere ou aceite `idempotency_key` no gateway; persista-a em `transactions` com restrição UNIQUE.
- Persista as linhas de `transactions`, `ledger_entries` e `outbox` dentro de uma única transação de banco de dados (escrita atômica).
- O worker de Outbox publica eventos e marca `outbox` como `PUBLISHED` ou `FAILED`.

## Checklist de implementação (por componente)

- `transaction-gw`: validar idempotência, tratamento de headers, rejeição rápida (Proprietário: Gateway Team)
- `ledger-core`: garantir atomicidade da transação DB + linha de outbox, concorrência otimista para saldos (Proprietário: Ledger Team)
- `outbox-worker`: implementar retries, publicação em DLQ, métricas (Proprietário: Plataforma)
- `ledger-reconciler`: validação em blocos (chunked), relatórios, orientação de remediação (Proprietário: Ledger Team)
- `notification-service`: publicação segura em DLQ e rate limiting (Proprietário: Notification Team)

3. Esquemas e testes de contrato

- Armazene esquemas em `/schema/json/...` ou use Avro/Protobuf com um registro.
- Adicione testes de contrato ao CI que validem a compatibilidade entre produtor e consumidor.
- Requisitos do CI **atualmente ativos**: testes unitários (`go test ./...` em todos os módulos), validação de sintaxe de JSON schemas e workflow de link-check em markdown.
- **Roadmap (não implementado):** testes de integração com `testcontainers`, lint de esquemas contra um schema registry.

4. Comandos locais de dev e teste

### Executar testes unitários (todos os módulos)

```bash
# De uma vez (usando go.work)
go test ./...

# Por aplicativo
cd apps/transaction-gw && go test ./...
cd apps/ledger-core && go test ./...
cd apps/ledger-reconciler && go test ./...
cd apps/notification-service && go test ./...
cd apps/rate-limiter && go test ./...
cd apps/ledger-backoffice && go test ./...
cd shared && go test ./...
```

### Executar testes de integração com testcontainers (requer docker)
O pipeline de CI define os comandos, mas você pode rodar localmente garantindo que o Docker esteja ativo.

5. Módulo `shared`

O diretório `shared` contém código compartilhado entre todos os aplicativos, incluindo utilitários de configuração, logging, saúde e clientes de infraestrutura.

- [Guia do Módulo Shared](reference/shared-module.md)

6. Observabilidade e Alertas

- Métricas para instrumentar: `transactions_processed_total`, `transactions_processed_errors_total`, `transaction_processing_duration_seconds`, `outbox_pending`, `dlq_messages_total`, `consumer_lag{topic,partition}`, `reconciler_discrepancies_total`.
- Tracing: propagar `trace_id` e `idempotency_key` em headers e logs.
- Alertas (exemplos): `LedgerConsumerHighPartitionLag`, `LedgerDlqGrowth`, `LedgerOutboxPendingHigh`, `LedgerReconciliationDiscrepancy`, `LedgerTransactionErrorRate`.

7. Runbooks (resumo)

- Reprocessamento de DLQ: siga `docs/playbooks/dlq-playbook.md`. Regras principais:
  - Reprocessar em pequenos lotes com `FOR UPDATE SKIP LOCKED`.
  - Anexar metadados `reprocessed_by` e registrar ações no sistema de tickets.
  - Parar se a taxa de erro dobrar ou se for observado impacto a jusante (downstream).

- Lag de consumidor / partições quentes (hot partitions): monitore o lag por partição, dimensione os consumidores, considere o sharding de contas quentes em subcontas ou o roteamento para partições alternativas.

- Reconciliador: execute a validação em blocos, classifique as discrepâncias (transitórias/operacionais/bug) e siga o playbook de mitigação.

8. Checklist de Migração e Rollout

- Criar ticket de migração com plano de snapshot de dados.
- Implantar esquema/DDL em staging e executar testes de migração.
- Implantar workers e habilitar feature flags para rollout faseado.
- Executar canary e monitorar métricas por 30-60 minutos antes de promover.

9. Segurança e segredos

- Use Vault ou KMS na nuvem para credenciais de DB e Kafka.
- Rotacione credenciais e registre eventos de acesso para auditoria.

10. Scripts e automação

- `scripts/prepare_pr.sh` cria uma branch e abre um PR via `gh`, se disponível.
- Adicione scripts de auxílio em `scripts/` para tarefas operacionais comuns (exportar failed_events, pequenos replays).

11. Contatos e escalonamento

- Ledger Team: proprietário do ledger-core
- Plataforma/SRE: proprietário do outbox worker, alertas e infra
- API Guild: proprietário de esquemas e contratos
- Produto/Compliance: proprietário de decisões de retenção e UX

Referências

- `business.md` — produto e SLAs
- `system-design.md` — DDL, contratos, política de retry e parâmetros operacionais
- `docs/playbooks/dlq-playbook.md` — runbook detalhado de DLQ
- `docs/reference/` — guias de referência técnica (idempotência, observabilidade, contratos)
