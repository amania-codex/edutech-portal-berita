package main

import (
	"log"
	"news-portal-backend/database"
	"news-portal-backend/models"
	"news-portal-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database.Connect()

	seedData()

	app := fiber.New(fiber.Config{
		AppName: "EduTech News Portal API v1.0",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4321", // Astro dev server
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	routes.Setup(app)

	log.Fatal(app.Listen(":3000"))
}

func seedData() {
	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin12345"), bcrypt.DefaultCost)
		admin := models.User{
			Name:     "Admin EduTech",
			Email:    "admin@edutech.id",
			Password: string(hash),
			Role:     "admin",
		}
		database.DB.Create(&admin)
		log.Println("Seed data User berhasil ditambahkan.")
	}

	var count int64
	database.DB.Model(&models.Article{}).Count(&count)
	if count == 0 {
		var author models.User
		database.DB.Where("email = ?", "admin@edutech.id").First(&author)

		articles := []models.Article{
			{
				Title:       "Monitoring KKN UAP, Program Mahasiswa di Pekon Madaraya Didorong Berkelanjutan dan Sejalan dengan SDGs",
				Slug:        "monitoring-kkn-uap-pekon-madaraya-sdgs",
				Excerpt:     "Program KKN UAP di Pekon Madaraya mendapat perhatian khusus dari universitas.",
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
			database.DB.Create(&a)
		}
		log.Println("Seed data berhasil ditambahkan.")
	}
}