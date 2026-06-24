package validation

import (
	"testing"

	"github.com/google/uuid"
)

func TestTransactionRequestValidator_Valid(t *testing.T) {
	validator := NewTransactionRequestValidator()

	source := uuid.New()
	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := &TransactionRequestDTO{
		SourceAccount:  source.String(),
		TargetAccount:  target.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
		Description:    "payment for services",
	}

	result := validator.Validate(dto)

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestTransactionRequestValidator_MissingSourceAccount(t *testing.T) {
	validator := NewTransactionRequestValidator()

	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := &TransactionRequestDTO{
		SourceAccount:  "",
		TargetAccount:  target.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for missing source account")
	}

	if len(result.Errors) == 0 {
		t.Errorf("expected at least one error")
	}

	if result.Errors[0].Code != "MISSING_SOURCE_ACCOUNT" {
		t.Errorf("expected error code MISSING_SOURCE_ACCOUNT, got %s", result.Errors[0].Code)
	}
}

func TestTransactionRequestValidator_InvalidSourceAccount(t *testing.T) {
	validator := NewTransactionRequestValidator()

	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := &TransactionRequestDTO{
		SourceAccount:  "not-a-uuid",
		TargetAccount:  target.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for invalid source account")
	}

	if result.Errors[0].Code != "INVALID_SOURCE_ACCOUNT_FORMAT" {
		t.Errorf("expected error code INVALID_SOURCE_ACCOUNT_FORMAT, got %s", result.Errors[0].Code)
	}
}

func TestTransactionRequestValidator_InvalidAmount(t *testing.T) {
	validator := NewTransactionRequestValidator()

	source := uuid.New()
	target := uuid.New()
	idempotencyKey := uuid.New()

	dto := &TransactionRequestDTO{
		SourceAccount:  source.String(),
		TargetAccount:  target.String(),
		Amount:         -5000,
		IdempotencyKey: idempotencyKey.String(),
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for negative amount")
	}

	if result.Errors[0].Code != "NEGATIVE_AMOUNT" {
		t.Errorf("expected error code NEGATIVE_AMOUNT, got %s", result.Errors[0].Code)
	}
}

func TestTransactionRequestValidator_SameSourceAndTarget(t *testing.T) {
	validator := NewTransactionRequestValidator()

	account := uuid.New()
	idempotencyKey := uuid.New()

	dto := &TransactionRequestDTO{
		SourceAccount:  account.String(),
		TargetAccount:  account.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for same source and target accounts")
	}

	foundError := false
	for _, err := range result.Errors {
		if err.Code == "SAME_ACCOUNTS" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("expected SAME_ACCOUNTS error")
	}
}

func TestTransactionRequestValidator_LongDescription(t *testing.T) {
	validator := NewTransactionRequestValidator()

	source := uuid.New()
	target := uuid.New()
	idempotencyKey := uuid.New()

	longDesc := ""
	for i := 0; i < 1001; i++ {
		longDesc += "a"
	}

	dto := &TransactionRequestDTO{
		SourceAccount:  source.String(),
		TargetAccount:  target.String(),
		Amount:         5000,
		IdempotencyKey: idempotencyKey.String(),
		Description:    longDesc,
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for long description")
	}

	if result.Errors[0].Code != "DESCRIPTION_TOO_LONG" {
		t.Errorf("expected error code DESCRIPTION_TOO_LONG, got %s", result.Errors[0].Code)
	}
}

func TestAccountRequestValidator_Valid(t *testing.T) {
	validator := NewAccountRequestValidator()

	accountID := uuid.New()

	dto := &AccountRequestDTO{
		AccountID:      accountID.String(),
		BalanceInCents: 100000,
	}

	result := validator.Validate(dto)

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestAccountRequestValidator_MissingAccountID(t *testing.T) {
	validator := NewAccountRequestValidator()

	dto := &AccountRequestDTO{
		AccountID:      "",
		BalanceInCents: 100000,
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for missing account ID")
	}

	if result.Errors[0].Code != "MISSING_ACCOUNT_ID" {
		t.Errorf("expected error code MISSING_ACCOUNT_ID, got %s", result.Errors[0].Code)
	}
}

func TestAccountRequestValidator_NegativeBalance(t *testing.T) {
	validator := NewAccountRequestValidator()

	accountID := uuid.New()

	dto := &AccountRequestDTO{
		AccountID:      accountID.String(),
		BalanceInCents: -100000,
	}

	result := validator.Validate(dto)

	if result.Valid {
		t.Errorf("expected invalid result for negative balance")
	}

	if result.Errors[0].Code != "NEGATIVE_BALANCE" {
		t.Errorf("expected error code NEGATIVE_BALANCE, got %s", result.Errors[0].Code)
	}
}
