package redis

import (
	"context"
	"rate-limiter/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call('get', key)

if current and tonumber(current) >= limit then
    return 0
else
    if not current then
        redis.call('set', key, 1)
        redis.call('expire', key, window)
    else
        redis.call('incr', key)
    end
    return 1
end
`

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, config domain.RateLimitConfig) (bool, error) {
	// Timeout agressivo para não travar a aplicação em caso de lentidão no Redis
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	result, err := r.client.Eval(ctx, rateLimitScript, []string{key}, config.Limit, config.Window).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
