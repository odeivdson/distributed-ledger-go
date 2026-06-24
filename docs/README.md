# 📚 Distributed Ledger — Documentação Consolidada & Interativa

**Bem-vindo!** Esta é a documentação central do projeto Distributed Ledger. Use o índice abaixo para encontrar rapidamente o que você precisa.

> **🎯 Primeiro acesso?** Comece com [QUICKSTART.md](QUICKSTART.md) ou escolha seu papel em [Documentação por Persona](#-documentação-por-persona).

---

## 🚀 Quick Start (5 passos - 30 min)

Inicie o projeto rapidamente:

👉 **[QUICKSTART.md](QUICKSTART.md)** — Passo a passo executável:
1. Clonar repositório
2. Subir Docker Compose
3. Fazer primeira transação
4. Entender o fluxo
5. Rodar testes

---

## 📊 Documentação por Persona

Escolha seu papel para um caminho customizado:

### 👨‍💻 **Engineer (Desenvolvedor)**
Você quer implementar features, corrigir bugs ou contribuir.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [QUICKSTART.md](QUICKSTART.md) | 30 min | Rodar projeto localmente |
| [STRUCTURE.md](STRUCTURE.md) | 20 min | Entender layout do monorepo |
| [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) | 20 min | Visualizar fluxos com diagramas |
| [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) | 30 min | Padrões de erro + exemplos Go |
| [system-design.md](system-design.md) | 40 min | Design técnico profundo |
| [reference/technical-contracts.md](reference/technical-contracts.md) | 20 min | APIs e eventos |

**Checklist para começar:**
- [ ] Rodou `docker-compose up -d`
- [ ] Fez primeira transação com curl
- [ ] Leu STRUCTURE.md e ARCHITECTURE_FLOWS.md
- [ ] Entendeu ERROR_HANDLING_PATTERNS.md

---

### 🚨 **SRE / Operações**
Você quer monitorar, responder a incidentes e fazer deploy.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [STRUCTURE.md](STRUCTURE.md) § 1-2 | 15 min | Componentes do sistema |
| [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) | 30 min | Políticas de operação |
| [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) | 25 min | Reprocessar eventos com falha |
| [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) | 30 min | Consumer lag, hot partitions |
| [reference/observability.md](reference/observability.md) | 20 min | Métricas e alertas |

**Checklist para responder incidentes:**
- [ ] Conhece todos os playbooks
- [ ] Consegue reprocessar DLQ
- [ ] Sabe escalar hot partition
- [ ] Tem acesso a Grafana/Prometheus

---

### 📊 **Product / Stakeholder**
Você quer entender o que o sistema faz e quando está pronto.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [business.md](business.md) | 15 min | Visão, escopo, SLAs |
| [STRUCTURE.md](STRUCTURE.md) § 1 | 10 min | Arquitetura de alto nível |
| [system-design.md](system-design.md) § 1 | 10 min | Design técnico simplificado |
| [reference/faq.md](reference/faq.md) | 20 min | Decisões arquiteturais |

**Checklist de lançamento:**
- [ ] Entendeu SLAs e critérios de aceite
- [ ] Aprovoupolítica de retenção (30d idempotency_key, 90d failed_events)
- [ ] Validou plano de rollout com equipe
- [ ] Acordou critério de sucesso da feature

---

### 🔌 **API Consumer (Cliente Externo)**
Você quer integrar com a API do ledger.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [reference/technical-contracts.md](reference/technical-contracts.md) | 15 min | Endpoints, payloads, exemplos |
| [QUICKSTART.md](QUICKSTART.md) § 3 | 5 min | Fazer primeira chamada |
| [reference/idempotency-guide.md](reference/idempotency-guide.md) | 10 min | Como usar idempotency_key |

**Checklist para integração:**
- [ ] Tem `idempotency_key` em todas as requisições
- [ ] Trata status 202 (pendente) vs 200 (completo)
- [ ] Implementou retry com backoff
- [ ] Validou payload contra JSON schema em `/schema/`

---

## 📖 Índice Completo

### 🏗️ Estrutura & Arquitetura

| Documento | Objetivo |
|-----------|----------|
| **[STRUCTURE.md](STRUCTURE.md)** | Layout completo, componentes, padrões |
| **[system-design.md](system-design.md)** | Design técnico, DDL, eventos |
| **[ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md)** | 8 diagramas Mermaid de fluxos |

### 🎯 Quick References

| Documento | Objetivo |
|-----------|----------|
| **[QUICKSTART.md](QUICKSTART.md)** | 5 passos para rodar localmente |
| **[business.md](business.md)** | Visão, SLAs, stakeholders |
| **[ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md)** | Matriz de erros + exemplos Go |

### 🛠️ Implementação & Desenvolvimento

| Documento | Objetivo |
|-----------|----------|
| **[dev-team.md](dev-team.md)** | Workflow, testes, deployment |
| **[reference/shared-module.md](reference/shared-module.md)** | Módulo shared |
| **[reference/technical-contracts.md](reference/technical-contracts.md)** | APIs e eventos |

### 📋 Políticas & Conformidade

| Documento | Objetivo |
|-----------|----------|
| **[reference/operational-compliance-policy.md](reference/operational-compliance-policy.md)** | Políticas unificadas |
| **[reference/idempotency-guide.md](reference/idempotency-guide.md)** | Idempotência em detalhes |
| **[reference/observability.md](reference/observability.md)** | Monitoramento |

### 🚨 Operações & Runbooks

| Documento | Objetivo |
|-----------|----------|
| **[playbooks/dlq-playbook.md](playbooks/dlq-playbook.md)** | Recuperar de falhas |
| **[playbooks/operations-runbooks.md](playbooks/operations-runbooks.md)** | Operações comuns |

### 📚 Referência Rápida

| Documento | Objetivo |
|-----------|----------|
| **[reference/faq.md](reference/faq.md)** | Decisões de design |

---

## 🔍 Encontre por Tópico

### Transações & Fluxo
- [QUICKSTART.md](QUICKSTART.md) § 3-4 — Fazer e entender transação
- [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) § 1-2 — Fluxo sucesso e erro

### Idempotência
- [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) § 3 — Política completa
- [reference/idempotency-guide.md](reference/idempotency-guide.md) — Exemplos SQL
- [reference/technical-contracts.md](reference/technical-contracts.md) — Como usar na API

### Error Handling & DLQ
- [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) — Matriz e padrões Go
- [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) § 3-4 — Diagramas de erro
- [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) — Reprocessamento operacional

### Hot Partitions & Consumer Lag
- [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) § 6 — Detecção e mitigação
- [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) § 1 — Runbook detalhado

### Reconciliação & Auditoria
- [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) § 7 — Fluxo de reconciliação
- [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) § 2 — Discrepâncias

### Observabilidade
- [reference/observability.md](reference/observability.md) — Métricas e alertas
- [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) § 5 — Políticas

### Segurança & Compliance
- [business.md](business.md) § 5-7 — Retenção e compliance
- [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) § 6 — Políticas

---

## 🎯 Tarefas Comuns (Como fazer...?)

| Tarefa | Documento | Seção |
|--------|-----------|-------|
| Rodar projeto localmente | [QUICKSTART.md](QUICKSTART.md) | Passo 1-2 |
| Fazer primeira transação | [QUICKSTART.md](QUICKSTART.md) | Passo 3 |
| Entender fluxo completo | [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) | § 1-2 |
| Tratar erro corretamente | [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) | § 3-5 |
| Reprocessar DLQ | [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) | Passo a passo |
| Lidar com consumer lag | [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) § 1 | Procedimento |
| Usar idempotência na API | [reference/technical-contracts.md](reference/technical-contracts.md) | § 1 |
| Entender estrutura do código | [STRUCTURE.md](STRUCTURE.md) | Layouts & padrões |
| Debugar transação falhada | [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) + [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) | Combinar |

---

## 📞 Suporte & Links Rápidos

| Questão | Resposta |
|---------|----------|
| "Como começo?" | [QUICKSTART.md](QUICKSTART.md) |
| "Qual é meu caminho?" | Escolha seu papel acima [👆](#-documentação-por-persona) |
| "Não entendo o fluxo" | Leia [ARCHITECTURE_FLOWS.md](ARCHITECTURE_FLOWS.md) |
| "Onde está o código?" | Consulte [STRUCTURE.md](STRUCTURE.md) |
| "Como tratar erros?" | Estude [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) |
| "Sistema em incidente" | [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) |
| "API não funciona" | [reference/technical-contracts.md](reference/technical-contracts.md) |

---

## 📋 Status da Documentação

**Fase 1 (Completa)** ✅
- [x] QUICKSTART.md (5 passos executáveis)
- [x] ARCHITECTURE_FLOWS.md (8 diagramas)
- [x] ERROR_HANDLING_PATTERNS.md (matriz + exemplos)

**Fase 2 (Consolidação)** 🔄
- [x] STRUCTURE.md (layout & padrões)
- [x] Consolidar idempotency-guide.md (resumo + exemplos)
- [x] Limpar redundâncias de observability
- [ ] Validar índices e links cruzados

**Fase 3 (Planejada)** 📅
- [ ] IMPLEMENTATION_EXAMPLES.md
- [ ] DEPLOYMENT_CHECKLIST.md
- [ ] VERSIONING_POLICY.md

---

## 🔄 Versionamento da Documentação

- **Versão:** 2.1
- **Data:** 2026-06-24
- **Tipo:** Fase 1 + Fase 2 (consolidação)
- **Próxima revisão:** 2026-08-01

---

## 📝 Como Contribuir com Documentação

1. **Identifique o gap** — Qual informação falta?
2. **Escolha o documento** — Qual é o melhor lugar para adicionar?
3. **Mantenha estilo** — Use Markdown, seções numeradas, links cruzados
4. **Valide links** — Rode CI para verificar link-check
5. **Faça PR** — Abra revisão para consolidar

**Proprietários:**
- Estrutura & Arquitetura: Staff Engineering
- Documentação de Operações: SRE
- Documentação de Produto: Product Manager
- Documentação Técnica: API Guild

---

**Obrigado por ler!** 🎉  
Quer contribuir? Abra uma issue ou PR no repositório.
