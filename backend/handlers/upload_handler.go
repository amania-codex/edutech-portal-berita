package handlers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadImage handles image uploads and saves them to the public/uploads directory
func UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak ada file yang diunggah",
		})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format file tidak didukung. Gunakan JPG, PNG, GIF, atau WebP",
		})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ukuran file maksimum 10MB",
		})
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d%s", timestamp, ext)
	savePath := fmt.Sprintf("./public/uploads/%s", filename)

	// Save file
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan file: " + err.Error(),
		})
	}

	// Return the public URL
	url := fmt.Sprintf("/uploads/%s", filename)
	return c.JSON(fiber.Map{
		"url":      url,
		"filename": filename,
		"size":     file.Size,
	})
}
