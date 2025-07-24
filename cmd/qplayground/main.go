package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the automation config JSON file")
	flag.Parse()

	if configPath == "" {
		log.Fatal("Please provide a config file path using -config flag")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	// Read and parse config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var exportedConfig ExportedAutomationConfig
	if err := json.Unmarshal(configData, &exportedConfig); err != nil {
		log.Fatalf("Failed to parse config JSON: %v", err)
	}

	fmt.Printf("🚀 Starting automation: %s\n", exportedConfig.Automation.Name)
	if exportedConfig.Automation.Description != "" {
		fmt.Printf("📝 Description: %s\n", exportedConfig.Automation.Description)
	}

	// Create run timestamp
	runTimestamp := time.Now().Unix()
	runID := fmt.Sprintf("run-%d", runTimestamp)
	timestampStr := time.Now().Format("20060102-150405")

	// Create directories for results
	resultsDir := filepath.Join("cmd", "qplayground", "results", timestampStr)
	logsDir := filepath.Join(resultsDir, "logs")
	filesDir := filepath.Join(resultsDir, "files")
	reportsDir := filepath.Join(resultsDir, "reports")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	if err := os.MkdirAll(filesDir, 0755); err != nil {
		log.Fatalf("Failed to create files directory: %v", err)
	}

	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		log.Fatalf("Failed to create reports directory: %v", err)
	}

	// Initialize logger
	logger := NewLocalLogger(logsDir, runID)
	defer logger.Close()

	// Initialize storage service
	storageService := NewLocalStorageService(filesDir)

	// Initialize runner
	runner := NewRunner(storageService, logger, reportsDir)

	// Run automation
	ctx := context.Background()
	logs, outputFiles, err := runner.RunAutomation(ctx, &exportedConfig, runID)

	if err != nil {
		logger.Error("❌ Automation failed", "error", err)
		fmt.Printf("❌ Automation failed: %v\n", err)
		
		// Generate reports even on failure
		if err := generateReports(reportsDir, &exportedConfig, logs, outputFiles, timestampStr, err); err != nil {
			fmt.Printf("⚠️ Failed to generate reports: %v\n", err)
		}
		
		os.Exit(1)
	} else {
		logger.Info("✅ Automation completed successfully")
		fmt.Printf("✅ Automation completed successfully!\n")
		
		// Generate reports on success
		if err := generateReports(reportsDir, &exportedConfig, logs, outputFiles, timestampStr, nil); err != nil {
			fmt.Printf("⚠️ Failed to generate reports: %v\n", err)
		} else {
			fmt.Printf("📊 Reports generated successfully!\n")
		}
		
		fmt.Printf("📁 Results saved to: %s\n", resultsDir)
		fmt.Printf("📁 Logs: %s\n", logsDir)
		fmt.Printf("📁 Files: %s\n", filesDir)
		fmt.Printf("📁 Reports: %s\n", reportsDir)
	}
}

func generateReports(reportsDir string, config *ExportedAutomationConfig, logs []map[string]interface{}, outputFiles []string, timestamp string, executionError error) error {
	reporter := NewReporter(reportsDir, config, logs, outputFiles, timestamp, executionError)
	
	// Generate CSV report
	if err := reporter.GenerateCSVReport(); err != nil {
		return fmt.Errorf("failed to generate CSV report: %w", err)
	}
	
	// Generate JSON report
	if err := reporter.GenerateJSONReport(); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}
	
	// Generate HTML report
	if err := reporter.GenerateHTMLReport(); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}
	
	return nil
}