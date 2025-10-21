package limiter

import (
	"context"
	"fmt"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/storage"
)

type RateLimiter struct {
	storage storage.Storage
	config  *config.Config
}

type LimitResult struct {
	Allowed    bool
	Remaining  int64
	ResetTime  time.Time
	RetryAfter time.Duration
}

func NewRateLimiter(storage storage.Storage, config *config.Config) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		config:  config,
	}
}

func (rl *RateLimiter) CheckIP(ctx context.Context, ip string) (*LimitResult, error) {
	key := fmt.Sprintf("ip:%s", ip)
	limit := rl.config.RateLimit.PerIP
	blockTime := rl.config.RateLimit.BlockTime

	return rl.checkLimit(ctx, key, limit, blockTime)
}

func (rl *RateLimiter) CheckToken(ctx context.Context, token string) (*LimitResult, error) {
	tokenConfig, exists := rl.config.APITokens[token]
	if !exists {
		return &LimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now(),
			RetryAfter: time.Hour,
		}, nil
	}

	key := fmt.Sprintf("token:%s", token)
	return rl.checkLimit(ctx, key, tokenConfig.Limit, tokenConfig.BlockTime)
}

func (rl *RateLimiter) checkLimit(ctx context.Context, key string, limit int64, blockTime time.Duration) (*LimitResult, error) {
	blockedKey := key + ":blocked"
	ttl, err := rl.storage.TTL(ctx, blockedKey)
	if err != nil {
		return nil, fmt.Errorf("error checking blocked status: %w", err)
	}

	if ttl > 0 {
		return &LimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(ttl),
			RetryAfter: ttl,
		}, nil
	}
	count, err := rl.storage.Increment(ctx, key, time.Second)
	if err != nil {
		return nil, fmt.Errorf("error incrementing counter: %w", err)
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	if count > limit {
		if err := rl.storage.Set(ctx, blockedKey, 1, blockTime); err != nil {
			return nil, fmt.Errorf("error setting block: %w", err)
		}

		return &LimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(blockTime),
			RetryAfter: blockTime,
		}, nil
	}

	keyTTL, err := rl.storage.TTL(ctx, key)
	if err != nil {
		keyTTL = time.Second
	}

	return &LimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetTime:  time.Now().Add(keyTTL),
		RetryAfter: 0,
	}, nil
}

func (rl *RateLimiter) Check(ctx context.Context, ip, token string) (*LimitResult, error) {
	if token != "" {
		if _, exists := rl.config.APITokens[token]; exists {
			return rl.CheckToken(ctx, token)
		}
	}

	return rl.CheckIP(ctx, ip)
}
