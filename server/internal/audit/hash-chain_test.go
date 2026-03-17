package audit

import "testing"

func makeEntry(userID, action, resourceType, resourceID, eventType, status string) Entry {
	return Entry{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventType:    eventType,
		Status:       status,
	}
}

func TestComputeHash(t *testing.T) {
	e := makeEntry("u1", "create", "secret", "s1", "action", "success")
	h1 := ComputeHash(e, "")
	h2 := ComputeHash(e, "")
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}
	// Different prevHash produces different hash
	h3 := ComputeHash(e, "prevhash")
	if h1 == h3 {
		t.Error("different prevHash should produce different hash")
	}
}

func TestComputeHash_DifferentFields(t *testing.T) {
	e1 := makeEntry("u1", "create", "secret", "s1", "action", "success")
	e2 := makeEntry("u2", "create", "secret", "s1", "action", "success")
	h1 := ComputeHash(e1, "")
	h2 := ComputeHash(e2, "")
	if h1 == h2 {
		t.Error("different userID should produce different hash")
	}
}

func TestComputeHash_NonEmpty(t *testing.T) {
	e := makeEntry("u1", "delete", "secret", "s1", "action", "success")
	h := ComputeHash(e, "")
	if h == "" {
		t.Error("hash should not be empty")
	}
	if len(h) != 64 {
		t.Errorf("SHA-256 hex should be 64 chars, got %d", len(h))
	}
}

func TestVerifyChain(t *testing.T) {
	entries := make([]Entry, 3)
	prevHash := ""
	for i := range entries {
		entries[i] = makeEntry("u1", "create", "secret", "s1", "action", "success")
		hash := ComputeHash(entries[i], prevHash)
		entries[i].HashPrev = hash
		prevHash = hash
	}

	ok, failIdx := VerifyChain(entries)
	if !ok {
		t.Errorf("valid chain should verify, failed at %d", failIdx)
	}
	if failIdx != -1 {
		t.Errorf("failIdx should be -1 for valid chain, got %d", failIdx)
	}
}

func TestVerifyChain_TamperDetected(t *testing.T) {
	entries := make([]Entry, 3)
	prevHash := ""
	for i := range entries {
		entries[i] = makeEntry("u1", "create", "secret", "s1", "action", "success")
		hash := ComputeHash(entries[i], prevHash)
		entries[i].HashPrev = hash
		prevHash = hash
	}
	// Tamper middle entry
	entries[1].Action = "tampered_action"
	ok, failIdx := VerifyChain(entries)
	if ok {
		t.Error("tampered chain should not verify")
	}
	if failIdx != 1 {
		t.Errorf("failIdx=%d, want 1", failIdx)
	}
}

func TestVerifyChain_Empty(t *testing.T) {
	ok, failIdx := VerifyChain([]Entry{})
	if !ok {
		t.Error("empty chain should be valid")
	}
	if failIdx != -1 {
		t.Errorf("failIdx=%d, want -1", failIdx)
	}
}

func TestVerifyChain_SingleEntry(t *testing.T) {
	e := makeEntry("u1", "create", "secret", "s1", "action", "success")
	hash := ComputeHash(e, "")
	e.HashPrev = hash
	ok, failIdx := VerifyChain([]Entry{e})
	if !ok {
		t.Errorf("single valid entry should verify, failIdx=%d", failIdx)
	}
}
