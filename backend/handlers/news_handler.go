package handlers

import (
	"fmt"
	"html"
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

// newsQueryFilters menerapkan filter umum (status, kategori, tag, pencarian)
// ke query artikel. Dipakai oleh GetNews untuk data & total.
func newsQueryFilters(c *fiber.Ctx, q *gorm.DB) *gorm.DB {
	status := c.Query("status")
	category := c.Query("category")
	if category == "" {
		category = c.Params("category")
	}
	tag := strings.TrimSpace(c.Query("tag"))
	search := strings.TrimSpace(c.Query("q"))

	switch {
	case status == "all" && isAdminUser(c):
		// admin: tanpa filter status
	case status != "" && status != "all" && isAdminUser(c):
		q = q.Where("status = ?", status)
	default:
		q = q.Where("is_published = ?", true)
	}

	if category != "" && strings.ToLower(category) != "semua" {
		q = applyCategoryFilter(q, category)
	}
	if tag != "" {
		q = q.Where("CONCAT(',', REPLACE(tags, ', ', ','), ',') LIKE ?", "%,"+tag+",%")
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("(title LIKE ? OR tags LIKE ? OR excerpt LIKE ?)", like, like, like)
	}
	return q
}

// applyCategoryFilter memfilter berdasarkan slug ATAU nama kategori, termasuk
// seluruh sub-kategori (mis. "pendidikan" ikut memuat "purbalingga").
func applyCategoryFilter(q *gorm.DB, category string) *gorm.DB {
	var root models.Category
	if err := database.DB.Where("slug = ? OR name = ?", category, category).First(&root).Error; err != nil {
		return q.Where("(category = ? OR category_slug = ?)", category, category)
	}

	var all []models.Category
	database.DB.Find(&all)
	ids := []uint{root.ID}
	names := []string{root.Name}
	slugs := []string{root.Slug}
	for i := 0; i < len(ids); i++ {
		for _, cat := range all {
			if cat.ParentID != nil && *cat.ParentID == ids[i] {
				ids = append(ids, cat.ID)
				names = append(names, cat.Name)
				slugs = append(slugs, cat.Slug)
			}
		}
	}
	return q.Where(
		"(category IN ? OR category_slug IN ? OR id IN (SELECT article_id FROM article_categories WHERE category_id IN ?))",
		names, slugs, ids,
	)
}

// GetNews — GET /api/news
// Query: limit, offset (atau page), status, category (slug/nama), tag, q, lite=1 (tanpa konten)
// Total data dikirim lewat header X-Total-Count.
func GetNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 20)
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := c.QueryInt("offset", 0)
	if page := c.QueryInt("page", 0); page > 1 {
		offset = (page - 1) * limit
	}

	base := newsQueryFilters(c, database.DB.Model(&models.Article{}))

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	c.Set("X-Total-Count", fmt.Sprintf("%d", total))

	q := base.Session(&gorm.Session{}).Order("created_at desc, id desc").Limit(limit).Offset(offset)
	if c.Query("lite") == "1" {
		q = q.Omit("content")
	}
	if err := q.Find(&articles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil berita."})
	}
	attachCategories(articles)
	return c.JSON(articles)
}

// attachCategories mengisi Article.Categories (slug semua kategori) dengan satu query.
func attachCategories(articles []models.Article) {
	if len(articles) == 0 {
		return
	}
	index := make(map[uint]int, len(articles))
	ids := make([]uint, 0, len(articles))
	for i, a := range articles {
		index[a.ID] = i
		ids = append(ids, a.ID)
	}
	var rows []struct {
		ArticleID uint
		Slug      string
	}
	database.DB.Table("article_categories ac").
		Select("ac.article_id, c.slug").
		Joins("JOIN categories c ON c.id = ac.category_id AND c.deleted_at IS NULL").
		Where("ac.article_id IN ?", ids).
		Scan(&rows)
	for _, r := range rows {
		i := index[r.ArticleID]
		articles[i].Categories = append(articles[i].Categories, r.Slug)
	}
}

// ResolveNewsSlug — GET /api/news/resolve/:slug
// Mencari artikel berdasarkan slug aktif atau slug lama (redirect WordPress).
func ResolveNewsSlug(c *fiber.Ctx) error {
	slug := strings.Trim(c.Params("slug"), "/ ")
	var article models.Article
	if err := database.DB.Select("id", "slug").
		Where("slug = ? AND is_published = ?", slug, true).First(&article).Error; err == nil {
		return c.JSON(fiber.Map{"slug": article.Slug})
	}
	var redirect models.ArticleRedirect
	if err := database.DB.Where("old_slug = ?", slug).First(&redirect).Error; err == nil {
		if err := database.DB.Select("id", "slug").
			Where("id = ? AND is_published = ?", redirect.ArticleID, true).First(&article).Error; err == nil {
			return c.JSON(fiber.Map{"slug": article.Slug})
		}
	}
	return c.Status(404).JSON(fiber.Map{"message": "Berita tidak ditemukan."})
}

// TrackNewsView — POST /api/news/:id/view
// Menambah jumlah tayangan (halaman berita di-render statis, jadi dihitung dari browser).
func TrackNewsView(c *fiber.Ctx) error {
	res := database.DB.Model(&models.Article{}).
		Where("id = ? AND is_published = ?", c.Params("id"), true).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false})
	}
	return c.JSON(fiber.Map{"success": true})
}

// GetFeaturedNews — GET /api/news/featured
func GetFeaturedNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	database.DB.Omit("content").Where("is_featured = ? AND is_published = ?", true, true).
		Order("created_at desc").Limit(limit).Find(&articles)
	return c.JSON(articles)
}

