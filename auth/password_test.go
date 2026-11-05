package auth

import "testing"

func TestHashPasswordDoesNotStoreThePlainPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct-horse-battery" {
		t.Fatal("password must be stored as a bcrypt hash")
	}
	if !CheckPassword(hash, "correct-horse-battery") {
		t.Fatal("correct password was rejected")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("incorrect password was accepted")
	}
}
