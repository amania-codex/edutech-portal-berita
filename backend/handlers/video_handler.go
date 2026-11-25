package handlers

import (
	"net/url"
	"regexp"
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

var youtubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// parseYouTubeID menerima alamat YouTube apa pun (watch, youtu.be, shorts, live, embed) atau ID-nya saja.
func parseYouTubeID(raw string) string {
	raw = strings.TrimSpace(raw)
	if youtubeIDPattern.MatchString(raw) {
		return raw
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	host = strings.TrimPrefix(host, "m.")
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	id := ""
	switch host {
	case "youtu.be":
		id = parts[0]
	case "youtube.com", "music.youtube.com", "youtube-nocookie.com":
		if v := u.Query().Get("v"); v != "" {
			id = v
		} else if len(parts) >= 2 && (parts[0] == "shorts" || parts[0] == "live" || parts[0] == "embed" || parts[0] == "v") {
			id = parts[1]
		}
	}
	if youtubeIDPattern.MatchString(id) {
		return id
	}
	return ""
}

// GetVideos — GET /api/videos (publik): video yang tampil di situs
func GetVideos(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 12)
	if limit <= 0 || limit > 100 {
		limit = 12
	}
	videos := make([]models.Video, 0)
	if err := database.DB.Where("is_published = ?", true).Order("sort_order asc, created_at desc").Limit(limit).Find(&videos).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil video."})
	}
	return c.JSON(videos)
}

// GetAdminVideos — GET /api/admin/videos
func GetAdminVideos(c *fiber.Ctx) error {
	videos := make([]models.Video, 0)
	if err := database.DB.Order("sort_order asc, created_at desc").Find(&videos).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil video."})
	}
	return c.JSON(videos)
}

type videoInput struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	IsPublished *bool  `json:"isPublished"`
}

func (in *videoInput) parse() (string, string) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	id := parseYouTubeID(in.URL)
	switch {
	case id == "":
		return "", "Alamat YouTube tidak dikenali. Salin alamat video dari YouTube (mis. https://youtu.be/...)."
	case in.Title == "" || len([]rune(in.Title)) > 255:
		return "", "Judul video wajib diisi (maksimal 255 karakter)."
	case len([]rune(in.Description)) > 2000:
		return "", "Keterangan maksimal 2000 karakter."
	case in.SortOrder < 0 || in.SortOrder > 9999:
		return "", "Urutan harus 0–9999."
	}
	return id, ""
}

// CreateVideo — POST /api/admin/videos
func CreateVideo(c *fiber.Ctx) error {
	var in videoInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid."})
	}
	id, msg := in.parse()
	if msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	video := models.Video{Title: in.Title, YoutubeID: id, Description: in.Description, SortOrder: in.SortOrder, IsPublished: in.IsPublished == nil || *in.IsPublished}
	if err := database.DB.Create(&video).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan video."})
	}
	logActivity(c, "buat", "video", video.ID, video.Title, "https://youtu.be/"+video.YoutubeID)
	return c.Status(201).JSON(video)
}

// UpdateVideo — PUT /api/admin/videos/:id
func UpdateVideo(c *fiber.Ctx) error {
	var video models.Video
	if err := database.DB.First(&video, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Video tidak ditemukan."})
	}
	var in videoInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid."})
	}
	id, msg := in.parse()
	if msg != "" {
		return c.Status(400).JSON(fiber.Map{"message": msg})
	}
	updates := map[string]interface{}{"title": in.Title, "youtube_id": id, "description": in.Description, "sort_order": in.SortOrder}
	if in.IsPublished != nil {
		updates["is_published"] = *in.IsPublished
	}
	if err := database.DB.Model(&models.Video{}).Where("id = ?", video.ID).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui video."})
	}
	database.DB.First(&video, video.ID)
	logActivity(c, "ubah", "video", video.ID, video.Title, "")
	return c.JSON(video)
}

// DeleteVideo — DELETE /api/admin/videos/:id
func DeleteVideo(c *fiber.Ctx) error {
	var video models.Video
	if err := database.DB.First(&video, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Video tidak ditemukan."})
	}
	if err := database.DB.Unscoped().Delete(&video).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal menghapus video."})
	}
	logActivity(c, "hapus", "video", video.ID, video.Title, "")
	return c.JSON(fiber.Map{"message": "Video dihapus."})
}
