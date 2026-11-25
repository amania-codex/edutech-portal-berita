package handlers

import (
	"fmt"
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RefreshClaims menyamakan data sesi dengan database: peran yang diganti admin
// langsung berlaku, dan akun yang dinonaktifkan/dihapus langsung tidak bisa dipakai.
func RefreshClaims(claims *Claims) bool {
	if claims == nil || claims.UserID == 0 {
		return false
	}
	var u models.User
	if err := database.DB.Select("id", "name", "email", "role", "suspended").Where("id = ?", claims.UserID).Limit(1).Find(&u).Error; err != nil || u.ID == 0 || u.Suspended {
		return false
	}
	claims.Role = u.Role
	claims.Name = u.Name
	claims.Email = u.Email
	return true
}

func currentClaims(c *fiber.Ctx) *Claims {
	claims, _ := c.Locals("user").(*Claims)
	return claims
}

// canManageAllNews — admin & editor boleh mengubah berita siapa pun; penulis hanya miliknya.
func canManageAllNews(claims *Claims) bool {
	return claims != nil && (claims.Role == models.RoleAdmin || claims.Role == models.RoleEditor)
}

func canEditArticle(claims *Claims, article *models.Article) bool {
	if canManageAllNews(claims) {
		return true
	}
	return claims != nil && claims.Role == models.RolePenulis && article.AuthorID != 0 && article.AuthorID == claims.UserID
}

// logActivity mencatat aksi ke riwayat. Gagal mencatat tidak membatalkan aksi.
func logActivity(c *fiber.Ctx, action, targetType string, targetID uint, title, detail string) {
	claims := currentClaims(c)
	entry := models.ActivityLog{
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		TargetTitle: truncate(title, 255),
		Detail:      detail,
		IP:          c.IP(),
	}
	if claims != nil {
		entry.UserID = claims.UserID
		entry.UserName = claims.Name
		entry.UserRole = claims.Role
	}
	database.DB.Create(&entry)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// GetActivity — GET /api/admin/activity
// Query: page, limit, user_id, type (berita|pengguna|video), target_id, action.
// Admin & editor melihat semua; penulis melihat aksinya sendiri dan aksi orang lain pada beritanya.
func GetActivity(c *fiber.Ctx) error {
	claims := currentClaims(c)
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	q := database.DB.Model(&models.ActivityLog{})
	if !canManageAllNews(claims) {
		q = q.Where("(user_id = ? OR (target_type = 'berita' AND target_id IN (SELECT id FROM articles WHERE author_id = ?)))", claims.UserID, claims.UserID)
	}
	if v := c.QueryInt("user_id", 0); v > 0 {
		q = q.Where("user_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("type")); v != "" {
		q = q.Where("target_type = ?", v)
	}
	if v := c.QueryInt("target_id", 0); v > 0 {
		q = q.Where("target_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("action")); v != "" {
		q = q.Where("action = ?", v)
	}

	var total int64
	q = q.Session(&gorm.Session{}) // aman dipakai ulang untuk hitung & ambil data
	q.Count(&total)
	c.Set("X-Total-Count", fmt.Sprintf("%d", total))

	logs := make([]models.ActivityLog, 0)
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil riwayat aktivitas."})
	}
	return c.JSON(logs)
}
