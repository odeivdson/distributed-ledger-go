package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccount_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   AccountStatus
		expected bool
	}{
		{
			name:     "Active account",
			status:   StatusActive,
			expected: true,
		},
		{
			name:     "Inactive account",
			status:   StatusInactive,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{
				ID:     uuid.New(),
				Status: tt.status,
			}
			if got := acc.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAccount_Credit_ValidateAmount(t *testing.T) {
	acc := &Account{
		ID:             uuid.New(),
		Status:         StatusActive,
		BalanceInCents: 1000,
		Version:        1,
		UpdatedAt:      time.Now(),
	}

	err := acc.Credit(100)
	if err != nil {
		t.Fatalf("Credit() with valid amount returned error: %v", err)
	}
	if acc.BalanceInCents != 1100 {
		t.Errorf("Balance after credit = %d, want %d", acc.BalanceInCents, 1100)
	}
}

func TestAccount_Debit_InsufficientBalance(t *testing.T) {
	acc := &Account{
		ID:             uuid.New(),
		Status:         StatusActive,
		BalanceInCents: 500,
		Version:        1,
		UpdatedAt:      time.Now(),
	}

	err := acc.Debit(1000)
	if err == nil {
		t.Fatal("Debit() with insufficient balance should return error")
	}
	if err != ErrInsufficientBalance {
		t.Errorf("Expected ErrInsufficientBalance, got %v", err)
	}
}

func TestAccount_Debit_ValidAmount(t *testing.T) {
	acc := &Account{
		ID:             uuid.New(),
		Status:         StatusActive,
		BalanceInCents: 1000,
		Version:        1,
		UpdatedAt:      time.Now(),
	}

	err := acc.Debit(500)
	if err != nil {
		t.Fatalf("Debit() with valid amount returned error: %v", err)
	}
	if acc.BalanceInCents != 500 {
		t.Errorf("Balance after debit = %d, want %d", acc.BalanceInCents, 500)
	}
}
