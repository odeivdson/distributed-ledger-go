package http

import (
	"log/slog"
	"net"
	"net/http"
	"rate-limiter/domain"
	"rate-limiter/ports"
)

type Middleware struct {
	limiter ports.RateLimiter
	config  domain.RateLimitConfig
}

func NewMiddleware(limiter ports.RateLimiter, config domain.RateLimitConfig) *Middleware {
	return &Middleware{
		limiter: limiter,
		config:  config,
	}
}

func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		allowed, err := m.limiter.Allow(r.Context(), ip, m.config)
		if err != nil {
			slog.Error("Erro ao processar rate limit", "error", err, "ip", ip)
			// Em caso de erro catastrófico em ambos os limitadores, permitimos a requisição (fail-open)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
