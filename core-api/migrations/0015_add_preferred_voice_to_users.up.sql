-- Preferência de voz por usuário (sistema de seleção de voz com preview,
-- pedido do usuário) — 2 colunas fixas em vez de 1 tabela genérica
-- idioma->voice_id, porque hoje só existem 2 idiomas suportados (ver
-- Service.resolveVoice em core-api/internal/tts/service.go) e o valor
-- guardado aqui é só o voice_id estável do catálogo curado (voice.go), não
-- o providerVoiceID cru do motor. NULL = nenhuma preferência salva, cai no
-- default do idioma (mesmo comportamento de GetReferenceAudio com
-- voiceID="").
ALTER TABLE users ADD COLUMN preferred_voice_ja TEXT;
ALTER TABLE users ADD COLUMN preferred_voice_en TEXT;
