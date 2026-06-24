package http

import (
	"encoding/json"
	"net/http"
	"transaction-gw/internal/domain"
	"transaction-gw/internal/usecase"

	"github.com/google/uuid"
)

type TransactionHandler struct {
	usecase *usecase.SubmitTransactionUseCase
}

func NewTransactionHandler(uc *usecase.SubmitTransactionUseCase) *TransactionHandler {
	return &TransactionHandler{
		usecase: uc,
	}
}

type TransactionRequestDTO struct {
	SourceAccount  string `json:"source_account_id"`
	TargetAccount  string `json:"target_account_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
	Description    string `json:"description"`
}

// Handle envia uma nova transação para processamento
// @Summary Enviar transação
// @Description Envia uma nova transação financeira para o sistema através do gateway.
// @Tags transactions
// @Accept  json
// @Produce  json
// @Param X-Idempotency-Key header string true "Chave única UUID para idempotência"
// @Param transaction body TransactionRequestDTO true "Detalhes da transação"
// @Success 202 {object} map[string]string
// @Failure 400 {string} string "Requisição inválida ou header faltando"
// @Failure 422 {object} map[string]interface{} "Validação falhada"
// @Failure 429 {string} string "Rate limit excedido"
// @Failure 500 {string} string "Erro interno"
// @Router /transactions [post]
func (h *TransactionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var dto TransactionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if dto.IdempotencyKey == "" {
		dto.IdempotencyKey = r.Header.Get("X-Idempotency-Key")
	}

	if dto.IdempotencyKey == "" {
		http.Error(w, "idempotency_key is required", http.StatusBadRequest)
		return
	}

	sourceID, err := uuid.Parse(dto.SourceAccount)
	if err != nil {
		http.Error(w, "Invalid source account id", http.StatusBadRequest)
		return
	}

	targetID, err := uuid.Parse(dto.TargetAccount)
	if err != nil {
		http.Error(w, "Invalid target account id", http.StatusBadRequest)
		return
	}

	tx := domain.NewTransactionRequest(
		sourceID,
		targetID,
		dto.Amount,
		dto.IdempotencyKey,
		dto.Description,
	)

	if err := h.usecase.Execute(r.Context(), tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"transaction_id": tx.ID.String(),
		"status":         "accepted",
	})
}
