package routes

import (
	"news-portal-backend/handlers"
	"news-portal-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// Auth routes
	api.Post("/auth/register", handlers.Register)
	api.Post("/auth/login", handlers.Login)
	api.Post("/auth/logout", handlers.Logout)
	api.Get("/auth/me", middleware.Protected(), handlers.GetMe)

	// News routes (Public)
	api.Get("/news", handlers.GetNews)
	api.Get("/news/featured", handlers.GetFeaturedNews)
	api.Get("/news/trending", handlers.GetTrendingNews)
	api.Get("/news/slug/:slug", handlers.GetNewsBySlug)
	api.Get("/news/:id", handlers.GetNewsById)
	api.Get("/categories", handlers.GetCategories)

	// News routes (Admin)
	api.Post("/news", middleware.Protected(), middleware.AdminOnly(), handlers.CreateNews)
	api.Put("/news/:id", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateNews)
	api.Delete("/news/:id", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteNews)

	// Comment routes
	api.Get("/comments", middleware.Protected(), middleware.AdminOnly(), handlers.GetComments)
	api.Get("/comments/news/:newsId", handlers.GetCommentsByNews)
	api.Post("/comments", middleware.Protected(), handlers.CreateComment)
	api.Delete("/comments/:id", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteComment)
}