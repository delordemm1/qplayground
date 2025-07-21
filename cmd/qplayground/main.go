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

	// Create directories for logs and files
	logsDir := filepath.Join("cmd", "qplayground", "logs", runID)
	filesDir := filepath.Join("cmd", "qplayground", "files", runID)

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	if err := os.MkdirAll(filesDir, 0755); err != nil {
		log.Fatalf("Failed to create files directory: %v", err)
	}

	// Initialize logger
	logger := NewLocalLogger(logsDir, runID)

	// Initialize storage service
	storageService := NewLocalStorageService(filesDir)

	// Initialize runner
	runner := NewRunner(storageService, logger)

	// Run automation
	ctx := context.Background()
	err = runner.RunAutomation(ctx, &exportedConfig, runID)

	if err != nil {
		logger.Error("❌ Automation failed", "error", err)
		fmt.Printf("❌ Automation failed: %v\n", err)
		os.Exit(1)
	} else {
		logger.Info("✅ Automation completed successfully")
		fmt.Printf("✅ Automation completed successfully!\n")
		fmt.Printf("📁 Logs saved to: %s\n", logsDir)
		fmt.Printf("📁 Files saved to: %s\n", filesDir)
	}
}