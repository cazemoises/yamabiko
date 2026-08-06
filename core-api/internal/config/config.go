package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	JWTSecret             string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	STTServiceURL         string
	CORSAllowedOrigins    []string
	CORSAllowLocalNetwork bool
	VoicevoxURL           string
	PiperAddress          string
	AudioCacheDir         string
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

	// PIPER_URL é um endereço TCP (host:port), não uma URL HTTP — o Piper fala
	// Wyoming (protocolo TCP puro, ver core-api/internal/tts/piper_client.go),
	// mantido com o nome "_URL" só pra bater com o padrão das outras env vars
	// de serviço externo (STT_SERVICE_URL, VOICEVOX_URL).
	piperAddress := os.Getenv("PIPER_URL")
	if piperAddress == "" {
		return nil, fmt.Errorf("PIPER_URL não configurado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	audioCacheDir := os.Getenv("AUDIO_CACHE_DIR")
	if audioCacheDir == "" {
		audioCacheDir = "audio-cache"
	}

	return &Config{
		Port:                  port,
		DatabaseURL:           dbURL,
		JWTSecret:             secret,
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		STTServiceURL:         sttServiceURL,
		CORSAllowedOrigins:    corsAllowedOrigins(),
		CORSAllowLocalNetwork: corsAllowLocalNetwork(),
		VoicevoxURL:           voicevoxURL,
		PiperAddress:          piperAddress,
		AudioCacheDir:         audioCacheDir,
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

// corsAllowLocalNetwork lê CORS_ALLOW_LOCAL_NETWORK — opt-in explícito (não
// hardcoda nenhum IP) pra aceitar, além de CORS_ALLOWED_ORIGINS, qualquer
// origin http:// cujo host seja um IP de rede privada (RFC1918,
// net.IP.IsPrivate) ou loopback — pedido do usuário pra testar o `web` a
// partir do celular na mesma Wi-Fi sem precisar fixar o IP da máquina (que
// muda a cada rede/DHCP). Default false: sem essa env var, o comportamento
// de CORS não muda em nada — é puramente aditivo e pensado só pra dev local,
// nunca deveria ser setado em produção.
func corsAllowLocalNetwork() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("CORS_ALLOW_LOCAL_NETWORK")))
	return v == "true" || v == "1"
}
