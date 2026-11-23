package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Waktu event diisi dalam WIB
var wib = time.FixedZone("WIB", 7*60*60)

// Event yang boleh dilihat publik
var publicEventStatuses = []string{"published", "completed"}

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

func eventQuery(c *fiber.Ctx) *gorm.DB {
	q := database.DB.Preload("Speakers").Preload("BankAccounts", "is_active = ?", true).Preload("Materials")
	if !isAdminUser(c) {
		q = q.Where("status IN ?", publicEventStatuses)
	}
	return q
}

// GetEvents — GET /api/events (publik: hanya yang terbit; admin dengan ?status=all melihat semua)
func GetEvents(c *fiber.Ctx) error {
	events := make([]models.Event, 0)
	q := database.DB.Preload("Speakers").Preload("BankAccounts", "is_active = ?", true).Preload("Materials")
	if !(c.Query("status") == "all" && isAdminUser(c)) {
		q = q.Where("status IN ?", publicEventStatuses)
	}
	if err := q.Order("start_time asc").Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil daftar agenda & event"})
	}
	return c.JSON(fiber.Map{"success": true, "data": events})
}

func countActiveRegistrations(eventID uint) int64 {
	var n int64
	database.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", eventID, []string{"verified", "waiting", "pending"}).Count(&n)
	return n
}

// GetEventBySlug — GET /api/events/slug/:slug
func GetEventBySlug(c *fiber.Ctx) error {
	var event models.Event
	if err := eventQuery(c).Where("slug = ?", c.Params("slug")).First(&event).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}
	registered := countActiveRegistrations(event.ID)
	return c.JSON(fiber.Map{
		"success":          true,
		"data":             event,
		"registered_count": registered,
		"is_quota_full":    event.Quota > 0 && registered >= int64(event.Quota),
	})
}

// GetEventById — GET /api/events/:id
func GetEventById(c *fiber.Ctx) error {
	var event models.Event
	if err := eventQuery(c).First(&event, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": event})
}

// GetAdminEvents — GET /api/admin/events (admin: semua event + jumlah pendaftar)
func GetAdminEvents(c *fiber.Ctx) error {
	events := make([]models.Event, 0)
	if err := database.DB.Preload("Speakers").Preload("BankAccounts").Preload("Materials").Order("start_time desc").Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil event"})
	}
	type row struct {
		models.Event
		Registered int64 `json:"registeredCount"`
	}
	out := make([]row, 0, len(events))
	for _, e := range events {
		out = append(out, row{e, countActiveRegistrations(e.ID)})
	}
	return c.JSON(fiber.Map{"success": true, "data": out})
}

