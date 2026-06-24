package ports

import (
	"context"
	"rate-limiter/domain"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, config domain.RateLimitConfig) (bool, error)
}
