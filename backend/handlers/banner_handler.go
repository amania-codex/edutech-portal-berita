package handlers

import (
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

// Posisi poster di situs
var bannerPlacements = map[string]bool{"header": true, "tengah": true, "kiri": true, "kanan": true}

// GetActiveBanners — GET /api/banners (Publik)
func GetActiveBanners(c *fiber.Ctx) error {
	banners := make([]models.Banner, 0)
	if err := database.DB.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data poster"})
	}
	return c.JSON(banners)
}

// GetAllBanners — GET /api/admin/banners (Admin)
func GetAllBanners(c *fiber.Ctx) error {
	banners := make([]models.Banner, 0)
	if err := database.DB.Order("created_at desc").Find(&banners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil daftar poster"})
	}
	return c.JSON(banners)
}

type bannerInput struct {
	Title     string `json:"title"`
	ImageUrl  string `json:"image_url"`
	TargetUrl string `json:"target_url"`
	Placement string `json:"placement"`
	IsActive  *bool  `json:"is_active"`
}

func (in *bannerInput) validate() string {
	in.Title = strings.TrimSpace(in.Title)
	in.ImageUrl = strings.TrimSpace(in.ImageUrl)
	in.TargetUrl = strings.TrimSpace(in.TargetUrl)
	if in.Placement == "" {
		in.Placement = "kanan"
	}
	switch {
	case in.Title == "" || in.ImageUrl == "":
		return "Judul poster dan gambar wajib diisi"
	case !bannerPlacements[in.Placement]:
		return "Posisi poster tidak dikenal"
	case !safeURL(in.ImageUrl, false):
		return "Alamat gambar tidak valid"
	case !safeURL(in.TargetUrl, true):
		return "Tautan tujuan harus diawali http://, https://, atau /"
	}
	return ""
}

// safeURL: kosong (bila boleh), path situs, atau http(s). Menolak javascript: dan sejenisnya.
func safeURL(u string, allowEmpty bool) bool {
	if u == "" {
		return allowEmpty
	}
	l := strings.ToLower(u)
	return strings.HasPrefix(l, "https://") || strings.HasPrefix(l, "http://") || (strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "//"))
}

// CreateBanner — POST /api/admin/banners (Admin)
func CreateBanner(c *fiber.Ctx) error {
	var input bannerInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := input.validate(); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	banner := models.Banner{Title: input.Title, ImageUrl: input.ImageUrl, TargetUrl: input.TargetUrl, Placement: input.Placement, IsActive: input.IsActive == nil || *input.IsActive}
	if err := database.DB.Create(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan poster"})
	}
	return c.Status(201).JSON(banner)
}

// UpdateBanner — PUT /api/admin/banners/:id (Admin)
func UpdateBanner(c *fiber.Ctx) error {
	var banner models.Banner
	if err := database.DB.First(&banner, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Poster tidak ditemukan"})
	}
	var input bannerInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := input.validate(); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	updates := map[string]interface{}{"title": input.Title, "image_url": input.ImageUrl, "target_url": input.TargetUrl, "placement": input.Placement}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if err := database.DB.Model(&banner).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui poster"})
	}
	database.DB.First(&banner, banner.ID)
	return c.JSON(banner)
}

// ToggleBannerStatus — PUT /api/admin/banners/:id/toggle (Admin)
func ToggleBannerStatus(c *fiber.Ctx) error {
	var banner models.Banner
	if err := database.DB.First(&banner, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Poster tidak ditemukan"})
	}
	if err := database.DB.Model(&banner).Update("is_active", !banner.IsActive).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui status poster"})
	}
	banner.IsActive = !banner.IsActive
	return c.JSON(banner)
}

// DeleteBanner — DELETE /api/admin/banners/:id (Admin)
func DeleteBanner(c *fiber.Ctx) error {
	var banner models.Banner
	if err := database.DB.First(&banner, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Poster tidak ditemukan"})
	}
	if err := database.DB.Delete(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus poster"})
	}
	return c.JSON(fiber.Map{"message": "Poster berhasil dihapus"})
}
