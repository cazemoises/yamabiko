package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yamabiko/core-api/internal/attempts"
	"github.com/yamabiko/core-api/internal/auth"
	"github.com/yamabiko/core-api/internal/config"
	"github.com/yamabiko/core-api/internal/dashboard"
	"github.com/yamabiko/core-api/internal/db"
	"github.com/yamabiko/core-api/internal/exercises"
	"github.com/yamabiko/core-api/internal/httpserver"
	"github.com/yamabiko/core-api/internal/phonetics"
	"github.com/yamabiko/core-api/internal/scenarios"
	"github.com/yamabiko/core-api/internal/srs"
	"github.com/yamabiko/core-api/internal/sttclient"
	"github.com/yamabiko/core-api/internal/tts"
	"github.com/yamabiko/core-api/internal/users"
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

	srsRepo := srs.NewPostgresRepository(pool)
	phoneticsRepo := phonetics.NewPostgresRepository(pool)
	usersRepo := users.NewPostgresRepository(pool)
	usersHandler := users.NewHandler(usersRepo, srsRepo, phoneticsRepo, exercisesRepo)
	dashboardHandler := dashboard.NewHandler(phoneticsRepo)

	sttClient := sttclient.New(cfg.STTServiceURL)
	attemptsRepo := attempts.NewPostgresRepository(pool)
	attemptsService := attempts.NewService(attemptsRepo, sttClient, exercisesRepo, srsRepo, usersRepo, phoneticsRepo)
	attemptsHandler := attempts.NewHandler(attemptsService)

	voicevoxClient := tts.NewVoicevoxClient(cfg.VoicevoxURL, cfg.VoicevoxSpeakerID)
	piperClient := tts.NewPiperClient(cfg.PiperAddress, cfg.PiperVoice)
	ttsClients := map[string]tts.TTSClient{"ja": voicevoxClient, "en": piperClient}
	ttsService := tts.NewService(ttsClients, exercisesRepo, cfg.AudioCacheDir)
	ttsHandler := tts.NewHandler(ttsService)

	scenariosRepo := scenarios.NewPostgresRepository(pool)
	scenariosHandler := scenarios.NewHandler(scenariosRepo, exercisesRepo)

	router := httpserver.NewRouter(authHandler, issuer, exercisesHandler, attemptsHandler, usersHandler, dashboardHandler, ttsHandler, scenariosHandler, cfg.CORSAllowedOrigins)

	log.Printf("core-api ouvindo na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
