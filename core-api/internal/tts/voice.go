package tts

import "strings"

// Voice é a entrada de catálogo exposta pela API — id estável (o que o
// frontend/preferência do usuário guardam), nome amigável em PT-BR e o
// idioma que ela fala. ProviderVoiceID (não exportado no JSON) é o que
// efetivamente vai pro motor de TTS: um speaker id numérico (string) pro
// VOICEVOX, ou um nome de modelo pro Piper — detalhe de implementação que o
// resto do sistema não precisa conhecer.
type Voice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`

	providerVoiceID string
}

// curatedVoices é a curadoria pedida pelo usuário: 6-8 speakers do VOICEVOX
// (não os 43 crus de GET /speakers — a maioria é personagem/mascote, sem
// serventia como referência de pronúncia num app sério) variando
// gênero/tom perceptível, e os modelos do Piper baixados pra en-US. O
// primeiro de cada idioma é o default (ver DefaultVoiceID).
//
// Speaker ids do VOICEVOX conferidos ao vivo via GET /speakers (nomes e
// estilos reais, não adivinhados) — ver BUILD_STATE.md pro raciocínio
// completo de cada escolha.
var curatedVoices = []Voice{
	{ID: "ja-announcer-neutral", Name: "Locutor(a) Neutro(a)", Language: "ja-JP", providerVoiceID: "30"}, // No.7 - アナウンス (default, já em uso)
	{ID: "ja-male-deep", Name: "Voz Masculina Grave", Language: "ja-JP", providerVoiceID: "13"},          // 青山龍星 - ノーマル
	{ID: "ja-male-young", Name: "Voz Masculina Jovem", Language: "ja-JP", providerVoiceID: "11"},         // 玄野武宏 - ノーマル
	{ID: "ja-male-formal", Name: "Voz Masculina Formal", Language: "ja-JP", providerVoiceID: "52"},       // 雀松朱司 - ノーマル
	{ID: "ja-female-natural", Name: "Voz Feminina Natural", Language: "ja-JP", providerVoiceID: "8"},     // 春日部つむぎ - ノーマル
	{ID: "ja-female-clear", Name: "Voz Feminina Clara", Language: "ja-JP", providerVoiceID: "9"},         // 波音リツ - ノーマル
	{ID: "ja-female-mature", Name: "Voz Feminina Madura", Language: "ja-JP", providerVoiceID: "20"},      // もち子さん - ノーマル

	{ID: "en-lessac", Name: "Lessac (voz masculina neutra)", Language: "en-US", providerVoiceID: "en_US-lessac-medium"}, // default, já em uso
	{ID: "en-amy", Name: "Amy (voz feminina)", Language: "en-US", providerVoiceID: "en_US-amy-medium"},
	{ID: "en-ryan", Name: "Ryan (voz masculina jovem)", Language: "en-US", providerVoiceID: "en_US-ryan-medium"},
}

// VoicesForLanguage devolve o catálogo curado pro subtag primário do idioma
// (ex: "ja"/"ja-JP" ambos casam com as vozes ja-JP) — nunca os 43 speakers
// crus do VOICEVOX. Sem cópia defensiva necessária: Voice é usado só por
// valor a partir daqui, e o slice do catálogo nunca é mutado.
func VoicesForLanguage(language string) []Voice {
	primary := primaryLanguageSubtag(language)
	result := make([]Voice, 0, len(curatedVoices))
	for _, v := range curatedVoices {
		if primaryLanguageSubtag(v.Language) == primary {
			result = append(result, v)
		}
	}
	return result
}

// DefaultVoiceID é a 1ª voz curada de cada idioma (ver curatedVoices) — é
// o que GetReferenceAudio usa quando o usuário não escolheu nenhuma
// preferência, ou quando a preferência salva não bate com nenhuma voz do
// catálogo atual (ex: catálogo mudou desde que o usuário escolheu).
func DefaultVoiceID(language string) string {
	voices := VoicesForLanguage(language)
	if len(voices) == 0 {
		return ""
	}
	return voices[0].ID
}

// findVoice resolve um voice_id (id estável, não o providerVoiceID interno)
// pra sua entrada completa de catálogo. ok=false pra id desconhecido.
func findVoice(voiceID string) (Voice, bool) {
	for _, v := range curatedVoices {
		if v.ID == voiceID {
			return v, true
		}
	}
	return Voice{}, false
}

func primaryLanguageSubtag(language string) string {
	return strings.ToLower(strings.SplitN(language, "-", 2)[0])
}
