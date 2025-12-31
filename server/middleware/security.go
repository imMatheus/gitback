package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

// CreateRateLimiter creates a rate limiter with custom configuration
func CreateRateLimiter(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
				Error: "Too many requests. Please try again later.",
				Code:  "RATE_LIMIT_EXCEEDED",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		
		return c.Next()
	}
}

// InputValidation validates request inputs for security issues
func InputValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for common injection attempts
		body := string(c.Body())
		if containsMaliciousPatterns(body) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: "Invalid request format",
				Code:  "INVALID_INPUT",
			})
		}
		
		return c.Next()
	}
}

func containsMaliciousPatterns(input string) bool {
	maliciousPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"eval(",
		"document.cookie",
		"../",
		"..\\",
		"file://",
		"data:",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}