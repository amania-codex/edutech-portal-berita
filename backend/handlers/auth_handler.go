package handlers

import (
	"fmt"
	"log"
	"math/rand"
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
	userClaims := c.Locals("user").(*Claims)
	var dbUser models.User
	if err := database.DB.First(&dbUser, userClaims.UserID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"id":         dbUser.ID,
		"name":       dbUser.Name,
		"email":      dbUser.Email,
		"role":       dbUser.Role,
		"avatar":     dbUser.Avatar,
		"bio":        dbUser.Bio,
		"created_at": dbUser.CreatedAt,
	})
}

// ChangePassword - PUT /api/auth/password
func ChangePassword(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*Claims)
	
	type Input struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format permintaan tidak valid."})
	}

	if len(input.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Password baru minimal 6 karakter."})
	}

	var dbUser models.User
	if err := database.DB.First(&dbUser, userClaims.UserID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "User tidak ditemukan."})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(input.OldPassword)); err != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Password lama salah."})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memproses password."})
	}

	dbUser.Password = string(hash)
	if err := database.DB.Save(&dbUser).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan password baru."})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Password berhasil diubah."})
}

// UpdateProfile - PUT /api/auth/profile
func UpdateProfile(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*Claims)
	
	type Input struct {
		Name   string `json:"name"`
		Bio    string `json:"bio"`
		Avatar string `json:"avatar"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format permintaan tidak valid."})
	}

	if len(input.Name) < 2 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Nama minimal 2 karakter."})
	}

	var dbUser models.User
	if err := database.DB.First(&dbUser, userClaims.UserID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "User tidak ditemukan."})
	}

	dbUser.Name = input.Name
	dbUser.Bio = input.Bio
	if input.Avatar != "" {
		dbUser.Avatar = input.Avatar
	}

	if err := database.DB.Save(&dbUser).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan profil."})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Profil berhasil diperbarui."})
}

// ForgotPassword - POST /api/auth/forgot-password
func ForgotPassword(c *fiber.Ctx) error {
	type Input struct {
		Email string `json:"email"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Return success even if not found to prevent email enumeration
		return c.JSON(fiber.Map{"success": true, "message": "Jika email terdaftar, kode telah dikirim."})
	}

	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	exp := time.Now().Add(15 * time.Minute)

	user.ResetCode = code
	user.ResetCodeExp = &exp
	database.DB.Save(&user)

	// Simulate sending email
	log.Printf("=========================================\n")
	log.Printf("EMAIL KE: %s\n", user.Email)
	log.Printf("KODE RESET PASSWORD ANDA: %s\n", code)
	log.Printf("Berlaku hingga 15 menit ke depan.\n")
	log.Printf("=========================================\n")

	return c.JSON(fiber.Map{"success": true, "message": "Jika email terdaftar, kode telah dikirim."})
}

// ResetPassword - POST /api/auth/reset-password
func ResetPassword(c *fiber.Ctx) error {
	type Input struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid"})
	}

	if len(input.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Password baru minimal 6 karakter"})
	}

	var user models.User
	if err := database.DB.Where("email = ? AND reset_code = ?", input.Email, input.Code).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Kode tidak valid atau email salah."})
	}

	if user.ResetCodeExp == nil || time.Now().After(*user.ResetCodeExp) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Kode telah kedaluwarsa."})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	user.Password = string(hash)
	user.ResetCode = ""
	user.ResetCodeExp = nil
	database.DB.Save(&user)

	return c.JSON(fiber.Map{"success": true, "message": "Password berhasil direset. Silakan login."})
}