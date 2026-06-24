package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf("host=localhost port=5432 user=postgres password=postgres dbname=ledger_test sslmode=disable")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	cleanupTestDB(t, db)

	query := `
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

		CREATE INDEX IF NOT EXISTS idx_processed_transactions_idempotency_key ON processed_transactions(idempotency_key);
		CREATE INDEX IF NOT EXISTS idx_processed_transactions_expires_at ON processed_transactions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_processed_transactions_status ON processed_transactions(response_status);
	`

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS processed_transactions;")
}

func TestIdempotencyRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIdempotencyRepository(db)
	ctx := context.Background()

	pt := &ProcessedTransactionDTO{
		ID:             uuid.New(),
		IdempotencyKey: uuid.New().String(),
		RequestHash:    "test-hash",
		ResponseStatus: "ACCEPTED",
		ResponseBody:   `{"status":"accepted"}`,
	}

	err := repo.Create(ctx, pt)
	if err != nil {
		t.Fatalf("failed to create processed transaction: %v", err)
	}

	retrieved, err := repo.GetByKey(ctx, pt.IdempotencyKey)
	if err != nil {
		t.Fatalf("failed to retrieve processed transaction: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected processed transaction, got nil")
	}

	if retrieved.IdempotencyKey != pt.IdempotencyKey {
		t.Errorf("expected idempotency key %s, got %s", pt.IdempotencyKey, retrieved.IdempotencyKey)
	}
}

func TestIdempotencyRepository_GetByKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIdempotencyRepository(db)
	ctx := context.Background()

	retrieved, err := repo.GetByKey(ctx, "non-existent-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved != nil {
		t.Fatal("expected nil for non-existent key")
	}
}

func TestIdempotencyRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIdempotencyRepository(db)
	ctx := context.Background()

	pt := &ProcessedTransactionDTO{
		ID:             uuid.New(),
		IdempotencyKey: uuid.New().String(),
		RequestHash:    "test-hash",
		ResponseStatus: "PENDING",
	}

	err := repo.Create(ctx, pt)
	if err != nil {
		t.Fatalf("failed to create processed transaction: %v", err)
	}

	pt.ResponseStatus = "SUCCESS"
	pt.ResponseBody = `{"status":"success"}`

	err = repo.Update(ctx, pt)
	if err != nil {
		t.Fatalf("failed to update processed transaction: %v", err)
	}

	retrieved, err := repo.GetByKey(ctx, pt.IdempotencyKey)
	if err != nil {
		t.Fatalf("failed to retrieve updated transaction: %v", err)
	}

	if retrieved.ResponseStatus != "SUCCESS" {
		t.Errorf("expected status SUCCESS, got %s", retrieved.ResponseStatus)
	}
}

func TestIdempotencyRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewIdempotencyRepository(db)
	ctx := context.Background()

	pt := &ProcessedTransactionDTO{
		ID:             uuid.New(),
		IdempotencyKey: uuid.New().String(),
		RequestHash:    "test-hash",
		ResponseStatus: "ACCEPTED",
	}

	err := repo.Create(ctx, pt)
	if err != nil {
		t.Fatalf("failed to create expired transaction: %v", err)
	}

	count, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("failed to delete expired transactions: %v", err)
	}

	if count != 0 {
		t.Errorf("expected to delete 0 non-expired transactions, deleted %d", count)
	}
}

func TestComputeRequestHash(t *testing.T) {
	data := []byte("test request data")
	hash := ComputeRequestHash(data)

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	hash2 := ComputeRequestHash(data)
	if hash != hash2 {
		t.Error("expected same hash for same data")
	}

	hash3 := ComputeRequestHash([]byte("different data"))
	if hash == hash3 {
		t.Error("expected different hash for different data")
	}
}
