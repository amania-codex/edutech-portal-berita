package database

import (
	"fmt"
	"log"
	"os"

	"news-portal-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func Connect() {
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "news_portal")

	dsnBase := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort)
	dbBase, err := gorm.Open(mysql.Open(dsnBase), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL!\n", err)
	}

	dbBase.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName))

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!\n", err)
	}

	log.Println("Database connection successfully opened")

	err = database.AutoMigrate(
		&models.User{}, 
		&models.Article{}, 
		&models.Comment{}, 
		&models.Category{},
		&models.ReadingHistory{},
		&models.Bookmark{},
		&models.Notification{},
		&models.CommentLike{},
		&models.Banner{},
		&models.Event{},
		&models.Speaker{},
		&models.Material{},
		&models.BankAccount{},
		&models.Registration{},
	)
	if err != nil {
		log.Fatal("Failed to auto-migrate database!\n", err)
	}
	log.Println("Database Migrated")

	DB = database

	// Check if seeding is disabled in production
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" && os.Getenv("SEED_DB") != "true" {
		log.Println("Production mode active: Database seeding skipped.")
		return
	}

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