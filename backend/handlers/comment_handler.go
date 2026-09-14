package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GetComments â€” GET /api/comments
func GetComments(c *fiber.Ctx) error {
	var comments []models.Comment

	// Fetch comments and join with news to get news title
	database.DB.Table("comments").
		Select("comments.*, articles.title as news_title").
		Joins("left join articles on comments.news_id = articles.id").
		Order("comments.created_at desc").
		Find(&comments)

	return c.JSON(comments)
}

// GetCommentsByNews â€” GET /api/comments/news/:newsId
func GetCommentsByNews(c *fiber.Ctx) error {
	newsId := c.Params("newsId")
	var comments []models.Comment
	database.DB.Where("news_id = ?", newsId).Order("created_at desc").Find(&comments)
	return c.JSON(comments)
}

// CreateComment â€” POST /api/comments
func CreateComment(c *fiber.Ctx) error {
	type Input struct {
		NewsID   uint   `json:"news_id"`
		UserID   uint   `json:"user_id"`
		UserName string `json:"user_name"`
		Content  string `json:"content"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format tidak valid."})
	}

	if len(input.Content) < 3 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Komentar terlalu pendek."})
	}

	analysis := AnalyzeSentiment(input.Content)

	comment := models.Comment{
		NewsID:         input.NewsID,
		UserID:         input.UserID,
		UserName:       input.UserName,
		Content:        input.Content,
		Sentiment:      analysis.Sentiment,
		SentimentScore: analysis.Score,
		IsFlagged:      analysis.IsFlagged,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan komentar."})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": comment})
}

// DeleteComment â€” DELETE /api/comments/:id (admin)
func DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.Comment{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghapus komentar."})
	}
	return c.JSON(fiber.Map{"success": true})
}