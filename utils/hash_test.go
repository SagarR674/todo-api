package utils

import "testing"

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("s3cret-passw0rd")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "s3cret-passw0rd" {
		t.Fatal("hash must not equal the plaintext")
	}
	if !CheckPassword(hash, "s3cret-passw0rd") {
		t.Error("CheckPassword should accept the correct password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword should reject an incorrect password")
	}
}

func TestHashPassword_SaltsAreUnique(t *testing.T) {
	a, _ := HashPassword("same-input")
	b, _ := HashPassword("same-input")
	if a == b {
		t.Fatal("two hashes of the same password should differ (random salt)")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	if CheckPassword("not-a-bcrypt-hash", "whatever") {
		t.Fatal("CheckPassword should return false for a malformed hash")
	}
}
