package handlers

import (
	"fmt"
	"net/mail"
	"strings"

	"news-portal-backend/database"
	"news-portal-backend/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var roleLabels = map[string]string{
	models.RoleAdmin:   "Admin",
	models.RoleEditor:  "Editor",
	models.RolePenulis: "Penulis",
	models.RoleUser:    "Pembaca",
}

type staffUserRow struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	Suspended    bool    `json:"suspended"`
	CreatedAt    string  `json:"createdAt"`
	LastLoginAt  *string `json:"lastLoginAt"`
	ArticleCount int64   `json:"articleCount"`
}

// GetUsers — GET /api/admin/users?role=staff|admin|editor|penulis|user&q=
func GetUsers(c *fiber.Ctx) error {
	q := database.DB.Table("users").Where("users.deleted_at IS NULL")
	switch role := c.Query("role"); role {
	case "", "all":
	case "staff":
		q = q.Where("users.role IN ?", []string{models.RoleAdmin, models.RoleEditor, models.RolePenulis})
	default:
		q = q.Where("users.role = ?", role)
	}
	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + s + "%"
		q = q.Where("(users.name LIKE ? OR users.email LIKE ?)", like, like)
	}
	rows := make([]staffUserRow, 0)
	err := q.Select(`users.id, users.name, users.email, users.role, users.suspended,
		DATE_FORMAT(users.created_at, '%Y-%m-%dT%H:%i:%s') AS created_at,
		DATE_FORMAT(users.last_login_at, '%Y-%m-%dT%H:%i:%s') AS last_login_at,
		(SELECT COUNT(*) FROM articles a WHERE a.author_id = users.id AND a.deleted_at IS NULL) AS article_count`).
		Order("FIELD(users.role, 'admin', 'editor', 'penulis', 'user'), users.name").
		Limit(500).
		Scan(&rows).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil daftar pengguna."})
	}
	return c.JSON(rows)
}

type userInput struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	Suspended *bool  `json:"suspended"`
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && strings.Contains(s, ".")
}

// CreateUser — POST /api/admin/users (Admin): tambah akun redaksi/pembaca
func CreateUser(c *fiber.Ctx) error {
	var in userInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid."})
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Role == "" {
		in.Role = models.RolePenulis
	}
	switch {
	case len([]rune(in.Name)) < 2 || len([]rune(in.Name)) > 100:
		return c.Status(400).JSON(fiber.Map{"message": "Nama wajib diisi (2–100 karakter)."})
	case !validEmail(in.Email):
		return c.Status(400).JSON(fiber.Map{"message": "Alamat email tidak valid."})
	case len(in.Password) < 8:
		return c.Status(400).JSON(fiber.Map{"message": "Password minimal 8 karakter."})
	case roleLabels[in.Role] == "":
		return c.Status(400).JSON(fiber.Map{"message": "Peran tidak dikenal."})
	}
	var count int64
	database.DB.Unscoped().Model(&models.User{}).Where("email = ?", in.Email).Count(&count)
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"message": "Email sudah dipakai akun lain."})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memproses password."})
	}
	user := models.User{Name: in.Name, Email: in.Email, Password: string(hash), Role: in.Role}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat akun."})
	}
	logActivity(c, "tambah", "pengguna", user.ID, user.Name, fmt.Sprintf("Akun baru %s sebagai %s", user.Email, roleLabels[user.Role]))
	return c.Status(201).JSON(fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "suspended": false})
}

// UpdateUser — PUT /api/admin/users/:id (Admin): ubah nama/email/peran, nonaktifkan, atau reset password
func UpdateUser(c *fiber.Ctx) error {
	me := currentClaims(c)
	var user models.User
	if err := database.DB.First(&user, c.Params("id")).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Pengguna tidak ditemukan."})
	}
	var in userInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Input tidak valid."})
	}

	updates := map[string]interface{}{}
	var changes []string
	if name := strings.TrimSpace(in.Name); name != "" && name != user.Name {
		if len([]rune(name)) < 2 || len([]rune(name)) > 100 {
			return c.Status(400).JSON(fiber.Map{"message": "Nama 2–100 karakter."})
		}
		updates["name"] = name
		changes = append(changes, fmt.Sprintf("nama: %s → %s", user.Name, name))
	}
	if email := strings.ToLower(strings.TrimSpace(in.Email)); email != "" && email != user.Email {
		if !validEmail(email) {
			return c.Status(400).JSON(fiber.Map{"message": "Alamat email tidak valid."})
		}
		var count int64
		database.DB.Unscoped().Model(&models.User{}).Where("email = ? AND id <> ?", email, user.ID).Count(&count)
		if count > 0 {
			return c.Status(409).JSON(fiber.Map{"message": "Email sudah dipakai akun lain."})
		}
		updates["email"] = email
		changes = append(changes, fmt.Sprintf("email: %s → %s", user.Email, email))
	}

	removesAdmin := false
	if in.Role != "" && in.Role != user.Role {
		if roleLabels[in.Role] == "" {
			return c.Status(400).JSON(fiber.Map{"message": "Peran tidak dikenal."})
		}
		if user.ID == me.UserID {
			return c.Status(400).JSON(fiber.Map{"message": "Anda tidak dapat mengubah peran akun sendiri."})
		}
		updates["role"] = in.Role
		changes = append(changes, fmt.Sprintf("peran: %s → %s", roleLabels[user.Role], roleLabels[in.Role]))
		removesAdmin = user.Role == models.RoleAdmin
	}
	if in.Suspended != nil && *in.Suspended != user.Suspended {
		if user.ID == me.UserID {
			return c.Status(400).JSON(fiber.Map{"message": "Anda tidak dapat menonaktifkan akun sendiri."})
		}
		updates["suspended"] = *in.Suspended
		if *in.Suspended {
			changes = append(changes, "akun dinonaktifkan")
			removesAdmin = removesAdmin || user.Role == models.RoleAdmin
		} else {
			changes = append(changes, "akun diaktifkan kembali")
		}
	}
	if removesAdmin {
		var admins int64
		database.DB.Model(&models.User{}).Where("role = ? AND suspended = ? AND id <> ?", models.RoleAdmin, false, user.ID).Count(&admins)
		if admins == 0 {
			return c.Status(400).JSON(fiber.Map{"message": "Harus ada minimal satu admin aktif."})
		}
	}
	if in.Password != "" {
		if len(in.Password) < 8 {
			return c.Status(400).JSON(fiber.Map{"message": "Password minimal 8 karakter."})
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal memproses password."})
		}
		updates["password"] = string(hash)
		changes = append(changes, "password direset oleh admin")
	}

	if len(updates) == 0 {
		return c.JSON(fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "suspended": user.Suspended})
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memperbarui pengguna."})
	}
	database.DB.First(&user, user.ID)
	logActivity(c, "ubah", "pengguna", user.ID, user.Name, strings.Join(changes, "; "))
	return c.JSON(fiber.Map{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "suspended": user.Suspended})
}
