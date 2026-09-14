package middleware

import (
	"fmt"
	"news-portal-backend/handlers"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected â€” JWT Middleware
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("auth_token")
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Akses ditolak. Token tidak ditemukan."})
		}

		token, err := jwt.ParseWithClaims(tokenString, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("edutech-super-secret-key-change-in-production-2026"), nil // TODO: Env
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Token tidak valid atau kedaluwarsa."})
		}

		claims, ok := token.Claims.(*handlers.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Klaim token tidak valid."})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

// AdminOnly â€” Role Guard
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(*handlers.Claims)
		if !ok || claims.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Akses ditolak. Hanya untuk administrator."})
		}
		return c.Next()
	}
}