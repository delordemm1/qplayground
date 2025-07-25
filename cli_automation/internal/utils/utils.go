package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/google/uuid"
)

func UtilGenerateUUID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to random UUID if V7 fails
		id = uuid.New()
	}
	return id.String()
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
