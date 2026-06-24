package http

import (
	"encoding/json"
	"net/http"
	"transaction-gw/internal/domain"
	"transaction-gw/internal/usecase"

	"github.com/google/uuid"
)

type AccountHandler struct {
	usecase *usecase.CreateAccountUseCase
}

func NewAccountHandler(uc *usecase.CreateAccountUseCase) *AccountHandler {
	return &AccountHandler{usecase: uc}
}

type AccountRequestDTO struct {
	AccountID      string `json:"account_id"`
	BalanceInCents int64  `json:"balance_in_cents"`
}

// Create cria uma nova conta
// @Summary Criar conta
// @Description Cria uma nova conta no sistema com saldo inicial.
// @Tags accounts
// @Accept  json
// @Produce  json
// @Param account body AccountRequestDTO true "Detalhes da conta"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Requisição inválida"
// @Failure 422 {object} map[string]interface{} "Validação falhada"
// @Failure 500 {string} string "Erro interno"
// @Router /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var dto AccountRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(dto.AccountID)
	if err != nil {
		http.Error(w, "Invalid account id", http.StatusBadRequest)
		return
	}

	account := domain.NewAccount(id, dto.BalanceInCents)
	if err := h.usecase.Execute(r.Context(), account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"account_id": account.ID.String(),
		"status":     "created",
	})
}
