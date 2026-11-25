package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GetAdminStats — GET /api/admin/stats
// Ringkasan angka untuk halaman depan dashboard. Penulis hanya melihat angka beritanya sendiri.
func GetAdminStats(c *fiber.Ctx) error {
	claims := currentClaims(c)
	all := canManageAllNews(claims)
	articles := database.DB.Model(&models.Article{})
	commentQ := database.DB.Model(&models.Comment{})
	if !all {
		articles = articles.Where("author_id = ?", claims.UserID)
		commentQ = commentQ.Where("news_id IN (SELECT id FROM articles WHERE author_id = ? AND deleted_at IS NULL)", claims.UserID)
	}
	var news struct {
		Total     int64 `json:"total"`
		Published int64 `json:"published"`
		Draft     int64 `json:"draft"`
		Views     int64 `json:"views"`
	}
	if err := articles.
		Select("COUNT(*) AS total, COALESCE(SUM(is_published), 0) AS published, COALESCE(SUM(NOT is_published), 0) AS draft, COALESCE(SUM(view_count), 0) AS views").
		Scan(&news).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil statistik."})
	}

	var comments struct {
		Total    int64 `json:"total"`
		Flagged  int64 `json:"flagged"`
		Positive int64 `json:"positive"`
		Neutral  int64 `json:"neutral"`
		Negative int64 `json:"negative"`
	}
	if err := commentQ.
		Select("COUNT(*) AS total, COALESCE(SUM(is_flagged), 0) AS flagged, COALESCE(SUM(sentiment = 'positive'), 0) AS positive, COALESCE(SUM(sentiment = 'neutral'), 0) AS neutral, COALESCE(SUM(sentiment = 'negative'), 0) AS negative").
		Scan(&comments).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil statistik."})
	}

	var subscribers int64
	if claims.Role == models.RoleAdmin {
		database.DB.Model(&models.NewsletterSubscriber{}).Count(&subscribers)
	}

	return c.JSON(fiber.Map{"news": news, "comments": comments, "subscribers": subscribers})
}
