package database

import (
	"fmt"
	"log"
	"os"

	"news-portal-backend/models"

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

	// Kolom menu kategori baru pertama kali dibuat → isi menu bawaan setelah migrasi
	menuIsNew := !database.Migrator().HasColumn(&models.Category{}, "menu_order")

	err = database.AutoMigrate(
		&models.User{}, 
		&models.Article{}, 
		&models.Comment{}, 
		&models.Category{},
		&models.ArticleCategory{},
		&models.ArticleRedirect{},
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
		&models.NewsletterSubscriber{},
		&models.Setting{},
		&models.Page{},
		&models.ActivityLog{},
		&models.Video{},
	)
	if err != nil {
		log.Fatal("Failed to auto-migrate database!\n", err)
	}
	log.Println("Database Migrated")

	DB = database

	// Arsip berita dimuat otomatis bila database masih kosong (juga di produksi)
	SeedNews()
	if menuIsNew {
		seedMenuDefaults()
	}

	// Di produksi hanya membuat akun bila ADMIN_PASSWORD diisi (lihat seeder.go)
	SeedAdmin()
}