type eventInput struct {
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	Venue           string `json:"venue"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Quota           int    `json:"quota"`
	Price           int    `json:"price"`
	PremiumPrice    *int   `json:"premium_price"`
	CertificateLink string `json:"certificate_link"`
	ImageUrl        string `json:"image_url"`
	Status          string `json:"status"`
	BankAccountIDs  []uint `json:"bank_account_ids"`
	Speakers        []struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Photo string `json:"photo"`
	} `json:"speakers"`
	Materials []struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		FilePath string `json:"file_path"`
		Link     string `json:"link"`
	} `json:"materials"`

	start, end time.Time
}

// parseEventTime: "2006-01-02T15:04" dari form (WIB) atau RFC3339
func parseEventTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, s, wib); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("format waktu tidak valid")
}

func (in *eventInput) validate(id uint) string {
	in.Title = strings.TrimSpace(in.Title)
	in.Venue = strings.TrimSpace(in.Venue)
	in.Slug = generateSlug(in.Slug)
	if in.Slug == "" {
		in.Slug = generateSlug(in.Title)
	}
	if in.Status == "" {
		in.Status = "published"
	}
	var err error
	if in.start, err = parseEventTime(in.StartTime); err != nil {
		return "Waktu mulai tidak valid."
	}
	if in.end, err = parseEventTime(in.EndTime); err != nil {
		return "Waktu selesai tidak valid."
	}
	switch {
	case len(in.Title) < 3 || len(in.Title) > 255:
		return "Judul event 3–255 karakter."
	case in.Slug == "":
		return "Slug tidak valid."
	case strings.TrimSpace(in.Description) == "":
		return "Deskripsi event wajib diisi."
	case in.Venue == "":
		return "Lokasi / platform wajib diisi."
	case !in.end.After(in.start):
		return "Waktu selesai harus setelah waktu mulai."
	case in.Quota < 0 || in.Price < 0 || (in.PremiumPrice != nil && *in.PremiumPrice < 0):
		return "Kuota dan harga tidak boleh negatif."
	case in.Status != "published" && in.Status != "draft" && in.Status != "completed" && in.Status != "cancelled":
		return "Status event tidak dikenal."
	case !safeURL(strings.TrimSpace(in.ImageUrl), true) || !safeURL(strings.TrimSpace(in.CertificateLink), true):
		return "Tautan gambar/sertifikat harus diawali http://, https://, atau /"
	}
	if in.PremiumPrice != nil && *in.PremiumPrice == 0 {
		in.PremiumPrice = nil
	}
	for _, m := range in.Materials {
		if !safeURL(strings.TrimSpace(m.Link), true) || !safeURL(strings.TrimSpace(m.FilePath), true) {
			return "Tautan materi harus diawali http://, https://, atau /"
		}
	}
	var clash int64
	database.DB.Unscoped().Model(&models.Event{}).Where("slug = ? AND id <> ?", in.Slug, id).Count(&clash)
	if clash > 0 {
		return "Slug event sudah dipakai."
	}
	return ""
}

// saveEvent menyimpan event beserta pemateri, materi, dan rekening (diganti seluruhnya).
func saveEvent(event *models.Event, in *eventInput) error {
	event.Title, event.Slug, event.Description, event.Venue = in.Title, in.Slug, in.Description, in.Venue
	event.StartTime, event.EndTime, event.Quota, event.Price, event.PremiumPrice = in.start, in.end, in.Quota, in.Price, in.PremiumPrice
	event.CertificateLink, event.ImageURL, event.Status = strings.TrimSpace(in.CertificateLink), strings.TrimSpace(in.ImageUrl), in.Status
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Speakers", "Materials", "BankAccounts", "Registrations").Save(event).Error; err != nil {
			return err
		}
		// premium_price NULL perlu ditulis eksplisit
		if err := tx.Model(event).UpdateColumn("premium_price", event.PremiumPrice).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("event_id = ?", event.ID).Delete(&models.Speaker{}).Error; err != nil {
			return err
		}
		for _, s := range in.Speakers {
			if strings.TrimSpace(s.Name) == "" {
				continue
			}
			if err := tx.Create(&models.Speaker{EventID: event.ID, Name: strings.TrimSpace(s.Name), Role: strings.TrimSpace(s.Role), Photo: strings.TrimSpace(s.Photo)}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("event_id = ?", event.ID).Delete(&models.Material{}).Error; err != nil {
			return err
		}
		for _, m := range in.Materials {
			if strings.TrimSpace(m.Title) == "" {
				continue
			}
			kind := m.Type
			if kind != "video" {
				kind = "file"
			}
			if err := tx.Create(&models.Material{EventID: event.ID, Title: strings.TrimSpace(m.Title), Type: kind, FilePath: strings.TrimSpace(m.FilePath), Link: strings.TrimSpace(m.Link)}).Error; err != nil {
				return err
			}
		}
		var banks []models.BankAccount
		if len(in.BankAccountIDs) > 0 {
			if err := tx.Where("id IN ?", in.BankAccountIDs).Find(&banks).Error; err != nil {
				return err
			}
		}
		return tx.Model(event).Association("BankAccounts").Replace(banks)
	})
}

// CreateEvent — POST /api/admin/events (admin)
func CreateEvent(c *fiber.Ctx) error {
	var in eventInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format payload tidak valid"})
	}
	if msg := in.validate(0); msg != "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": msg})
	}
	var event models.Event
	if err := saveEvent(&event, &in); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan event"})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "message": "Event berhasil dibuat", "data": event})
}

// UpdateEvent — PUT /api/admin/events/:id (admin)
func UpdateEvent(c *fiber.Ctx) error {
	var event models.Event
	if err := database.DB.First(&event, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}
	var in eventInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Input tidak valid"})
	}
	if msg := in.validate(event.ID); msg != "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": msg})
	}
	if err := saveEvent(&event, &in); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui event"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Event berhasil diperbarui", "data": event})
}

// DeleteEvent — DELETE /api/admin/events/:id (admin)
func DeleteEvent(c *fiber.Ctx) error {
	res := database.DB.Delete(&models.Event{}, c.Params("id"))
	if res.Error != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghapus event"})
	}
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Event berhasil dihapus"})
}

// RegisterEvent — POST /api/events/:id/register (publik / pengguna)
func RegisterEvent(c *fiber.Ctx) error {
	var event models.Event
	if err := database.DB.Where("status = ?", "published").First(&event, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Event tidak ditemukan atau pendaftaran ditutup"})
	}
	if time.Now().After(event.EndTime) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Event sudah selesai, pendaftaran ditutup."})
	}

	var input struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Tier         string `json:"tier"` // free, premium
		PaymentProof string `json:"payment_proof"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Format data tidak valid"})
	}
	input.Name, input.Email = strings.TrimSpace(input.Name), strings.ToLower(strings.TrimSpace(input.Email))
	if len(input.Name) < 2 || len(input.Name) > 255 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Nama wajib diisi"})
	}
	if a, err := mail.ParseAddress(input.Email); err != nil || a.Address != input.Email {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Email tidak valid"})
	}
	if !safeURL(strings.TrimSpace(input.PaymentProof), true) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Bukti pembayaran tidak valid"})
	}

	tier := "free"
	if input.Tier == "premium" && event.PremiumPrice != nil && *event.PremiumPrice > 0 {
		tier = "premium"
	}
	// Biaya dihitung di server, bukan dari isian pendaftar
	amount := float64(event.Price)
	if tier == "premium" {
		amount = float64(*event.PremiumPrice)
	}
	status := "verified"
	if amount > 0 {
		status = "waiting" // perlu verifikasi admin
		if strings.TrimSpace(input.PaymentProof) == "" {
			status = "pending"
		}
	}

	var dup int64
	database.DB.Model(&models.Registration{}).Where("event_id = ? AND email = ? AND status <> ?", event.ID, input.Email, "rejected").Count(&dup)
	if dup > 0 {
		return c.Status(409).JSON(fiber.Map{"success": false, "error": "Email ini sudah terdaftar di event ini. Cek menu Tiket Saya."})
	}
	if event.Quota > 0 && countActiveRegistrations(event.ID) >= int64(event.Quota) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Mohon maaf, kuota peserta untuk event ini sudah penuh."})
	}

	var userID *uint
	if claims := optionalClaims(c); claims != nil {
		uid := claims.UserID
		userID = &uid
	}

	reg := models.Registration{
		TicketCode: generateTicketCode(), UserID: userID, EventID: event.ID,
		Name: input.Name, Email: input.Email, Phone: strings.TrimSpace(input.Phone),
		PaymentProof: strings.TrimSpace(input.PaymentProof), Status: status, Tier: tier, TotalAmount: amount,
	}
	if err := database.DB.Create(&reg).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memproses pendaftaran"})
	}
	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Pendaftaran berhasil disimpan",
		"ticket_code": reg.TicketCode,
		"status":      status,
		"data":        reg,
	})
}

