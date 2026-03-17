package crypto

import "fmt"

// StorageKey returns the MinIO object key for a secret blob.
func StorageKey(userID, secretID string) string {
	return fmt.Sprintf("secrets/%s/%s", userID, secretID)
}
