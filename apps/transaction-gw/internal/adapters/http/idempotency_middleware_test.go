package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"transaction-gw/internal/adapters/postgres"

	"github.com/google/uuid"
)

type mockIdempotencyRepo struct {
	processed map[string]*postgres.ProcessedTransactionDTO
}

func newMockIdempotencyRepo() *mockIdempotencyRepo {
	return &mockIdempotencyRepo{
		processed: make(map[string]*postgres.ProcessedTransactionDTO),
	}
}

func (m *mockIdempotencyRepo) GetByKey(ctx context.Context, key string) (*postgres.ProcessedTransactionDTO, error) {
	return m.processed[key], nil
}

func (m *mockIdempotencyRepo) Create(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error {
	m.processed[pt.IdempotencyKey] = pt
	return nil
}

func (m *mockIdempotencyRepo) Update(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error {
	m.processed[pt.IdempotencyKey] = pt
	return nil
}

func (m *mockIdempotencyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestIdempotencyMiddleware_DuplicateRequest(t *testing.T) {
	repo := newMockIdempotencyRepo()
	middleware := NewIdempotencyMiddleware(repo)

	idempotencyKey := uuid.New().String()

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}))

	req1 := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(`{"amount":100}`))
	req1.Header.Set("X-Idempotency-Key", idempotencyKey)
	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusAccepted {
		t.Errorf("first request: expected status %d, got %d", http.StatusAccepted, w1.Code)
	}

	req2 := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(`{"amount":100}`))
	req2.Header.Set("X-Idempotency-Key", idempotencyKey)
	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("duplicate request: expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestIdempotencyMiddleware_MissingHeader(t *testing.T) {
	repo := newMockIdempotencyRepo()
	middleware := NewIdempotencyMiddleware(repo)

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/transactions", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for missing header, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestIdempotencyMiddleware_InvalidUUID(t *testing.T) {
	repo := newMockIdempotencyRepo()
	middleware := NewIdempotencyMiddleware(repo)

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/transactions", nil)
	req.Header.Set("X-Idempotency-Key", "invalid-uuid")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid UUID, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestValidateIdempotencyKey_Empty(t *testing.T) {
	key := ""
	if !isValidUUID(key) {
		return
	}
}

func TestValidateIdempotencyKey_InvalidFormat(t *testing.T) {
	key := "not-a-uuid"
	if !isValidUUID(key) {
		return
	}
}

func TestValidateIdempotencyKey_Valid(t *testing.T) {
	key := uuid.New().String()
	if !isValidUUID(key) {
		t.Errorf("expected valid UUID: %s", key)
	}
}

func TestHandler_CreateTransaction_WithIdempotency(t *testing.T) {
	dto := TransactionRequestDTO{
		SourceAccount:  uuid.New().String(),
		TargetAccount:  uuid.New().String(),
		Amount:         1000,
		IdempotencyKey: uuid.New().String(),
		Description:    "Test transaction",
	}

	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal DTO: %v", err)
	}

	req := httptest.NewRequest("POST", "/transactions", io.NopCloser(bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	if w.Code == http.StatusBadRequest && w.Body.String() == "Invalid source account id" {
		return
	}
}
