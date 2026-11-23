package handlers

import (
	"net/mail"
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
)

// Subscribe — POST /api/newsletter {email}
// Mendaftar ulang dengan email yang sama dianggap berhasil (tanpa duplikat).
func Subscribe(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid."})
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 150 || !strings.Contains(email[strings.LastIndex(email, "@"):], ".") {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Alamat email tidak valid."})
	}
	sub := models.NewsletterSubscriber{Email: email}
	if err := database.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&sub).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan. Silakan coba lagi."})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Terima kasih! Email Anda telah terdaftar."})
}

// GetSubscribers — GET /api/admin/newsletter (admin)
func GetSubscribers(c *fiber.Ctx) error {
	subs := make([]models.NewsletterSubscriber, 0)
	if err := database.DB.Order("created_at desc").Find(&subs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data pelanggan."})
	}
	return c.JSON(subs)
}
