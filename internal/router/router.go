package router

import (
	"github.com/falashlion/urban-sanctuary-api/internal/handler"
	"github.com/falashlion/urban-sanctuary-api/internal/middleware"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/cache"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Config holds all dependencies needed to build the router.
type Config struct {
	AuthHandler     *handler.AuthHandler
	UserHandler     *handler.UserHandler
	PropertyHandler *handler.PropertyHandler
	BookingHandler  *handler.BookingHandler
	PaymentHandler  *handler.PaymentHandler
	AdminHandler    *handler.AdminHandler
	HealthHandler   *handler.HealthHandler

	Redis        *cache.RedisClient
	AccessSecret string
	CORSOrigins  []string
	Log          zerolog.Logger
}

// Setup creates and configures the Gin router with all routes.
func Setup(cfg Config) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(cfg.Log))
	router.Use(middleware.Logger(cfg.Log))
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(middleware.GlobalRateLimiter(cfg.Redis))

	// API v1
	v1 := router.Group("/api/v1")

	// Health check
	v1.GET("/health", cfg.HealthHandler.Health)

	// --- Auth routes (public) ---
	auth := v1.Group("/auth")
	auth.Use(middleware.AuthRateLimiter(cfg.Redis))
	{
		auth.POST("/register", cfg.AuthHandler.Register)
		auth.POST("/login", cfg.AuthHandler.Login)
		auth.POST("/refresh", cfg.AuthHandler.Refresh)
		auth.POST("/logout", cfg.AuthHandler.Logout)
		auth.POST("/otp/request", cfg.AuthHandler.RequestOTP)
		auth.POST("/otp/verify", cfg.AuthHandler.VerifyOTP)
		auth.POST("/password/reset-request", cfg.AuthHandler.PasswordResetRequest)
		auth.POST("/password/reset", cfg.AuthHandler.PasswordReset)
	}

	// 2FA routes (require auth)
	twoFA := auth.Group("/2fa", middleware.JWTAuth(cfg.AccessSecret))
	{
		twoFA.POST("/enable", cfg.AuthHandler.Enable2FA)
		twoFA.POST("/confirm", cfg.AuthHandler.Confirm2FA)
		twoFA.POST("/verify", cfg.AuthHandler.Verify2FA)
	}

	// --- User routes (authenticated) ---
	users := v1.Group("/users", middleware.JWTAuth(cfg.AccessSecret))
	{
		users.GET("/me", cfg.UserHandler.GetProfile)
		users.PATCH("/me", cfg.UserHandler.UpdateProfile)
		users.PATCH("/me/password", cfg.UserHandler.ChangePassword)
		users.GET("/me/bookings", cfg.UserHandler.GetBookings)
		users.GET("/me/loyalty", cfg.UserHandler.GetLoyalty)
		users.GET("/me/notifications", cfg.UserHandler.GetNotifications)
	}

	// --- Property routes (public listing, authenticated for CRUD) ---
	properties := v1.Group("/properties")
	{
		// Public routes
		properties.GET("", cfg.PropertyHandler.List)
		properties.GET("/:id", cfg.PropertyHandler.GetByID)
		properties.GET("/:id/availability", cfg.PropertyHandler.GetAvailability)

		// Authenticated routes (homeowner or admin)
		authProperties := properties.Group("", middleware.JWTAuth(cfg.AccessSecret), middleware.RequireOwnerOrAdmin())
		{
			authProperties.POST("", cfg.PropertyHandler.Create)
			authProperties.PATCH("/:id", cfg.PropertyHandler.Update)
			authProperties.POST("/:id/images", cfg.PropertyHandler.UploadImages)
			authProperties.DELETE("/:id/images/:imageId", cfg.PropertyHandler.DeleteImage)
		}

		// Admin-only routes
		adminProperties := properties.Group("", middleware.JWTAuth(cfg.AccessSecret), middleware.RequireAdmin())
		{
			adminProperties.PATCH("/:id/status", cfg.PropertyHandler.UpdateStatus)
		}
	}

	// --- Booking routes (authenticated) ---
	bookings := v1.Group("/bookings", middleware.JWTAuth(cfg.AccessSecret))
	{
		bookings.POST("", cfg.BookingHandler.Create)
		bookings.GET("/:id", cfg.BookingHandler.GetByID)
		bookings.POST("/:id/cancel", cfg.BookingHandler.Cancel)
		bookings.POST("/:id/review", cfg.BookingHandler.SubmitReview)
	}

	// --- Payment routes (authenticated) ---
	payments := v1.Group("/payments", middleware.JWTAuth(cfg.AccessSecret))
	{
		payments.POST("/initiate", cfg.PaymentHandler.Initiate)
		payments.GET("/:id/status", cfg.PaymentHandler.GetStatus)
	}

	// --- Webhook routes (public, validated by signature) ---
	webhooks := v1.Group("/webhooks")
	{
		webhooks.POST("/mtn", cfg.PaymentHandler.MTNWebhook)
		webhooks.POST("/orange", cfg.PaymentHandler.OrangeWebhook)
	}

	// --- Admin routes (admin only) ---
	admin := v1.Group("/admin", middleware.JWTAuth(cfg.AccessSecret), middleware.RequireAdmin())
	{
		admin.GET("/users", cfg.AdminHandler.ListUsers)
		admin.PATCH("/users/:id/role", cfg.AdminHandler.UpdateUserRole)
		admin.PATCH("/users/:id/status", cfg.AdminHandler.UpdateUserStatus)

		admin.GET("/properties", cfg.AdminHandler.ListProperties)
		admin.GET("/bookings", cfg.AdminHandler.ListBookings)

		admin.GET("/tickets", cfg.AdminHandler.ListTickets)
		admin.GET("/tickets/:id", cfg.AdminHandler.GetTicket)
		admin.POST("/tickets/:id/reply", cfg.AdminHandler.ReplyToTicket)
		admin.PATCH("/tickets/:id/status", cfg.AdminHandler.UpdateTicketStatus)

		admin.GET("/permissions", cfg.AdminHandler.GetPermissions)
		admin.PUT("/permissions", cfg.AdminHandler.UpdatePermissions)

		admin.GET("/config", cfg.AdminHandler.GetSiteConfig)
		admin.PUT("/config/:key", cfg.AdminHandler.UpdateSiteConfig)

		admin.GET("/audit-logs", cfg.AdminHandler.GetAuditLogs)
	}

	return router
}
