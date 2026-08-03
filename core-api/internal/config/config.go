package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	STTServiceURL      string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET não configurado")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL não configurado")
	}

	sttServiceURL := os.Getenv("STT_SERVICE_URL")
	if sttServiceURL == "" {
		return nil, fmt.Errorf("STT_SERVICE_URL não configurado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		JWTSecret:          secret,
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		STTServiceURL:      sttServiceURL,
		CORSAllowedOrigins: corsAllowedOrigins(),
	}, nil
}

// corsAllowedOrigins lê CORS_ALLOWED_ORIGINS (lista separada por vírgula) —
// em produção, setar essa env var pro domínio real do web. Sem ela, cai no
// default de desenvolvimento local (Vite dev server).
func corsAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:5173"}
	}

	origins := strings.Split(raw, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}
