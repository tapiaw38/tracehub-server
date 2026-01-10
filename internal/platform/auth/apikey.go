package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// ApiKeyPrefix is the prefix for all API keys
	ApiKeyPrefix = "th_"
	// ApiKeyLength is the length of the random part of the API key
	ApiKeyLength = 32
)

// GenerateApiKey generates a new API key with the format: th_<random>
func GenerateApiKey() (string, error) {
	randomBytes := make([]byte, ApiKeyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 and remove padding
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	apiKey := ApiKeyPrefix + encoded

	return apiKey, nil
}

// HashApiKey hashes an API key for storage in the database
func HashApiKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// ValidateApiKeyFormat validates the format of an API key
func ValidateApiKeyFormat(apiKey string) error {
	if !strings.HasPrefix(apiKey, ApiKeyPrefix) {
		return fmt.Errorf("invalid API key format: must start with %s", ApiKeyPrefix)
	}

	if len(apiKey) < len(ApiKeyPrefix)+20 {
		return fmt.Errorf("invalid API key format: too short")
	}

	return nil
}

// ExtractApiKey extracts the API key from the Authorization header
// Supports both "Bearer <key>" and plain key formats
func ExtractApiKey(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	// Check if it starts with "Bearer "
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	// Otherwise, treat the whole header as the API key
	return authHeader, nil
}
