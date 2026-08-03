package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/config"
	"github.com/yamabiko/core-api/internal/db"
	"github.com/yamabiko/core-api/internal/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config inválida: %v", err)
	}

	if err := db.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("falha ao rodar migrations: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao conectar no postgres: %v", err)
	}
	defer pool.Close()

	repo := auth.NewPostgresRepository(pool)
	issuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	service := auth.NewService(repo, issuer)
	handler := auth.NewHandler(service)

	router := httpserver.NewRouter(handler, issuer)

	log.Printf("core-api ouvindo na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
