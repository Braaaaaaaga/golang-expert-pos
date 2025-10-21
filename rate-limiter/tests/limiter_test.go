package tests

import (
	"context"
	"testing"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
	"rate-limiter/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterIP(t *testing.T) {
	// Setup
	store := storage.NewMemoryStorage()
	defer store.Close()

	cfg := &config.Config{
		RateLimit: struct {
			PerIP     int64
			BlockTime time.Duration
		}{
			PerIP:     5,
			BlockTime: 10 * time.Second,
		},
		APITokens: make(map[string]config.TokenConfig),
	}

	rl := limiter.NewRateLimiter(store, cfg)
	ctx := context.Background()

	// Test normal requests within limit
	for i := 0; i < 5; i++ {
		result, err := rl.CheckIP(ctx, "192.168.1.1")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, int64(5-i-1), result.Remaining)
	}

	// Test rate limit exceeded
	result, err := rl.CheckIP(ctx, "192.168.1.1")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
	assert.True(t, result.RetryAfter > 0)
}

func TestRateLimiterToken(t *testing.T) {
	// Setup
	store := storage.NewMemoryStorage()
	defer store.Close()

	cfg := &config.Config{
		RateLimit: struct {
			PerIP     int64
			BlockTime time.Duration
		}{
			PerIP:     5,
			BlockTime: 10 * time.Second,
		},
		APITokens: map[string]config.TokenConfig{
			"test-token": {
				Token:     "test-token",
				Limit:     3,
				BlockTime: 5 * time.Second,
			},
		},
	}

	rl := limiter.NewRateLimiter(store, cfg)
	ctx := context.Background()

	// Test token-based limiting
	for i := 0; i < 3; i++ {
		result, err := rl.CheckToken(ctx, "test-token")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, int64(3-i-1), result.Remaining)
	}

	// Test token rate limit exceeded
	result, err := rl.CheckToken(ctx, "test-token")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
}

func TestRateLimiterTokenOverridesIP(t *testing.T) {
	// Setup
	store := storage.NewMemoryStorage()
	defer store.Close()

	cfg := &config.Config{
		RateLimit: struct {
			PerIP     int64
			BlockTime time.Duration
		}{
			PerIP:     2, // Low IP limit
			BlockTime: 10 * time.Second,
		},
		APITokens: map[string]config.TokenConfig{
			"high-limit-token": {
				Token:     "high-limit-token",
				Limit:     10, // Higher token limit
				BlockTime: 5 * time.Second,
			},
		},
	}

	rl := limiter.NewRateLimiter(store, cfg)
	ctx := context.Background()

	// First, exhaust IP limit without token
	for i := 0; i < 2; i++ {
		result, err := rl.CheckIP(ctx, "192.168.1.1")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	// IP limit should be exceeded
	result, err := rl.CheckIP(ctx, "192.168.1.1")
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// But with token, should still be allowed (token overrides IP)
	result, err = rl.Check(ctx, "192.168.1.1", "high-limit-token")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimiterInvalidToken(t *testing.T) {
	// Setup
	store := storage.NewMemoryStorage()
	defer store.Close()

	cfg := &config.Config{
		RateLimit: struct {
			PerIP     int64
			BlockTime time.Duration
		}{
			PerIP:     5,
			BlockTime: 10 * time.Second,
		},
		APITokens: make(map[string]config.TokenConfig),
	}

	rl := limiter.NewRateLimiter(store, cfg)
	ctx := context.Background()

	// Test invalid token
	result, err := rl.CheckToken(ctx, "invalid-token")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
}

func TestMemoryStorage(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	ctx := context.Background()

	// Test increment
	count, err := store.Increment(ctx, "test-key", time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test get
	value, err := store.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, int64(1), value)

	// Test increment again
	count, err = store.Increment(ctx, "test-key", time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test TTL
	ttl, err := store.TTL(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= time.Second)

	// Test set
	err = store.Set(ctx, "another-key", 42, time.Minute)
	require.NoError(t, err)

	value, err = store.Get(ctx, "another-key")
	require.NoError(t, err)
	assert.Equal(t, int64(42), value)
}
