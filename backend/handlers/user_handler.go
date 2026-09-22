package handlers

import (
	"news-portal-backend/database"
	"news-portal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Get userID from JWT
func getUserID(c *fiber.Ctx) uint {
	user := c.Locals("user").(*Claims)
	return user.UserID
}

// GetReadingHistory — GET /api/user/history
func GetReadingHistory(c *fiber.Ctx) error {
	userID := getUserID(c)

	var histories []models.ReadingHistory
	database.DB.Where("user_id = ?", userID).Order("read_at desc").Limit(20).Find(&histories)

	// Populate virtual fields
	for i, h := range histories {
		var article models.Article
		if err := database.DB.Select("title", "slug", "image_url").First(&article, h.NewsID).Error; err == nil {
			histories[i].NewsTitle = article.Title
			histories[i].NewsSlug = article.Slug
			histories[i].NewsImage = article.ImageURL
		}
	}

	return c.JSON(histories)
}

// AddReadingHistory — POST /api/user/history
func AddReadingHistory(c *fiber.Ctx) error {
	userID := getUserID(c)
	
	type Input struct {
		NewsID uint `json:"news_id"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid input"})
	}

	var history models.ReadingHistory
	// Check if already exists to update read_at instead of duplicating
	if err := database.DB.Where("user_id = ? AND news_id = ?", userID, input.NewsID).First(&history).Error; err == nil {
		history.ReadAt = time.Now()
		database.DB.Save(&history)
		return c.JSON(history)
	}

	history = models.ReadingHistory{
		UserID: userID,
		NewsID: input.NewsID,
		ReadAt: time.Now(),
	}
	database.DB.Create(&history)
	return c.JSON(history)
}

// GetBookmarks — GET /api/user/bookmarks
func GetBookmarks(c *fiber.Ctx) error {
	userID := getUserID(c)

	var bookmarks []models.Bookmark
	database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&bookmarks)

	for i, b := range bookmarks {
		var article models.Article
		if err := database.DB.Select("title", "slug", "image_url").First(&article, b.NewsID).Error; err == nil {
			bookmarks[i].NewsTitle = article.Title
			bookmarks[i].NewsSlug = article.Slug
			bookmarks[i].NewsImage = article.ImageURL
		}
	}

	return c.JSON(bookmarks)
}

// ToggleBookmark — POST /api/user/bookmarks
func ToggleBookmark(c *fiber.Ctx) error {
	userID := getUserID(c)
	
	type Input struct {
		NewsID uint `json:"news_id"`
	}
	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid input"})
	}

	var bookmark models.Bookmark
	if err := database.DB.Where("user_id = ? AND news_id = ?", userID, input.NewsID).First(&bookmark).Error; err == nil {
		// Exists, so un-bookmark (delete)
		database.DB.Delete(&bookmark)
		return c.JSON(fiber.Map{"bookmarked": false, "message": "Berita dihapus dari bookmark"})
	}

	// Create
	bookmark = models.Bookmark{
		UserID: userID,
		NewsID: input.NewsID,
	}
	database.DB.Create(&bookmark)
	return c.JSON(fiber.Map{"bookmarked": true, "message": "Berita disimpan"})
}

// CheckBookmark — GET /api/user/bookmarks/:news_id
func CheckBookmark(c *fiber.Ctx) error {
	userID := getUserID(c)
	newsID := c.Params("news_id")

	var bookmark models.Bookmark
	if err := database.DB.Where("user_id = ? AND news_id = ?", userID, newsID).First(&bookmark).Error; err == nil {
		return c.JSON(fiber.Map{"bookmarked": true})
	}
	return c.JSON(fiber.Map{"bookmarked": false})
}

// GetNotifications — GET /api/user/notifications
func GetNotifications(c *fiber.Ctx) error {
	userID := getUserID(c)

	var notifs []models.Notification
	database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifs)

	for i, n := range notifs {
		var sender models.User
		if err := database.DB.Select("name", "avatar").First(&sender, n.SenderID).Error; err == nil {
			notifs[i].SenderName = sender.Name
			notifs[i].SenderAvatar = sender.Avatar
		}
	}

	return c.JSON(notifs)
}

// ReadNotification — POST /api/user/notifications/:id/read
func ReadNotification(c *fiber.Ctx) error {
	userID := getUserID(c)
	id := c.Params("id")

	var notif models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&notif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Notifikasi tidak ditemukan"})
	}

	notif.IsRead = true
	database.DB.Save(&notif)

	return c.JSON(fiber.Map{"success": true})
}

// ReadAllNotifications — POST /api/user/notifications/read-all
func ReadAllNotifications(c *fiber.Ctx) error {
	userID := getUserID(c)
	database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true)
	return c.JSON(fiber.Map{"success": true})
}
