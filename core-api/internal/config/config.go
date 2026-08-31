package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	STTServiceURL         string
	CORSAllowedOrigins    []string
	CORSAllowLocalNetwork bool
	VoicevoxURL           string
	PiperAddress          string
	AudioCacheDir         string
	// TLSCertFile/TLSKeyFile (env TLS_CERT_FILE/TLS_KEY_FILE) — opcional,
	// sobe com HTTPS em vez de HTTP puro quando os 2 estão setados (ver
	// cmd/api/main.go). Necessário pra acessar o `web` via HTTPS real (ex:
	// Tailscale, ver BUILD_STATE.md) sem "mixed content": um fetch() de uma
	// página HTTPS pra uma API http:// é bloqueado pelo próprio browser
	// (confirmado ao vivo — Chrome recusa com "Mixed Content... This
	// request has been blocked"), então servir só o `web` via HTTPS não
	// basta, o core-api também precisa falar TLS no mesmo hostname.
	TLSCertFile string
	TLSKeyFile  string
	// DevFakeRemoteEmail/DevFakeRemoteName (env DEV_FAKE_REMOTE_EMAIL/
	// DEV_FAKE_REMOTE_NAME) — só pra dev local sem Pangolin na frente (ver
	// httpserver.NewRouter): quando setado, injeta esses valores como
	// Remote-Email/Remote-Name em toda requisição antes de RequireAuth,
	// simulando o que o Pangolin faria em produção. NUNCA setar em
	// produção — lá o header real só existe se a requisição realmente
	// atravessou o Pangolin.
	DevFakeRemoteEmail string
	DevFakeRemoteName  string
}

// TLSEnabled é true só quando os 2 arquivos estão configurados — usar só um
// dos dois é erro de configuração (ver main.go, que recusa a subir nesse
// caso em vez de silenciosamente cair pra HTTP).
func (c *Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

func Load() (*Config, error) {
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

	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")
	if (tlsCertFile == "") != (tlsKeyFile == "") {
		return nil, fmt.Errorf("TLS_CERT_FILE e TLS_KEY_FILE precisam estar os 2 setados ou os 2 vazios (veio cert=%q, key=%q)", tlsCertFile, tlsKeyFile)
	}

	return &Config{
		Port:                  port,
		DatabaseURL:           dbURL,
		STTServiceURL:         sttServiceURL,
		CORSAllowedOrigins:    corsAllowedOrigins(),
		CORSAllowLocalNetwork: corsAllowLocalNetwork(),
		VoicevoxURL:           voicevoxURL,
		PiperAddress:          piperAddress,
		AudioCacheDir:         audioCacheDir,
		TLSCertFile:           tlsCertFile,
		TLSKeyFile:            tlsKeyFile,
		DevFakeRemoteEmail:    os.Getenv("DEV_FAKE_REMOTE_EMAIL"),
		DevFakeRemoteName:     os.Getenv("DEV_FAKE_REMOTE_NAME"),
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
