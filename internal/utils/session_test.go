package utils

import (
	"testing"
	"time"
)

func TestSessionHelpers(t *testing.T) {
	key, err := GenerateSecretKey(16)
	if err != nil {
		t.Fatalf("generate secret key: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected minimum 32-byte key, got %d", len(key))
	}

	if got := FormatSessionKey([]byte{0x01, 0x02, 0x0a}); got != "01020a" {
		t.Fatalf("unexpected formatted key: %s", got)
	}

	secret := []byte("01234567890123456789012345678901")
	token, err := SignSession(secret, 42, time.Minute)
	if err != nil {
		t.Fatalf("sign session: %v", err)
	}

	claims, err := VerifySession(secret, token)
	if err != nil {
		t.Fatalf("verify session: %v", err)
	}
	if claims.AdminID != 42 {
		t.Fatalf("unexpected admin id: %d", claims.AdminID)
	}

	if _, err := VerifySession(secret, token+"tamper"); err == nil {
		t.Fatal("expected tampered token to fail")
	}

	expired, err := SignSession(secret, 7, -time.Minute)
	if err != nil {
		t.Fatalf("sign expired session: %v", err)
	}
	if _, err := VerifySession(secret, expired); err == nil {
		t.Fatal("expected expired session to fail")
	}
}
