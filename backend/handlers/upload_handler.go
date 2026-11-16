package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
)

// UploadImage handles image uploads, compresses, and saves them
func UploadImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak ada file yang diunggah",
		})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
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
	if fileHeader.Size > 10*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ukuran file maksimum 10MB",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membuka file",
		})
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal membaca format gambar",
		})
	}

	// Pastikan direktori tujuan ada
	uploadDir := "./public/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mempersiapkan direktori penyimpanan",
		})
	}

	// Kompresi dan resize (lebar maksimum 1600px)
	if img.Bounds().Dx() > 1600 {
		img = imaging.Resize(img, 1600, 0, imaging.Lanczos)
	}

	timestamp := time.Now().UnixNano()
	var filename string
	var savePath string
	var saveErr error

	if ext == ".png" {
		filename = fmt.Sprintf("%d.png", timestamp)
		savePath = filepath.Join(uploadDir, filename)
		saveErr = imaging.Save(img, savePath)
	} else {
		filename = fmt.Sprintf("%d.jpg", timestamp)
		savePath = filepath.Join(uploadDir, filename)
		saveErr = imaging.Save(img, savePath, imaging.JPEGQuality(85))
	}

	if saveErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan gambar: " + saveErr.Error(),
		})
	}

	// Return the public URL
	url := fmt.Sprintf("/uploads/%s", filename)
	return c.JSON(fiber.Map{
		"url":      url,
		"filename": filename,
		"size":     fileHeader.Size,
	})
}
