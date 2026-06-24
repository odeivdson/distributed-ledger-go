# 📚 API Reference

**Data:** 2026-06-24  
**Versão da API:** v1.0.0  

A fonte de verdade para a documentação da API é o arquivo OpenAPI 3.0: `docs/openapi/openapi.yaml`.

---

## 🔍 Como Visualizar a Documentação Interativa

Para interagir com os endpoints e ver todos os schemas (modelos), recomendamos o uso do **ReDoc**:

```bash
docker run -d \
  -p 8888:80 \
  -e SPEC_URL=/openapi.yaml \
  -v $(pwd)/docs/openapi/openapi.yaml:/usr/share/nginx/html/openapi.yaml \
  redocly/redoc
```
Após rodar o comando acima, acesse: [http://localhost:8888](http://localhost:8888)

Alternativamente, importe `docs/openapi/openapi.yaml` no **Postman** ou acesse o Swagger automático para desenvolvimento em [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

---

## 🧪 Teste Rápido (cURL)

Todas as requisições que alteram estado necessitam do header `X-Request-ID` para rastreamento (já a transação em si exige `idempotency_key` no corpo da requisição para prevenir duplicidade).

### 1. Criar uma Transação (POST /v1/transactions)
```bash
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: $(uuidgen)" \
  -d '{
    "source_account_id": "550e8400-e29b-41d4-a716-446655440000",
    "target_account_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "amount": 10000,
    "idempotency_key": "txn-001-'$(date +%s)'",
    "description": "Exemplo de transação via curl"
  }'
```
**Resposta (202 Accepted):**
A transação entra em processamento assíncrono. Retorna um `transaction_id`.

### 2. Consultar Status (GET /v1/transactions/{id})
```bash
curl -X GET http://localhost:8080/v1/transactions/ID_DA_TRANSACAO \
  -H "X-Request-ID: $(uuidgen)"
```
**Respostas Possíveis:** `PENDING`, `COMPLETED`, `FAILED`.

### 3. Criar Conta (POST /v1/accounts)
Para testes, as contas podem ser criadas apontando saldo inicial.
```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: $(uuidgen)" \
  -d '{
    "account_id": "550e8400-e29b-41d4-a716-446655440000",
    "balance_in_cents": 50000
  }'
```

---

## 🔐 Erros Comuns
- **400 Invalid JSON Schema:** O body não obedece ao schema (ex: `amount` ausente ou negativo).
- **409 Conflict:** `idempotency_key` já foi utilizada (A transação anterior é retornada para manter a consistência).
- **503 Service Unavailable:** Falha temporária em algum recurso (Kafka, DB). A requisição pode ser re-tentada.
