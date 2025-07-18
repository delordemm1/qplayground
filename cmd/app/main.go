package main

import (
	"context"
	"encoding/gob"
	"log"
	"log/slog"
	"net/http"

	"github.com/delordemm1/qplayground/internal/controller/web"
	"github.com/delordemm1/qplayground/internal/core/config"
	"github.com/delordemm1/qplayground/internal/modules/auth"
	"github.com/delordemm1/qplayground/internal/modules/automation"
	"github.com/delordemm1/qplayground/internal/modules/media"
	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/organization"
	"github.com/delordemm1/qplayground/internal/modules/project"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	// Import plugin packages so their init() functions run and register actions
	_ "github.com/delordemm1/qplayground/internal/plugins/playwright"
	_ "github.com/delordemm1/qplayground/internal/plugins/r2"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	inertia "github.com/romsar/gonertia/v2"
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

	// Initialize Redis
	redisClient := config.InitRedis()
	defer redisClient.Close()

	// Initialize SSE Manager
	sseManager := automation.NewSSEManager()
	defer sseManager.Shutdown()

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

	// STORAGE Dependencies
	r2Storage, err := storage.NewR2Storage()
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}
	storageService := storage.NewStorageService(r2Storage)

	// MEDIA Dependencies
	imageProcessor := media.NewBimgProcessor()
	mediaService := media.NewMediaService(imageProcessor)
	_ = mediaService

	// ORGANIZATION Dependencies
	organizationRepo := organization.NewOrganizationRepository(pool)
	organizationService := organization.NewOrganizationService(organizationRepo)

	// PROJECT Dependencies
	projectRepo := project.NewProjectRepository(pool)
	projectService := project.NewProjectService(projectRepo)

	// AUTOMATION Dependencies
	automationRepo := automation.NewAutomationRepository(pool)
	runCache := automation.NewRedisRunCache(redisClient)
	automationService := automation.NewAutomationService(automationRepo, runCache)
	automationRunner := automation.NewRunner(automationRepo, storageService, notificationService, sseManager)

	// Initialize automation scheduler
	scheduler := automation.NewScheduler(automationRepo, automationService, runCache, automationRunner, sseManager)

	// AUTH Dependencies (updated to include organization service)
	authRepo := auth.NewAuthRepository(pool)
	authService := auth.NewAuthService(authRepo, notificationService, sessionManager, organizationService)

	// Initialize middleware
	siteMiddleware := web.NewSiteMiddleware(i, sessionManager)
	authMiddleware := web.NewAuthMiddleware(i, sessionManager, authService)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	// Apply session middleware first
	r.Use(sessionManager.LoadAndSave)

	// Apply flash message sharing middleware
	r.Use(siteMiddleware.FlashMessageSharingMiddleware)

	// Apply Inertia middleware
	r.Use(i.Middleware)

	// Sync Redis with database on startup
	syncRedisRunsOnStartup(pool, runCache)

	// Start automation scheduler
	scheduler.Start(context.Background())
	defer scheduler.Stop()

	// Public routes (guest only)
	publicRouter := web.NewPublicRouter(web.NewPublicHandler(i, sessionManager))
	r.Mount("/", publicRouter)

	// Authentication routes
	authHandler := web.NewAuthHandler(i, sessionManager, authService)
	authRouter := web.NewAuthRouter(authHandler)
	r.Mount("/auth", authRouter)

	// Protected routes (authenticated users only)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.OnlyUser)

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
		projectHandler := web.NewProjectHandler(i, sessionManager, projectService, automationService)
		projectRouter := web.NewProjectRouter(projectHandler)
		r.Mount("/projects", projectRouter)

		// Automation routes (nested under projects)
		r.Route("/projects/{projectId}/automations", func(r chi.Router) {
			automationHandler := web.NewAutomationHandler(i, sessionManager, automationService, projectService, scheduler, sseManager)
			automationRouter := web.NewAutomationRouter(automationHandler)
			r.Mount("/", automationRouter)
			// Nested routes for steps and actions
			r.Route("/{id}/steps/{stepId}/actions", func(r chi.Router) {
				r.Post("/", automationHandler.CreateAction)
				r.Put("/{actionId}", automationHandler.UpdateAction)
				r.Delete("/{actionId}", automationHandler.DeleteAction)
			})
		})

		// Mount SSE server for automation events
		r.Mount("/events/", sseManager.GetServer())

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

// syncRedisRunsOnStartup syncs all runs from database to Redis on application startup
func syncRedisRunsOnStartup(pool *pgxpool.Pool, runCache automation.RunCache) {
	ctx := context.Background()

	// Query all runs from database
	query := `
		SELECT id, automation_id, status, start_time, end_time, logs_json, output_files_json, error_message, created_at, updated_at
		FROM automation_runs
		WHERE status IN ('pending', 'running', 'queued', 'completed', 'failed', 'cancelled')
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		slog.Error("Failed to query runs for Redis sync", "error", err)
		return
	}
	defer rows.Close()

	var runs []*automation.AutomationRun
	for rows.Next() {
		var run automation.AutomationRun
		var startTime, endTime, createdAt, updatedAt pgtype.Timestamp
		var logsJSON, outputFilesJSON, errorMessage pgtype.Text

		err := rows.Scan(
			&run.ID, &run.AutomationID, &run.Status, &startTime, &endTime,
			&logsJSON, &outputFilesJSON, &errorMessage, &createdAt, &updatedAt,
		)
		if err != nil {
			slog.Error("Failed to scan run for Redis sync", "error", err)
			continue
		}

		// Convert pgtype timestamps
		if startTime.Valid {
			run.StartTime = &startTime.Time
		}
		if endTime.Valid {
			run.EndTime = &endTime.Time
		}
		if logsJSON.Valid {
			run.LogsJSON = logsJSON.String
		}
		if outputFilesJSON.Valid {
			run.OutputFilesJSON = outputFilesJSON.String
		}
		if errorMessage.Valid {
			run.ErrorMessage = errorMessage.String
		}
		run.CreatedAt = createdAt.Time
		run.UpdatedAt = updatedAt.Time

		runs = append(runs, &run)
	}

	// Upsert all runs to Redis
	err = runCache.UpsertAllRuns(ctx, runs)
	if err != nil {
		slog.Error("Failed to upsert runs to Redis", "error", err)
		return
	}

	slog.Info("Successfully synced runs to Redis on startup", "count", len(runs))
}
