package tts_test

import (
	"testing"

	"github.com/yamabiko/core-api/internal/tts"
)

func TestVoicesForLanguage_JapaneseReturnsCuratedSubsetNotRawSpeakers(t *testing.T) {
	voices := tts.VoicesForLanguage("ja-JP")

	if len(voices) < 12 || len(voices) > 20 {
		t.Fatalf("esperava entre 12 e 20 vozes curadas pra ja-JP (bem menos que os 43 speakers crus), veio %d", len(voices))
	}
	for _, v := range voices {
		if v.Language != "ja-JP" {
			t.Fatalf("esperava só vozes ja-JP, veio %q pra voz %q", v.Language, v.ID)
		}
		if v.ID == "" || v.Name == "" {
			t.Fatalf("voz sem id/nome: %+v", v)
		}
	}
}

func TestVoicesForLanguage_MatchesByPrimarySubtag(t *testing.T) {
	byFullTag := tts.VoicesForLanguage("ja-JP")
	byPrimaryOnly := tts.VoicesForLanguage("ja")

	if len(byFullTag) != len(byPrimaryOnly) {
		t.Fatalf("esperava o mesmo resultado pra 'ja-JP' e 'ja', veio %d vs %d", len(byFullTag), len(byPrimaryOnly))
	}
}

func TestVoicesForLanguage_EnglishReturnsAtLeastThreeVoices(t *testing.T) {
	voices := tts.VoicesForLanguage("en-US")

	// Só 3 porque é o total de vozes en-US de locutor único em qualidade
	// "high" disponíveis no repositório oficial do Piper (lessac/ryan/ljspeech).
	if len(voices) < 3 {
		t.Fatalf("esperava pelo menos 3 vozes en-US, veio %d", len(voices))
	}
	for _, v := range voices {
		if v.Language != "en-US" {
			t.Fatalf("esperava só vozes en-US, veio %q pra voz %q", v.Language, v.ID)
		}
	}
}

func TestVoicesForLanguage_UnknownLanguageReturnsEmpty(t *testing.T) {
	voices := tts.VoicesForLanguage("fr-FR")
	if len(voices) != 0 {
		t.Fatalf("esperava lista vazia pra idioma sem curadoria, veio %d vozes", len(voices))
	}
}

func TestDefaultVoiceID_JapaneseIsTheNeutralAnnouncer(t *testing.T) {
	// speaker=30 (No.7 - アナウンス) é o default histórico (mantido, não trocado
	// nesta curadoria) — a 1ª entrada do catálogo ja-JP deve continuar sendo ela.
	if got := tts.DefaultVoiceID("ja-JP"); got != "ja-announcer-neutral" {
		t.Fatalf("esperava default ja-announcer-neutral, veio %q", got)
	}
}

func TestDefaultVoiceID_EnglishIsLessac(t *testing.T) {
	if got := tts.DefaultVoiceID("en-US"); got != "en-lessac" {
		t.Fatalf("esperava default en-lessac, veio %q", got)
	}
}

func TestDefaultVoiceID_UnknownLanguageReturnsEmpty(t *testing.T) {
	if got := tts.DefaultVoiceID("fr-FR"); got != "" {
		t.Fatalf("esperava string vazia pra idioma sem curadoria, veio %q", got)
	}
}

func TestVoicesForLanguage_AllVoiceIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, lang := range []string{"ja-JP", "en-US"} {
		for _, v := range tts.VoicesForLanguage(lang) {
			if seen[v.ID] {
				t.Fatalf("voice_id duplicado no catálogo: %q", v.ID)
			}
			seen[v.ID] = true
		}
	}
}
