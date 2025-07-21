package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// UtilGenerateUUID generates a new UUID
func UtilGenerateUUID() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		panic("Failed to generate UUID")
	}
	return uuid.String()
}

// UtilGenerateRandomString generates a random hex string
func UtilGenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// UtilTimestamp returns current timestamp
func UtilTimestamp() time.Time {
	return time.Now()
}