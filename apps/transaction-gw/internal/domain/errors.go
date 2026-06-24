package domain

type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

var (
	ErrInvalidIdempotencyKey = &ValidationError{
		Code:    "INVALID_IDEMPOTENCY_KEY",
		Message: "idempotency key is invalid or missing",
	}

	ErrDuplicateTransaction = &ValidationError{
		Code:    "DUPLICATE_TRANSACTION",
		Message: "transaction with this idempotency key was already processed",
	}

	ErrAccountNotActive = &ValidationError{
		Code:    "ACCOUNT_INACTIVE",
		Message: "account is not active",
	}

	ErrInsufficientBalance = &ValidationError{
		Code:    "INSUFFICIENT_BALANCE",
		Message: "account has insufficient balance",
	}

	ErrAccountNotFound = &ValidationError{
		Code:    "ACCOUNT_NOT_FOUND",
		Message: "account not found",
	}

	ErrTransactionNotFound = &ValidationError{
		Code:    "TRANSACTION_NOT_FOUND",
		Message: "transaction not found",
	}
)
