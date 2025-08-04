package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/delordemm1/qplayground/internal/core/config"
	"github.com/delordemm1/qplayground/internal/modules/automation"
	"github.com/delordemm1/qplayground/internal/modules/notification"
	"github.com/delordemm1/qplayground/internal/modules/storage"
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"

	// Import plugin packages so their init() functions run and register actions
	_ "github.com/delordemm1/qplayground/internal/plugins/api"
	_ "github.com/delordemm1/qplayground/internal/plugins/playwright"
	_ "github.com/delordemm1/qplayground/internal/plugins/r2"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize logger
	platform.InitLogger()

	// Install Playwright
	if err := playwright.Install(); err != nil {
		log.Fatalf("Could not install playwright: %v", err)
	}

	// Define command-line flags
	var (
		automationID              = flag.String("automation-id", "", "ID of the automation to run (required)")
		runID                     = flag.String("run-id", "", "ID of the automation run (required)")
		isSubRun                  = flag.Bool("is-sub-run", false, "Whether this is an individual sub-run")
		subRunIndex               = flag.Int("sub-run-index", 0, "Index of this sub-run (required if --is-sub-run is true)")
		consolidate               = flag.Bool("consolidate", false, "Whether to consolidate sub-run results")
		totalExpectedRuns         = flag.Int("total-expected-runs", 0, "Total number of sub-runs expected (required if --consolidate is true)")
		outputDir                 = flag.String("output-dir", "", "Local directory for temporary files")
		overrideMaxConcurrentRuns = flag.Int("override-max-concurrent-runs", 0, "Override max concurrent runs")
		overrideRunMode           = flag.String("override-run-mode", "", "Override run mode (sequential/parallel)")
		overrideRunCount          = flag.Int("override-run-count", 0, "Override run count")
		overrideRunDelay          = flag.Int("override-run-delay", 0, "Override run delay in milliseconds")
	)
	flag.Parse()

	// Validate required flags
	if *automationID == "" {
		log.Fatal("--automation-id is required")
	}
	if *runID == "" {
		log.Fatal("--run-id is required")
	}
	if *isSubRun && *subRunIndex < 0 {
		log.Fatal("--sub-run-index must be >= 0 when --is-sub-run is true")
	}
	if *consolidate && *totalExpectedRuns <= 0 {
		log.Fatal("--total-expected-runs must be > 0 when --consolidate is true")
	}

	// Initialize database
	pool := config.InitDatabase()
	defer pool.Close()

	// Initialize Redis
	redisClient := config.InitRedis()
	defer redisClient.Close()

	// Initialize storage service
	var objectStorage storage.ObjectStorage
	switch platform.ENV_STORAGE_PROVIDER {
	case "gcp":
		objectStorage, err = storage.NewGcpStorage()
	case "r2":
		fallthrough
	default:
		objectStorage, err = storage.NewR2Storage()
	}
	if err != nil {
		log.Fatalf("Failed to initialize %s storage: %v", platform.ENV_STORAGE_PROVIDER, err)
	}
	storageService := storage.NewStorageService(objectStorage)

	// Initialize notification service
	notificationService := notification.NewMailService()

	// Initialize automation components
	automationRepo := automation.NewAutomationRepository(pool)
	runCache := automation.NewRedisRunCache(redisClient)
	automationService := automation.NewAutomationService(automationRepo, runCache, pool)

	// Create SSE manager (though it won't be used in CLI)
	sseManager := automation.NewSSEManager()
	defer sseManager.Shutdown()

	automationRunner := automation.NewRunner(automationRepo, storageService, notificationService, sseManager)

	ctx := context.Background()

	// Build overrides
	var overrides *automation.RunOverrides
	if *overrideMaxConcurrentRuns > 0 || *overrideRunMode != "" || *overrideRunCount > 0 || *overrideRunDelay > 0 {
		overrides = &automation.RunOverrides{}
		if *overrideMaxConcurrentRuns > 0 {
			overrides.MaxConcurrentRuns = overrideMaxConcurrentRuns
		}
		if *overrideRunMode != "" {
			overrides.RunMode = overrideRunMode
		}
		if *overrideRunCount > 0 {
			overrides.RunCount = overrideRunCount
		}
		if *overrideRunDelay > 0 {
			overrides.RunDelay = overrideRunDelay
		}
	}

	if *consolidate {
		// Consolidation mode
		slog.Info("Starting consolidation", "run_id", *runID, "total_expected_runs", *totalExpectedRuns)

		err := automationRunner.ConsolidateSubRuns(ctx, *runID)
		if err != nil {
			log.Fatalf("Consolidation failed: %v", err)
		}

		slog.Info("Consolidation completed successfully", "run_id", *runID)
		return
	}

	// Fetch automation and run
	automationObj, err := automationService.GetAutomationByID(ctx, *automationID)
	if err != nil {
		log.Fatalf("Failed to get automation: %v", err)
	}

	run, err := automationService.GetRunByID(ctx, *runID)
	if err != nil {
		log.Fatalf("Failed to get run: %v", err)
	}

	// Determine project ID from automation
	projectID := automationObj.ProjectID

	// Execute the automation
	if *isSubRun {
		slog.Info("Starting sub-run execution",
			"automation_id", *automationID,
			"run_id", *runID,
			"sub_run_index", *subRunIndex,
			"output_dir", *outputDir)

		_, _, err = automationRunner.RunAutomation(ctx, projectID, run, true, *subRunIndex, overrides)
	} else {
		slog.Info("Starting single run execution",
			"automation_id", *automationID,
			"run_id", *runID,
			"output_dir", *outputDir)

		_, _, err = automationRunner.RunAutomation(ctx, projectID, run, false, 0, overrides)
	}

	if err != nil {
		log.Fatalf("Automation execution failed: %v", err)
	}

	if *isSubRun {
		slog.Info("Sub-run completed successfully", "sub_run_index", *subRunIndex)
	} else {
		slog.Info("Single run completed successfully")
	}
}
