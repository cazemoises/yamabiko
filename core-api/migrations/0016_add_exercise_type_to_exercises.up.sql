-- 7 tipos de exercício novos além do áudio original (pedido do usuário,
-- substitui/amplia uma versão anterior do pedido que só cobria 5).
-- exercise_type usa CHECK em vez de um tipo ENUM nativo do Postgres — mesmo
-- padrão já usado pra difficulty (ver 0002_create_exercises) — porque é mais
-- simples de estender depois (ALTER TYPE ... ADD VALUE tem restrições
-- transacionais chatas, DROP/RECREATE CONSTRAINT não). Default
-- 'audio_pronunciation' cobre os exercícios existentes sem precisar de
-- backfill: são todos de áudio hoje.
--
-- type_data é NULLABLE de propósito: 'audio_pronunciation' e 'dictation'
-- não usam type_data nenhum (dictation reaproveita expected_transcript/
-- expected_romaji já existentes, igual ao áudio). A estrutura de type_data
-- pros outros 5 tipos é responsabilidade da camada de aplicação
-- (core-api/internal/exercises/validation/), não validada em SQL.
ALTER TABLE exercises ADD COLUMN exercise_type TEXT NOT NULL DEFAULT 'audio_pronunciation'
    CHECK (exercise_type IN (
        'audio_pronunciation',
        'multiple_choice_translation',
        'word_order',
        'verb_conjugation',
        'dictation',
        'free_translation',
        'matching_pairs',
        'true_false'
    ));

ALTER TABLE exercises ADD COLUMN type_data JSONB;
