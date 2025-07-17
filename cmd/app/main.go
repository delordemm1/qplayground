package main

import (
	"encoding/gob"
	"log"
	"log/slog"
	"net/http"

	"github.com/delordemm1/qplayground/internal/controller/web"
	"github.com/delordemm1/qplayground/internal/core/config"
	"github.com/delordemm1/qplayground/internal/modules/automation"
	"github.com/delordemm1/qplayground/internal/modules/media"
	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/organization"
	"github.com/delordemm1/qplayground/internal/modules/project"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/delordemm1/qplayground/internal/modules/auth"

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

	// ORGANIZATION Dependencies
	organizationRepo := organization.NewOrganizationRepository(pool)
	organizationService := organization.NewOrganizationService(organizationRepo)

	// PROJECT Dependencies
	projectRepo := project.NewProjectRepository(pool)
	projectService := project.NewProjectService(projectRepo)

	// AUTOMATION Dependencies
	automationRepo := automation.NewAutomationRepository(pool)
	automationService := automation.NewAutomationService(automationRepo)

	// AUTH Dependencies (updated to include organization service)
	authRepo := auth.NewAuthRepository(pool)
	authService := auth.NewAuthService(authRepo, notificationService, sessionManager, organizationService)

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

	// Public routes (guest only)
	publicRouter := web.NewPublicRouter(web.NewPublicHandler(i, sessionManager))
	r.Mount("/", publicRouter)

	// Authentication routes
	authHandler := web.NewAuthHandler(i, sessionManager, authService)
	authRouter := web.NewAuthRouter(authHandler)
	r.Mount("/auth", authRouter)

	// Protected routes (authenticated users only)
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.OnlyUser)

		// Dashboard
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			err := i.Render(w, r, "dashboard", inertia.Props{
				"user": getUserFromContext(r.Context()),
			})
			if err != nil {
				platform.UtilHandleServerErr(w, err)
				return
			}
		})

		// Organization routes
		organizationHandler := web.NewOrganizationHandler(i, sessionManager, organizationService)
		organizationRouter := web.NewOrganizationRouter(organizationHandler)
		r.Mount("/organizations", organizationRouter)

		// Project routes
		projectHandler := web.NewProjectHandler(i, sessionManager, projectService)
		projectRouter := web.NewProjectRouter(projectHandler)
		r.Mount("/projects", projectRouter)

		// Automation routes (nested under projects)
		r.Route("/projects/{projectId}/automations", func(r chi.Router) {
			automationHandler := web.NewAutomationHandler(i, sessionManager, automationService, projectService)
			automationRouter := web.NewAutomationRouter(automationHandler)
			r.Mount("/", automationRouter)
		})
	})

	slog.Info("Server starting on :8084")
	if err := http.ListenAndServe(":8084", r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

// Helper function to get user from context (simplified for now)
func getUserFromContext(ctx context.Context) *auth.User {
	// This is a simplified version - in production you'd fetch the full user from database
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil
	}
	
	// Return a minimal user object for now
	// In production, you'd fetch this from the database
	return &auth.User{
		ID: userID,
		// Add other fields as needed
	}
}

// func initializeServices() {

// }
