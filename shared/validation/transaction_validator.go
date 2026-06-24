package validation

import (
	"regexp"

	"github.com/google/uuid"
)

type TransactionRequestValidator struct{}

func NewTransactionRequestValidator() *TransactionRequestValidator {
	return &TransactionRequestValidator{}
}

type TransactionRequestDTO struct {
	SourceAccount  string `json:"source_account_id"`
	TargetAccount  string `json:"target_account_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
	Description    string `json:"description"`
}

func (v *TransactionRequestValidator) Validate(data interface{}) *ValidationResult {
	result := NewValidationResult()

	dto, ok := data.(*TransactionRequestDTO)
	if !ok {
		result.AddError("", "invalid data type", "INVALID_TYPE")
		return result
	}

	v.validateSourceAccount(dto.SourceAccount, result)
	v.validateTargetAccount(dto.TargetAccount, result)
	v.validateAmount(dto.Amount, result)
	v.validateIdempotencyKey(dto.IdempotencyKey, result)
	v.validateDescription(dto.Description, result)

	if result.HasErrors() {
		return result
	}

	v.validateBusinessRules(dto, result)
	return result
}

func (v *TransactionRequestValidator) validateSourceAccount(source string, result *ValidationResult) {
	if source == "" {
		result.AddError("source_account_id", "source account ID is required", "MISSING_SOURCE_ACCOUNT")
		return
	}

	if _, err := uuid.Parse(source); err != nil {
		result.AddError("source_account_id", "source account ID must be a valid UUID", "INVALID_SOURCE_ACCOUNT_FORMAT")
	}
}

func (v *TransactionRequestValidator) validateTargetAccount(target string, result *ValidationResult) {
	if target == "" {
		result.AddError("target_account_id", "target account ID is required", "MISSING_TARGET_ACCOUNT")
		return
	}

	if _, err := uuid.Parse(target); err != nil {
		result.AddError("target_account_id", "target account ID must be a valid UUID", "INVALID_TARGET_ACCOUNT_FORMAT")
	}
}

func (v *TransactionRequestValidator) validateAmount(amount int64, result *ValidationResult) {
	if amount == 0 {
		result.AddError("amount", "amount is required", "MISSING_AMOUNT")
		return
	}

	if amount < 0 {
		result.AddError("amount", "amount must be positive", "NEGATIVE_AMOUNT")
		return
	}

	if amount <= 0 {
		result.AddError("amount", "amount must be greater than zero", "INVALID_AMOUNT")
	}
}

func (v *TransactionRequestValidator) validateIdempotencyKey(key string, result *ValidationResult) {
	if key == "" {
		result.AddError("idempotency_key", "idempotency key is required", "MISSING_IDEMPOTENCY_KEY")
		return
	}

	if _, err := uuid.Parse(key); err != nil {
		result.AddError("idempotency_key", "idempotency key must be a valid UUID", "INVALID_IDEMPOTENCY_KEY_FORMAT")
	}

	if len(key) > 255 {
		result.AddError("idempotency_key", "idempotency key must not exceed 255 characters", "IDEMPOTENCY_KEY_TOO_LONG")
	}
}

func (v *TransactionRequestValidator) validateDescription(desc string, result *ValidationResult) {
	if len(desc) > 1000 {
		result.AddError("description", "description must not exceed 1000 characters", "DESCRIPTION_TOO_LONG")
	}

	if desc != "" && !isValidDescription(desc) {
		result.AddError("description", "description contains invalid characters", "INVALID_DESCRIPTION")
	}
}

func (v *TransactionRequestValidator) validateBusinessRules(dto *TransactionRequestDTO, result *ValidationResult) {
	source, _ := uuid.Parse(dto.SourceAccount)
	target, _ := uuid.Parse(dto.TargetAccount)

	if source == target {
		result.AddError("", "source and target accounts must be different", "SAME_ACCOUNTS")
	}
}

func isValidDescription(desc string) bool {
	// Aceita apenas caracteres alfanuméricos, espaço, pontuação comum
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9\s\-._,;:()]+$`)
	return validPattern.MatchString(desc)
}

type AccountRequestValidator struct{}

func NewAccountRequestValidator() *AccountRequestValidator {
	return &AccountRequestValidator{}
}

type AccountRequestDTO struct {
	AccountID      string `json:"account_id"`
	BalanceInCents int64  `json:"balance_in_cents"`
}

func (v *AccountRequestValidator) Validate(data interface{}) *ValidationResult {
	result := NewValidationResult()

	dto, ok := data.(*AccountRequestDTO)
	if !ok {
		result.AddError("", "invalid data type", "INVALID_TYPE")
		return result
	}

	v.validateAccountID(dto.AccountID, result)
	v.validateBalance(dto.BalanceInCents, result)

	return result
}

func (v *AccountRequestValidator) validateAccountID(accountID string, result *ValidationResult) {
	if accountID == "" {
		result.AddError("account_id", "account ID is required", "MISSING_ACCOUNT_ID")
		return
	}

	if _, err := uuid.Parse(accountID); err != nil {
		result.AddError("account_id", "account ID must be a valid UUID", "INVALID_ACCOUNT_ID_FORMAT")
	}
}

func (v *AccountRequestValidator) validateBalance(balance int64, result *ValidationResult) {
	if balance < 0 {
		result.AddError("balance_in_cents", "balance must not be negative", "NEGATIVE_BALANCE")
		return
	}

	if balance > 9223372036854775807 { // Max int64
		result.AddError("balance_in_cents", "balance exceeds maximum value", "BALANCE_OVERFLOW")
	}
}
