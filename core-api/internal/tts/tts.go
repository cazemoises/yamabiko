// Package tts gera áudio de referência de pronúncia a partir de texto,
// escolhendo o motor certo por idioma — VOICEVOX pra ja-JP, Piper pra en-US
// (Sec. pedida pelo usuário). O resto do core-api só vê a interface
// TTSClient (texto -> WAV); cada client concreto (voicevox_client.go,
// piper_client.go) é o único ponto do core-api que conhece o protocolo de
// fio do respectivo motor (HTTP pro VOICEVOX, Wyoming/TCP puro pro Piper).
package tts

import "context"

// TTSClient sintetiza texto em áudio (WAV) numa voz específica do motor —
// providerVoiceID é o identificador cru que o motor entende (speaker id do
// VOICEVOX, nome de modelo do Piper), resolvido a partir do voice_id
// estável do catálogo (Sec. voice.go) por quem chama.
type TTSClient interface {
	Synthesize(ctx context.Context, text, providerVoiceID string) ([]byte, error)
}
