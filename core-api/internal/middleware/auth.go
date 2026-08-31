package middleware

import (
	"github.com/yamabiko/core-api/internal/auth"
)

// RequireAuth e UserIDFromContext vivem em internal/auth (ver
// internal/auth/context.go). Reexportados aqui pra não exigir mudança nos
// chamadores existentes (users, tts, dashboard, attempts, httpserver).
var RequireAuth = auth.RequireAuth

var UserIDFromContext = auth.UserIDFromContext
