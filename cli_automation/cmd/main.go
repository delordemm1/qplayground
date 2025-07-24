package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/delordemm1/qplayground-cli/internal/automation"
	"github.com/delordemm1/qplayground-cli/internal/notification"
	"github.com/delordemm1/qplayground-cli/internal/storage"
	"github.com/delordemm1/qplayground-cli/internal/utils"

	// Import plugin packages so their init() functions run and register actions
	_ "github.com/delordemm1/qplayground-cli/internal/plugins/playwright"
)

func main() {
	// Parse command line arguments
	var configPath = flag.String("config-path", "", "Path to the automation configuration JSON file")
	var outputDir = flag.String("output-dir", "", "Directory to save reports and screenshots")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("--config-path is required")
	}
	if *outputDir == "" {
		log.Fatal("--output-dir is required")
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

	// Initialize services
	localStorage := storage.NewLocalFileStorage(*outputDir)
	storageService := storage.NewStorageService(localStorage)
	notificationService := notification.NewNoOpMailService()

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

	slog.Info("Starting automation execution",
		"automation_name", automationObj.Name,
		"run_id", run.ID,
		"output_dir", *outputDir)

	// Execute automation
	ctx := context.Background()
	err = runner.RunAutomation(ctx, automationObj, run)

	if err != nil {
		slog.Error("Automation execution failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Automation execution completed successfully",
		"run_id", run.ID,
		"status", run.Status,
		"output_dir", *outputDir)
}

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