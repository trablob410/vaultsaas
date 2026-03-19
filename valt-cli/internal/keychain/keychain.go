package keychain

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const service = "valt-cli"
const userKey = "auth-token"

// GetToken returns the stored auth token, checking env var first.
func GetToken() (string, error) {
	if v := os.Getenv("VALT_AUTH_TOKEN"); v != "" {
		return v, nil
	}
	token, err := keyring.Get(service, userKey)
	if err != nil {
		return "", nil // not found = not logged in
	}
	return token, nil
}

// SetToken stores the auth token in the OS keychain.
func SetToken(token string) error {
	if err := keyring.Set(service, userKey, token); err != nil {
		return fmt.Errorf("storing token in keychain: %w", err)
	}
	return nil
}

// DeleteToken removes the stored token.
func DeleteToken() error {
	return keyring.Delete(service, userKey)
}