// GetUserTickets — GET /api/user/tickets
func GetUserTickets(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*Claims)
	if !ok || userClaims == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Akses ditolak"})
	}
	registrations := make([]models.Registration, 0)
	if err := database.DB.Preload("Event").Preload("Event.Materials").
		Where("user_id = ? OR email = ?", userClaims.UserID, strings.ToLower(userClaims.Email)).
		Order("created_at desc").Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil data tiket"})
	}
	return c.JSON(fiber.Map{"success": true, "data": registrations})
}

// GetAdminRegistrations — GET /api/admin/registrations (admin)
func GetAdminRegistrations(c *fiber.Ctx) error {
	registrations := make([]models.Registration, 0)
	query := database.DB.Preload("Event").Order("created_at desc")
	if eventID := c.QueryInt("event_id", 0); eventID > 0 {
		query = query.Where("event_id = ?", eventID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&registrations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memuat pendaftaran"})
	}
	return c.JSON(fiber.Map{"success": true, "data": registrations})
}

// UpdateRegistrationStatus — PUT /api/admin/registrations/:id/status (admin)
func UpdateRegistrationStatus(c *fiber.Ctx) error {
	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Status tidak valid"})
	}
	switch input.Status {
	case "verified", "rejected", "waiting", "pending":
	default:
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Status tidak valid"})
	}
	var reg models.Registration
	if err := database.DB.First(&reg, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Pendaftaran tidak ditemukan"})
	}
	if err := database.DB.Model(&reg).Update("status", input.Status).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal memperbarui status pendaftaran"})
	}
	if reg.UserID != nil && (input.Status == "verified" || input.Status == "rejected") {
		var event models.Event
		database.DB.Select("title").First(&event, reg.EventID)
		msg := fmt.Sprintf("Pendaftaran Anda untuk \"%s\" telah diverifikasi. Tiket %s aktif.", event.Title, reg.TicketCode)
		if input.Status == "rejected" {
			msg = fmt.Sprintf("Pendaftaran Anda untuk \"%s\" (tiket %s) ditolak. Hubungi admin untuk bantuan.", event.Title, reg.TicketCode)
		}
		CreateSystemNotification(*reg.UserID, msg, "")
	}
	return c.JSON(fiber.Map{"success": true, "message": "Status pendaftaran berhasil diperbarui"})
}

