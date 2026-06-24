package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"shared/validation"
)

type ValidatingHandler struct {
	validator validation.Validator
	handler   http.Handler
}

func NewValidatingHandler(validator validation.Validator, handler http.Handler) *ValidatingHandler {
	return &ValidatingHandler{
		validator: validator,
		handler:   handler,
	}
}

func (vh *ValidatingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		vh.handler.ServeHTTP(w, r)
		return
	}

	var dto interface{}
	if r.URL.Path == "/transactions" {
		dto = &validation.TransactionRequestDTO{}
	} else if r.URL.Path == "/accounts" {
		dto = &validation.AccountRequestDTO{}
	} else {
		vh.handler.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":  false,
			"errors": []map[string]string{
				{
					"field":   "body",
					"message": "failed to read request body",
					"code":    "READ_ERROR",
				},
			},
		})
		return
	}

	if err := json.Unmarshal(body, dto); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":  false,
			"errors": []map[string]string{
				{
					"field":   "body",
					"message": "invalid JSON format",
					"code":    "INVALID_JSON",
				},
			},
		})
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if txDto, ok := dto.(*validation.TransactionRequestDTO); ok {
		if txDto.IdempotencyKey == "" {
			txDto.IdempotencyKey = r.Header.Get("X-Idempotency-Key")
		}
	}

	result := vh.validator.Validate(dto)
	if !result.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(result)
		return
	}

	vh.handler.ServeHTTP(w, r)
}
