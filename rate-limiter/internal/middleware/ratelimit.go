package middleware

import (
	"net"
	"net/http"
	"strings"

	"rate-limiter/internal/limiter"

	"github.com/gin-gonic/gin"
)

func RateLimiterMiddleware(rateLimiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		token := c.GetHeader("API_KEY")

		result, err := rateLimiter.Check(c.Request.Context(), ip, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
		c.Header("X-RateLimit-Reset", result.ResetTime.Format("2006-01-02T15:04:05Z07:00"))

		if !result.Allowed {
			c.Header("Retry-After", string(rune(int(result.RetryAfter.Seconds()))))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" && isValidIP(ip) {
				return ip
			}
		}
	}

	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && isValidIP(xRealIP) {
		return xRealIP
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
