package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

var pairingToken string

func GenerateAuthToken(tokenLength int) (string, error) {
	tokenBytes := make([]byte, tokenLength)

	_, err := rand.Read(tokenBytes)

	if err != nil {
		log.Printf("Error generating random token bytes: %v", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	token := hex.EncodeToString(tokenBytes) // Converting raw bytes to a string

	pairingToken = token

	return token, nil
}

func VerifyAuthToken(token string) bool {
	if pairingToken == "" || token == "" {
		return false
	}
	return token == pairingToken
}

// Insecure? probably, fast? definitely
func IsPaired() bool {
	return pairingToken != ""
}
