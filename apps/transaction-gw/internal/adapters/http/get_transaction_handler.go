package http

import (
	"context"
	"encoding/json"
	"net/http"
	"transaction-gw/internal/usecase"

	"github.com/google/uuid"
)

type GetTransactionUseCaseInterface interface {
	Execute(ctx context.Context, transactionID uuid.UUID) (*usecase.GetTransactionOutput, error)
}

type GetTransactionHandler struct {
	usecase GetTransactionUseCaseInterface
}

func NewGetTransactionHandler(uc GetTransactionUseCaseInterface) *GetTransactionHandler {
	return &GetTransactionHandler{usecase: uc}
}

// GetByID retorna os detalhes de uma transação
// @Summary Obter transação
// @Description Retorna os detalhes de uma transação específica pelo ID.
// @Tags transactions
// @Produce  json
// @Param id path string true "ID da transação (UUID)"
// @Success 200 {object} GetTransactionOutput
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Transação não encontrada"
// @Failure 500 {string} string "Erro interno"
// @Router /transactions/{id} [get]
func (h *GetTransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	transactionIDStr := r.PathValue("id")
	if transactionIDStr == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	output, err := h.usecase.Execute(r.Context(), transactionID)
	if err != nil {
		if err.Error() == "TRANSACTION_NOT_FOUND: transaction not found" {
			http.Error(w, "Transaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

type GetTransactionOutput struct {
	ID             string `json:"transaction_id"`
	SourceAccount  string `json:"source_account_id"`
	TargetAccount  string `json:"target_account_id"`
	Amount         int64  `json:"amount"`
	Description    string `json:"description"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt      string `json:"created_at"`
}
