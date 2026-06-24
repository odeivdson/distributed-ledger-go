package local

import (
	"context"
	"rate-limiter/domain"
	"sync"
	"time"
)

type item struct {
	count     int
	expiresAt time.Time
}

type InMemoryRateLimiter struct {
	store sync.Map
}

func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{}
}

func (l *InMemoryRateLimiter) Allow(ctx context.Context, key string, config domain.RateLimitConfig) (bool, error) {
	now := time.Now()
	
	val, ok := l.store.Load(key)
	if !ok {
		l.store.Store(key, &item{
			count:     1,
			expiresAt: now.Add(time.Duration(config.Window) * time.Second),
		})
		return true, nil
	}

	it := val.(*item)
	if now.After(it.expiresAt) {
		it.count = 1
		it.expiresAt = now.Add(time.Duration(config.Window) * time.Second)
		return true, nil
	}

	if it.count >= config.Limit {
		return false, nil
	}

	it.count++
	return true, nil
}
