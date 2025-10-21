package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
	"rate-limiter/internal/middleware"
	"rate-limiter/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, *storage.MemoryStorage) {
	gin.SetMode(gin.TestMode)

	store := storage.NewMemoryStorage()

	cfg := &config.Config{
		RateLimit: struct {
			PerIP     int64
			BlockTime time.Duration
		}{
			PerIP:     3,
			BlockTime: 5 * time.Second,
		},
		APITokens: map[string]config.TokenConfig{
			"valid-token": {
				Token:     "valid-token",
				Limit:     10,
				BlockTime: 3 * time.Second,
			},
		},
	}

	rl := limiter.NewRateLimiter(store, cfg)

	router := gin.New()
	router.Use(middleware.RateLimiterMiddleware(rl))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return router, store
}

func TestMiddlewareIPRateLimit(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// Make requests within limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "you have reached the maximum number of requests")
}

func TestMiddlewareTokenRateLimit(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// Make requests with valid token (higher limit)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", "valid-token")
		req.RemoteAddr = "192.168.1.1:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("API_KEY", "valid-token")
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestMiddlewareHeaders(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check rate limit headers are set
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestMiddlewareInvalidToken(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// Make request with invalid token - should fall back to IP limiting
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", "invalid-token")
		req.RemoteAddr = "192.168.1.3:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Should be rate limited based on IP limit
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("API_KEY", "invalid-token")
	req.RemoteAddr = "192.168.1.3:12345"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
