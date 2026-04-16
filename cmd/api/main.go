package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/config"
	"github.com/falashlion/urban-sanctuary-api/internal/handler"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/cache"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/payment"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/falashlion/urban-sanctuary-api/internal/router"
	"github.com/falashlion/urban-sanctuary-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	// --- Logger ---
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Caller().
		Str("service", "urban-sanctuary-api").
		Logger()

	// --- Config ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		log = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("service", "urban-sanctuary-api").
			Logger()
	}

	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Msg("Starting Urban Sanctuary API")

	// --- Database ---
	dbPool, err := repository.NewDBPool(cfg.DB.DSN, cfg.DB.MaxConns, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbPool.Close()

	// --- Redis ---
	var redisClient *cache.RedisClient
	if cfg.Redis.URL != "" {
		redisClient, err = cache.NewRedisClient(cfg.Redis.URL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, running without cache")
		} else {
			defer redisClient.Close()
		}
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepository(dbPool)
	authRepo := repository.NewAuthRepository(dbPool)
	propRepo := repository.NewPropertyRepository(dbPool)
	bookingRepo := repository.NewBookingRepository(dbPool)
	paymentRepo := repository.NewPaymentRepository(dbPool)
	adminRepo := repository.NewAdminRepository(dbPool)

	// --- Payment Clients ---
	mtnClient := payment.NewMTNMoMoClient(
		cfg.Payment.MTN.APIKey, cfg.Payment.MTN.APISecret,
		cfg.Payment.MTN.BaseURL, cfg.Payment.MTN.CallbackURL, log,
	)
	orangeClient := payment.NewOrangeMoneyClient(
		cfg.Payment.Orange.APIKey, cfg.Payment.Orange.APISecret,
		cfg.Payment.Orange.BaseURL, cfg.Payment.Orange.CallbackURL, log,
	)

	// --- Services ---
	authSvc := service.NewAuthService(userRepo, authRepo, redisClient, cfg, log)
	userSvc := service.NewUserService(userRepo, bookingRepo, adminRepo, log)
	propSvc := service.NewPropertyService(propRepo, userRepo, authRepo, nil, log) // S3 optional
	bookingSvc := service.NewBookingService(bookingRepo, propRepo, userRepo, authRepo, log)
	paymentSvc := service.NewPaymentService(paymentRepo, bookingRepo, userRepo, mtnClient, orangeClient, log)
	adminSvc := service.NewAdminService(userRepo, propRepo, bookingRepo, adminRepo, log)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authSvc, log)
	userHandler := handler.NewUserHandler(userSvc, log)
	propHandler := handler.NewPropertyHandler(propSvc, log)
	bookingHandler := handler.NewBookingHandler(bookingSvc, log)
	paymentHandler := handler.NewPaymentHandler(paymentSvc, log)
	adminHandler := handler.NewAdminHandler(adminSvc, log)
	healthHandler := handler.NewHealthHandler(dbPool, redisClient)

	// --- Router ---
	r := router.Setup(router.Config{
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		PropertyHandler: propHandler,
		BookingHandler:  bookingHandler,
		PaymentHandler:  paymentHandler,
		AdminHandler:    adminHandler,
		HealthHandler:   healthHandler,
		Redis:           redisClient,
		AccessSecret:    cfg.JWT.AccessSecret,
		CORSOrigins:     cfg.CORS.Origins,
		Log:             log,
	})

	// --- Server ---
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info().Msg("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Server forced to shutdown")
		}
	}()

	log.Info().Str("addr", srv.Addr).Msg("Server listening")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed to start")
	}

	log.Info().Msg("Server exited")
}
