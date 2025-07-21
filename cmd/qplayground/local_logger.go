package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LocalLogger implements logging to both terminal and file
type LocalLogger struct {
	runID   string
	logFile *os.File
	logger  *slog.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewLocalLogger creates a new local logger
func NewLocalLogger(logsDir, runID string) *LocalLogger {
	logFilePath := filepath.Join(logsDir, "automation.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to create log file: %v", err))
	}

	// Create a simple text logger for terminal output
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &LocalLogger{
		runID:   runID,
		logFile: logFile,
		logger:  logger,
	}
}

// Info logs an info message
func (l *LocalLogger) Info(msg string, args ...interface{}) {
	l.log("INFO", msg, args...)
}

// Error logs an error message
func (l *LocalLogger) Error(msg string, args ...interface{}) {
	l.log("ERROR", msg, args...)
}

// Warn logs a warning message
func (l *LocalLogger) Warn(msg string, args ...interface{}) {
	l.log("WARN", msg, args...)
}

// Debug logs a debug message
func (l *LocalLogger) Debug(msg string, args ...interface{}) {
	l.log("DEBUG", msg, args...)
}

// With returns a logger with additional context
func (l *LocalLogger) With(args ...interface{}) *LocalLogger {
	return l // Simplified implementation
}

// log handles the actual logging
func (l *LocalLogger) log(level, msg string, args ...interface{}) {
	// Log to terminal
	switch level {
	case "INFO":
		l.logger.Info(msg, args...)
	case "ERROR":
		l.logger.Error(msg, args...)
	case "WARN":
		l.logger.Warn(msg, args...)
	case "DEBUG":
		l.logger.Debug(msg, args...)
	}

	// Convert args to map
	data := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				data[key] = args[i+1]
			}
		}
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Data:      data,
	}

	// Write to file as JSON
	if l.logFile != nil {
		jsonData, _ := json.Marshal(entry)
		l.logFile.WriteString(string(jsonData) + "\n")
		l.logFile.Sync()
	}
}

// Close closes the log file
func (l *LocalLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}