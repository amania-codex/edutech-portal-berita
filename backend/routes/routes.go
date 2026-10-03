package routes

import (
	"news-portal-backend/handlers"
	"news-portal-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// Serve uploaded files statically
	app.Static("/uploads", "./public/uploads")

	// Upload route (protected for all users to allow avatar upload)
	api.Post("/upload", middleware.Protected(), handlers.UploadImage)

	// Auth routes
	api.Post("/auth/register", handlers.Register)
	api.Post("/auth/login", handlers.Login)
	api.Post("/auth/logout", handlers.Logout)
	api.Post("/auth/forgot-password", handlers.ForgotPassword)
	api.Post("/auth/reset-password", handlers.ResetPassword)
	api.Get("/auth/me", middleware.Protected(), handlers.GetMe)
	api.Put("/auth/password", middleware.Protected(), handlers.ChangePassword)
	api.Put("/auth/profile", middleware.Protected(), handlers.UpdateProfile)

	// News routes (Public)
	api.Get("/news", handlers.GetNews)
	api.Get("/news/featured", handlers.GetFeaturedNews)
	api.Get("/news/trending", handlers.GetTrendingNews)
	api.Get("/news/slug/:slug", handlers.GetNewsBySlug)
	api.Get("/news/:id", handlers.GetNewsById)
	api.Get("/categories", handlers.GetAllCategories)
	api.Post("/categories", middleware.Protected(), middleware.AdminOnly(), handlers.CreateCategory)
	api.Put("/categories/:id", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateCategory)
	api.Delete("/categories/:id", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteCategory)
	// News routes (Admin)
	api.Post("/news", middleware.Protected(), middleware.AdminOnly(), handlers.CreateNews)
	api.Put("/news/:id", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateNews)
	api.Delete("/news/:id", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteNews)

	// Comment routes
	api.Get("/comments", middleware.Protected(), middleware.AdminOnly(), handlers.GetComments)
	api.Get("/comments/news/:newsId", handlers.GetCommentsByNews)
	api.Post("/comments", middleware.Protected(), handlers.CreateComment)
	api.Post("/comments/:id/like", middleware.Protected(), handlers.LikeComment)
	api.Delete("/comments/:id", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteComment)

	// User dashboard routes
	api.Get("/user/history", middleware.Protected(), handlers.GetReadingHistory)
	api.Post("/user/history", middleware.Protected(), handlers.AddReadingHistory)
	api.Get("/user/bookmarks", middleware.Protected(), handlers.GetBookmarks)
	api.Post("/user/bookmarks", middleware.Protected(), handlers.ToggleBookmark)
	api.Get("/user/bookmarks/:news_id", middleware.Protected(), handlers.CheckBookmark)
	api.Get("/user/notifications", middleware.Protected(), handlers.GetNotifications)
	api.Post("/user/notifications/read-all", middleware.Protected(), handlers.ReadAllNotifications)
	api.Post("/user/notifications/:id/read", middleware.Protected(), handlers.ReadNotification)
}