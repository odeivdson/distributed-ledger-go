package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type MetricsMiddleware struct {
	registry *MetricsRegistry
}

func NewMetricsMiddleware(registry *MetricsRegistry) *MetricsMiddleware {
	return &MetricsMiddleware{registry: registry}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		m.registry.TransactionDuration.Observe(duration)
		m.registry.TransactionTotal.Inc()

		statusStr := strconv.Itoa(rw.statusCode)
		m.registry.TransactionStatus.WithLabelValues(statusStr, r.Method).Inc()

		if rw.statusCode >= 400 {
			m.registry.ErrorsTotal.WithLabelValues(statusStr).Inc()
		}
	})
}
