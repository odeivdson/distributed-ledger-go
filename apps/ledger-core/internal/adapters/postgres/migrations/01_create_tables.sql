-- Tabela de Contas e Controle de Saldo Consolidado
CREATE TABLE IF NOT EXISTS accounts_balance (
                                  account_id UUID PRIMARY KEY,
                                  balance_in_cents BIGINT NOT NULL DEFAULT 0,
                                  version BIGINT NOT NULL DEFAULT 0, -- Controle de concorrência otimista
                                  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                  CONSTRAINT balance_cannot_be_negative CHECK (balance_in_cents >= 0)
);

-- Tabela Geral de Transações (Agrupador)
CREATE TABLE IF NOT EXISTS transactions (
                              id UUID PRIMARY KEY,
                              idempotency_key VARCHAR(255) UNIQUE NOT NULL,
                              description TEXT,
                              created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Tabela Core do Ledger: Lançamentos Contábeis Imutáveis (Append-Only)
CREATE TABLE IF NOT EXISTS ledger_entries (
                                id UUID PRIMARY KEY,
                                transaction_id UUID NOT NULL REFERENCES transactions(id),
                                account_id UUID NOT NULL REFERENCES accounts_balance(account_id),
                                entry_type VARCHAR(10) NOT NULL, -- 'CREDIT' ou 'DEBIT'
                                amount_in_cents BIGINT NOT NULL,
                                created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                CONSTRAINT entry_type_enum CHECK (entry_type IN ('CREDIT', 'DEBIT')),
                                CONSTRAINT amount_positive CHECK (amount_in_cents > 0)
);

-- Índices Estratégicos para Alta Performance
CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_date ON ledger_entries(account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key);