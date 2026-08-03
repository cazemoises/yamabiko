package auth_test

import (
	"testing"

	"github.com/yamabiko/core-api/internal/auth"
)

func TestHashPassword_CheckPassword_RoundTrip(t *testing.T) {
	hash, err := auth.HashPassword("senha123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "senha123" {
		t.Fatal("hash não deveria ser igual à senha em texto puro")
	}
	if !auth.CheckPassword(hash, "senha123") {
		t.Fatal("esperava CheckPassword true pra senha correta")
	}
	if auth.CheckPassword(hash, "senhaerrada") {
		t.Fatal("esperava CheckPassword false pra senha errada")
	}
}
