package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"shared/validation"

	"github.com/google/uuid"
)

func TestValidatingHandler_ValidTransaction(t *testing.T) {
	source := uuid.New()
	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := validation.TransactionRequestDTO{
		SourceAccount:  source.String(),
		TargetAccount:  target.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
		Description:    "test payment",
	}

	body, _ := json.Marshal(dto)

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	validator := validation.NewTransactionRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
}

func TestValidatingHandler_InvalidTransaction_MissingAmount(t *testing.T) {
	source := uuid.New()
	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := validation.TransactionRequestDTO{
		SourceAccount:  source.String(),
		TargetAccount:  target.String(),
		Amount:         0, // Invalid
		IdempotencyKey: idempotencyKey.String(),
		Description:    "test payment",
	}

	body, _ := json.Marshal(dto)

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	validator := validation.NewTransactionRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("POST", "/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	var result validation.ValidationResult
	json.NewDecoder(w.Body).Decode(&result)

	if result.Valid {
		t.Errorf("expected invalid result")
	}

	if len(result.Errors) == 0 {
		t.Errorf("expected at least one error")
	}
}

func TestValidatingHandler_InvalidJSON(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	validator := validation.NewTransactionRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("POST", "/transactions", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)

	if result["valid"].(bool) {
		t.Errorf("expected invalid result")
	}
}

func TestValidatingHandler_ValidAccount(t *testing.T) {
	accountID := uuid.New()

	dto := validation.AccountRequestDTO{
		AccountID:      accountID.String(),
		BalanceInCents: 100000,
	}

	body, _ := json.Marshal(dto)

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	validator := validation.NewAccountRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestValidatingHandler_InvalidAccount_NegativeBalance(t *testing.T) {
	accountID := uuid.New()

	dto := validation.AccountRequestDTO{
		AccountID:      accountID.String(),
		BalanceInCents: -100000,
	}

	body, _ := json.Marshal(dto)

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	validator := validation.NewAccountRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestValidatingHandler_GetRequest_SkipsValidation(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	validator := validation.NewTransactionRequestValidator()
	handler := NewValidatingHandler(validator, innerHandler)

	req := httptest.NewRequest("GET", "/transactions/123", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
