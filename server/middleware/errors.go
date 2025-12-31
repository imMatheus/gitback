package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// RecoveryMiddleware recovers from panics and returns proper error responses
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC: %v\nStack trace:\n%s", r, debug.Stack())
				
				err := c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Error: "Internal server error",
					Code:  "INTERNAL_ERROR",
				})
				if err != nil {
					log.Printf("Failed to send error response: %v", err)
				}
			}
		}()
		return c.Next()
	}
}

// ValidationError creates a validation error response
func ValidationError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error: message,
		Code:  "VALIDATION_ERROR",
	})
}

// NotFoundError creates a not found error response
func NotFoundError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Error: message,
		Code:  "NOT_FOUND",
	})
}

// InternalError creates an internal server error response
func InternalError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: message,
		Code:  "INTERNAL_ERROR",
	})
}

// TimeoutError creates a timeout error response
func TimeoutError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusRequestTimeout).JSON(ErrorResponse{
		Error: message,
		Code:  "TIMEOUT",
	})
}