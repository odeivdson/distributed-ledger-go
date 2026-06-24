package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"transaction-gw/internal/adapters/postgres"

	"github.com/google/uuid"
)

type IdempotencyRepositoryI interface {
	GetByKey(ctx context.Context, key string) (*postgres.ProcessedTransactionDTO, error)
	Create(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error
	Update(ctx context.Context, pt *postgres.ProcessedTransactionDTO) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type IdempotencyMiddleware struct {
	repo IdempotencyRepositoryI
}

func NewIdempotencyMiddleware(repo IdempotencyRepositoryI) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{repo: repo}
}

type idempotencyResponseWriter struct {
	http.ResponseWriter
	status      int
	body        *bytes.Buffer
	wroteHeader bool
}

func (w *idempotencyResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (m *IdempotencyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get("X-Idempotency-Key")

		if idempotencyKey == "" {
			http.Error(w, "X-Idempotency-Key header is required", http.StatusBadRequest)
			return
		}

		if !isValidUUID(idempotencyKey) {
			http.Error(w, "X-Idempotency-Key must be a valid UUID", http.StatusBadRequest)
			return
		}

		existing, err := m.repo.GetByKey(r.Context(), idempotencyKey)
		if err != nil {
			http.Error(w, "failed to check idempotency", http.StatusInternalServerError)
			return
		}

		if existing != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(existing.ResponseBody))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		requestHash := computeRequestHash(body)

		pt := &postgres.ProcessedTransactionDTO{
			ID:             uuid.New(),
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
		}

		if err := m.repo.Create(r.Context(), pt); err != nil {
			http.Error(w, "failed to store idempotency key", http.StatusInternalServerError)
			return
		}

		rw := &idempotencyResponseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(rw, r)

		pt.ResponseStatus = http.StatusText(rw.status)
		pt.ResponseBody = rw.body.String()

		if err := m.repo.Update(r.Context(), pt); err != nil {
			return
		}
	})
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func computeRequestHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
