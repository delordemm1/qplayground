package platform

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func UtilGenerateUUID() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		slog.Error(
			"CRITICAL SYSTEM ERROR: Failed to generate a new UUID.",
			"error", err,
			"reason", "This indicates a severe issue with the system's random number generator or cryptographic capabilities, making basic entity creation impossible. The application must stop.",
		)
	}
	return uuid.String()
}

func UtilGenerateRandomState(length int) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to a base64 URL-safe string
	state := base64.URLEncoding.EncodeToString(randomBytes)

	return state, nil
}

func UtilGenerateRandomString(length int) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to a hex string
	hex := hex.EncodeToString(randomBytes)
	return hex, nil
}

// UtilStrPtr converts a string to a string pointer, returning nil if the string is empty
func UtilStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Perf logs performance timing
func Perf(msg string, start time.Time) {
	slog.Info(msg, "duration", time.Since(start))
}