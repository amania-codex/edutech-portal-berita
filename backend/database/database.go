package database

import "news-portal-backend/models"
import "gorm.io/driver/mysql"
import "gorm.io/gorm"
import "log"
import "golang.org/x/crypto/bcrypt"

var DB *gorm.DB

func Connect() {
	dsnBase := "root:@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
	dbBase, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL!\n", err)
	}

	dbBase.Exec("CREATE DATABASE IF NOT EXISTS news_portal;")

	dsn := "root:@tcp(127.0.0.1:3306)/news_portal?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!\n", err)
	}

	log.Println("Database connection successfully opened")

	err = database.AutoMigrate(&models.User{}, &models.Article{}, &models.Comment{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database!\n", err)
	}
	log.Println("Database Migrated")

	DB = database

	// Seed admin user
	var count int64
	DB.Model(&models.User{}).Where("email = ?", "admin@edutech.id").Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
		DB.Create(&models.User{
			Name:     "Administrator",
			Email:    "admin@edutech.id",
			Password: string(hash),
			Role:     "admin",
		})
		log.Println("Admin user seeded.")
	}
}