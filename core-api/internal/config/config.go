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
	VoicevoxURL        string
	VoicevoxSpeakerID  string
	AudioCacheDir      string
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

	voicevoxURL := os.Getenv("VOICEVOX_URL")
	if voicevoxURL == "" {
		return nil, fmt.Errorf("VOICEVOX_URL não configurado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	voicevoxSpeakerID := os.Getenv("VOICEVOX_SPEAKER_ID")
	if voicevoxSpeakerID == "" {
		// speaker 30 = "No.7 - アナウンス" (estilo locutor/anúncio) — voz neutra e
		// adulta, otimizada pra leitura clara de texto, em vez do default puro do
		// VOICEVOX (speaker 1, ずんだもん/Zundamon, voz de mascote/personagem — não
		// serve como referência de pronúncia num app de aprendizado). Ver
		// BUILD_STATE.md pro raciocínio completo da escolha.
		voicevoxSpeakerID = "30"
	}

	audioCacheDir := os.Getenv("AUDIO_CACHE_DIR")
	if audioCacheDir == "" {
		audioCacheDir = "audio-cache"
	}

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		JWTSecret:          secret,
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		STTServiceURL:      sttServiceURL,
		CORSAllowedOrigins: corsAllowedOrigins(),
		VoicevoxURL:        voicevoxURL,
		VoicevoxSpeakerID:  voicevoxSpeakerID,
		AudioCacheDir:      audioCacheDir,
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