// GetTrendingNews — GET /api/news/trending
// Artikel bertanda trending; jika belum ada, pakai yang terbanyak dibaca 30 hari terakhir.
func GetTrendingNews(c *fiber.Ctx) error {
	var articles []models.Article
	limit := c.QueryInt("limit", 5)
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	database.DB.Omit("content").Where("is_trending = ? AND is_published = ?", true, true).
		Order("view_count desc").Limit(limit).Find(&articles)
	if len(articles) == 0 {
		articles = mostViewed(limit, 30)
	}
	return c.JSON(articles)
}

// mostViewed mengambil artikel terbanyak dibaca dalam rentang hari tertentu,
// dengan fallback ke seluruh waktu jika rentang tersebut kosong.
func mostViewed(limit, days int) []models.Article {
	var articles []models.Article
	database.DB.Omit("content").
		Where("is_published = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", true, days).
		Order("view_count desc").Limit(limit).Find(&articles)
	if len(articles) == 0 {
		database.DB.Omit("content").Where("is_published = ?", true).
			Order("view_count desc").Limit(limit).Find(&articles)
	}
	return articles
}

// GetPopularNews — GET /api/news/popular
// Artikel paling banyak dilihat dalam rentang waktu tertentu (days=30 default)
func GetPopularNews(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 6)
	days := c.QueryInt("days", 30)
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}
	if days <= 0 {
		days = 30
	}
	return c.JSON(mostViewed(limit, days))
}

// GetNewsByCategories — GET /api/news/by-categories?cats=pendidikan,artikel&limit=5
// Untuk CategoryBlock: ambil artikel dari beberapa kategori sekaligus (slug atau nama)
func GetNewsByCategories(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 5)
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	result := fiber.Map{}
	for _, cat := range strings.Split(c.Query("cats"), ",") {
		cat = strings.TrimSpace(cat)
		if cat == "" {
			continue
		}
		var articles []models.Article
		applyCategoryFilter(database.DB.Omit("content"), cat).
			Where("is_published = ?", true).
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
	database.DB.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
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

// syncArticleCategory menjaga article_categories tetap sejalan dengan kategori utama:
// kategori utama lama dilepas, kategori utama baru ditambahkan, sisanya dibiarkan.
func syncArticleCategory(articleID uint, oldSlug, newSlug string) {
	if articleID == 0 || newSlug == "" || oldSlug == newSlug {
		return
	}
	if oldSlug != "" {
		database.DB.Exec("DELETE ac FROM article_categories ac JOIN categories c ON c.id = ac.category_id WHERE ac.article_id = ? AND c.slug = ?", articleID, oldSlug)
	}
	database.DB.Exec("INSERT IGNORE INTO article_categories (article_id, category_id) SELECT ?, id FROM categories WHERE slug = ? AND deleted_at IS NULL", articleID, newSlug)
}

// categorySlugFor mengembalikan slug kategori berdasarkan nama (untuk artikel dari dashboard).
func categorySlugFor(name string) string {
	if name == "" {
		return ""
	}
	var cat models.Category
	if err := database.DB.Where("name = ? OR slug = ?", name, name).First(&cat).Error; err == nil {
		return cat.Slug
	}
	return generateSlug(name)
}

var htmlTagRegex = regexp.MustCompile("<[^>]*>")

// plainText — teks polos dari HTML: tag dibuang, entitas (&nbsp; &amp;) diterjemahkan.
func plainText(content string) string {
	text := html.UnescapeString(htmlTagRegex.ReplaceAllString(content, " "))
	return strings.Join(strings.Fields(text), " ") // Fields juga memecah spasi non-breaking
}

// estimateReadTime — estimasi menit baca (200 kata per menit, minimal 1).
func estimateReadTime(content string) int {
	return max(1, (len(strings.Fields(plainText(content)))+199)/200)
}

func generateExcerpt(content string, customExcerpt string) string {
	if strings.TrimSpace(customExcerpt) != "" {
		return strings.TrimSpace(customExcerpt)
	}
	plain := plainText(content)
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
		Category:     input.Category,
		CategorySlug: categorySlugFor(input.Category),
		ReadTime:     estimateReadTime(input.Content),
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
	syncArticleCategory(article.ID, "", article.CategorySlug)
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

	oldCategorySlug, oldArticleSlug := article.CategorySlug, article.Slug
	updates := make(map[string]interface{})
	if strings.TrimSpace(input.Title) != "" {
		updates["title"] = input.Title
	}
	if strings.TrimSpace(input.Content) != "" {
		updates["content"] = input.Content
		updates["read_time"] = estimateReadTime(input.Content)
		updates["excerpt"] = generateExcerpt(input.Content, input.Excerpt)
	} else if strings.TrimSpace(input.Excerpt) != "" {
		updates["excerpt"] = input.Excerpt
	}
	if input.ImageURL != "" {
		updates["image_url"] = input.ImageURL
	}
	if input.Category != "" {
		updates["category"] = input.Category
		updates["category_slug"] = categorySlugFor(input.Category)
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
		if newSlug, ok := updates["category_slug"].(string); ok {
			syncArticleCategory(article.ID, oldCategorySlug, newSlug)
		}
		// Link lama tetap hidup bila slug diganti
		if newSlug, ok := updates["slug"].(string); ok {
			database.DB.Exec("DELETE FROM article_redirects WHERE old_slug = ?", newSlug)
			database.DB.Exec("INSERT INTO article_redirects (old_slug, article_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE article_id = VALUES(article_id)", oldArticleSlug, article.ID)
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