// GetBankAccounts — GET /api/bank-accounts (publik: yang aktif; admin ?all=1: semua)
func GetBankAccounts(c *fiber.Ctx) error {
	accounts := make([]models.BankAccount, 0)
	query := database.DB.Order("created_at asc")
	if !(c.Query("all") == "1" && isAdminUser(c)) {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&accounts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal mengambil daftar rekening"})
	}
	return c.JSON(fiber.Map{"success": true, "data": accounts})
}

// CreateBankAccount — POST /api/admin/bank-accounts (admin)
func CreateBankAccount(c *fiber.Ctx) error {
	var input struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountHolder string `json:"account_holder"`
		Logo          string `json:"logo"`
		IsActive      *bool  `json:"is_active"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Input tidak valid"})
	}
	input.BankName, input.AccountNumber, input.AccountHolder = strings.TrimSpace(input.BankName), strings.TrimSpace(input.AccountNumber), strings.TrimSpace(input.AccountHolder)
	if input.BankName == "" || input.AccountNumber == "" || input.AccountHolder == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Nama bank, nomor rekening, dan atas nama wajib diisi"})
	}
	if !safeURL(strings.TrimSpace(input.Logo), true) {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Logo harus tautan http(s) atau /uploads/"})
	}
	bank := models.BankAccount{
		BankName: input.BankName, AccountNumber: input.AccountNumber, AccountHolder: input.AccountHolder,
		Logo: strings.TrimSpace(input.Logo), IsActive: input.IsActive == nil || *input.IsActive,
	}
	if err := database.DB.Create(&bank).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menyimpan rekening bank"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Rekening bank berhasil ditambahkan", "data": bank})
}

// DeleteBankAccount — DELETE /api/admin/bank-accounts/:id (admin)
func DeleteBankAccount(c *fiber.Ctx) error {
	res := database.DB.Delete(&models.BankAccount{}, c.Params("id"))
	if res.Error != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Gagal menghapus rekening"})
	}
	if res.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Rekening tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Rekening berhasil dihapus"})
}
