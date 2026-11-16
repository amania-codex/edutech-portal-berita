package main

import (
	"log"
	"news-portal-backend/database"
	"news-portal-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.Connect()

	app := fiber.New(fiber.Config{
		AppName:   "EduTech News Portal API v1.0",
		BodyLimit: 50 * 1024 * 1024, // 50MB for Quill base64 images
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4321", // Astro dev server
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	routes.Setup(app)

	log.Fatal(app.Listen(":3000"))
}