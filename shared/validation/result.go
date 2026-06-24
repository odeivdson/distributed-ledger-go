package validation

import (
	"encoding/json"
	"fmt"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ValidationResult struct {
	Valid  bool               `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

func (vr *ValidationResult) Error() string {
	if vr.Valid {
		return ""
	}
	if len(vr.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation error: %s", vr.Errors[0].Message)
}

func (vr *ValidationResult) HasErrors() bool {
	return !vr.Valid && len(vr.Errors) > 0
}

func (vr *ValidationResult) AddError(field, message, code string) {
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	vr.Valid = false
}

func (vr *ValidationResult) ToJSON() []byte {
	data, _ := json.Marshal(vr)
	return data
}

func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}
}

type Validator interface {
	Validate(data interface{}) *ValidationResult
}
