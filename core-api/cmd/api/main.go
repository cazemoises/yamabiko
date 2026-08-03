package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamabiko/core-api/internal/attempts"
	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/config"
	"github.com/yamabiko/core-api/internal/db"
	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/httpserver"
	"github.com/yamabiko/core-api/internal/sttclient"
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

	authRepo := auth.NewPostgresRepository(pool)
	issuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authService := auth.NewService(authRepo, issuer)
	authHandler := auth.NewHandler(authService)

	exercisesRepo := exercises.NewPostgresRepository(pool)
	exercisesHandler := exercises.NewHandler(exercisesRepo)

	sttClient := sttclient.New(cfg.STTServiceURL)
	attemptsRepo := attempts.NewPostgresRepository(pool)
	attemptsService := attempts.NewService(attemptsRepo, sttClient, exercisesRepo)
	attemptsHandler := attempts.NewHandler(attemptsService)

	router := httpserver.NewRouter(authHandler, issuer, exercisesHandler, attemptsHandler)

	log.Printf("core-api ouvindo na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
