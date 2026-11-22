package main

import (
	"log"
	"os"

	"news-portal-backend/database"
	"news-portal-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// allowedOrigins — alamat frontend yang boleh mengakses API.
// Produksi: CORS_ORIGINS="https://edutech.id,https://www.edutech.id"
func allowedOrigins() string {
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		return v
	}
	return "http://localhost:4321,http://localhost:4322,http://127.0.0.1:4321,http://127.0.0.1:4322,http://localhost:80,http://localhost"
}

func main() {
	loadDotEnv(".env")
	database.Connect()

	app := fiber.New(fiber.Config{
		AppName:   "EduTech News Portal API v1.0",
		BodyLimit: 50 * 1024 * 1024, // 50MB for Quill base64 images

		// Di belakang nginx (satu server), pakai IP asli pengunjung untuk rate limiter.
		// nginx: proxy_set_header X-Real-IP $remote_addr;
		// Header hanya dipercaya bila datang dari proxy lokal, jadi tidak bisa dipalsukan pengunjung.
		ProxyHeader:             "X-Real-IP",
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1"},
		EnableIPValidation:      true,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
		ExposeHeaders:    "X-Total-Count",
	}))

	routes.Setup(app)

	log.Fatal(app.Listen(listenAddr()))
}