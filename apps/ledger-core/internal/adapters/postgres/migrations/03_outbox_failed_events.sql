-- Adicionar colunas de status, metadata e updated_at na tabela transactions se não existirem
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

-- Criar Tabela Outbox para envio confiável de eventos (Outbox Pattern)
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE NULL
);

-- Criar Índices Estratégicos no Outbox
CREATE INDEX IF NOT EXISTS idx_outbox_status_created ON outbox (status, created_at);

-- Criar Tabela para Armazenamento de Eventos que Excederam Retentativas (DLQ)
CREATE TABLE IF NOT EXISTS failed_events (
    id UUID PRIMARY KEY,
    source_topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    error TEXT NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    first_error_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_error_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB NULL
);
