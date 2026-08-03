package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	STTServiceURL   string
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
		Port:            port,
		DatabaseURL:     dbURL,
		JWTSecret:       secret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		STTServiceURL:   sttServiceURL,
	}, nil
}
