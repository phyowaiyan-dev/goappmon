package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type SessionClaims struct {
	AdminID   int64 `json:"admin_id"`
	ExpiresAt int64 `json:"expires_at"`
}

func SignSession(secret []byte, adminID int64, ttl time.Duration) (string, error) {
	claims := SessionClaims{
		AdminID:   adminID,
		ExpiresAt: time.Now().Add(ttl).UTC().Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func VerifySession(secret []byte, token string) (SessionClaims, error) {
	parts := splitToken(token)
	if len(parts) != 2 {
		return SessionClaims{}, errors.New("invalid session token")
	}

	expectedMAC := hmac.New(sha256.New, secret)
	_, _ = expectedMAC.Write([]byte(parts[0]))
	expectedSignature := expectedMAC.Sum(nil)

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return SessionClaims{}, errors.New("invalid session signature")
	}
	if !hmac.Equal(signature, expectedSignature) {
		return SessionClaims{}, errors.New("invalid session signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return SessionClaims{}, errors.New("invalid session payload")
	}

	var claims SessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return SessionClaims{}, err
	}
	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return SessionClaims{}, errors.New("session expired")
	}
	return claims, nil
}

func GenerateSecretKey(size int) ([]byte, error) {
	if size < 32 {
		size = 32
	}
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func splitToken(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

func FormatSessionKey(secret []byte) string {
	return fmt.Sprintf("%x", secret)
}
