package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for any token that is missing, malformed,
// tampered with, or expired.
var ErrInvalidToken = errors.New("invalid or expired token")

// JWTManager issues and verifies HS256 JWTs for authenticated users.
type JWTManager struct {
	secret []byte
	expiry time.Duration
}

// NewJWTManager builds a manager from the configured secret and token lifetime.
func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), expiry: expiry}
}

// Claims is the JWT payload. The subject is the user ID.
type Claims struct {
	UserID uint `json:"uid"`
	jwt.RegisteredClaims
}

// Generate returns a signed token for the given user ID.
func (m *JWTManager) Generate(userID uint) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies a token string and returns its claims.
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
