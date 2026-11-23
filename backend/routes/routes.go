package routes

import (
	"time"

	"news-portal-backend/handlers"
	"news-portal-backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// ID wajib angka (:id<int>). GORM memperlakukan ID berupa teks sebagai potongan SQL mentah,
	// jadi tanpa batasan ini parameter URL bisa dipakai untuk SQL injection.

	// Rate limiter for auth endpoints (max 20 requests per minute)
	authLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Terlalu banyak percobaan. Silakan coba lagi setelah 1 menit.",
			})
		},
	})

	// Rate limiter for image upload (max 30 uploads per minute)
	uploadLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Batas kuota unggah tercapai. Silakan coba sesaat lagi.",
			})
		},
	})

	// Rate limiter penghitung tayangan (max 60 per menit per IP)
	viewLimiter := limiter.New(limiter.Config{Max: 60, Expiration: 1 * time.Minute})

	// Serve uploaded files statically
	// (gambar hasil migrasi WordPress ada di ./public/uploads/wp/<tahun>/<bulan>/...)
	app.Static("/uploads", "./public/uploads")

	// Upload route (protected for all users to allow avatar upload)
	api.Post("/upload", uploadLimiter, middleware.Protected(), handlers.UploadImage)
	api.Post("/upload-proof", uploadLimiter, handlers.UploadImage)

	// Auth routes
	api.Post("/auth/register", authLimiter, handlers.Register)
	api.Post("/auth/login", authLimiter, handlers.Login)
	api.Post("/auth/logout", handlers.Logout)
	api.Post("/auth/forgot-password", authLimiter, handlers.ForgotPassword)
	api.Post("/auth/reset-password", authLimiter, handlers.ResetPassword)
	api.Get("/auth/me", middleware.Protected(), handlers.GetMe)
	api.Get("/auth/profile", middleware.Protected(), handlers.GetMe) // dipakai layout frontend
	api.Put("/auth/password", middleware.Protected(), handlers.ChangePassword)
	api.Put("/auth/profile", middleware.Protected(), handlers.UpdateProfile)

	// News routes (Public)
	api.Get("/news", handlers.GetNews)
	api.Get("/news/featured", handlers.GetFeaturedNews)
	api.Get("/news/trending", handlers.GetTrendingNews)
	api.Get("/news/popular", handlers.GetPopularNews)
	api.Get("/news/by-categories", handlers.GetNewsByCategories)
	api.Get("/news/resolve/:slug", handlers.ResolveNewsSlug)
	api.Get("/news/slug/:slug", handlers.GetNewsBySlug)
	api.Post("/news/:id<int>/view", viewLimiter, handlers.TrackNewsView)
	api.Get("/news/category/:category", handlers.GetNews)
	api.Get("/news/:id<int>", handlers.GetNewsById)
	api.Get("/categories", handlers.GetAllCategories)
	api.Get("/settings", handlers.GetSettings)
	api.Put("/admin/settings", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateSettings)
	api.Get("/pages", handlers.GetPages)
	api.Get("/pages/slug/:slug", handlers.GetPageBySlug)
	api.Post("/admin/pages", middleware.Protected(), middleware.AdminOnly(), handlers.CreatePage)
	api.Put("/admin/pages/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.UpdatePage)
	api.Delete("/admin/pages/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeletePage)
	api.Post("/categories", middleware.Protected(), middleware.AdminOnly(), handlers.CreateCategory)
	api.Put("/categories/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateCategory)
	api.Delete("/categories/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteCategory)
	// News routes (Admin)
	api.Post("/news", middleware.Protected(), middleware.AdminOnly(), handlers.CreateNews)
	api.Put("/news/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateNews)
	api.Delete("/news/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteNews)

	api.Post("/newsletter", limiter.New(limiter.Config{Max: 10, Expiration: 1 * time.Minute}), handlers.Subscribe)
	api.Get("/admin/newsletter", middleware.Protected(), middleware.AdminOnly(), handlers.GetSubscribers)
	api.Get("/admin/stats", middleware.Protected(), middleware.AdminOnly(), handlers.GetAdminStats)

	// Banner / Poster routes
	api.Get("/banners", handlers.GetActiveBanners)
	api.Get("/admin/banners", middleware.Protected(), middleware.AdminOnly(), handlers.GetAllBanners)
	api.Post("/admin/banners", middleware.Protected(), middleware.AdminOnly(), handlers.CreateBanner)
	api.Put("/admin/banners/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateBanner)
	api.Put("/admin/banners/:id<int>/toggle", middleware.Protected(), middleware.AdminOnly(), handlers.ToggleBannerStatus)
	api.Delete("/admin/banners/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteBanner)

	// Comment routes
	api.Get("/comments", middleware.Protected(), middleware.AdminOnly(), handlers.GetComments)
	api.Get("/comments/news/:newsId<int>", handlers.GetCommentsByNews)
	api.Post("/comments", middleware.Protected(), handlers.CreateComment)
	api.Post("/comments/:id<int>/like", middleware.Protected(), handlers.LikeComment)
	api.Delete("/comments/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteComment)

	// User dashboard routes
	api.Get("/user/history", middleware.Protected(), handlers.GetReadingHistory)
	api.Post("/user/history", middleware.Protected(), handlers.AddReadingHistory)
	api.Delete("/user/history/:id<int>", middleware.Protected(), handlers.DeleteReadingHistoryItem)
	api.Delete("/user/history", middleware.Protected(), handlers.ClearReadingHistory)

	api.Get("/user/bookmarks", middleware.Protected(), handlers.GetBookmarks)
	api.Post("/user/bookmarks", middleware.Protected(), handlers.ToggleBookmark)
	api.Get("/user/bookmarks/:news_id<int>", middleware.Protected(), handlers.CheckBookmark)

	api.Get("/user/notifications", middleware.Protected(), handlers.GetNotifications)
	api.Post("/user/notifications/read-all", middleware.Protected(), handlers.ReadAllNotifications)
	api.Post("/user/notifications/:id<int>/read", middleware.Protected(), handlers.ReadNotification)
	api.Delete("/user/notifications/:id<int>", middleware.Protected(), handlers.DeleteNotification)
	api.Delete("/user/notifications", middleware.Protected(), handlers.ClearAllNotifications)

	// User Tickets
	api.Get("/user/tickets", middleware.Protected(), handlers.GetUserTickets)

	// Event routes (Public)
	api.Get("/events", handlers.GetEvents)
	api.Get("/events/slug/:slug", handlers.GetEventBySlug)
	api.Get("/events/:id<int>", handlers.GetEventById)
	api.Post("/events/:id<int>/register", handlers.RegisterEvent)

	// Bank Account routes
	api.Get("/bank-accounts", handlers.GetBankAccounts)

	// Admin Event & Registration routes
	api.Get("/admin/events", middleware.Protected(), middleware.AdminOnly(), handlers.GetAdminEvents)
	api.Post("/admin/events", middleware.Protected(), middleware.AdminOnly(), handlers.CreateEvent)
	api.Put("/admin/events/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateEvent)
	api.Delete("/admin/events/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteEvent)

	api.Get("/admin/registrations", middleware.Protected(), middleware.AdminOnly(), handlers.GetAdminRegistrations)
	api.Put("/admin/registrations/:id<int>/status", middleware.Protected(), middleware.AdminOnly(), handlers.UpdateRegistrationStatus)

	api.Post("/admin/bank-accounts", middleware.Protected(), middleware.AdminOnly(), handlers.CreateBankAccount)
	api.Delete("/admin/bank-accounts/:id<int>", middleware.Protected(), middleware.AdminOnly(), handlers.DeleteBankAccount)
}