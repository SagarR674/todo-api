package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-at-least-16-chars-long"

func TestJWTManager_GenerateAndParse(t *testing.T) {
	m := NewJWTManager(testSecret, time.Hour)

	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty token")
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
}

func TestJWTManager_Parse_Rejects(t *testing.T) {
	m := NewJWTManager(testSecret, time.Hour)
	valid, _ := m.Generate(1)

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"garbage", "not-a-jwt"},
		{"two segments", "aaa.bbb"},
		{"tampered payload", valid[:len(valid)-4] + "AAAA"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := m.Parse(tc.token); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

func TestJWTManager_Parse_ExpiredToken(t *testing.T) {
	m := NewJWTManager(testSecret, -time.Minute) // already expired

	token, err := m.Generate(7)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := m.Parse(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestJWTManager_Parse_WrongSecret(t *testing.T) {
	issuer := NewJWTManager(testSecret, time.Hour)
	verifier := NewJWTManager("a-completely-different-secret-value", time.Hour)

	token, _ := issuer.Generate(1)
	if _, err := verifier.Parse(token); err == nil {
		t.Fatal("expected token signed with another secret to be rejected")
	}
}

func TestJWTManager_Parse_RejectsNoneAlg(t *testing.T) {
	m := NewJWTManager(testSecret, time.Hour)

	// Craft a token with alg "none".
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: 1})
	unsigned, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	if _, err := m.Parse(unsigned); err == nil {
		t.Fatal("expected alg=none token to be rejected")
	}
}
