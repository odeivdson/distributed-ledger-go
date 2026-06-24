# 🔐 Acesso aos Sistemas - Credenciais e Links

**Data:** 2026-06-24  
**Status:** ✅ Todos os sistemas estão acessíveis localmente

---

## 🔑 Credenciais de Acesso

### Grafana (Monitoramento Visual)
- **URL:** [http://localhost:3000](http://localhost:3000)
- **Usuário:** `admin`
- **Senha:** `admin`
- **Função:** Dashboards de métricas, alertas visuais

### PostgreSQL (Banco de Dados)
- **Host:** localhost
- **Porta:** 5432
- **Usuário:** `staff_eng`
- **Senha:** `super_secret_password`
- **Database:** `ledger_db`
- **Comando de acesso:**
  ```bash
  psql -h localhost -U staff_eng -d ledger_db
  ```

### Kafka (Message Broker)
- **Host (Interno):** kafka:29092
- **Host (Externo):** localhost:9092
- **Tópicos principais:**
  - `transactions` - Transações em processamento
  - `notifications` - Notificações para clientes

### Redis (Cache)
- **Host:** localhost
- **Porta:** 6379
- **Sem autenticação** (padrão)

---

## 🌐 Endpoints de Acesso

### Serviços Principais

| Serviço | URL | Porta | Descrição |
|---------|-----|-------|-----------|
| **Transaction Gateway** | [http://localhost:8080](http://localhost:8080) | 8080 | API de transações |
| **Ledger Core** | [http://localhost:8082](http://localhost:8082) | 8082 | Processamento de ledger |
| **Backoffice** | [http://localhost:8081](http://localhost:8081) | 8081 | Dashboard administrativo |
| **Notification Service** | [http://localhost:8083](http://localhost:8083) | 8083 | Serviço de notificações |
| **Ledger Reconciler** | [http://localhost:8084](http://localhost:8084) | 8084 | Reconciliação de saldos |

### Monitoramento & Observabilidade

| Sistema | URL | Porta | Descrição |
|---------|-----|-------|-----------|
| **Prometheus** | [http://localhost:9090](http://localhost:9090) | 9090 | Métricas & Queries PromQL |
| **Grafana** | [http://localhost:3000](http://localhost:3000) | 3000 | Dashboards Visuais |
| **Alertmanager** | [http://localhost:9093](http://localhost:9093) | 9093 | Gerenciador de Alertas |

---

## 📊 Queries Úteis no Prometheus

### Health Check dos Serviços
```promql
up{job=~"transaction-gw|ledger-core|notification-service"}
```

### Taxa de Transações por Segundo
```promql
rate(transactions_total[1m])
```

### Latência p95
```promql
histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m]))
```

### Erros em Processamento
```promql
rate(processing_errors_total[5m])
```

---

## 📈 Dashboards Grafana Disponíveis

Login com `admin/admin` para acessar:

1. **Overview Dashboard**
   - Status de todos os serviços
   - Taxa de transações
   - Erros em tempo real

2. **Ledger Core Performance**
   - Latência de processamento
   - Taxa de sucesso/falha
   - Utilização de recursos

3. **Transaction Gateway**
   - Requisições recebidas
   - Taxa de rejeição
   - Tempo de resposta

4. **System Health**
   - Memória disponível
   - CPU dos containers
   - Disk usage

---

## 🔧 Verificar Status dos Containers

```bash
# Ver todos os containers rodando
docker-compose ps

# Resultado esperado (8 serviços):
# NAME                    STATUS
# kafka                   Up 5 minutes
# postgres                Up 5 minutes
# redis                   Up 5 minutes
# transaction-gw          Up 5 minutes
# ledger-core             Up 5 minutes
# notification-service    Up 5 minutes
# ledger-reconciler       Up 5 minutes
# ledger-backoffice       Up 5 minutes
# prometheus              Up 5 minutes
# alertmanager            Up 5 minutes
# grafana                 Up 5 minutes
```

---

## 📋 Variáveis de Ambiente

Todas as credenciais estão em `.env`:

```bash
# Banco de Dados
DB_USER=staff_eng
DB_PASSWORD=super_secret_password
DB_NAME=ledger_db

# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin

# Kafka
KAFKA_BROKERS=kafka:29092

# Redis
REDIS_ADDR=redis:6379

# Portas
GW_PORT=8080
BACKOFFICE_PORT=8081
CORE_PORT=8082
NOTIF_PORT=8083
RECON_PORT=8084
PROMETHEUS_PORT=9090
ALERTMANAGER_PORT=9093
GRAFANA_PORT=3000
```

---

## ⚠️ Segurança em Produção

**⚠️ IMPORTANTE:** As credenciais acima são APENAS para desenvolvimento local.

Para produção:
- [ ] Use senhas fortes (mínimo 32 caracteres)
- [ ] Implemente mTLS entre serviços
- [ ] Configure firewalls para restringir acesso
- [ ] Use Secret Management (Vault, AWS Secrets Manager)
- [ ] Ative logs de auditoria em todos os acessos
- [ ] Configure rate limiting no gateway
- [ ] Implemente zero-trust networking

Veja [`reference/operational-compliance-policy.md`](reference/operational-compliance-policy.md) para detalhes.

---

## 🆘 Troubleshooting de Acesso

### "Connection refused" ao acessar Grafana
```bash
# Verificar se Grafana está rodando
docker-compose ps | grep grafana

# Se não estiver, reiniciar
docker-compose restart grafana

# Ver logs
docker-compose logs grafana
```

### Prometheus sem dados
```bash
# Verificar targets no Prometheus
curl http://localhost:9090/api/v1/targets

# Se algum estiver DOWN, verificar logs do serviço
docker-compose logs ledger-core
```

### PostgreSQL recusa conexão
```bash
# Verificar se postgres está rodando
docker-compose ps | grep postgres

# Verificar credentials
cat .env | grep DB_

# Tentar conexão
psql -h localhost -U staff_eng -d ledger_db
```

---

## ✅ Checklist de Acesso

- [ ] Conseguir acessar [http://localhost:8080](http://localhost:8080) (Gateway)
- [ ] Conseguir fazer login no Grafana com admin/admin
- [ ] Ver dados no Prometheus [http://localhost:9090](http://localhost:9090)
- [ ] Conseguir conectar ao PostgreSQL
- [ ] Ver logs em tempo real: `docker-compose logs -f`
- [ ] Todos os 8+ containers rodando (docker-compose ps)

---

**Última atualização:** 2026-06-24  
**Próxima revisão:** 2026-07-15
