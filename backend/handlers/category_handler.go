package handlers

import (
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetAllCategories — GET /api/categories
func GetAllCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := database.DB.Order("name asc").Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil kategori"})
	}
	// ?with_count=1 → jumlah artikel terbit per kategori (termasuk sub-kategori)
	if c.Query("with_count") == "1" {
		for i := range categories {
			applyCategoryFilter(database.DB.Model(&models.Article{}), categories[i].Slug).
				Where("is_published = ?", true).
				Count(&categories[i].Count)
		}
	}
	return c.JSON(categories)
}

type categoryInput struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	ParentID  *uint  `json:"parentId"`
	MenuOrder *int   `json:"menuOrder"`
	IsTopic   *bool  `json:"isTopic"`
}

// validate memeriksa input kategori; id = 0 untuk kategori baru.
func (in *categoryInput) validate(id uint) string {
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len(in.Name) > 100 {
		return "Nama kategori 2–100 karakter."
	}
	in.Slug = generateSlug(in.Slug)
	if in.Slug == "" {
		in.Slug = generateSlug(in.Name)
	}
	if in.Slug == "" {
		return "Slug tidak valid."
	}
	if in.MenuOrder != nil && (*in.MenuOrder < 0 || *in.MenuOrder > 99) {
		return "Urutan menu 0–99."
	}
	var clash int64
	database.DB.Model(&models.Category{}).Where("(name = ? OR slug = ?) AND id <> ?", in.Name, in.Slug, id).Count(&clash)
	if clash > 0 {
		return "Nama atau slug kategori sudah dipakai."
	}
	if in.ParentID != nil && *in.ParentID == 0 {
		in.ParentID = nil
	}
	if in.ParentID != nil {
		// induk harus ada dan bukan kategori ini sendiri atau turunannya
		for cur, depth := *in.ParentID, 0; cur != 0; depth++ {
			if cur == id || depth > 20 {
				return "Induk kategori tidak boleh kategori ini sendiri atau sub-kategorinya."
			}
			var p models.Category
			if err := database.DB.Select("id", "parent_id").First(&p, cur).Error; err != nil {
				return "Induk kategori tidak ditemukan."
			}
			if p.ParentID == nil {
				break
			}
			cur = *p.ParentID
		}
	}
	return ""
}

// CreateCategory — POST /api/categories
func CreateCategory(c *fiber.Ctx) error {
	var input categoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := input.validate(0); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	category := models.Category{Name: input.Name, Slug: input.Slug, ParentID: input.ParentID}
	if input.MenuOrder != nil {
		category.MenuOrder = *input.MenuOrder
	}
	if input.IsTopic != nil {
		category.IsTopic = *input.IsTopic
	}
	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat kategori"})
	}
	return c.JSON(category)
}

// UpdateCategory — PUT /api/categories/:id
func UpdateCategory(c *fiber.Ctx) error {
	var category models.Category
	if err := database.DB.First(&category, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Kategori tidak ditemukan"})
	}
	var input categoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}
	if msg := input.validate(category.ID); msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}

	oldName, oldSlug := category.Name, category.Slug
	updates := map[string]interface{}{"name": input.Name, "slug": input.Slug, "parent_id": input.ParentID}
	if input.MenuOrder != nil {
		updates["menu_order"] = *input.MenuOrder
	}
	if input.IsTopic != nil {
		updates["is_topic"] = *input.IsTopic
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&category).Updates(updates).Error; err != nil {
			return err
		}
		// artikel yang menyimpan nama/slug lama ikut diperbarui
		if oldName != input.Name || oldSlug != input.Slug {
			return tx.Model(&models.Article{}).Where("category = ? OR category_slug = ?", oldName, oldSlug).
				Updates(map[string]interface{}{"category": input.Name, "category_slug": input.Slug}).Error
		}
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui kategori"})
	}
	database.DB.First(&category, category.ID)
	return c.JSON(category)
}

// DeleteCategory — DELETE /api/categories/:id
// Kategori dihapus permanen (namanya bisa dipakai lagi). Artikelnya tidak ikut terhapus;
// sub-kategorinya naik menjadi kategori utama.
func DeleteCategory(c *fiber.Ctx) error {
	var category models.Category
	if err := database.DB.First(&category, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Kategori tidak ditemukan"})
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Category{}).Where("parent_id = ?", category.ID).Update("parent_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Where("category_id = ?", category.ID).Delete(&models.ArticleCategory{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&category).Error
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus kategori"})
	}
	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
