package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestTransactionRequest_Validate(t *testing.T) {
	source := uuid.New()
	target := uuid.New()

	tests := []struct {
		name    string
		tx      *TransactionRequest
		wantErr error
	}{
		{
			name: "valid transaction",
			tx: &TransactionRequest{
				SourceAccount:  source,
				TargetAccount:  target,
				AmountInCents:  100,
				IdempotencyKey: "key-1",
			},
			wantErr: nil,
		},
		{
			name: "missing source account",
			tx: &TransactionRequest{
				TargetAccount:  target,
				AmountInCents:  100,
				IdempotencyKey: "key-1",
			},
			wantErr: ErrAccountIDRequired,
		},
		{
			name: "same source and target",
			tx: &TransactionRequest{
				SourceAccount:  source,
				TargetAccount:  source,
				AmountInCents:  100,
				IdempotencyKey: "key-1",
			},
			wantErr: ErrInvalidAccount,
		},
		{
			name: "zero amount",
			tx: &TransactionRequest{
				SourceAccount:  source,
				TargetAccount:  target,
				AmountInCents:  0,
				IdempotencyKey: "key-1",
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "missing idempotency key",
			tx: &TransactionRequest{
				SourceAccount:  source,
				TargetAccount:  target,
				AmountInCents:  100,
			},
			wantErr: ErrIdempotencyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tx.Validate(); err != tt.wantErr {
				t.Errorf("TransactionRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
