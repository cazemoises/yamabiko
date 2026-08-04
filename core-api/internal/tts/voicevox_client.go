package tts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type VoicevoxClient struct {
	baseURL    string
	speakerID  string
	httpClient *http.Client
}

func NewVoicevoxClient(baseURL, speakerID string) *VoicevoxClient {
	return &VoicevoxClient{baseURL: baseURL, speakerID: speakerID, httpClient: &http.Client{}}
}

// Synthesize converte texto japonês em áudio (WAV) via VOICEVOX, seguindo o
// fluxo padrão de 2 passos da API: POST /audio_query monta a "receita"
// prosódica (fonemas, pitch, duração) a partir do texto, e POST /synthesis
// renderiza o WAV a partir dessa receita — os dois passos existem separados
// pra permitir editar a prosódia entre eles, mas aqui usamos o default.
func (c *VoicevoxClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	query, err := c.audioQuery(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("voicevox audio_query: %w", err)
	}
	audio, err := c.synthesis(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("voicevox synthesis: %w", err)
	}
	return audio, nil
}

func (c *VoicevoxClient) audioQuery(ctx context.Context, text string) ([]byte, error) {
	params := url.Values{"text": {text}, "speaker": {c.speakerID}}
	endpoint := c.baseURL + "/audio_query?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doExpectingOK(req)
}

func (c *VoicevoxClient) synthesis(ctx context.Context, audioQuery []byte) ([]byte, error) {
	endpoint := c.baseURL + "/synthesis?" + url.Values{"speaker": {c.speakerID}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(audioQuery))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doExpectingOK(req)
}

func (c *VoicevoxClient) doExpectingOK(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}
	return body, nil
}
