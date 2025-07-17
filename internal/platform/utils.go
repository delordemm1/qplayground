package platform

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// pghelpers
func UtilTimeToPGTimestamp(t time.Time) pgtype.Timestamp {
	var ts pgtype.Timestamp
	_ = ts.Scan(t)
	return ts
}
func UtilStrPtrToPGText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func UtilGetIDUUID(id string) string {
	if id == "" {
		return UtilGenerateUUID()
	}
	return id
}

// UtilStrPtr converts a string to a string pointer, returning nil if the string is empty
func UtilStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// http

func UtilHandleServerErr(w http.ResponseWriter, err error) {
	log.Printf("http error: %s\n", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("server error"))
}
