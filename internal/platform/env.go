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
		panic("Missing environment variable: " + key)
	}
	return value
}

func mustHaveEnvInt(key string) int {
	value := mustHaveEnv(key)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic("Invalid integer value for environment variable: " + key)
	}
	return intValue
}

var (
	ENV_LOG_LEVEL                  = mustHaveEnv("LOG_LEVEL")
	ENV_APP_URL                    = mustHaveEnv("APP_URL")
	ENV_DATABASE_URL               = mustHaveEnv("DATABASE_URL")
	ENV_GOOGLE_OAUTH_CLIENT_ID     = mustHaveEnv("GOOGLE_OAUTH_CLIENT_ID")
	ENV_GOOGLE_OAUTH_CLIENT_SECRET = mustHaveEnv("GOOGLE_OAUTH_CLIENT_SECRET")
	
	// SMTP Configuration
	ENV_SMTP_HOST     = mustHaveEnv("SMTP_SERVER")
	ENV_SMTP_PORT     = mustHaveEnvInt("SMTP_PORT")
	ENV_SMTP_USERNAME = mustHaveEnv("SMTP_USERNAME")
	ENV_SMTP_PASSWORD = mustHaveEnv("SMTP_PASSWORD")
	ENV_SMTP_FROM     = os.Getenv("SMTP_FROM_EMAIL")
	
	// Cloudflare R2 Configuration
	ENV_CLOUDFLARE_ACCOUNT_ID = mustHaveEnv("CLOUDFLARE_ACCOUNT_ID")
	ENV_R2_ACCESS_KEY_ID      = mustHaveEnv("R2_ACCESS_KEY_ID")
	ENV_R2_SECRET_ACCESS_KEY  = mustHaveEnv("R2_SECRET_ACCESS_KEY")
	ENV_R2_BUCKET_NAME        = mustHaveEnv("R2_BUCKET_NAME")
	ENV_R2_PUBLIC_URL         = mustHaveEnv("R2_PUBLIC_URL")
)

func init() {
	// Set default SMTP_FROM if not provided
	if ENV_SMTP_FROM == "" {
		ENV_SMTP_FROM = ENV_SMTP_USERNAME
	}
}