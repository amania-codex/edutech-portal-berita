package handlers

import (
	"fmt"
	"news-portal-backend/database"
	"news-portal-backend/models"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetNews â€” GET /api/news
func GetNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	status := c.Query("status")
	category := c.Query("category")

	q := database.DB.Order("created_at desc").Limit(limit).Offset(offset)
	if status != "" {
		q = q.Where("status = ?", status)
	} else {
		q = q.Where("is_published = ?", true)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	q.Find(&articles)
	return c.JSON(articles)
}

// GetFeaturedNews â€” GET /api/news/featured
func GetFeaturedNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	database.DB.Where("is_featured = ? AND is_published = ?", true, true).
		Order("created_at desc").Limit(limit).Find(&articles)
	return c.JSON(articles)
}

// GetTrendingNews â€” GET /api/news/trending
func GetTrendingNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	database.DB.Where("is_trending = ? AND is_published = ?", true, true).
		Order("view_count desc").Limit(limit).Find(&articles)
	return c.JSON(articles)
}

// GetNewsBySlug â€” GET /api/news/slug/:slug
func GetNewsBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var article models.Article
	if err := database.DB.Where("slug = ? AND is_published = ?", slug, true).First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Berita tidak ditemukan."})
	}
	database.DB.Model(&article).Update("view_count", article.ViewCount+1)
	return c.JSON(article)
}

// GetNewsById â€” GET /api/news/:id
func GetNewsById(c *fiber.Ctx) error {
	id := c.Params("id")
	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Berita tidak ditemukan."})
	}
	return c.JSON(article)
}

// GetCategories â€” GET /api/categories
func GetCategories(c *fiber.Ctx) error {
	var categories []string
	database.DB.Model(&models.Article{}).Where("is_published = ?", true).
		Distinct("category").Pluck("category", &categories)
	return c.JSON(categories)
}

// CreateNews â€” POST /api/news (admin)
func CreateNews(c *fiber.Ctx) error {
	type Input struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Excerpt     string `json:"excerpt"`
		Content     string `json:"content"`
		ImageURL    string `json:"imageUrl"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		Author      string `json:"author"`
		Status      string `json:"status"`
		IsFeatured  bool   `json:"isFeatured"`
		IsTrending  bool   `json:"isTrending"`
		IsPublished bool   `json:"isPublished"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid."})
	}
	if strings.TrimSpace(input.Title) == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Judul wajib diisi."})
	}
	if strings.TrimSpace(input.Content) == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Konten wajib diisi."})
	}

	// Auto-generate slug if empty
	if input.Slug == "" {
		input.Slug = generateSlug(input.Title)
	}

	// Ensure unique slug
	slug := input.Slug
	counter := 1
	for {
		var existing models.Article
		if database.DB.Where("slug = ?", slug).First(&existing).Error != nil {
			break
		}
		slug = input.Slug + "-" + fmt.Sprintf("%d", counter)
		counter++
	}

	isPublished := input.Status == "published"
	article := models.Article{
		Title:       input.Title,
		Slug:        slug,
		Excerpt:     input.Excerpt,
		Content:     input.Content,
		ImageURL:    input.ImageURL,
		Category:    input.Category,
		Tags:        input.Tags,
		Author:      input.Author,
		Status:      input.Status,
		IsFeatured:  input.IsFeatured,
		IsTrending:  input.IsTrending,
		IsPublished: isPublished,
	}

	if claims, ok := c.Locals("user").(*Claims); ok {
		article.AuthorID = claims.UserID
		if article.Author == "" {
			article.Author = claims.Name
		}
	}

	if err := database.DB.Create(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan berita."})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": article})
}

// UpdateNews â€” PUT /api/news/:id (admin)
func UpdateNews(c *fiber.Ctx) error {
	id := c.Params("id")
	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Berita tidak ditemukan."})
	}

	type Input struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Excerpt     string `json:"excerpt"`
		Content     string `json:"content"`
		ImageURL    string `json:"imageUrl"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		Author      string `json:"author"`
		Status      string `json:"status"`
		IsFeatured  bool   `json:"isFeatured"`
		IsTrending  bool   `json:"isTrending"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid."})
	}

	updates := map[string]interface{}{
		"title":        input.Title,
		"excerpt":      input.Excerpt,
		"content":      input.Content,
		"image_url":    input.ImageURL,
		"category":     input.Category,
		"tags":         input.Tags,
		"author":       input.Author,
		"status":       input.Status,
		"is_featured":  input.IsFeatured,
		"is_trending":  input.IsTrending,
		"is_published": input.Status == "published",
	}
	if input.Slug != "" {
		updates["slug"] = input.Slug
	}

	if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui berita."})
	}
	return c.JSON(fiber.Map{"success": true, "data": article})
}

// DeleteNews â€” DELETE /api/news/:id (admin)
func DeleteNews(c *fiber.Ctx) error {
	id := c.Params("id")
	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Berita tidak ditemukan."})
	}
	database.DB.Delete(&article)
	return c.JSON(fiber.Map{"success": true})
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s := result.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}