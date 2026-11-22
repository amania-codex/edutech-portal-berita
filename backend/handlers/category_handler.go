package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
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

// CreateCategory — POST /api/categories
func CreateCategory(c *fiber.Ctx) error {
	type Input struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}

	if input.Slug == "" {
		input.Slug = generateSlug(input.Name)
	}

	category := models.Category{
		Name: input.Name,
		Slug: input.Slug,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat kategori"})
	}

	return c.JSON(category)
}

// UpdateCategory — PUT /api/categories/:id
func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	
	type Input struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid"})
	}

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Kategori tidak ditemukan"})
	}

	oldName, oldSlug := category.Name, category.Slug

	category.Name = input.Name
	if input.Slug != "" {
		category.Slug = input.Slug
	} else {
		category.Slug = generateSlug(input.Name)
	}

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui kategori"})
	}

	// Update category references in articles
	if oldName != category.Name || oldSlug != category.Slug {
		database.DB.Model(&models.Article{}).Where("category = ?", oldName).
			Updates(map[string]interface{}{"category": category.Name, "category_slug": category.Slug})
	}

	return c.JSON(category)
}

// DeleteCategory — DELETE /api/categories/:id
func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Kategori tidak ditemukan"})
	}

	// Nullify or keep as is? Let's just delete the category record, the strings in articles will stay (as historical data)
	// But it won't be in the dropdown anymore.
	
	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus kategori"})
	}

	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
