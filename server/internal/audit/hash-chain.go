package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// hashPayload is the canonical preimage of a v2 chain hash. Every persisted
// field of Entry participates, so tampering with any column (including
// ip_address, user_agent, metadata or event_time) breaks the chain.
type hashPayload struct {
	Version      int    `json:"v"`
	PrevHash     string `json:"prev"`
	EventTime    string `json:"t"`
	UserID       string `json:"uid"`
	Action       string `json:"act"`
	ResourceType string `json:"rtype"`
	ResourceID   string `json:"rid"`
	EventType    string `json:"etype"`
	Status       string `json:"status"`
	IPAddress    string `json:"ip"`
	UserAgent    string `json:"ua"`
	RegionCode   string `json:"region"`
	Metadata     string `json:"meta"`
}

// canonicalTime renders EventTime for hashing. PostgreSQL timestamptz keeps
// microsecond precision, so entries are truncated before hashing to keep the
// preimage identical before and after a DB round-trip.
func canonicalTime(t time.Time) string {
	return t.Truncate(time.Microsecond).UTC().Format(time.RFC3339Nano)
}

// ComputeHash generates the SHA-256 chain hash of an entry, binding it to the
// previous entry's hash and to every field of the entry itself.
func ComputeHash(e Entry, prevHash string) string {
	payload, err := json.Marshal(hashPayload{
		Version:      2,
		PrevHash:     prevHash,
		EventTime:    canonicalTime(e.EventTime),
		UserID:       e.UserID,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		EventType:    e.EventType,
		Status:       e.Status,
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
		RegionCode:   e.RegionCode,
		Metadata:     e.Metadata,
	})
	if err != nil {
		// Marshal of a plain string struct cannot fail.
		panic(fmt.Sprintf("audit: marshal hash payload: %v", err))
	}
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

// computeHashV1 reproduces the pre-000042 chain hash. It only covered
// user/action/resource/event_type/status; rows written before the v2 upgrade
// remain verifiable with it, but note that v1-era ip/metadata/timestamp are
// NOT covered for those rows.
func computeHashV1(e Entry, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		prevHash, e.UserID, e.Action, e.ResourceType, e.ResourceID, e.EventType, e.Status)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// VerifyChain verifies that a sequence of entries (ordered by seq) forms an
// unbroken hash chain. It returns (ok, firstBrokenIndex); ok is true when the
// whole sequence verifies (-1 index). Entries written under v1 verify with the
// v1 algorithm; any entry whose stored hash matches neither algorithm breaks
// the chain at its index.
func VerifyChain(entries []Entry) (bool, int) {
	prevHash := ""
	for i, e := range entries {
		expected := ComputeHash(e, prevHash)
		if e.HashPrev == expected {
			prevHash = expected
			continue
		}
		if e.HashPrev == computeHashV1(e, prevHash) {
			prevHash = e.HashPrev
			continue
		}
		return false, i
	}
	return true, -1
}
