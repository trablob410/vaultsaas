package crypto

import "testing"

func TestStorageKey(t *testing.T) {
	key := StorageKey("user123", "secret456")
	expected := "secrets/user123/secret456"
	if key != expected {
		t.Errorf("got %q, want %q", key, expected)
	}
}

func TestStorageKeyEmpty(t *testing.T) {
	key := StorageKey("", "")
	expected := "secrets//"
	if key != expected {
		t.Errorf("got %q, want %q", key, expected)
	}
}

func TestStorageKeyOnlyUser(t *testing.T) {
	key := StorageKey("user1", "")
	expected := "secrets/user1/"
	if key != expected {
		t.Errorf("got %q, want %q", key, expected)
	}
}

func TestStorageKeyOnlySecret(t *testing.T) {
	key := StorageKey("", "sec1")
	expected := "secrets//sec1"
	if key != expected {
		t.Errorf("got %q, want %q", key, expected)
	}
}
