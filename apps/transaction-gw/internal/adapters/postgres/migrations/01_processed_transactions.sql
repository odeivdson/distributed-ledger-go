-- Tabela para rastreamento de transações processadas no Gateway
-- Permite validação de idempotência e deduplicação
CREATE TABLE IF NOT EXISTS processed_transactions (
    id UUID PRIMARY KEY,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    request_hash VARCHAR(64) NOT NULL,
    response_status VARCHAR(20) NOT NULL,
    response_body TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '30 days')
);

-- Índice para busca rápida por idempotency_key
CREATE INDEX IF NOT EXISTS idx_processed_transactions_idempotency_key ON processed_transactions(idempotency_key);

-- Índice para limpeza de registros expirados
CREATE INDEX IF NOT EXISTS idx_processed_transactions_expires_at ON processed_transactions(expires_at);

-- Índice para busca por status
CREATE INDEX IF NOT EXISTS idx_processed_transactions_status ON processed_transactions(response_status);
