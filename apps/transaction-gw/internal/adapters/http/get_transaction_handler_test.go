package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"transaction-gw/internal/usecase"

	"github.com/google/uuid"
)

type MockGetTransactionUseCaseImpl struct {
	output *usecase.GetTransactionOutput
	err    error
}

func (m *MockGetTransactionUseCaseImpl) Execute(ctx context.Context, transactionID uuid.UUID) (*usecase.GetTransactionOutput, error) {
	return m.output, m.err
}

func TestGetTransactionHandler_UseCaseIntegration(t *testing.T) {
	txID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()

	uc := &MockGetTransactionUseCaseImpl{
		output: &usecase.GetTransactionOutput{
			ID:             txID.String(),
			SourceAccount:  sourceID.String(),
			TargetAccount:  targetID.String(),
			Amount:         5000,
			Description:    "test",
			IdempotencyKey: "test-key",
			CreatedAt:      "2026-06-24T14:30:00Z",
		},
	}

	handler := NewGetTransactionHandler(uc)
	if handler == nil {
		t.Errorf("handler should not be nil")
	}
}

func TestGetTransactionHandler_CreationWithInterface(t *testing.T) {
	var uci GetTransactionUseCaseInterface
	uc := &MockGetTransactionUseCaseImpl{}
	uci = uc

	handler := NewGetTransactionHandler(uci)
	if handler == nil {
		t.Errorf("handler should not be nil")
	}
}

func TestMockGetTransactionUseCase_Execute(t *testing.T) {
	txID := uuid.New()
	
	mock := &MockGetTransactionUseCaseImpl{
		output: &usecase.GetTransactionOutput{
			ID: txID.String(),
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	output, err := mock.Execute(req.Context(), txID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if output == nil {
		t.Errorf("output should not be nil")
	}
	if output.ID != txID.String() {
		t.Errorf("ID mismatch: expected %s, got %s", txID.String(), output.ID)
	}
}
