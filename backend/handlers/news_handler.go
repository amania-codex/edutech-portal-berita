package handlers

import (
	"fmt"
	"news-portal-backend/database"
	"news-portal-backend/models"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Helper to check if current request has authenticated admin role
func isAdminUser(c *fiber.Ctx) bool {
	if claims, ok := c.Locals("user").(*Claims); ok && claims != nil {
		return claims.Role == "admin"
	}
	tokenString := c.Cookies("auth_token")
	if tokenString == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if tokenString == "" {
		return false
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return GetJWTSecret(), nil
	})
	if err == nil && token.Valid {
		if claims, ok := token.Claims.(*Claims); ok {
			return claims.Role == "admin"
		}
	}
	return false
}

// GetNews — GET /api/news
func GetNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 20)
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := c.QueryInt("offset", 0)
	status := c.Query("status")
	category := c.Query("category")
	if category == "" {
		category = c.Params("category")
	}
	search := c.Query("q")

	q := database.DB.Order("created_at desc").Limit(limit).Offset(offset)
	if status != "" {
		if status != "published" && !isAdminUser(c) {
			q = q.Where("is_published = ?", true)
		} else {
			q = q.Where("status = ?", status)
		}
	} else {
		q = q.Where("is_published = ?", true)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if search != "" {
		q = q.Where("title LIKE ?", "%"+search+"%")
	}
	q.Find(&articles)
	return c.JSON(articles)
}

// GetFeaturedNews — GET /api/news/featured
func GetFeaturedNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	database.DB.Where("is_featured = ? AND is_published = ?", true, true).
		Order("created_at desc").Limit(limit).Find(&articles)
	return c.JSON(articles)
}

// GetTrendingNews — GET /api/news/trending
func GetTrendingNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	database.DB.Where("is_trending = ? AND is_published = ?", true, true).
		Order("view_count desc").Limit(limit).Find(&articles)
	return c.JSON(articles)
}

// GetPopularNews — GET /api/news/popular
// Mirip logika post_views_count di beritaxx: artikel paling banyak dilihat
// dalam rentang waktu tertentu (days=30 default)
func GetPopularNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 6)
	days := c.QueryInt("days", 30)
	if limit <= 0 { limit = 6 }
	if limit > 20 { limit = 20 }
	if days <= 0 { days = 30 }

	database.DB.
		Where("is_published = ? AND created_at >= datetime('now', ?)", true, fmt.Sprintf("-%d days", days)).
		Order("view_count desc").
		Limit(limit).
		Find(&articles)

	// Fallback jika hasil kosong (data lama): ambil semua tanpa filter tanggal
	if len(articles) == 0 {
		database.DB.
			Where("is_published = ?", true).
			Order("view_count desc").
			Limit(limit).
			Find(&articles)
	}

	return c.JSON(articles)
}

// GetNewsByCategories — GET /api/news/by-categories?cats=Rakornas,Event,Riset&limit=5
// Untuk CategoryBlock: ambil artikel dari beberapa kategori sekaligus
func GetNewsByCategories(c *fiber.Ctx) error {
	catsParam := c.Query("cats") // e.g. "Rakornas,Event,Riset"
	limit := c.QueryInt("limit", 5)
	if limit <= 0 { limit = 5 }
	if limit > 20 { limit = 20 }

	cats := strings.Split(catsParam, ",")
	result := fiber.Map{}

	for _, cat := range cats {
		cat = strings.TrimSpace(cat)
		if cat == "" { continue }
		var articles []models.Article
		database.DB.
			Where("category = ? AND is_published = ?", cat, true).
			Order("created_at desc").
			Limit(limit).
			Find(&articles)
		result[cat] = articles
	}

	return c.JSON(result)
}

// GetNewsBySlug — GET /api/news/slug/:slug
func GetNewsBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var article models.Article
	if err := database.DB.Where("slug = ? AND is_published = ?", slug, true).First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Berita tidak ditemukan."})
	}
	database.DB.Model(&article).Update("view_count", gorm.Expr("view_count + ?", 1))
	article.ViewCount++
	return c.JSON(article)
}

// GetNewsById — GET /api/news/:id
func GetNewsById(c *fiber.Ctx) error {
	id := c.Params("id")
	var article models.Article
	q := database.DB
	if !isAdminUser(c) {
		q = q.Where("is_published = ?", true)
	}
	if err := q.First(&article, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Berita tidak ditemukan."})
	}
	return c.JSON(article)
}


var htmlTagRegex = regexp.MustCompile("<[^>]*>")

func generateExcerpt(content string, customExcerpt string) string {
	if strings.TrimSpace(customExcerpt) != "" {
		return strings.TrimSpace(customExcerpt)
	}
	plain := htmlTagRegex.ReplaceAllString(content, " ")
	plain = strings.Join(strings.Fields(plain), " ")
	runes := []rune(plain)
	if len(runes) > 160 {
		return string(runes[:157]) + "..."
	}
	return string(runes)
}

// CreateNews — POST /api/news (admin)
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

	// Ensure unique slug (using Unscoped to prevent collision with soft-deleted rows)
	slug := input.Slug
	counter := 1
	for {
		var existing models.Article
		if database.DB.Unscoped().Where("slug = ?", slug).First(&existing).Error != nil {
			break
		}
		slug = input.Slug + "-" + fmt.Sprintf("%d", counter)
		counter++
	}

	isPublished := input.Status == "published"
	article := models.Article{
		Title:       input.Title,
		Slug:        slug,
		Excerpt:     generateExcerpt(input.Content, input.Excerpt),
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

// UpdateNews — PUT /api/news/:id (admin)
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
		IsFeatured  *bool  `json:"isFeatured"`
		IsTrending  *bool  `json:"isTrending"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid."})
	}

	updates := make(map[string]interface{})
	if strings.TrimSpace(input.Title) != "" {
		updates["title"] = input.Title
	}
	if strings.TrimSpace(input.Content) != "" {
		updates["content"] = input.Content
		updates["excerpt"] = generateExcerpt(input.Content, input.Excerpt)
	} else if strings.TrimSpace(input.Excerpt) != "" {
		updates["excerpt"] = input.Excerpt
	}
	if input.ImageURL != "" {
		updates["image_url"] = input.ImageURL
	}
	if input.Category != "" {
		updates["category"] = input.Category
	}
	if input.Tags != "" {
		updates["tags"] = input.Tags
	}
	if input.Author != "" {
		updates["author"] = input.Author
	}
	if input.Status != "" {
		updates["status"] = input.Status
		updates["is_published"] = (input.Status == "published")
	}
	if input.IsFeatured != nil {
		updates["is_featured"] = *input.IsFeatured
	}
	if input.IsTrending != nil {
		updates["is_trending"] = *input.IsTrending
	}

	// Validate slug uniqueness if updated
	if input.Slug != "" && input.Slug != article.Slug {
		var existing models.Article
		if database.DB.Unscoped().Where("slug = ? AND id != ?", input.Slug, article.ID).First(&existing).Error == nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Slug sudah digunakan oleh artikel lain."})
		}
		updates["slug"] = input.Slug
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui berita."})
		}
	}
	return c.JSON(fiber.Map{"success": true, "data": article})
}

// DeleteNews — DELETE /api/news/:id (admin)
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