package platform

import (
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

func isRunningTest() bool {
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") {
			return true
		}
	}
	return false
}

func mustHaveEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		if isRunningTest() {
			return "test"
		}
		// For CLI, we'll use defaults instead of panicking
		return ""
	}
	return value
}

func mustHaveEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

var (
	ENV_LOG_LEVEL = getEnvWithDefault("LOG_LEVEL", "info")
	ENV_APP_URL   = getEnvWithDefault("APP_URL", "http://localhost:8084")

	// SMTP Configuration (optional for CLI)
	ENV_SMTP_HOST     = mustHaveEnv("SMTP_SERVER")
	ENV_SMTP_PORT     = mustHaveEnvInt("SMTP_PORT", 587)
	ENV_SMTP_USERNAME = mustHaveEnv("SMTP_USERNAME")
	ENV_SMTP_PASSWORD = mustHaveEnv("SMTP_PASSWORD")
	ENV_SMTP_FROM     = getEnvWithDefault("SMTP_FROM_EMAIL", ENV_SMTP_USERNAME)

	// Cloudflare R2 Configuration
	ENV_CLOUDFLARE_ACCOUNT_ID = mustHaveEnv("CLOUDFLARE_ACCOUNT_ID")
	ENV_R2_ACCESS_KEY_ID      = mustHaveEnv("R2_ACCESS_KEY_ID")
	ENV_R2_SECRET_ACCESS_KEY  = mustHaveEnv("R2_SECRET_ACCESS_KEY")
	ENV_R2_BUCKET_NAME        = mustHaveEnv("R2_BUCKET_NAME")
	ENV_R2_PUBLIC_URL         = mustHaveEnv("R2_PUBLIC_URL")

	// GCP Storage Configuration
	ENV_GCP_BUCKET_SECRET       = mustHaveEnv("GCP_BUCKET_SECRET")
	ENV_GCP_BUCKET_NAME         = mustHaveEnv("GCP_BUCKET_NAME")
	ENV_GCP_BUCKET_ACCESS_KEY   = mustHaveEnv("GCP_BUCKET_ACCESS_KEY")
	ENV_GCP_BUCKET_ENDPOINT_URL = mustHaveEnv("GCP_BUCKET_ENDPOINT_URL")
	ENV_GCP_BUCKET_PUBLIC_URL   = mustHaveEnv("GCP_BUCKET_PUBLIC_URL")

	// Storage Configuration
	ENV_STORAGE_PROVIDER     = getEnvWithDefault("STORAGE_PROVIDER", "local")
	ENV_CLI_STORAGE_PROVIDER = getEnvWithDefault("CLI_STORAGE_PROVIDER", "local")

	// GitHub Actions specific
	ENV_GITHUB_RUN_ID     = mustHaveEnv("GITHUB_RUN_ID")
	ENV_GITHUB_RUN_NUMBER = mustHaveEnv("GITHUB_RUN_NUMBER")
	ENV_RUNNER_INDEX      = getEnvWithDefault("RUNNER_INDEX", "0")
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	// Set default SMTP_FROM if not provided
	if ENV_SMTP_FROM == "" {
		ENV_SMTP_FROM = ENV_SMTP_USERNAME
	}
}