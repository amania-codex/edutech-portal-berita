package database

import (
	"log"

	"news-portal-backend/models"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin seeds the default administrator account
func SeedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("email = ?", "admin@edutech.id").Count(&count)
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		if err != nil {
			log.Println("Failed to hash admin password:", err)
			return
		}

		admin := models.User{
			Name:     "Administrator",
			Email:    "admin@edutech.id",
			Password: string(hash),
			Role:     "admin",
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Println("Failed to seed admin user:", err)
			return
		}
		log.Println("Admin user seeded.")
	}
}
