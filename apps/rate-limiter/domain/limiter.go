package domain

import "errors"

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type RateLimitConfig struct {
	Limit  int
	Window int // em segundos
}
