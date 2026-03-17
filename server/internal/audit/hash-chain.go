package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ComputeHash generates a SHA-256 hash linking the entry to the previous hash.
func ComputeHash(e Entry, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		prevHash, e.UserID, e.Action, e.ResourceType, e.ResourceID, e.EventType, e.Status)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// VerifyChain verifies that a sequence of entries has a valid hash chain.
func VerifyChain(entries []Entry) (bool, int) {
	prevHash := ""
	for i, e := range entries {
		expected := ComputeHash(e, prevHash)
		if e.HashPrev != expected {
			return false, i
		}
		prevHash = expected
	}
	return true, -1
}
