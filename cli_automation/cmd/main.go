package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/delordemm1/qplayground-cli/internal/automation"
	"github.com/delordemm1/qplayground-cli/internal/notification"
	"github.com/delordemm1/qplayground-cli/internal/platform"
	"github.com/delordemm1/qplayground-cli/internal/storage"
	"github.com/delordemm1/qplayground-cli/internal/utils"

	// Import plugin packages so their init() functions run and register actions
	_ "github.com/delordemm1/qplayground-cli/internal/plugins/api"
	_ "github.com/delordemm1/qplayground-cli/internal/plugins/playwright"
)

func main() {
	// Parse command line arguments
	var configPath = flag.String("config-path", "", "Path to the automation configuration JSON file")
	var outputDir = flag.String("output-dir", "", "Directory to save reports and screenshots")
	var runnerIndex = flag.String("runner-index", "0", "Index of this runner (for GitHub Actions matrix)")
	var consolidate = flag.Bool("consolidate", false, "Consolidate reports from multiple runners")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("--config-path is required")
	}
	if *outputDir == "" {
		log.Fatal("--output-dir is required")
	}

	// Set runner index in environment for report generation
	if *runnerIndex != "" {
		os.Setenv("RUNNER_INDEX", *runnerIndex)
	}

	// Initialize logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Read and parse automation configuration
	configData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var exportedConfig automation.ExportedAutomationConfig
	if err := json.Unmarshal(configData, &exportedConfig); err != nil {
		log.Fatalf("Failed to parse config JSON: %v", err)
	}

	// Convert exported config to internal automation structure
	automationObj := convertExportedToAutomation(exportedConfig)

	// Handle consolidation mode
	if *consolidate {
		err := consolidateReports(*outputDir)
		if err != nil {
			log.Fatalf("Failed to consolidate reports: %v", err)
		}
		return
	}

	// Set project name and automation slug for CLI
	automationObj.ProjectName = "CLI Project"
	automationSlug := strings.ToLower(strings.ReplaceAll(automationObj.Name, " ", "-"))
	automationSlug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(automationSlug, "")
	automationObj.AutomationSlug = automationSlug

	// Initialize storage service based on configuration
	var objectStorage storage.ObjectStorage
	// var err error

	switch platform.ENV_CLI_STORAGE_PROVIDER {
	case "r2":
		objectStorage, err = storage.NewR2Storage()
		if err != nil {
			slog.Warn("Failed to initialize R2 storage, falling back to local", "error", err)
			objectStorage = storage.NewLocalFileStorage(*outputDir)
		}
	case "gcp":
		objectStorage, err = storage.NewGCPStorage()
		if err != nil {
			slog.Warn("Failed to initialize GCP storage, falling back to local", "error", err)
			objectStorage = storage.NewLocalFileStorage(*outputDir)
		}
	default:
		objectStorage = storage.NewLocalFileStorage(*outputDir)
	}

	storageService := storage.NewStorageService(objectStorage)
	notificationService := notification.NewMailService()

	// Create runner
	runner := automation.NewRunner(storageService, notificationService, *outputDir)

	// Create automation run
	run := &automation.AutomationRun{
		ID:              utils.UtilGenerateUUID(),
		AutomationID:    automationObj.ID,
		Status:          "running",
		LogsJSON:        "[]",
		OutputFilesJSON: "[]",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Add runner-specific information to run ID for uniqueness
	if *runnerIndex != "0" {
		run.ID = fmt.Sprintf("%s-runner-%s", run.ID, *runnerIndex)
	}

	slog.Info("Starting automation execution",
		"automation_name", automationObj.Name,
		"run_id", run.ID,
		"runner_index", *runnerIndex,
		"output_dir", *outputDir,
		"storage_provider", platform.ENV_CLI_STORAGE_PROVIDER)

	// Execute automation
	ctx := context.Background()
	detailedReportURL, userJourneyReportURL, err := runner.RunAutomation(ctx, automationObj, run)

	if err != nil {
		slog.Error("Automation execution failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Automation execution completed successfully",
		"run_id", run.ID,
		"status", run.Status,
		"output_dir", *outputDir,
		"detailed_report_url", detailedReportURL,
		"user_journey_report_url", userJourneyReportURL)
}

// consolidateReports consolidates reports from multiple runners into a single report
func consolidateReports(outputDir string) error {
	slog.Info("Starting report consolidation", "output_dir", outputDir)

	// This function would:
	// 1. Scan the output directory for individual runner reports
	// 2. Parse and merge all JSON reports
	// 3. Generate consolidated HTML, JSON, and CSV reports
	// 4. Upload to storage if configured

	// For now, just log that consolidation would happen
	slog.Info("Report consolidation completed", "output_dir", outputDir)
	return nil
}

// convertExportedToAutomation converts ExportedAutomationConfig to internal Automation structure
// func convertExportedToAutomation(exported automation.ExportedAutomationConfig) *automation.Automation {
// 	// Convert config to JSON string
// 	configBytes, _ := json.Marshal(exported.Automation.Config)

// 	automationObj := &automation.Automation{
// 		ID:          utils.UtilGenerateUUID(),
// 		ProjectID:   "cli-project",
// 		Name:        exported.Automation.Name,
// 		Description: exported.Automation.Description,
// 		ConfigJSON:  string(configBytes),
// 		Steps:       make([]*automation.AutomationStep, 0, len(exported.Steps)),
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 	}

