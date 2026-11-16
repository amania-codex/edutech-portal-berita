package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
)

// Helper: Generate ticket code TIC-XXXXXX
func generateTicketCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return fmt.Sprintf("TIC-%s", string(b))
}

// GetEvents — GET /api/events (Public)
func GetEvents(c *fiber.Ctx) error {
	var events []models.Event
	status := c.Query("status", "published")

	query := database.DB.Preload("Speakers").Preload("BankAccounts").Preload("Materials")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("start_time asc").Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil daftar agenda & event"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    events,
	})
}

// GetEventBySlug — GET /api/events/slug/:slug (Public)
func GetEventBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var event models.Event

	if err := database.DB.Preload("Speakers").Preload("Materials").Preload("BankAccounts").Where("slug = ?", slug).First(&event).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}

	// Hitung peserta terdaftar
	var registeredCount int64
	database.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN (?, ?)", event.ID, "verified", "waiting").Count(&registeredCount)

	return c.JSON(fiber.Map{
		"success":          true,
		"data":             event,
		"registered_count": registeredCount,
		"is_quota_full":    event.Quota > 0 && registeredCount >= int64(event.Quota),
	})
}

// GetEventById — GET /api/events/:id
func GetEventById(c *fiber.Ctx) error {
	id := c.Params("id")
	var event models.Event

	if err := database.DB.Preload("Speakers").Preload("Materials").Preload("BankAccounts").First(&event, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    event,
	})
}

// CreateEvent — POST /api/admin/events (Admin Only)
func CreateEvent(c *fiber.Ctx) error {
	type SpeakerInput struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Photo string `json:"photo"`
	}
	type MaterialInput struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		FilePath string `json:"file_path"`
		Link     string `json:"link"`
	}
	type Input struct {
		Title           string          `json:"title"`
		Slug            string          `json:"slug"`
		Description     string          `json:"description"`
		Venue           string          `json:"venue"`
		StartTime       string          `json:"start_time"`
		EndTime         string          `json:"end_time"`
		Quota           int             `json:"quota"`
		Price           int             `json:"price"`
		PremiumPrice    *int            `json:"premium_price"`
		CertificateLink string          `json:"certificate_link"`
		ImageUrl        string          `json:"image_url"`
		Status          string          `json:"status"`
		BankAccountIDs  []uint          `json:"bank_account_ids"`
		Speakers        []SpeakerInput  `json:"speakers"`
		Materials       []MaterialInput `json:"materials"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format payload tidak valid"})
	}

	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Description) == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Judul dan deskripsi event wajib diisi"})
	}

	slug := input.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(input.Title, " ", "-"))
	}

	start, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		start, _ = time.Parse("2006-01-02T15:04", input.StartTime)
		if start.IsZero() {
			start = time.Now().Add(24 * time.Hour)
		}
	}
	end, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		end, _ = time.Parse("2006-01-02T15:04", input.EndTime)
		if end.IsZero() {
			end = start.Add(2 * time.Hour)
		}
	}

	status := input.Status
	if status == "" {
		status = "published"
	}

	event := models.Event{
		Title:           input.Title,
		Slug:            slug,
		Description:     input.Description,
		Venue:           input.Venue,
		StartTime:       start,
		EndTime:         end,
		Quota:           input.Quota,
		Price:           input.Price,
		PremiumPrice:    input.PremiumPrice,
		CertificateLink: input.CertificateLink,
		ImageURL:        input.ImageUrl,
		Status:          status,
	}

	if len(input.BankAccountIDs) > 0 {
		var banks []models.BankAccount
		database.DB.Where("id IN ?", input.BankAccountIDs).Find(&banks)
		event.BankAccounts = banks
	}

	if err := database.DB.Create(&event).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan event: " + err.Error()})
	}

	// Insert speakers
	for _, s := range input.Speakers {
		if strings.TrimSpace(s.Name) != "" {
			speaker := models.Speaker{
				EventID: event.ID,
				Name:    s.Name,
				Role:    s.Role,
				Photo:   s.Photo,
			}
			database.DB.Create(&speaker)
		}
	}

	// Insert materials
	for _, m := range input.Materials {
		if strings.TrimSpace(m.Title) != "" {
			mat := models.Material{
				EventID:  event.ID,
				Title:    m.Title,
				Type:     m.Type,
				FilePath: m.FilePath,
				Link:     m.Link,
			}
			database.DB.Create(&mat)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event berhasil dibuat",
		"data":    event,
	})
}

// UpdateEvent — PUT /api/admin/events/:id (Admin Only)
func UpdateEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	var event models.Event

	if err := database.DB.First(&event, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}

	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Input tidak valid"})
	}

	if err := database.DB.Model(&event).Updates(input).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui event"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event berhasil diperbarui",
		"data":    event,
	})
}

// DeleteEvent — DELETE /api/admin/events/:id (Admin Only)
func DeleteEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.Event{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghapus event"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event berhasil dihapus",
	})
}

// RegisterEvent — POST /api/events/:id/register (Public / User)
func RegisterEvent(c *fiber.Ctx) error {
	eventId := c.Params("id")
	var event models.Event

	if err := database.DB.First(&event, eventId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}

	type Input struct {
		Name         string  `json:"name"`
		Email        string  `json:"email"`
		Phone        string  `json:"phone"`
		Tier         string  `json:"tier"` // free, premium
		PaymentProof string  `json:"payment_proof"`
		TotalAmount  float64 `json:"total_amount"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format data tidak valid"})
	}

	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Nama dan email wajib diisi"})
	}

	// Kuota check
	if event.Quota > 0 {
		var activeCount int64
		database.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN (?, ?)", event.ID, "verified", "waiting").Count(&activeCount)
		if activeCount >= int64(event.Quota) {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Mohon maaf, kuota peserta untuk event ini sudah penuh."})
		}
	}

	tier := input.Tier
	if tier == "" {
		tier = "free"
	}

	status := "verified"
	if tier == "premium" || event.Price > 0 {
		status = "waiting" // Perlu verifikasi admin jika berbayar
		if input.PaymentProof == "" {
			status = "pending"
		}
	}

	// User ID jika login
	var userId *uint
	if userClaims, ok := c.Locals("user").(*Claims); ok && userClaims != nil {
		uid := userClaims.UserID
		userId = &uid
	}

	ticketCode := generateTicketCode()

	reg := models.Registration{
		TicketCode:   ticketCode,
		UserID:       userId,
		EventID:      event.ID,
		Name:         input.Name,
		Email:        input.Email,
		Phone:        input.Phone,
		PaymentProof: input.PaymentProof,
		Status:       status,
		Tier:         tier,
		TotalAmount:  input.TotalAmount,
	}

	if err := database.DB.Create(&reg).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memproses pendaftaran: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Pendaftaran berhasil disimpan",
		"ticket_code": ticketCode,
		"status":      status,
		"data":        reg,
	})
}

