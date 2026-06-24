# Negócios — Visão, Escopo, SLAs e Stakeholders

Status: Aprovado pelo Staff (leitores: Produto, Compliance, Finanças, Eng)
Proprietário: Produto + Staff Engineering
Última atualização: 2026-06-23

Objetivo

Este documento captura a fundamentação de negócios, decisões de escopo, SLAs, stakeholders e critérios de aceitação para o produto Distributed Ledger.

1. Visão do Produto

- O ledger é a única fonte de verdade para movimentações financeiras na plataforma.
- Deve garantir integridade financeira (partidas dobradas), auditabilidade, recuperabilidade e rastreabilidade.
- O escopo inicial suporta uma única moeda base; multi-moeda foi adiada para um épico futuro.

2. Principais Capacidades

- Transações atômicas produzindo lançamentos de débito/crédito equilibrados.
- Ingestão idempotente para requisições de clientes (via `idempotency_key`).
- Lançamentos de ledger persistentes apenas de acréscimo (append-only) e projeção de saldos.
- Padrão Outbox + DLQ para publicação durável de eventos com runbooks para reprocessamento seguro.
- Reconciliador para detectar e classificar discrepâncias entre lançamentos de ledger e projeções.

3. Stakeholders e Responsabilidades

- Produto: definir SLAs, retenção de dados e critérios de aceitação.
- Compliance / Controle Financeiro: aprovar cronogramas de retenção, requisitos de auditoria e tratamento de PII.
- SRE / Plataforma: proprietários de runbooks, alertas, prontidão operacional e exercícios de DR.
- Engenharia (Ledger Team): implementar lógica central do ledger, idempotência, atomicidade do outbox.
- API Guild: proprietária de esquemas de eventos e governança de compatibilidade.

4. SLAs e Critérios de Aceitação de Negócios

- SLA Funcional (disponibilidade): endpoints de serviço alcançáveis 99,99% (monitoramento + runbooks).
- SLA de Processamento: 99,9% das transações processadas em <2s em condições normais; exceções devem seguir o fluxo de reprocessamento documentado.
- SLA de Correção: invariantes contábeis mantidos (soma débitos == soma créditos) para transações processadas.
- SLA de Reconciliação: discrepâncias descobertas pelo reconciliador devem ser classificadas em até 24 horas e atribuídas a um proprietário.

5. Retenção / Compliance (pendente de confirmação)

- Sugerido (staff): TTL de `idempotency_key` = 30 dias; retenção de `failed_events` = 90 dias; retenção de arquivamento/lançamentos de ledger conforme regras legais.
- Ação: Produto + Compliance devem confirmar estes valores; atualizar `reference/operational-compliance-policy.md` e registros legais.

6. Decisões de Escopo e Roadmap

- Multi-moeda: adiado; planejar um épico futuro que inclua mudanças de esquema, mitigação de sharding e UX do produto.
- Mitigação de hot-account: o produto concorda em permitir sharding/subcontas e throttling adaptativo para contas muito grandes.

7. Checklist de aceitação para lançamentos

- Esquemas de contrato publicados em `/schema/` e passíveis no CI.
- Exercício de DLQ/runbook executado em staging e documentado.
- Instrumentação (métricas/tracing) validada em staging.
- Aprovação do Produto sobre retenção e UX para os estados `PENDING`/`COMPLETED`/`FAILED`.

8. Regras de Validação de Negócios e Cenários

- **Partidas Dobradas:** cada transação deve resultar em lançamentos de débito e crédito equilibrados (soma débitos == soma créditos).
- **Immutability:** lançamentos históricos nunca são modificados ou excluídos. Correções são feitas via transações de compensação (estornos).
- **Idempotência:** requisições repetidas com a mesma `idempotency_key` não devem produzir efeitos financeiros duplicados.
- **Validação:**
    - Valores zero ou negativos são rejeitados.
    - Contas de origem e destino devem ser válidas e ativas.
    - Saldo suficiente deve ser verificado para operações de débito quando exigido pela política.
- **Cenários:**
    - *Transferência Interna:* Débito A, Crédito B, atualiza ambos os saldos.
    - *Estorno:* Cria uma nova transação que inverte o débito/crédito da original, mantendo uma referência a ela.
    - *Requisição Duplicada:* Identifica a chave e retorna o resultado anterior sem reprocessar.

Referências

- Veja `system-design.md` para detalhes técnicos e `dev-team.md` para runbooks operacionais e fluxo de trabalho do desenvolvedor.
