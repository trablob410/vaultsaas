package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordRoundtrip(t *testing.T) {
	hash, err := HashPassword("TestPassword123!")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash prefix wrong: %q", hash[:10])
	}
	match, err := VerifyPassword("TestPassword123!", hash)
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Error("correct password should match")
	}
}

func TestVerifyPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("CorrectHorse99!")
	if err != nil {
		t.Fatal(err)
	}
	match, err := VerifyPassword("WrongHorse99!", hash)
	if err != nil {
		t.Fatal(err)
	}
	if match {
		t.Error("wrong password should not match")
	}
}

func TestHashPassword_Unique(t *testing.T) {
	h1, err := HashPassword("SamePass99!")
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashPassword("SamePass99!")
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Error("two hashes of same password should differ (salt)")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	_, err := VerifyPassword("anypassword", "notavalidhash")
	if err == nil {
		t.Error("invalid hash format should return error")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatal(err)
	}
	match, err := VerifyPassword("", hash)
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Error("empty password roundtrip should match")
	}
}
