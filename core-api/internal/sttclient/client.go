// Package sttclient é o único ponto do core-api que fala com o stt-service.
// core-api nunca chama faster-whisper diretamente (Sec. 1 do CLAUDE.md).
package sttclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, httpClient: &http.Client{}}
}

type TranscriptionResult struct {
	Transcript string  `json:"transcript"`
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// Transcribe envia o áudio e o idioma-alvo (ex: "ja-JP", "en-US") pro
// stt-service — o Whisper transcreve melhor quando sabe de antemão em que
// idioma o áudio está, em vez de detectar automaticamente.
func (c *Client) Transcribe(ctx context.Context, filename string, audio io.Reader, language string) (*TranscriptionResult, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, audio); err != nil {
		return nil, err
	}
	if err := writer.WriteField("language", language); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/transcribe", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stt-service retornou status %d: %s", resp.StatusCode, respBody)
	}

	var result TranscriptionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
