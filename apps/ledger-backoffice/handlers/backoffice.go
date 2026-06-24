package handlers

import (
	"fmt"
	"html/template"
	"ledger-backoffice/repository"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
)

type BackofficeHandler struct {
	repo *repository.ReaderRepository
}

func NewBackofficeHandler(repo *repository.ReaderRepository) *BackofficeHandler {
	return &BackofficeHandler{repo: repo}
}

func (h *BackofficeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetDashboardData(r.Context())
	if err != nil {
		h.renderError(w, "Erro ao carregar dashboard", err)
		return
	}

	h.render(w, "dashboard.html", data)
}

func (h *BackofficeHandler) AccountDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	acc, err := h.repo.GetAccountDetail(r.Context(), id)
	if err != nil {
		h.renderError(w, "Erro ao carregar conta", err)
		return
	}

	if acc == nil {
		h.renderError(w, fmt.Sprintf("Conta %s não encontrada", id), nil)
		return
	}

	entries, err := h.repo.GetAuditTrail(r.Context(), id)
	if err != nil {
		h.renderError(w, "Erro ao carregar rastro de auditoria", err)
		return
	}

	h.render(w, "account.html", map[string]interface{}{
		"Account": acc,
		"Entries": entries,
	})
}

func (h *BackofficeHandler) DLQ(w http.ResponseWriter, r *http.Request) {
	volume, err := h.repo.GetDLQVolume(r.Context())
	if err != nil {
		h.renderError(w, "Erro ao carregar DLQ", err)
		return
	}

	h.render(w, "dlq.html", map[string]interface{}{
		"DLQVolume": volume,
	})
}

func (h *BackofficeHandler) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.New("layout").Funcs(template.FuncMap{
		"formatCents": func(cents int64) string {
			return fmt.Sprintf("%.2f", float64(cents)/100.0)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02/01/2006 15:04:05.000")
		},
	}).ParseFiles(
		filepath.Join("apps", "ledger-backoffice", "web", "templates", "layout.html"),
		filepath.Join("apps", "ledger-backoffice", "web", "templates", name),
	)

	if err != nil {
		slog.Error("Erro ao processar template", "error", err)
		http.Error(w, "Erro interno de renderização", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		slog.Error("Erro ao executar template", "error", err)
	}
}

func (h *BackofficeHandler) renderError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	http.Error(w, fmt.Sprintf("%s: %v", msg, err), http.StatusInternalServerError)
}
