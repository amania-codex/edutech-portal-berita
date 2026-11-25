package middleware

import (
	"fmt"
	"news-portal-backend/handlers"
	"news-portal-backend/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected — JWT Middleware
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
			return handlers.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Token tidak valid atau kedaluwarsa."})
		}

		claims, ok := token.Claims.(*handlers.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Klaim token tidak valid."})
		}

		// Peran terbaru & status akun selalu dibaca dari database
		if !handlers.RefreshClaims(claims) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Akun tidak aktif. Hubungi admin."})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

// AdminOnly — Role Guard
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(*handlers.Claims)
		if !ok || claims.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Akses ditolak. Hanya untuk administrator."})
		}
		return c.Next()
	}
}

// StaffOnly — admin, editor, dan penulis (staf redaksi)
func StaffOnly() fiber.Handler {
	return roleGuard("Akses ditolak. Hanya untuk tim redaksi.", models.RoleAdmin, models.RoleEditor, models.RolePenulis)
}

// EditorOnly — admin dan editor
func EditorOnly() fiber.Handler {
	return roleGuard("Akses ditolak. Hanya untuk admin dan editor.", models.RoleAdmin, models.RoleEditor)
}

func roleGuard(message string, roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(*handlers.Claims)
		if ok {
			for _, r := range roles {
				if claims.Role == r {
					return c.Next()
				}
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": message})
	}
}
