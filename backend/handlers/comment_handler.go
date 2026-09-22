package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"
	"strconv"
	"gorm.io/gorm"

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
	// Ambil hanya parent comment (ParentID is null)
	database.DB.Where("news_id = ? AND parent_id IS NULL", newsId).Preload("Replies").Order("created_at desc").Find(&comments)

	// Jika ada auth, cek like status. Karena JWT tidak selalu dipass di public route, biarkan frontend yg cek, atau pass user id dari query
	return c.JSON(comments)
}

// CreateComment â€” POST /api/comments
func CreateComment(c *fiber.Ctx) error {
	type Input struct {
		NewsID   uint   `json:"news_id"`
		UserID   uint   `json:"user_id"`
		UserName string `json:"user_name"`
		Content  string `json:"content"`
		ParentID *uint  `json:"parent_id"` // Optional
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
		ParentID:       input.ParentID,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan komentar."})
	}

	// Create notification if it's a reply
	if input.ParentID != nil {
		var parent models.Comment
		if err := database.DB.First(&parent, *input.ParentID).Error; err == nil {
			if parent.UserID != input.UserID { // Don't notify self
				var article models.Article
				database.DB.First(&article, input.NewsID)
				
				notif := models.Notification{
					UserID:    parent.UserID,
					SenderID:  input.UserID,
					Type:      "reply",
					Message:   "membalas komentar Anda",
					CommentID: comment.ID,
					NewsSlug:  article.Slug,
				}
				database.DB.Create(&notif)
			}
		}
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

// LikeComment — POST /api/comments/:id/like
func LikeComment(c *fiber.Ctx) error {
	commentID := c.Params("id")
	userID := getUserID(c) // Use same helper from user_handler (they are in same package)

	var like models.CommentLike
	if err := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error; err == nil {
		// Already liked, so unlike
		database.DB.Delete(&like)
		database.DB.Model(&models.Comment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count - 1"))
		return c.JSON(fiber.Map{"liked": false})
	}

	// Create like
	like = models.CommentLike{
		CommentID: parseUint(commentID),
		UserID:    userID,
	}
	database.DB.Create(&like)
	database.DB.Model(&models.Comment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count + 1"))

	// Create notification
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err == nil {
		if comment.UserID != userID {
			var article models.Article
			database.DB.First(&article, comment.NewsID)

			notif := models.Notification{
				UserID:    comment.UserID,
				SenderID:  userID,
				Type:      "like",
				Message:   "menyukai komentar Anda",
				CommentID: comment.ID,
				NewsSlug:  article.Slug,
			}
			database.DB.Create(&notif)
		}
	}

	return c.JSON(fiber.Map{"liked": true})
}

// Helper to parse string to uint
func parseUint(s string) uint {
	var val uint
	// simple parsing for ID
	importStr, _ := strconv.Atoi(s)
	val = uint(importStr)
	return val
}