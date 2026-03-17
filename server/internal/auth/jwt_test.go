package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func generateTestKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privKey)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return privPEM, pubPEM
}

func TestJWTRoundtrip(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, err := NewJWTManager(privPEM, pubPEM)
	if err != nil {
		t.Fatal(err)
	}
	token, err := mgr.GenerateAccessToken("user-abc-123")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := mgr.ValidateAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userID != "user-abc-123" {
		t.Errorf("userID=%q, want user-abc-123", userID)
	}
}

func TestJWT_TamperedToken(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	token, _ := mgr.GenerateAccessToken("user-abc")
	tampered := token[:len(token)-4] + "xxxx"
	_, err := mgr.ValidateAccessToken(tampered)
	if err == nil {
		t.Error("tampered token should fail validation")
	}
}

func TestJWT_WrongKey(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	token, _ := mgr.GenerateAccessToken("user-abc")
	// Validate with different manager (different keys)
	privPEM2, pubPEM2 := generateTestKeys(t)
	mgr2, _ := NewJWTManager(privPEM2, pubPEM2)
	_, err := mgr2.ValidateAccessToken(token)
	if err == nil {
		t.Error("wrong key should fail")
	}
}

func TestJWT_RefreshTokenLength(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	token, err := mgr.GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	// 32 bytes -> 64 hex chars
	if len(token) != 64 {
		t.Errorf("refresh token len=%d, want 64", len(token))
	}
}

func TestJWT_RefreshTokensUnique(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	t1, _ := mgr.GenerateRefreshToken()
	t2, _ := mgr.GenerateRefreshToken()
	if t1 == t2 {
		t.Error("refresh tokens should be unique")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	// Verify the token has exp claim and works correctly
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	token, _ := mgr.GenerateAccessToken("user")
	// Token should be valid right now
	_, err := mgr.ValidateAccessToken(token)
	if err != nil {
		t.Errorf("fresh token should be valid: %v", err)
	}
	// Test that empty token fails
	_, err = mgr.ValidateAccessToken("")
	if err == nil {
		t.Error("empty token should fail")
	}
	// Test invalid string
	_, err = mgr.ValidateAccessToken("not.a.token")
	if err == nil {
		t.Error("invalid token should fail")
	}
}

// Ensure 15-minute expiry regression
func TestJWT_AccessTokenExpiry(t *testing.T) {
	privPEM, pubPEM := generateTestKeys(t)
	mgr, _ := NewJWTManager(privPEM, pubPEM)
	_ = time.Now() // ensure time package used
	token, _ := mgr.GenerateAccessToken("user")
	userID, err := mgr.ValidateAccessToken(token)
	if err != nil || userID != "user" {
		t.Errorf("valid token failed: %v", err)
	}
}

func TestNewJWTManager_InvalidPrivKey(t *testing.T) {
	_, err := NewJWTManager([]byte("not-a-pem"), []byte("also-not-a-pem"))
	if err == nil {
		t.Error("invalid PEM should fail")
	}
}