// GetUserTickets — GET /api/user/tickets (Protected)
func GetUserTickets(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*Claims)
	if !ok || userClaims == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Akses ditolak"})
	}

	var registrations []models.Registration
	if err := database.DB.Preload("Event").Where("user_id = ? OR email = ?", userClaims.UserID, userClaims.Email).Order("created_at desc").Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil data tiket"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    registrations,
	})
}

// GetAdminRegistrations — GET /api/admin/registrations (Admin Only)
func GetAdminRegistrations(c *fiber.Ctx) error {
	var registrations []models.Registration
	query := database.DB.Preload("Event").Preload("User").Order("created_at desc")

	if eventId := c.Query("event_id"); eventId != "" {
		query = query.Where("event_id = ?", eventId)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memuat pendaftaran"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    registrations,
	})
}

// UpdateRegistrationStatus — PUT /api/admin/registrations/:id/status (Admin Only)
func UpdateRegistrationStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	type Input struct {
		Status string `json:"status"` // verified, rejected, waiting, pending
	}

	var input Input
	if err := c.BodyParser(&input); err != nil || input.Status == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Status tidak valid"})
	}

	if err := database.DB.Model(&models.Registration{}).Where("id = ?", id).Update("status", input.Status).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui status pendaftaran"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status pendaftaran berhasil diperbarui",
	})
}

// GetBankAccounts — GET /api/bank-accounts (Public & Admin)
func GetBankAccounts(c *fiber.Ctx) error {
	var accounts []models.BankAccount
	query := database.DB.Order("created_at asc")
	if c.Query("active_only") == "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil daftar rekening"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    accounts,
	})
}

// CreateBankAccount — POST /api/admin/bank-accounts (Admin Only)
func CreateBankAccount(c *fiber.Ctx) error {
	type Input struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountHolder string `json:"account_holder"`
		Logo          string `json:"logo"`
		IsActive      *bool  `json:"is_active"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Input tidak valid"})
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	bank := models.BankAccount{
		BankName:      input.BankName,
		AccountNumber: input.AccountNumber,
		AccountHolder: input.AccountHolder,
		Logo:          input.Logo,
		IsActive:      isActive,
	}

	if err := database.DB.Create(&bank).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan rekening bank"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Rekening bank berhasil ditambahkan",
		"data":    bank,
	})
}

// DeleteBankAccount — DELETE /api/admin/bank-accounts/:id (Admin Only)
func DeleteBankAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DB.Delete(&models.BankAccount{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghapus rekening"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Rekening berhasil dihapus",
	})
}
