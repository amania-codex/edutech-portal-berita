package handlers

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"news-portal-backend/database"
	"news-portal-backend/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GetJWTSecret returns the secret key byte slice from env or fallback
func GetJWTSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if os.Getenv("APP_ENV") == "production" && s == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable MUST be set in production mode!")
	}
	if s != "" {
		return []byte(s)
	}
	return []byte("edutech-super-secret-key-change-in-production-2026")
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
	return token.SignedString(GetJWTSecret())
}

// Register — POST /api/auth/register
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

	// Trigger welcome system notification
	CreateSystemNotification(user.ID, "Selamat datang di EduTech Portal Berita! Anda kini dapat membaca artikel, menyimpan berita favorit, dan berpartisipasi dalam diskusi komentar.", "")

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"user":    fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
		"token":   token,
	})
}

// Login — POST /api/auth/login
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
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"user":    fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
		"token":   token,
	})
}

// Logout — POST /api/auth/logout
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

// GetMe — GET /api/auth/me
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

	// Trigger security system notification
	CreateSystemNotification(dbUser.ID, "Keamanan Akun: Kata sandi Anda baru saja berhasil diperbarui. Jika Anda tidak melakukan tindakan ini, segera amankan akun Anda.", "")

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

	// Trigger profile system notification
	CreateSystemNotification(dbUser.ID, "Informasi biodata & profil akun Anda berhasil diperbarui.", "")

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

	// Generate 6-digit cryptographically secure code
	b := make([]byte, 3)
	if _, err := cryptorand.Read(b); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghasilkan kode"})
	}
	num := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	code := fmt.Sprintf("%06d", num)
	exp := time.Now().Add(15 * time.Minute)

	// Store SHA-256 hash of the code
	codeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))
	user.ResetCode = codeHash
	user.ResetCodeExp = &exp
	user.ResetAttempts = 0
	database.DB.Save(&user)

	// Simulate sending email (logged securely)
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
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Kode tidak valid atau email salah."})
	}

	if user.ResetCode == "" || user.ResetCodeExp == nil || time.Now().After(*user.ResetCodeExp) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Kode telah kedaluwarsa atau belum diminta."})
	}

	if user.ResetAttempts >= 5 {
		user.ResetCode = ""
		user.ResetCodeExp = nil
		database.DB.Save(&user)
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Batas percobaan terlampaui. Silakan minta kode baru."})
	}

	inputHash := fmt.Sprintf("%x", sha256.Sum256([]byte(input.Code)))
	if user.ResetCode != inputHash {
		user.ResetAttempts++
		database.DB.Save(&user)
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Kode verifikasi salah."})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	user.Password = string(hash)
	user.ResetCode = ""
	user.ResetCodeExp = nil
	user.ResetAttempts = 0
	database.DB.Save(&user)

	return c.JSON(fiber.Map{"success": true, "message": "Password berhasil direset. Silakan login."})
}