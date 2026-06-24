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
| [ARCHITECTURE.md](ARCHITECTURE.md) | 40 min | Entender arquitetura, layout e fluxos |
| [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) | 30 min | Padrões de erro + exemplos Go |
| [API_REFERENCE.md](API_REFERENCE.md) | 20 min | APIs, Swagger e exemplos de requisição |

**Checklist para começar:**
- [ ] Rodou `docker-compose up -d`
- [ ] Fez primeira transação com curl
- [ ] Leu ARCHITECTURE.md
- [ ] Entendeu ERROR_HANDLING_PATTERNS.md

---

### 🚨 **SRE / Operações**
Você quer monitorar, responder a incidentes e fazer deploy.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | 30 min | Componentes do sistema e fluxos de erro |
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
| [ARCHITECTURE.md](ARCHITECTURE.md) | 20 min | Design técnico simplificado |
| [reference/faq.md](reference/faq.md) | 20 min | Decisões arquiteturais |

**Checklist de lançamento:**
- [ ] Entendeu SLAs e critérios de aceite
- [ ] Aprovou política de retenção (30d idempotency_key, 90d failed_events)
- [ ] Validou plano de rollout com equipe
- [ ] Acordou critério de sucesso da feature

---

### 🔌 **API Consumer (Cliente Externo)**
Você quer integrar com a API do ledger.

| Documento | Tempo | Objetivo |
|-----------|-------|----------|
| [API_REFERENCE.md](API_REFERENCE.md) | 15 min | Endpoints, payloads, exemplos |
| [QUICKSTART.md](QUICKSTART.md) | 5 min | Fazer primeira chamada |
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
| **[ARCHITECTURE.md](ARCHITECTURE.md)** | Layout completo, design técnico, componentes, diagramas Mermaid |

### 🎯 Quick References

| Documento | Objetivo |
|-----------|----------|
| **[QUICKSTART.md](QUICKSTART.md)** | 5 passos para rodar localmente |
| **[business.md](business.md)** | Visão, SLAs, stakeholders |
| **[ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md)** | Matriz de erros + exemplos Go |
| **[API_REFERENCE.md](API_REFERENCE.md)** | Referência central da API e Swagger |

### 🛠️ Implementação & Desenvolvimento

| Documento | Objetivo |
|-----------|----------|
| **[dev-team.md](dev-team.md)** | Workflow, testes, deployment |
| **[reference/shared-module.md](reference/shared-module.md)** | Módulo shared |

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
- [QUICKSTART.md](QUICKSTART.md) — Fazer e entender transação
- [ARCHITECTURE.md](ARCHITECTURE.md) — Fluxo sucesso e erro

### Idempotência
- [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) — Política completa
- [reference/idempotency-guide.md](reference/idempotency-guide.md) — Exemplos SQL
- [API_REFERENCE.md](API_REFERENCE.md) — Como usar na API

### Error Handling & DLQ
- [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) — Matriz e padrões Go
- [ARCHITECTURE.md](ARCHITECTURE.md) — Diagramas de erro
- [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) — Reprocessamento operacional

### Hot Partitions & Consumer Lag
- [ARCHITECTURE.md](ARCHITECTURE.md) — Detecção e mitigação
- [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) — Runbook detalhado

### Reconciliação & Auditoria
- [ARCHITECTURE.md](ARCHITECTURE.md) — Fluxo de reconciliação
- [playbooks/operations-runbooks.md](playbooks/operations-runbooks.md) — Discrepâncias

### Observabilidade
- [reference/observability.md](reference/observability.md) — Métricas e alertas

### Segurança & Compliance
- [business.md](business.md) — Retenção e compliance
- [reference/operational-compliance-policy.md](reference/operational-compliance-policy.md) — Políticas

---

## 📞 Suporte & Links Rápidos

| Questão | Resposta |
|---------|----------|
| "Como começo?" | [QUICKSTART.md](QUICKSTART.md) |
| "Qual é meu caminho?" | Escolha seu papel acima |
| "Não entendo o fluxo" | Leia [ARCHITECTURE.md](ARCHITECTURE.md) |
| "Onde está o código?" | Consulte [ARCHITECTURE.md](ARCHITECTURE.md) |
| "Como tratar erros?" | Estude [ERROR_HANDLING_PATTERNS.md](ERROR_HANDLING_PATTERNS.md) |
| "Sistema em incidente" | [playbooks/dlq-playbook.md](playbooks/dlq-playbook.md) |
| "API não funciona" | [API_REFERENCE.md](API_REFERENCE.md) |

---

## 🔄 Versionamento da Documentação

- **Versão:** 3.0
- **Data:** 2026-06-24
- **Tipo:** Fase 3 (Consolidação & Simplificação)

---

**Obrigado por ler!** 🎉  
Quer contribuir? Abra uma issue ou PR no repositório.
