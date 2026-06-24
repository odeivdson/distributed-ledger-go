-- GAP-001: Adicionar status à tabela de contas para validação de conta ativa
ALTER TABLE accounts_balance ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'ACTIVE';

-- Validação de status permitido (com tratamento de constraint duplicada)
-- First check and drop if exists, then create
ALTER TABLE accounts_balance DROP CONSTRAINT IF EXISTS check_account_status;
ALTER TABLE accounts_balance ADD CONSTRAINT check_account_status CHECK (status IN ('ACTIVE', 'INACTIVE'));

-- Índice para busca de contas ativas
CREATE INDEX IF NOT EXISTS idx_accounts_balance_status ON accounts_balance(status);

-- Update existing accounts to be ACTIVE (backward compatibility)
UPDATE accounts_balance SET status = 'ACTIVE' WHERE status IS NULL;
