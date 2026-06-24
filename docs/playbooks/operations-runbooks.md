# Runbooks Operacionais — Consumer Lag, Hot Partitions, Reconciliador, Backup & Restore

Status: Pronto para revisão
Proprietário: SRE / Ops
Última atualização: 2026-06-23

Este arquivo reúne runbooks operacionais complementares usados por Ops/SRE/Eng para incidentes comuns.
Referencie `../reference/operational-compliance-policy.md` e `dlq-playbook.md` para políticas e DLQ.

---

## 1. Lag de Consumidor / Partição Quente (Consumer Lag / Hot Partition)

Objetivo: detectar, diagnosticar e mitigar lag elevado em consumidores Kafka e pontos de calor (hotspots) por account_id.

### 1.1 Sinais de alerta
- Métrica: `consumer_lag{topic="transactions"}` alta em uma partição por > 5m
- Aumento de `transaction_processing_duration_seconds`
- Backlog crescente no outbox / DLQ associado

### 1.2 Procedimento rápido
1. Identificar partição(ões) com lag:

```bash
# Prometheus / Grafana: verificar painel de lag por partição
# Exemplo kafka-consumer-group (local) para checar lag
kafka-consumer-groups --bootstrap-server $KAFKA_BOOTSTRAP_SERVERS --describe --group ledger-core-group
```

2. Mapear partições para `account_id` (se possível): verificar nomeação/chaves nas mensagens ou usar código do produtor para identificar a chave.
3. Verificar se a conta é uma "super conta" (super account) ou um caso legítimo de aumento de volume.

### 1.3 Mitigações imediatas
- Se o tópico com lag for causado por consumidores lentos:
  - reiniciar o pod do consumidor (canary) ou aumentar as réplicas dos consumidores (se forem stateless e particionados).
  - aumentar o paralelismo do worker (se for seguro, preservando a ordenação por chave).
- Para partição quente por conta única:
  - ativar o throttling por conta (rate limiter) para reduzir a ingestão temporariamente.
  - contatar o proprietário do Produto/Conta para coordenar janelas de backlog.
  - considerar o roteamento temporário para uma partição alternativa (se houver suporte no produtor).

### 1.4 Escalonamento
- Ops: mitigação (reiniciar, escalar, aplicar throttle)
- SRE: investigar infraestrutura (E/S, CPU, GC, rede)
- Eng: análise de arquitetura para padrão de sharding/subcontas

---

## 2. Reconciliador — Discrepância de saldo

Objetivo: validar e tratar discrepâncias detectadas entre a projeção de saldo e a soma dos lançamentos.

### 2.1 Detecção
- Métrica: `reconciler_discrepancies_total > 0`
- O disparo do alerta crítico inicia o runbook de investigação.

### 2.2 Investigação inicial
1. Isolar o conjunto de contas com discrepância.
2. Verificar logs e `trace_id` relacionados às transações nas janelas de cálculo.
3. Rodar query de validação (exemplo):

```sql
-- calcular saldo a partir dos lançamentos
SELECT account_id, SUM(CASE WHEN entry_type='CREDIT' THEN amount_in_cents ELSE -amount_in_cents END) AS calc_balance
FROM ledger_entries
WHERE created_at <= now() - interval '1 minute'
GROUP BY account_id
ORDER BY account_id;

-- comparar com accounts_balance
SELECT a.account_id, a.balance_in_cents, c.calc_balance
FROM accounts_balance a
JOIN ( /* subquery acima */ ) c USING (account_id)
WHERE a.balance_in_cents != c.calc_balance;
```

4. Classificar a discrepância: transitória (em voo/in-flight), operacional (falha na publicação do outbox) ou bug (persistente).

### 2.3 Tratamento
- Transitória: aguardar reconciliação automática (mensagens em voo), executar o reconciliador novamente.
- Operacional: investigar outbox/DLQ e reprocessar eventos faltantes conforme `dlq-playbook.md`.
- Bug: envolver engenharia, bloquear implantações relacionadas ao componente afetado e gerar correção via estorno ou transação corretiva com auditoria.

### 2.4 Comunicação
- Notificar Produto/Compliance com resumo e impacto financeiro estimado.
- Criar ticket com severidade e cronograma (timeline) de resolução.

---

## 3. Backup & Restore — Postgres

Objetivo: garantir backups regulares e validar procedimentos de restauração (restore) e reprocessamento do outbox.

### 3.1 Backup
- Regras: backups diários incrementais + full semanal (ajustar conforme política de retenção de dados)
- Ferramenta recomendada: `pg_dump` / arquivamento de WAL / snapshot gerenciado (conforme a infraestrutura)

Exemplo de snapshot (GCP/AWS/etc): use o snapshot do provedor ou `pg_basebackup`.

### 3.2 Validação de restauração (Staging)
1. Restaurar o dump no ambiente de staging.
2. Executar verificações de integridade (sanity checks): contagens básicas (`transactions`, `ledger_entries`), índices e restrições (constraints).
3. Executar teste de reprocessamento de outbox: iniciar o worker e validar se o outbox pendente é republicado com sucesso.

### 3.3 Procedimento de emergência
- Em caso de perda de dados: restaurar o último backup full + aplicar WAL até o ponto desejado.
- Notificar Compliance e Produto imediatamente.

---

## 4. Operações de emergência e comunicação

- Ao detectar risco material (discrepância financeira, perda de mensagens em massa):
  1. Declarar incidente (Severidade P0/P1) e abrir ponte de conferência (conference bridge).
  2. Registrar cronograma e ações no postmortem/ticket.
  3. Escalar conforme a matriz de responsabilidade (Ops -> SRE -> Eng -> Produto -> Compliance).

---

## 5. Links e referências
- `../reference/operational-compliance-policy.md`
- `dlq-playbook.md`
