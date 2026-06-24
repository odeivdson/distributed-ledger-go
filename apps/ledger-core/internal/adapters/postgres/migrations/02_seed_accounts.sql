-- Criar algumas contas para teste inicial
INSERT INTO accounts_balance (account_id, balance_in_cents, version)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 100000, 0), -- Conta com R$ 1000,00
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 50000, 0)   -- Conta com R$ 500,00
ON CONFLICT (account_id) DO NOTHING;
