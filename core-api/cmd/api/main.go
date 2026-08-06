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

	voicevoxClient := tts.NewVoicevoxClient(cfg.VoicevoxURL)
	piperClient := tts.NewPiperClient(cfg.PiperAddress)
	ttsClients := map[string]tts.TTSClient{"ja": voicevoxClient, "en": piperClient}
	ttsService := tts.NewService(ttsClients, exercisesRepo, usersRepo, cfg.AudioCacheDir)
	ttsHandler := tts.NewHandler(ttsService)

	scenariosRepo := scenarios.NewPostgresRepository(pool)
	scenariosHandler := scenarios.NewHandler(scenariosRepo, exercisesRepo)

	router := httpserver.NewRouter(authHandler, issuer, exercisesHandler, attemptsHandler, usersHandler, dashboardHandler, ttsHandler, scenariosHandler, cfg.CORSAllowedOrigins, cfg.CORSAllowLocalNetwork)

	if cfg.TLSEnabled() {
		// HTTPS real (ex: certificado Tailscale, ver BUILD_STATE.md) —
		// necessário porque uma página HTTPS não consegue chamar uma API
		// http:// (mixed content, bloqueado pelo próprio browser antes de a
		// requisição sair, confirmado ao vivo). Mesmo cert/key do `web`
		// (Vite), só montados também no container do core-api.
		log.Printf("core-api ouvindo HTTPS na porta %s (TLS_CERT_FILE=%s)", cfg.Port, cfg.TLSCertFile)
		if err := http.ListenAndServeTLS(":"+cfg.Port, cfg.TLSCertFile, cfg.TLSKeyFile, router); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("core-api ouvindo na porta %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
