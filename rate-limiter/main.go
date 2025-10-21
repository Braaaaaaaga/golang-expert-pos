package main

import (
	"context"
	"log"
	"net/http"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
	"rate-limiter/internal/middleware"
	"rate-limiter/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting Rate Limiter server on port %s", cfg.Server.Port)
	log.Printf("Rate limit per IP: %d req/s, Block time: %v", cfg.RateLimit.PerIP, cfg.RateLimit.BlockTime)
	log.Printf("Configured API tokens: %d", len(cfg.APITokens))

	var store storage.Storage

	redisStore := storage.NewRedisStorage(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	ctx := context.Background()

	if err := redisStore.Ping(ctx); err != nil {
		log.Printf("Redis connection failed, using in-memory storage: %v", err)
		store = storage.NewMemoryStorage()
	} else {
		log.Println("Connected to Redis successfully")
		store = redisStore
	}

	rateLimiter := limiter.NewRateLimiter(store, cfg)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Rate limiter is running",
		})
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Request successful",
			"ip":      c.ClientIP(),
			"token":   c.GetHeader("API_KEY"),
		})
	})

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
