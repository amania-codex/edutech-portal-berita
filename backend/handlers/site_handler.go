package handlers

import (
	"net/mail"
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ---------- Pengaturan situs ----------

// Kunci pengaturan yang boleh diisi admin. Semuanya tampil publik (tidak ada rahasia di sini).
var settingKeys = map[string]int{
	"site_description": 300,
	"contact_email":    150,
	"contact_phone":    50,
	"contact_address":  300,
	"facebook_url":     300,
	"x_url":            300,
	"instagram_url":    300,
	"youtube_url":      300,
	"tiktok_url":       300,
}

// GetSettings — GET /api/settings (publik)
func GetSettings(c *fiber.Ctx) error {
	var rows []models.Setting
	if err := database.DB.Find(&rows).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil pengaturan"})
	}
	out := make(map[string]string, len(settingKeys))
	for k := range settingKeys {
		out[k] = ""
	}
	for _, r := range rows {
		if _, ok := settingKeys[r.Key]; ok {
			out[r.Key] = r.Value
		}
	}
	return c.JSON(out)
}

// UpdateSettings — PUT /api/admin/settings {kunci: nilai} (admin)
func UpdateSettings(c *fiber.Ctx) error {
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	rows := make([]models.Setting, 0, len(input))
	for k, v := range input {
		max, ok := settingKeys[k]
		if !ok {
			return c.Status(400).JSON(fiber.Map{"message": "Pengaturan tidak dikenal: " + k})
		}
		v = strings.TrimSpace(v)
		if len([]rune(v)) > max {
			return c.Status(400).JSON(fiber.Map{"message": "Isian terlalu panjang: " + k})
		}
		if strings.HasSuffix(k, "_url") && v != "" && !strings.HasPrefix(strings.ToLower(v), "https://") && !strings.HasPrefix(strings.ToLower(v), "http://") {
			return c.Status(400).JSON(fiber.Map{"message": "Tautan media sosial harus diawali https://"})
		}
		if k == "contact_email" && v != "" {
			if a, err := mail.ParseAddress(v); err != nil || a.Address != v {
				return c.Status(400).JSON(fiber.Map{"message": "Email kontak tidak valid"})
			}
		}
		rows = append(rows, models.Setting{Key: k, Value: v})
	}
	if len(rows) > 0 {
		if err := database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan pengaturan"})
		}
	}
	return GetSettings(c)
}

// ---------- Halaman (Tentang Kami, Kontak, Pedoman, dll.) ----------

// GetPages — GET /api/pages (publik: judul & slug halaman terbit)
func GetPages(c *fiber.Ctx) error {
	pages := make([]models.Page, 0)
	q := database.DB.Order("menu_order asc, title asc")
	if !(c.Query("status") == "all" && isAdminUser(c)) {
		q = q.Where("is_published = ?", true).Omit("content")
	}
	if err := q.Find(&pages).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil halaman"})
	}
	return c.JSON(pages)
}

// GetPageBySlug — GET /api/pages/slug/:slug
func GetPageBySlug(c *fiber.Ctx) error {
	var page models.Page
	q := database.DB.Where("slug = ?", c.Params("slug"))
	if !isAdminUser(c) {
		q = q.Where("is_published = ?", true)
	}
	if err := q.First(&page).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Halaman tidak ditemukan"})
	}
	return c.JSON(page)
}

type pageInput struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Content     string `json:"content"`
	MenuOrder   int    `json:"menuOrder"`
	IsPublished *bool  `json:"isPublished"`
}

func (in *pageInput) validate(id uint) string {
	in.Title = strings.TrimSpace(in.Title)
	in.Slug = generateSlug(in.Slug)
	if in.Slug == "" {
		in.Slug = generateSlug(in.Title)
	}
	switch {
	case len(in.Title) < 2 || len(in.Title) > 150:
		return "Judul halaman 2–150 karakter."
	case in.Slug == "":
		return "Slug tidak valid."
	case len(in.Content) > 200000:
		return "Isi halaman terlalu panjang."
	case in.MenuOrder < 0 || in.MenuOrder > 99:
		return "Urutan 0–99."
	}
	var clash int64
	database.DB.Model(&models.Page{}).Where("slug = ? AND id <> ?", in.Slug, id).Count(&clash)
	if clash > 0 {
		return "Slug halaman sudah dipakai."
	}
	return ""
}

// CreatePage — POST /api/admin/pages
func CreatePage(c *fiber.Ctx) error {
	var in pageInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := in.validate(0); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	page := models.Page{Title: in.Title, Slug: in.Slug, Content: in.Content, MenuOrder: in.MenuOrder, IsPublished: in.IsPublished == nil || *in.IsPublished}
	if err := database.DB.Create(&page).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan halaman"})
	}
	return c.Status(201).JSON(page)
}

// UpdatePage — PUT /api/admin/pages/:id
func UpdatePage(c *fiber.Ctx) error {
	var page models.Page
	if err := database.DB.First(&page, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Halaman tidak ditemukan"})
	}
	var in pageInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := in.validate(page.ID); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	updates := map[string]interface{}{"title": in.Title, "slug": in.Slug, "content": in.Content, "menu_order": in.MenuOrder}
	if in.IsPublished != nil {
		updates["is_published"] = *in.IsPublished
	}
	if err := database.DB.Model(&page).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui halaman"})
	}
	database.DB.First(&page, page.ID)
	return c.JSON(page)
}

// DeletePage — DELETE /api/admin/pages/:id (permanen, slug bisa dipakai lagi)
func DeletePage(c *fiber.Ctx) error {
	res := database.DB.Session(&gorm.Session{}).Unscoped().Delete(&models.Page{}, c.Params("id"))
	if res.Error != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus halaman"})
	}
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Halaman tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"message": "Halaman dihapus"})
}
