package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GetActiveBanners — GET /api/banners (Publik)
func GetActiveBanners(c *fiber.Ctx) error {
	var banners []models.Banner
	if err := database.DB.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data poster"})
	}
	return c.JSON(banners)
}

// GetAllBanners — GET /api/admin/banners (Admin)
func GetAllBanners(c *fiber.Ctx) error {
	var banners []models.Banner
	if err := database.DB.Order("created_at desc").Find(&banners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil daftar poster"})
	}
	return c.JSON(banners)
}

// CreateBanner — POST /api/admin/banners (Admin)
func CreateBanner(c *fiber.Ctx) error {
	type Input struct {
		Title     string `json:"title"`
		ImageUrl  string `json:"image_url"`
		TargetUrl string `json:"target_url"`
		IsActive  *bool  `json:"is_active"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}

	if input.Title == "" || input.ImageUrl == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Judul poster dan gambar wajib diisi"})
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	banner := models.Banner{
		Title:     input.Title,
		ImageUrl:  input.ImageUrl,
		TargetUrl: input.TargetUrl,
		IsActive:  isActive,
	}

	if err := database.DB.Create(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan poster"})
	}

	return c.Status(201).JSON(banner)
}

// ToggleBannerStatus — PUT /api/admin/banners/:id/toggle (Admin)
func ToggleBannerStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var banner models.Banner
	if err := database.DB.First(&banner, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Poster tidak ditemukan"})
	}

	banner.IsActive = !banner.IsActive
	if err := database.DB.Save(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui status poster"})
	}

	return c.JSON(banner)
}

// DeleteBanner — DELETE /api/admin/banners/:id (Admin)
func DeleteBanner(c *fiber.Ctx) error {
	id := c.Params("id")
	var banner models.Banner
	if err := database.DB.First(&banner, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Poster tidak ditemukan"})
	}

	if err := database.DB.Delete(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus poster"})
	}

	return c.JSON(fiber.Map{"message": "Poster berhasil dihapus"})
}
