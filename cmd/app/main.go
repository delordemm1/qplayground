package main

import (
	"encoding/gob"
	"log"
	"log/slog"
	"net/http"

	"github.com/delordemm1/qplayground/internal/controller/web"
	"github.com/delordemm1/qplayground/internal/core/config"
	"github.com/delordemm1/qplayground/internal/modules/media"
	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize logger
	platform.InitLogger()

	// Initialize database
	pool := config.InitDatabase()
	defer pool.Close()

	// Initialize session manager
	sessionConfig := config.DefaultSessionConfig()
	sessionManager := config.InitSession(pool, sessionConfig)

	// Initialize Inertia
	gob.Register(platform.FlashMessage{})
	inertiaConfig := config.DefaultInertiaConfig()
	i := config.InitInertia(inertiaConfig, sessionManager)

	// Initialize services
	// initializeServices()
	// NOTIFICATION Dependencies
	notificationService := notification.NewMailService()
	_ = notificationService

	// STORAGE Dependencies
	r2Storage, err := storage.NewR2Storage()
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}
	storageService := storage.NewStorageService(r2Storage)
	_ = storageService

	// MEDIA Dependencies
	imageProcessor := media.NewBimgProcessor()
	mediaService := media.NewMediaService(imageProcessor)
	_ = mediaService

	// Initialize middleware
	appMiddleware := web.NewSiteMiddleware(i, sessionManager)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	// Apply session middleware first
	r.Use(sessionManager.LoadAndSave)

	// Apply flash message sharing middleware
	r.Use(appMiddleware.FlashMessageSharingMiddleware)

	// Apply Inertia middleware
	r.Use(i.Middleware)

	// Public routes
	publicRouter := web.NewPublicRouter(web.NewPublicHandler(i, sessionManager))
	r.Mount("/", publicRouter)

	slog.Info("Server starting on :8084")
	if err := http.ListenAndServe(":8084", r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

// func initializeServices() {

// }
