-- Suporte multi-idioma (pedido do usuário): rotula em que idioma cada attempt
-- foi tentado e em que idioma cada padrão fonético recorrente foi observado.
-- Aditivo, default 'ja-JP' pra bater com o produto ja-JP-only histórico (mesma
-- decisão da migration 0009 pra exercises) — não muda comportamento existente.
ALTER TABLE attempts ADD COLUMN language TEXT NOT NULL DEFAULT 'ja-JP';

ALTER TABLE phonetic_error_patterns ADD COLUMN language TEXT NOT NULL DEFAULT 'ja-JP';

-- pattern_type já é implicitamente exclusivo por idioma (nenhuma categoria da
-- taxonomia japonesa e da inglesa compartilha nome), mas a constraint composta
-- deixa isso explícito e evita colisão futura se um pattern_type genérico vier
-- a ser reaproveitado entre idiomas.
ALTER TABLE phonetic_error_patterns DROP CONSTRAINT phonetic_error_patterns_user_id_pattern_type_key;
ALTER TABLE phonetic_error_patterns ADD CONSTRAINT phonetic_error_patterns_user_id_pattern_type_language_key
    UNIQUE (user_id, pattern_type, language);
