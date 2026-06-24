# Quick Start — 5 Passos para Começar com Distributed Ledger

**Tempo estimado:** 30 minutos
**Pré-requisitos:** Git, Docker/Docker Compose, Go 1.22+

Este guia prático o levará de zero até rodar o projeto localmente e fazer seu primeiro teste.

---

## ⚡ Passo 1: Clonar e Explorar (5 min)

```bash
# Clonar o repositório
git clone https://github.com/seu-org/distributed-ledger-go.git
cd distributed-ledger-go

# Explorar a estrutura
tree -L 2 apps/
# Resultado esperado:
# apps/
# ├── ledger-core/          ← Processa transações
# ├── transaction-gw/       ← Gateway HTTP
# ├── notification-service/ ← Envia notificações
# ├── ledger-reconciler/    ← Auditoria de saldos
# ├── ledger-backoffice/    ← Dashboard administrativo
# └── rate-limiter/         ← Proteção contra DDoS
```

---

## ⚡ Passo 2: Subir a Infraestrutura com Docker (10 min)

```bash
# Subir Kafka, Postgres, Redis e todos os apps
docker-compose up -d

# Verificar se tudo está rodando
docker ps | grep distributed

# Resultado esperado: 8 containers em "Up" status
# - postgres
# - kafka
# - redis
# - transaction-gw
# - ledger-core
# - notification-service
# - ledger-reconciler
# - ledger-backoffice
```

