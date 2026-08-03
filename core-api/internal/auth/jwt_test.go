package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yamabiko/core-api/internal/auth"
)

func TestJWTIssuer_IssueAndParseAccessToken(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret", time.Minute, time.Hour)
	userID := uuid.New()

	token, err := issuer.IssueAccessToken(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := issuer.Parse(token, auth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("esperava userID %v, veio %v", userID, claims.UserID)
	}
}

func TestJWTIssuer_ParseExpiredToken(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret", -time.Minute, time.Hour)
	token, err := issuer.IssueAccessToken(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := issuer.Parse(token, auth.TokenTypeAccess); err == nil {
		t.Fatal("esperava erro pra token expirado")
	}
}

func TestJWTIssuer_ParseWithWrongSecretFails(t *testing.T) {
	issuer := auth.NewJWTIssuer("secret-a", time.Minute, time.Hour)
	otherIssuer := auth.NewJWTIssuer("secret-b", time.Minute, time.Hour)

	token, err := issuer.IssueAccessToken(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := otherIssuer.Parse(token, auth.TokenTypeAccess); err == nil {
		t.Fatal("esperava erro pra token assinado com outro secret")
	}
}
