package database

import (
	"log"
	"os"

	"news-portal-backend/models"

	"golang.org/x/crypto/bcrypt"
)

// Password bawaan untuk pengembangan lokal. Repo ini publik, jadi tidak boleh dipakai di produksi.
const defaultAdminPassword = "admin123"

// SeedAdmin membuat akun administrator bila belum ada.
//
// ADMIN_EMAIL dan ADMIN_PASSWORD menentukan akunnya. Di produksi ADMIN_PASSWORD wajib diisi
// (minimal 10 karakter); tanpa itu tidak ada akun yang dibuat. Akun yang sudah ada tidak diubah,
// kecuali bila passwordnya masih password bawaan: diganti dengan ADMIN_PASSWORD.
func SeedAdmin() {
	email := getEnv("ADMIN_EMAIL", "admin@edutech.id")
	password := os.Getenv("ADMIN_PASSWORD")
	production := os.Getenv("APP_ENV") == "production"
	strong := len(password) >= 10 && password != defaultAdminPassword

	var admin models.User
	if err := DB.Where("email = ?", email).Limit(1).Find(&admin).Error; err != nil {
		log.Println("Failed to check admin user:", err)
		return
	}

	if admin.ID != 0 {
		if !production || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(defaultAdminPassword)) != nil {
			return
		}
		if !strong {
			log.Printf("PERINGATAN: akun %s masih memakai password bawaan. Isi ADMIN_PASSWORD (minimal 10 karakter) lalu nyalakan ulang.", email)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			log.Println("Failed to hash admin password:", err)
			return
		}
		if err := DB.Model(&admin).Update("password", string(hash)).Error; err != nil {
			log.Println("Failed to update admin password:", err)
			return
		}
		log.Printf("Password bawaan akun %s sudah diganti.", email)
		return
	}

	if production && !strong {
		log.Println("Akun admin tidak dibuat: ADMIN_PASSWORD kosong atau kurang dari 10 karakter.")
		return
	}
	if password == "" {
		password = defaultAdminPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Println("Failed to hash admin password:", err)
		return
	}
	admin = models.User{Name: "Administrator", Email: email, Password: string(hash), Role: "admin"}
	if err := DB.Create(&admin).Error; err != nil {
		log.Println("Failed to seed admin user:", err)
		return
	}
	log.Println("Admin user seeded.")
}
