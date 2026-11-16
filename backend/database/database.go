package database

import (
	"fmt"
	"log"
	"os"
	"time"

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

	// Seed regular user
	var userCount int64
	DB.Model(&models.User{}).Where("email = ?", "user@edutech.id").Count(&userCount)
	if userCount == 0 {
		hashUser, _ := bcrypt.GenerateFromPassword([]byte("user123"), 10)
		DB.Create(&models.User{
			Name:     "Regular User",
			Email:    "user@edutech.id",
			Password: string(hashUser),
			Role:     "user",
		})
		log.Println("Regular user seeded.")
	}

	// Seed categories
	var catCount int64
	DB.Model(&models.Category{}).Count(&catCount)
	if catCount == 0 {
		initialCategories := []models.Category{
			{Name: "Teknologi", Slug: "teknologi"},
			{Name: "Pendidikan", Slug: "pendidikan"},
			{Name: "Kampus", Slug: "kampus"},
			{Name: "Event", Slug: "event"},
			{Name: "Riset", Slug: "riset"},
			{Name: "Loker", Slug: "loker"},
			{Name: "Video", Slug: "video"},
			{Name: "Artikel", Slug: "artikel"},
			{Name: "Headline", Slug: "headline"},
			{Name: "Beasiswa", Slug: "beasiswa"},
		}
		for _, cat := range initialCategories {
			DB.Create(&cat)
		}
		log.Println("Categories seeded.")
	}

	// Seed banner if none exists
	var bannerCount int64
	DB.Model(&models.Banner{}).Count(&bannerCount)
	if bannerCount == 0 {
		DB.Create(&models.Banner{
			Title:     "SENIKO 2026 - Seminar Nasional Informatika dan Komputer",
			ImageUrl:  "https://images.unsplash.com/photo-1540575467063-178a50c2df87?auto=format&fit=crop&w=1600&q=80",
			TargetUrl: "https://seniko.id",
			IsActive:  true,
		})
		log.Println("Default banner seeded.")
	}

	// Seed sample article if none exists
	var articleCount int64
	DB.Model(&models.Article{}).Count(&articleCount)
	if articleCount == 0 {
		var author models.User
		DB.Where("email = ?", "admin@edutech.id").First(&author)

		articles := []models.Article{
			{
				Title:       "Monitoring KKN UAP, Program Mahasiswa di Pekon Madaraya Didorong Berkelanjutan dan Sejalan dengan SDGs",
				Slug:        "monitoring-kkn-uap-pekon-madaraya-sdgs",
				Content:     "Program Kuliah Kerja Nyata (KKN) Universitas Aisyah Pringsewu (UAP) di Pekon Madaraya mendapatkan monitoring intensif...",
				ImageURL:    "https://images.unsplash.com/photo-1529156069898-49953e39b3ac?auto=format&fit=crop&w=1200&q=80",
				Category:    "Pendidikan",
				AuthorID:    author.ID,
				Author:      author.Name,
				Tags:        "KKN,UAP,SDGs,Mahasiswa",
				ViewCount:   229,
				IsFeatured:  true,
				IsTrending:  true,
				IsPublished: true,
				Status:      "published",
			},
		}
		for _, a := range articles {
			DB.Create(&a)
		}
		log.Println("Sample article seeded.")
	}

	// Seed 100+ comprehensive educational articles (min 10 per category)
	SeedArticles()

	// Seed Bank Accounts from apabaeid_amaniaedutech if none exists
	var bankCount int64
	DB.Model(&models.BankAccount{}).Count(&bankCount)
	if bankCount == 0 {
		bca := models.BankAccount{
			BankName:      "BCA",
			AccountNumber: "12424124",
			AccountHolder: "PT Edutech",
			Logo:          "https://upload.wikimedia.org/wikipedia/commons/5/5c/Bank_Central_Asia.svg",
			IsActive:      true,
		}
		DB.Create(&bca)
		log.Println("Default bank account seeded.")
	}

	// Seed sample Event if none exists
	var eventCount int64
	DB.Model(&models.Event{}).Count(&eventCount)
	if eventCount == 0 {
		var bcaAccount models.BankAccount
		DB.First(&bcaAccount)

		now := time.Now()
		startTime := now.Add(7 * 24 * time.Hour)
		endTime := startTime.Add(3 * time.Hour)
		premiumPrice := 50000

		sampleEvent := models.Event{
			Title:           "Workshop AI & Future Education: Transformasi Pembelajaran Digital",
			Slug:            "workshop-ai-future-education-transformasi-digital",
			Description:     "Pelajari implementasi praktis teknologi kecerdasan buatan (AI) terkini dalam pendidikan tinggi dan sekolah menengah. Diselenggarakan secara daring interaktif via Zoom.",
			Venue:           "Zoom Meeting & Live Streaming YouTube",
			StartTime:       startTime,
			EndTime:         endTime,
			Quota:           150,
			Price:           0,
			PremiumPrice:    &premiumPrice,
			CertificateLink: "https://edutech.id/certificates/ai-future-education",
			ImageURL:        "https://images.unsplash.com/photo-1540575467063-178a50c2df87?auto=format&fit=crop&w=1200&q=80",
			Status:          "published",
			Speakers: []models.Speaker{
				{
					Name: "Dr. Nurul Hidayat, M.Kom.",
					Role: "Pakar Kecerdasan Buatan & Dosen Universitas Jenderal Soedirman",
				},
				{
					Name: "Ir. Ahmad Fauzi, S.T., M.T.",
					Role: "Head of Technology Edutech ID",
				},
			},
			Materials: []models.Material{
				{
					Title: "Slide Presentasi Pengenalan AI dalam Kurikulum",
					Type:  "file",
					Link:  "https://example.com/slides-ai-education.pdf",
				},
			},
			BankAccounts: []models.BankAccount{bcaAccount},
		}
		DB.Create(&sampleEvent)
		log.Println("Sample event, speakers, and materials seeded.")
	}
}