**Portas disponíveis (clique nos links):**
- **API Gateway:** [http://localhost:8080](http://localhost:8080) — Transaction Gateway
- **Backoffice Dashboard:** [http://localhost:8081](http://localhost:8081) — Admin Dashboard
- **Prometheus:** [http://localhost:9090](http://localhost:9090) — Métricas & Queries
- **Alertmanager:** [http://localhost:9093](http://localhost:9093) — Gerenciador de Alertas
- **Grafana:** [http://localhost:3000](http://localhost:3000) — Dashboards Visuais (admin/admin)
- **PostgreSQL:** localhost:5432 — Banco de Dados
- **Kafka:** kafka:29092 (interno) / localhost:9092 (externo)

---

## ⚡ Passo 3: Fazer sua Primeira Transação (5 min)

```bash
# Terminal 1: Verificar health check do Gateway
curl -X GET http://localhost:8080/health
# Resultado esperado: {"status":"healthy"}

# Terminal 1: Enviar uma transação
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": "550e8400-e29b-41d4-a716-446655440000",
    "target_account_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "amount": 10000,
    "idempotency_key": "txn-001-'$(date +%s)'",
    "description": "Minha primeira transação"
  }'

# Resultado esperado: 202 Accepted
# {
#   "transaction_id": "uuid...",
#   "status": "PENDING",
#   "idempotency_key": "txn-001-...",
#   "created_at": "2026-06-24T10:00:00Z"
# }
```

---

## ⚡ Passo 4: Entender o Fluxo da Transação (7 min)

```bash
# Cada transação segue este fluxo:
# 1. Cliente envia POST /transactions → Gateway recebe
# 2. Gateway valida + registra em transactions table + cria outbox entry
# 3. Gateway publica no Kafka topic "transactions"
# 4. ledger-core consome do Kafka
# 5. ledger-core cria ledger_entries (débito + crédito)
# 6. outbox-worker publica evento para downstream
# 7. notification-service envia notificação
# 8. Dashboard mostra resultado final

# Monitorar logs em tempo real
docker logs -f ledger-core --tail 50

# Você verá algo como:
# 2026-06-24T10:00:01Z INFO Processando transação source_account=550e... target_account=6ba7... amount=10000
# 2026-06-24T10:00:02Z INFO Lançamentos criados com sucesso entries_count=2
# 2026-06-24T10:00:02Z INFO Trabalhador de Outbox publicando eventos...
```

---

## ⚡ Passo 5: Rodar Testes (3 min)

```bash
# Testar um módulo específico
cd apps/ledger-core
go test ./...

# Testar todos os módulos
cd ../.. && go test ./apps/... ./shared/...

# Resultado esperado: todos os testes passam
# ok   transaction-gw/internal/domain  0.343s
# ?    ledger-core/cmd/app             [no test files]
# ...

# Se houver erro, veja a seção "Troubleshooting" abaixo
```

---

## 📊 Próximos Passos

Agora que o projeto está rodando, recomendamos:

### Entender a Arquitetura (30 min)
Leia [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) para ver:
- Fluxo de transação bem-sucedida
- Fluxo de erro e DLQ
- Fluxo de hot partition

### Criar seu Primeiro Caso de Uso (1 hora)
Veja as documentações de arquitetura para aprender:
- Como criar um novo usecase
- Como implementar error handling com Outbox
- Como escrever testes

### Entender Idempotência (20 min)
Leia [`docs/reference/operational-compliance-policy.md`](reference/operational-compliance-policy.md) para saber:
- Por que idempotência importa
- Como `idempotency_key` funciona
- Exemplo de requisição duplicada

### Explorar o Dashboard (10 min)
Abra [http://localhost:8081](http://localhost:8081) e explore:
- Dashboard executivo
- Audit trail de contas
- Monitor de DLQ

### Explorar a API (10 min)

#### Swagger UI (Parcial - gerado automaticamente)
- URL: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- Endpoints: POST /accounts, POST /transactions, GET /transactions/{id}

#### ReDoc + OpenAPI 3.0 (Completo) ⭐ **RECOMENDADO**
```bash
# Iniciar ReDoc em paralelo
docker run -d \
  -p 8888:80 \
  -e SPEC_URL=/openapi.yaml \
  -v $(pwd)/docs/openapi.yaml:/usr/share/nginx/html/openapi.yaml \
  redocly/redoc
```
- URL: [http://localhost:8888](http://localhost:8888)
- Endpoints: Todos (health, accounts, transactions, metrics)

**Veja também:** [`API_REFERENCE.md`](API_REFERENCE.md) para detalhes sobre documentação da API

---

## 🐛 Troubleshooting

### Docker não está rodando
```bash
# Verificar status
docker-compose ps

# Se houver erro, verificar logs
docker-compose logs postgres

# Recriar tudo do zero
docker-compose down -v
docker-compose up -d
```

### "Connection refused" na porta 8080
```bash
# Verificar se o gateway está rodando
docker ps | grep transaction-gw

# Se não estiver, verificar erro
docker-compose logs transaction-gw

# Verificar variáveis de ambiente
cat docker-compose.yml | grep -A 5 "transaction-gw"
```

### Testes falhando com "database connection error"
```bash
# Verificar se postgres está rodando
docker-compose logs postgres

# Se erro de migração, executar manualmente
docker-compose exec postgres psql -U staff_eng -d ledger_db < \
  apps/ledger-core/internal/adapters/postgres/migrations/01_create_tables.sql
```

### "Outbox pending" acumulando
```bash
# Verificar se o worker está rodando
docker logs ledger-core | grep "Trabalhador de Outbox"

# Se não houver logs, o worker pode estar falhando
docker-compose logs ledger-core --tail 100
```

---

## 📚 Documentação Essencial por Papel

| Papel | Próximo passo | Link |
|-------|--------------|------|
| **Engineer (Você quer fazer mudanças)** | Entender padrões de erro | [`ERROR_HANDLING_PATTERNS.md`](ERROR_HANDLING_PATTERNS.md) |
| **SRE/Ops (Você quer operar)** | Aprender runbooks | [`playbooks/dlq-playbook.md`](playbooks/dlq-playbook.md) |
| **Product (Você quer features)** | Entender SLAs | [`business.md`](business.md) |
| **API Consumer (Você usa a API)** | Contratos de API | [`reference/technical-contracts.md`](reference/technical-contracts.md) |
| **Monitoramento (Você cuida de observabilidade)** | Acessar sistemas | [`SYSTEM_ACCESS.md`](SYSTEM_ACCESS.md) |
| **Documentação API/Swagger** | Testar endpoints e ver specs | [`API_REFERENCE.md`](API_REFERENCE.md) |

---

## ✅ Checklist de Conclusão

- [ ] Docker está rodando (verificar `docker-compose ps`)
- [ ] Primeira transação foi enviada (curl POST /v1/transactions)
- [ ] Logs mostram processamento (ver `docker logs -f ledger-core`)
- [ ] Testes passam (rodar `go test ./...`)
- [ ] Dashboard acessível (abrir `http://localhost:8081/backoffice`)
- [ ] Leu [`ARCHITECTURE.md`](ARCHITECTURE.md)

Se todos os itens estão checkados, **parabéns! Você está pronto para contribuir** 🎉

---

## ❓ Perguntas Frequentes

**P: Posso rodar sem Docker?**
R: Sim, mas você precisa instalar Postgres, Kafka e Redis manualmente. Não recomendamos para novos devs. Veja [`dev-team.md`](dev-team.md) seção "Executar Localmente".

**P: Como debugar uma transação que falhou?**
R: Veja [`ERROR_HANDLING_PATTERNS.md`](ERROR_HANDLING_PATTERNS.md) e [`playbooks/dlq-playbook.md`](playbooks/dlq-playbook.md).

**P: Como criar um novo serviço no monorepo?**
R: Veja o guia de estrutura do projeto em [`ARCHITECTURE.md`](ARCHITECTURE.md) para melhores práticas.

**P: Onde vejo as métricas?**
R: Dashboard em `http://localhost:8081/backoffice` ou configure Prometheus/Grafana. Veja [`reference/observability.md`](reference/observability.md).

---

**Suporte:** Abra uma issue no GitHub ou contate o time de Staff Engineering.

**Última atualização:** 2026-06-24
**Próxima revisão:** 2026-07-15
