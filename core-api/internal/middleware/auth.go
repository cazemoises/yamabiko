package middleware

import (
	"github.com/yamabiko/core-api/internal/auth"
)

// RequireAuth e UserIDFromContext agora vivem em internal/auth (ver
// internal/auth/context.go — necessário pra POST /auth/pin-setup usar o
// próprio middleware sem criar import cycle, já que auth não pode
// importar middleware de volta). Reexportados aqui pra não exigir mudança
// nos chamadores existentes (users, tts, dashboard, attempts, httpserver).
var RequireAuth = auth.RequireAuth

var UserIDFromContext = auth.UserIDFromContext