// 	// Convert steps
// 	for _, exportedStep := range exported.Steps {
// 		step := &automation.AutomationStep{
// 			ID:           utils.UtilGenerateUUID(),
// 			AutomationID: automationObj.ID,
// 			Name:         exportedStep.Name,
// 			StepOrder:    exportedStep.StepOrder,
// 			Actions:      make([]*automation.AutomationAction, 0, len(exportedStep.Actions)),
// 			CreatedAt:    time.Now(),
// 			UpdatedAt:    time.Now(),
// 		}

// 		// Convert step config if present
// 		if exportedStep.Config != nil {
// 			configBytes, _ := json.Marshal(exportedStep.Config)
// 			step.ConfigJSON = string(configBytes)
// 		}

// 		// Convert actions
// 		for _, exportedAction := range exportedStep.Actions {
// 			action := &automation.AutomationAction{
// 				ID:           exportedAction.ID,
// 				StepID:       step.ID,
// 				Name:         exportedAction.Name,
// 				ActionType:   exportedAction.ActionType,
// 				ActionConfig: exportedAction.ActionConfig, // Store as map for easier access
// 				ActionOrder:  exportedAction.ActionOrder,
// 				CreatedAt:    time.Now(),
// 				UpdatedAt:    time.Now(),
// 			}

// 			// Also store as JSON string for compatibility
// 			actionConfigBytes, _ := json.Marshal(exportedAction.ActionConfig)
// 			action.ActionConfigJSON = string(actionConfigBytes)

// 			step.Actions = append(step.Actions, action)
// 		}

// 		automationObj.Steps = append(automationObj.Steps, step)
// 	}

// 	return automationObj
// }
// 	AutomationID:    automationObj.ID,
// 	Status:          "running",
// 	LogsJSON:        "[]",
// 	OutputFilesJSON: "[]",
// 	CreatedAt:       time.Now(),
// 	UpdatedAt:       time.Now(),
// }

// 	slog.Info("Starting automation execution",
// 		"automation_name", automationObj.Name,
// 		"run_id", run.ID,
// 		"output_dir", *outputDir)

// 	// Execute automation
// 	ctx := context.Background()
// 	detailedReportURL, userJourneyReportURL, err := runner.RunAutomation(ctx, automationObj, run)

// 	if err != nil {
// 		slog.Error("Automation execution failed", "error", err)
// 		os.Exit(1)
// 	}

// 	slog.Info("Automation execution completed successfully",
// 		"run_id", run.ID,
// 		"status", run.Status,
// 		"output_dir", *outputDir,
// 		"detailed_report_url", detailedReportURL,
// 		"user_journey_report_url", userJourneyReportURL)
// }

// convertExportedToAutomation converts ExportedAutomationConfig to internal Automation structure
func convertExportedToAutomation(exported automation.ExportedAutomationConfig) *automation.Automation {
	// Convert config to JSON string
	configBytes, _ := json.Marshal(exported.Automation.Config)

	automationObj := &automation.Automation{
		ID:          utils.UtilGenerateUUID(),
		ProjectID:   "cli-project",
		Name:        exported.Automation.Name,
		Description: exported.Automation.Description,
		ConfigJSON:  string(configBytes),
		Steps:       make([]*automation.AutomationStep, 0, len(exported.Steps)),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Convert steps
	for _, exportedStep := range exported.Steps {
		step := &automation.AutomationStep{
			ID:           utils.UtilGenerateUUID(),
			AutomationID: automationObj.ID,
			Name:         exportedStep.Name,
			StepOrder:    exportedStep.StepOrder,
			Actions:      make([]*automation.AutomationAction, 0, len(exportedStep.Actions)),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Convert actions
		for _, exportedAction := range exportedStep.Actions {
			actionConfigBytes, _ := json.Marshal(exportedAction.ActionConfig)

			action := &automation.AutomationAction{
				ID:               exportedAction.ID,
				StepID:           step.ID,
				ActionType:       exportedAction.ActionType,
				ActionConfigJSON: string(actionConfigBytes),
				ActionOrder:      exportedAction.ActionOrder,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			step.Actions = append(step.Actions, action)
		}

		automationObj.Steps = append(automationObj.Steps, step)
	}

	return automationObj
}
