package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "edutech-super-secret-key-change-in-production-2026"
}

type Claims struct {
	UserID uint   `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func generateToken(user models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Register â€” POST /api/auth/register
func Register(c *fiber.Ctx) error {
	type Input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format permintaan tidak valid."})
	}

	if len(input.Name) < 2 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Nama minimal 2 karakter."})
	}
	if len(input.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Password minimal 6 karakter."})
	}

	// Check duplicate email
	var existing models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return c.Status(409).JSON(fiber.Map{"success": false, "error": "Email sudah terdaftar."})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memproses password."})
	}

	user := models.User{Name: input.Name, Email: input.Email, Password: string(hash), Role: "user"}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal membuat akun."})
	}

	token, err := generateToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal membuat sesi."})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"user":    fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
		"token":   token,
	})
}

// Login â€” POST /api/auth/login
func Login(c *fiber.Ctx) error {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format permintaan tidak valid."})
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Email atau password salah."})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Email atau password salah."})
	}

	token, err := generateToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal membuat sesi."})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"user":    fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
		"token":   token,
	})
}

// Logout â€” POST /api/auth/logout
func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   -1,
	})
	return c.JSON(fiber.Map{"success": true})
}

// GetMe â€” GET /api/auth/me
func GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*Claims)
	return c.JSON(fiber.Map{
		"id":    user.UserID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}