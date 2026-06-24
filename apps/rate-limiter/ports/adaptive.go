package ports

import (
	"context"
	"log/slog"
	"rate-limiter/domain"
)

type AdaptiveRateLimiter struct {
	primary   RateLimiter
	secondary RateLimiter
}

func NewAdaptiveRateLimiter(primary, secondary RateLimiter) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		primary:   primary,
		secondary: secondary,
	}
}

func (a *AdaptiveRateLimiter) Allow(ctx context.Context, key string, config domain.RateLimitConfig) (bool, error) {
	allowed, err := a.primary.Allow(ctx, key, config)
	if err != nil {
		slog.Warn("Rate Limiter primário falhou, chaveando para secundário", "error", err)
		return a.secondary.Allow(ctx, key, config)
	}
	return allowed, nil
}
