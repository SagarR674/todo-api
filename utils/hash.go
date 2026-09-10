package utils

import "golang.org/x/crypto/bcrypt"

// BcryptCost is the work factor used when hashing passwords. It defaults to the
// bcrypt library's recommended cost; tests may lower it for speed.
var BcryptCost = bcrypt.DefaultCost

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	return string(bytes), err
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
