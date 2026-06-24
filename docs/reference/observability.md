# Guia de Observabilidade — Resumo e Referência

Este documento agora serve como resumo de observabilidade e aponta para o documento central de políticas operacionais e conformidade.

**Status do documento**

- **Status:** Referência Ativa
- **Proprietário:** SRE / Observabilidade
- **Última atualização:** 2026-06-23

**Uso recomendado**

Consulte `./operational-compliance-policy.md` como fonte de verdade para métricas, alertas e dashboards do ledger.

## Resumo

- Métricas de transações, erros, latência, lag de consumers, outbox e DLQ são obrigatórias.
- Dashboards devem cobrir ingestão, consumers, resiliência e hot partitions.
- As regras de alerta devem ser testadas em staging antes de promover para produção.

## Referência central

- `./operational-compliance-policy.md`

## Tracing

- Propagar `trace_id` em headers de eventos.
- Spans recomendados:
  - HTTP ingress (gateway)
  - Persist transaction (DB transaction)
  - Outbox publish
  - Consumer processing
  - External provider call (notification)

## Dashboards e alertas (exemplos)

- Dashboard de Ingestão: taxa de transações, latência, erros
- Dashboard de Consumers: lag por partição, throughput, tempo de processamento
- Dashboard de Resiliência: número de retries, mensagens em DLQ, outbox pendentes

### Regras de alerta sugeridas (Prometheus)

- `LedgerConsumerHighPartitionLag`
  - Expression: `max by (topic, partition) (consumer_lag{topic="transactions"}) > 1000`
  - For: `5m`
  - Severity: `critical`
  - Description: "Lag do consumidor de transações alto por mais de 5 minutos."

- `LedgerDlqGrowth`
  - Expression: `increase(dlq_messages_total[5m]) > 0`
  - For: `10m`
  - Severity: `warning`
  - Description: "Mensagens chegando na DLQ, investigar falhas de processamento ou validação."

- `LedgerOutboxPendingHigh`
  - Expression: `outbox_pending > 100`
  - For: `5m`
  - Severity: `warning`
  - Description: "Número excessivo de mensagens pendentes no outbox."

- `LedgerReconciliationDiscrepancy`
  - Expression: `reconciler_discrepancies_total > 0`
  - For: `0m`
  - Severity: `critical`
  - Description: "Discrepância detectada na reconciliação do ledger."

- `LedgerTransactionErrorRate`
  - Expression: `rate(transactions_processed_errors_total[5m]) / rate(transactions_processed_total[5m]) > 0.01`
  - For: `10m`
  - Severity: `warning`
  - Description: "Taxa de erro de transações excedeu 1% nas últimas 10 minutos."

## Logs

- Logs estruturados com `transaction_id`, `idempotency_key`, `trace_id`, `account_id`.
- Separar logs de auditoria (lancamentos) dos logs de aplicação.

## Checklist de Instrumentação

- Instrumentar todos os serviços com OpenTelemetry
- Exportar métricas Prometheus
- Configurar dashboards e alertas no Grafana
- Definir thresholds (limiares) em staging antes de promover para produção
- Centralizar regras de alerta em um repositório de configuração de monitoramento

---

Arquivo de referência. Ajustar thresholds e alertas conforme ambiente e baselines (linhas de base).