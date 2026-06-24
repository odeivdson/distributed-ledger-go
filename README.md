# Distributed Ledger Go — Sistema de Ledger Financeiro Distribuído

> Um sistema de ledger distribuído pronto para produção, construído em Go, Kafka e PostgreSQL. Implementa contabilidade de partidas dobradas com garantias de consistência estrita, projetado para transações financeiras de alta throughput.

[![Status Build](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Versão Go](https://img.shields.io/badge/Go-1.22+-blue)]()
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)]()
[![Apache Kafka](https://img.shields.io/badge/Kafka-3.6-blue)]()
[![Licença](https://img.shields.io/badge/license-MIT-green)]()

---

## 🎯 Visão Geral

**Distributed Ledger Go** é uma implementação de referência de um sistema de ledger financeiro distribuído que prioriza:

- **Consistência Estrita:** Contabilidade de partidas dobradas com garantias ACID
- **Idempotência:** Semântica At-Least-Once com detecção de duplicatas
- **Tolerância a Falhas:** Degradação graciosa com Outbox Pattern e reprocessamento de DLQ
- **Escalabilidade:** Arquitetura orientada a eventos suportando milhões de transações
- **Observabilidade:** Métricas compreensivas, logs estruturados e rastreamento distribuído
- **Qualidade de Código:** Engenharia em nível Staff com Arquitetura Hexagonal, princípios SOLID e Clean Architecture

Este projeto demonstra padrões empresariais para construir infraestrutura financeira crítica usando Go.

---

## 📚 Links Rápidos

- **🚀 Primeiros Passos:** [Quick start em 5 minutos](docs/QUICKSTART.md)
- **📖 Documentação Completa:** [Índice completo de docs](docs/README.md)
- **🏗️ Arquitetura:** [Arquitetura e Componentes do Sistema](docs/ARCHITECTURE.md)
- **⚙️ Guia de Implementação:** [Workflow de desenvolvimento](docs/dev-team.md)
- **🎓 Aprenda com Exemplos:** [Padrões de tratamento de erro](docs/ERROR_HANDLING_PATTERNS.md)

---

## 🚀 Primeiros Passos

O guia detalhado para iniciar a aplicação, configurar os containers e realizar a primeira transação está centralizado em nosso Quickstart:

👉 **[Guia rápido de inicialização →](docs/QUICKSTART.md)**

---

## 🏗️ Arquitetura

### Componentes do Sistema

```
┌─────────────────────────────────────────────────────────────┐
│                     Camada de Gateway API                   │
│  ┌──────────────────────┬──────────────────────────────┐   │
│  │   Transaction GW     │    Rate Limiter (Adaptativo) │   │
│  │  (HTTP + Kafka)      │  (Redis + Fallback Local)    │   │
│  └──────────────────────┴──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                   Stream de Eventos (Kafka)                 │
│  Tópicos: transactions, failed-events, reconciliation      │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────┐
│        Camada de Consumidores (Kafka → Lógica de Negócio)    │
│  ┌─────────────────┬──────────────┬──────────────────┐       │
│  │  Ledger Core    │ Reconciliador│  Notification    │       │
│  │ (Persist Txns)  │   (Audit)    │    Service       │       │
│  └─────────────────┴──────────────┴──────────────────┘       │
└──────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                     Camada de Dados                          │
│  ┌────────────────────────────────────────────────────┐    │
│  │PostgreSQL (transactions, ledger_entries, outbox)  │    │
│  │Redis (cache, rate limits, locks)                  │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Princípios Chave

- **Append-Only:** Sem UPDATE/DELETE em lançamentos—somente INSERT
- **Contabilidade de Partidas Dobradas:** Cada transação cria lançamentos balanceados de débito/crédito
- **Idempotência:** Requisições duplicadas retornam resultados em cache
- **Event Sourcing:** Trilha de auditoria completa via log de eventos imutável
- **Degradação Graciosa:** Padrão de Outbox local garante entrega mesmo se Kafka cair

👉 **[Detalhes da Arquitetura e Fluxos →](docs/ARCHITECTURE.md)**

---

## 🛠️ Stack Tecnológica

| Componente | Tecnologia | Versão | Propósito |
|-----------|-----------|--------|---------|
| **Linguagem** | Go | 1.22+ | Sistemas concorrentes de alto desempenho |
| **Mensageria** | Apache Kafka | 3.6+ (KRaft) | Stream de eventos, garantia de entrega |
| **Banco de Dados** | PostgreSQL | 16 | Transações ACID, ledger imutável |
| **Cache** | Redis | 7.2+ | Rate limiting, locks distribuídos |
| **Containerização** | Docker Compose | Últimas | Dev local & testing |
| **Observabilidade** | Prometheus/Grafana | Últimas | Métricas & alertas (opcional) |

---

## 📂 Estrutura do Projeto

O projeto segue **Arquitetura Hexagonal** em layout de monorepo:

```
distributed-ledger-go/
├── apps/                          # Microserviços
│   ├── transaction-gw/            # API HTTP ingress → publisher Kafka
│   ├── ledger-core/               # Consumer Kafka → persist transactions
│   ├── rate-limiter/              # Middleware HTTP rate limiting adaptativo
│   ├── notification-service/      # Consumer evento com worker pool
│   ├── ledger-reconciler/         # Auditor em batch (eventual consistency)
│   └── ledger-backoffice/         # Dashboard administrativo (server-rendered)
├── infra/                         # Configurações de infraestrutura e deploy
│   └── observability/             # Prometheus, Grafana, Alertmanager
├── shared/                        # Pacotes compartilhados
│   ├── contracts/                 # Contratos de eventos & schemas
│   ├── models/                    # DTOs & value objects
│   ├── clients/                   # Clientes de infraestrutura
│   ├── schema/                    # Schemas JSON & definições protobuf
│   └── middleware/                # Preocupações transversais
├── migrations/                    # Schema de banco & evolução
├── docs/                          # Documentação completa
│   ├── README.md                  # Hub de documentação
│   ├── QUICKSTART.md              # Onboarding em 5 passos
│   ├── ARCHITECTURE.md            # Arquitetura, componentes e fluxos
│   ├── API_REFERENCE.md           # Referência da API (Swagger/OpenAPI)
│   ├── ERROR_HANDLING_PATTERNS.md # Matriz de erros + exemplos Go
│   ├── business.md                # Visão, SLAs, stakeholders
│   ├── dev-team.md                # Workflow do desenvolvedor
│   ├── reference/                 # Contratos API, políticas
│   └── playbooks/                 # Runbooks operacionais
├── docker-compose.yml             # Ambiente de desenvolvimento local
└── .github/workflows/             # Pipelines CI/CD
```

👉 **[Guia completo de arquitetura →](docs/ARCHITECTURE.md)**

---

## 🎓 Recursos Principais

### ✅ Idempotência
Cada requisição externa inclui uma `idempotency_key` para detectar e deduplicas requisições:
```bash
POST /api/v1/transactions \
  -H "X-Idempotency-Key: uuid-gerado-cliente" \
  -d '{"account_id": "acc-001", "amount": 10000, ...}'

# Se enviado novamente com mesma chave → retorna resultado em cache (200 OK)
# Se enviado primeira vez → processa e retorna resultado (202 Accepted)
```

### ✅ Tratamento de Erro e Resiliência
- **Erros transitórios** → Retry automático com backoff exponencial
- **Erros permanentes** → Rota para DLQ para investigação manual
- **Falhas de infraestrutura** → Degradação graciosa com Outbox Pattern
- **Circuit breaker** → Proteção contra falhas em cascata

👉 **[Padrões de tratamento de erro →](docs/ERROR_HANDLING_PATTERNS.md)**

### ✅ Observabilidade
- **Logs estruturados** com IDs de rastreamento para tracing end-to-end
- **Métricas Prometheus** para throughput, latência e taxas de erro de transações
- **Rastreamento distribuído** via OpenTelemetry (opcional)
- **Dashboard administrativo** para inspeção de ledger em tempo real

### ✅ Escalabilidade
- **Scaling horizontal** de consumidores via particionamento Kafka
- **Paginação baseada em cursor** para processamento em batch eficiente
- **Connection pooling** para eficiência de banco de dados
- **Worker pools** com concorrência controlada

---

## 🚀 Executando o Projeto

### Desenvolvimento Local
```bash
# Iniciar todos os serviços
docker-compose up -d

# Rodar testes com race detector
go test -race ./...

# Executar um serviço específico em modo desenvolvimento
cd apps/transaction-gw
go run cmd/main.go

# Ver logs
docker-compose logs -f ledger-core
```

### Executando Testes
```bash
# Testes unitários
go test ./...

# Testes de integração (requer Docker)
go test -tags=integration ./...

# Com race detector (recomendado para serviços Go)
go test -race ./...

# Com cobertura
go test -cover ./...
```

### Acessando Serviços

| Serviço | URL | Propósito |
|---------|-----|----------|
| **API de Transações** | http://localhost:8080 | Enviar transações |
| **Dashboard Admin** | http://localhost:8888 | Ver estado do ledger |
| **Prometheus** | http://localhost:9090 | Métricas (se habilitado) |
| **Grafana** | http://localhost:3000 | Dashboards (se habilitado) |

---

## 📖 Documentação

Este projeto inclui documentação abrangente de tier-1:

### Para Todos
- **[Quick Start (30 min)](docs/QUICKSTART.md)** — Colocar o projeto em funcionamento
- **[Visão Geral](docs/README.md)** — Hub de documentação com navegação por role

### Para Engenheiros
- **[Arquitetura e Fluxos](docs/ARCHITECTURE.md)** — Deep dive técnico, componentes e diagramas Mermaid
- **[Padrões de Tratamento de Erro](docs/ERROR_HANDLING_PATTERNS.md)** — Matriz + exemplos Go

### Para Operações
- **[Políticas Operacionais](docs/reference/operational-compliance-policy.md)** — SLAs, idempotência, compliance
- **[DLQ Playbook](docs/playbooks/dlq-playbook.md)** — Reprocessamento de eventos falhados
- **[Runbooks de Operações](docs/playbooks/operations-runbooks.md)** — Lag, partições quentes, reconciliação

### Para Produto & Stakeholders
- **[Visão de Negócio](docs/business.md)** — Features, SLAs, roadmap
- **[FAQ](docs/reference/faq.md)** — Decisões de design e rationale

---

## 🏆 Excelência em Engenharia

Este projeto demonstra engenharia em nível Staff:

### Padrões Arquiteturais
- **Arquitetura Hexagonal** — Domain-driven design independente de frameworks
- **Event Sourcing** — Trilha de auditoria imutável e completa
- **CQRS (eventual consistency)** — Separação de modelos read/write
- **Outbox Pattern** — Outbox transacional com garantia de entrega
- **Circuit Breaker** — Resiliência contra falhas em cascata
- **Worker Pool** — Concorrência controlada para eficiência de recursos

### Qualidade de Código
- **Princípios SOLID** — Single responsibility, open/closed, etc.
- **Clean Architecture** — Camadas Domain, Application, Infrastructure
- **Testes Compreensivos** — Unitários, integração, contrato, end-to-end
- **Linting Estrito** — golangci-lint, vet, fmt
- **Race Detector** — Todos os testes rodam com flag `-race`

### Excelência Operacional
- **Observabilidade** — Logs estruturados, métricas, rastreamento distribuído
- **Runbooks** — Guias passo-a-passo para tarefas operacionais comuns
- **Graceful Shutdown** — Drenagem limpa de conexões e processamento de mensagens
- **Health Checks** — Probes de liveness e readiness
- **Segurança** — Scanning de secrets, auditoria de dependências, geração SBOM

---

## 🤝 Contribuindo

Bem-vindo contribuições! Por favor siga estas diretrizes:

### Antes de Começar
1. Leia [ARCHITECTURE.md](docs/ARCHITECTURE.md) para entender o layout do monorepo e os fluxos de eventos.
2. Revise [ERROR_HANDLING_PATTERNS.md](docs/ERROR_HANDLING_PATTERNS.md)
3. Verifique [dev-team.md](docs/dev-team.md) para convenções de workflow

### Processo de Contribuição
1. Faça fork do repositório
2. Crie um branch de feature: `git checkout -b feature/sua-feature`
3. Faça mudanças seguindo os padrões deste projeto
4. Execute testes: `go test -race ./...`
5. Submeta um pull request com descrição clara

### Padrões de Código
- ✅ Todos os testes passam com race detector: `go test -race ./...`
- ✅ Código formatado com `gofmt`
- ✅ Linting passa: `golangci-lint run`
- ✅ Novos pacotes incluem arquivos `*_test.go`
- ✅ Documentação atualizada para APIs públicas

---

## 🔒 Segurança

Segurança é uma preocupação de primeira classe:

- **Idempotência** previne ataques de replay
- **Validação de entrada** protege contra ataques de injeção
- **Gerenciamento de secrets** via variáveis de ambiente (nunca commitadas)
- **Scanning de dependências** via GitHub Dependabot
- **Geração SBOM** para transparência da cadeia de suprimentos

👉 Para problemas de segurança, por favor envie email para `security@example.com`

---

## 📊 Performance & Benchmarks

Características esperadas de performance (em hardware moderno):

| Métrica | Target | Notas |
|--------|--------|-------|
| **Throughput de Transações** | 10K+ TPS | Por partição Kafka |
| **Latência de Ingestion (p99)** | < 100ms | De API para banco de dados |
| **Latência de Processamento de Evento** | < 500ms | De Kafka para consumer |
| **Lookup de Idempotência** | < 10ms | Taxa de cache hit 99%+ |
| **Scan de Reconciliação** | O(1) por batch | Paginação baseada em cursor |

---

## 📝 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

## 🙏 Agradecimentos

Construído seguindo melhores práticas da indústria de:
- Sam Newman — *Building Microservices*
- Robert C. Martin — *Clean Architecture*
- Eric Evans — *Domain-Driven Design*
- Chris Richardson — *Microservices Patterns*

---

## 📞 Suporte & Recursos

- **Documentação:** [docs/README.md](docs/README.md)
- **Issues:** [GitHub Issues](https://github.com/seu-usuario/distributed-ledger-go/issues)
- **Discussões:** [GitHub Discussions](https://github.com/seu-usuario/distributed-ledger-go/discussions)

---

## 🎯 Próximos Passos

**Explore o codebase:**
1. Comece com [QUICKSTART.md](docs/QUICKSTART.md)
2. Leia [ARCHITECTURE.md](docs/ARCHITECTURE.md) para entender a estrutura de serviços e fluxos
3. Escolha sua área: [dev-team.md](docs/dev-team.md) para implementação, [playbooks/](docs/playbooks/) para operações

**Contribua:**
- Abra uma issue para bugs ou solicitações de features
- Submeta PRs seguindo as diretrizes de contribuição
- Melhore a documentação!

---

**Construído com ❤️ pelo Time de Engenharia**  
*Distributed Ledger Go — Infraestrutura Financeira Empresarial para a Cloud*
