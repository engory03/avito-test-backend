package utils

import "testing"

func TestHashPassword_CheckPasswordHash(t *testing.T) {
	h, err := HashPassword("my-secret-password")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPasswordHash("my-secret-password", h) {
		t.Fatal("expected match")
	}
	if CheckPasswordHash("other", h) {
		t.Fatal("expected mismatch")
	}
}
