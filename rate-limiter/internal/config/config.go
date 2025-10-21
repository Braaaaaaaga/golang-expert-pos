package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type TokenConfig struct {
	Token     string
	Limit     int64
	BlockTime time.Duration
}
type Config struct {
	Redis struct {
		Host     string
		Port     string
		Password string
		DB       int
	}
	RateLimit struct {
		PerIP     int64
		BlockTime time.Duration
	}
	Server struct {
		Port string
	}
	APITokens map[string]TokenConfig
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := &Config{}

	config.Redis.Host = getEnv("REDIS_HOST", "localhost")
	config.Redis.Port = getEnv("REDIS_PORT", "6379")
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")
	config.Redis.DB = getEnvInt("REDIS_DB", 0)

	config.RateLimit.PerIP = getEnvInt64("RATE_LIMIT_PER_IP", 10)
	config.RateLimit.BlockTime = time.Duration(getEnvInt("RATE_LIMIT_BLOCK_TIME_IP", 300)) * time.Second

	config.Server.Port = getEnv("SERVER_PORT", "8080")

	config.APITokens = parseAPITokens(getEnv("API_TOKENS", ""))

	return config
}

func parseAPITokens(tokensStr string) map[string]TokenConfig {
	tokens := make(map[string]TokenConfig)

	if tokensStr == "" {
		return tokens
	}

	tokenPairs := strings.Split(tokensStr, ",")
	for _, pair := range tokenPairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 3 {
			log.Printf("Warning: Invalid token format: %s", pair)
			continue
		}

		token := strings.TrimSpace(parts[0])
		limit, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			log.Printf("Warning: Invalid limit for token %s: %v", token, err)
			continue
		}

		blockTimeSeconds, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			log.Printf("Warning: Invalid block time for token %s: %v", token, err)
			continue
		}

		tokens[token] = TokenConfig{
			Token:     token,
			Limit:     limit,
			BlockTime: time.Duration(blockTimeSeconds) * time.Second,
		}
	}

	return tokens
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
