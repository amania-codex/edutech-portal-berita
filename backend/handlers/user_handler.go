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

	histories := make([]models.ReadingHistory, 0)
	database.DB.Table("reading_histories").
		Select("reading_histories.*, articles.title as news_title, articles.slug as news_slug, articles.image_url as news_image").
		Joins("LEFT JOIN articles ON reading_histories.news_id = articles.id").
		Where("reading_histories.user_id = ? AND reading_histories.deleted_at IS NULL", userID).
		Order("reading_histories.read_at desc").
		Limit(30).
		Scan(&histories)

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

	bookmarks := make([]models.Bookmark, 0)
	database.DB.Table("bookmarks").
		Select("bookmarks.*, articles.title as news_title, articles.slug as news_slug, articles.image_url as news_image").
		Joins("LEFT JOIN articles ON bookmarks.news_id = articles.id").
		Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL", userID).
		Order("bookmarks.created_at desc").
		Scan(&bookmarks)

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

	notifs := make([]models.Notification, 0)
	database.DB.Table("notifications").
		Select("notifications.*, users.name as sender_name, users.avatar as sender_avatar").
		Joins("LEFT JOIN users ON notifications.sender_id = users.id").
		Where("notifications.user_id = ? AND notifications.deleted_at IS NULL", userID).
		Order("notifications.created_at desc").
		Scan(&notifs)

	for i := range notifs {
		if notifs[i].Type == "system" || notifs[i].SenderID == 0 {
			notifs[i].SenderName = "Sistem EduTech"
		}
	}

	return c.JSON(notifs)
}

// DeleteReadingHistoryItem — DELETE /api/user/history/:id
func DeleteReadingHistoryItem(c *fiber.Ctx) error {
	userID := getUserID(c)
	id := c.Params("id")

	var history models.ReadingHistory
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&history).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Riwayat tidak ditemukan"})
	}

	database.DB.Delete(&history)
	return c.JSON(fiber.Map{"success": true, "message": "Riwayat berhasil dihapus"})
}

// ClearReadingHistory — DELETE /api/user/history
func ClearReadingHistory(c *fiber.Ctx) error {
	userID := getUserID(c)
	database.DB.Where("user_id = ?", userID).Delete(&models.ReadingHistory{})
	return c.JSON(fiber.Map{"success": true, "message": "Semua riwayat baca berhasil dibersihkan"})
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

// DeleteNotification — DELETE /api/user/notifications/:id
func DeleteNotification(c *fiber.Ctx) error {
	userID := getUserID(c)
	id := c.Params("id")

	var notif models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&notif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Notifikasi tidak ditemukan"})
	}

	database.DB.Delete(&notif)
	return c.JSON(fiber.Map{"success": true, "message": "Notifikasi berhasil dihapus"})
}

// ClearAllNotifications — DELETE /api/user/notifications
func ClearAllNotifications(c *fiber.Ctx) error {
	userID := getUserID(c)
	database.DB.Where("user_id = ?", userID).Delete(&models.Notification{})
	return c.JSON(fiber.Map{"success": true, "message": "Semua notifikasi berhasil dibersihkan"})
}

// Helper to create system notification
func CreateSystemNotification(userID uint, message string, linkSlug string) error {
	notif := models.Notification{
		UserID:   userID,
		SenderID: 0,
		Type:     "system",
		Message:  message,
		NewsSlug: linkSlug,
		IsRead:   false,
	}
	return database.DB.Create(&notif).Error
}

