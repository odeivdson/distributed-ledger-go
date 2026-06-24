# Runbook: Outbox & DLQ — Reprocessamento e Operação

Este runbook descreve passos operacionais para lidar com mensagens pendentes em `outbox` e itens em DLQ (`failed_events` / tópicos de DLQ).

**Status do documento**

- **Status:** Pronto para revisão
- **Proprietário:** SRE/OPS
- **Última atualização:** 2026-06-23

## Objetivos

- garantir entrega eventual de eventos via Outbox
- permitir reprocessamento seguro de mensagens que falharam
- fornecer playbook para operação e recuperação

## Pré-requisitos

- Acesso ao banco de dados (leitura/escrita limitada)
- Acesso controlado ao banco de dados (`psql`) com role de leitura/escrita para ops.
- Acesso ao cluster Kafka (`kafka-console-*` ou admin client).
- Variáveis de ambiente: `DATABASE_URL`, `KAFKA_BOOTSTRAP_SERVERS` configuradas na sessão.

Estados relevantes

- `outbox.status`: `PENDING`, `PUBLISHED`, `FAILED`.
- `failed_events`: registros com `payload`, `error`, `attempts`, `first_error_at`, `last_error_at`.

Detecção inicial

1) Alertas a monitorar (Grafana / Prometheus):
   - Aumento de `outbox_pending` acima da linha de base (baseline)
   - `dlq_messages_total` crescente
   - Disparo de `LedgerOutboxPendingHigh` ou `LedgerDlqGrowth`

2) Primeiro passo:
   - Verificar logs do worker de outbox e do serviço correspondente (ex: `ledger-core` ou `outbox-worker`).
   - Executar (psql):

```bash
psql "$DATABASE_URL" -c "SELECT count(*) FROM outbox WHERE status='PENDING';"
psql "$DATABASE_URL" -c "SELECT id, aggregate_type, event_type, attempts, created_at, last_error FROM outbox WHERE status='FAILED' ORDER BY created_at LIMIT 20;"
psql "$DATABASE_URL" -c "SELECT id, payload, error, attempts, first_error_at FROM failed_events ORDER BY first_error_at DESC LIMIT 20;"
```

Classificação de erros

- Transitório (infra/timeout): favorecer reprocessamento automático/manual com retries.
- Permanente (validação de negócio/esquema): requer intervenção humana e correção do payload/serviço.

Reprocessamento seguro — princípios

- Reprocessar em pequenos lotes (ex: 50-200 mensagens) com `FOR UPDATE SKIP LOCKED`.
- Validar contadores antes/depois e comparar métricas (taxa de sucesso, crescimento da DLQ).
- Anexar metadados `reprocessed_by`, `reprocessed_at`, `reprocess_reason` ao reprocessamento.

Procedimento passo a passo (reprocessar outbox PENDING)

1. Selecionar lote para reprocessar (LOCK):

```bash
psql "$DATABASE_URL" -c "BEGIN; SELECT id, event_type FROM outbox WHERE status='PENDING' ORDER BY created_at LIMIT 100 FOR UPDATE SKIP LOCKED;"
```

2. Republicar via worker interno ou script (exemplo simples usando kafka-console-producer):

```bash
# Exportar payloads para arquivos e publicar um a um usando --sync para garantir entrega visível
# (Use o tópico correto, por exemplo, "notifications" para eventos de outbox do ledger-core)
psql "$DATABASE_URL" -At -c "SELECT id, payload::text FROM outbox WHERE status='PENDING' ORDER BY created_at LIMIT 100" | while IFS='|' read -r id payload; do
  echo "$payload" | kafka-console-producer --broker-list "$KAFKA_BOOTSTRAP_SERVERS" --topic notifications
  # Se publicação OK, marcar como publicado
  psql "$DATABASE_URL" -c "UPDATE outbox SET status='PUBLISHED', attempts = attempts + 1, published_at = now() WHERE id = '$id';"
done
```

3. Validar: confirmar redução de `outbox_pending` e ausência de aumento de `failed_events`.

Reprocessamento de outbox FAILED (após inspeção)

1. Inspecionar `last_error` e `attempts`:

```bash
psql "$DATABASE_URL" -c "SELECT id, last_error, attempts FROM outbox WHERE status='FAILED' ORDER BY created_at LIMIT 50;"
```

2. Corrigir a causa (esquema, campos estranhos, infra). Se seguro, resetar o estado e reagendar (re-schedule):

```bash
psql "$DATABASE_URL" -c "UPDATE outbox SET status='PENDING', attempts = 0 WHERE id = '<id>' RETURNING id;"
```

3. Monitorar processo de republicação e verificar se `attempts` cresce e `status` torna-se `PUBLISHED`.

DLQ (`failed_events`) — reprocessamento manual

1. Listar itens recentes na DLQ:

```bash
psql "$DATABASE_URL" -c "SELECT id, source_topic, payload::text, error, attempts, metadata FROM failed_events ORDER BY first_error_at DESC LIMIT 50;"
```

2. Para erros transitórios, republicar o payload para o tópico original usando `kafka-console-producer`:

```bash
# Obtenha o source_topic da listagem anterior e publique no respectivo tópico
echo '<payload-json>' | kafka-console-producer --broker-list $KAFKA_BOOTSTRAP_SERVERS --topic notifications
```

3. Ao reprocessar, registrar `reprocessed_by` (login do operador) e `reprocessed_at` em `failed_events.metadata`.

Critérios de parada / rollback

- Se ao reprocessar houver aumento de erros em 2x em relação à baseline ou impacto em consumidores a jusante (downstream), parar o reprocessamento.
- Abrir incidente e escalar para Engenharia + Produto.

Escalonamento

1. Nível 1 (Ops): inspeciona e tenta reprocessamento de lote pequeno (small-batch).
2. Nível 2 (SRE/Plataforma): se o problema persistir ou afetar o throughput, executar investigação de infra (Kafka, rede, banco de dados).
3. Nível 3 (Engenharia): falhas de esquema, validação de negócio, lógica do produtor.

Checklist de pré-reprocessamento

- Confirmar proprietário (owner) e janela de manutenção (se necessário).
- Tirar snapshot de `outbox` e `failed_events` (exportar CSV/JSON).
- Validar ambiente de replay (staging vs prod).

Observabilidade e validações pós-reprocessamento

- Verificar métricas: `outbox_pending`, `dlq_messages_total`, `transactions_processed_errors_total`.
- Revisar logs correlacionando `trace_id` e `transaction_id`.

Segurança

- Use credenciais rotacionadas via Vault.
- Não exponha payloads sensíveis em logs públicos; aplique mascaramento (masking).

Comandos de apoio rápidos

```bash
# Listar pendentes
psql "$DATABASE_URL" -c "SELECT id, created_at, attempts FROM outbox WHERE status='PENDING' ORDER BY created_at LIMIT 100;"

# Marcar FAILED -> PENDING (reagendar)
psql "$DATABASE_URL" -c "UPDATE outbox SET status='PENDING', attempts = 0 WHERE status='FAILED' AND created_at < now() - interval '5 minutes' RETURNING id;"

# Exportar failed_events
psql "$DATABASE_URL" -c "COPY (SELECT id, payload::text, error FROM failed_events ORDER BY first_error_at DESC LIMIT 100) TO STDOUT WITH CSV" > failed_events_sample.csv
```

Notas finais

- Documentar todas as ações de reprocessamento no sistema de tickets e atualizar `reprocessed_by`/`reprocessed_at` no metadata.
- Este runbook deve ser referenciado a partir de `../reference/operational-compliance-policy.md`.
