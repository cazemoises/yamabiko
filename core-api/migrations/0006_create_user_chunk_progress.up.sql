-- "chunk" mapeia 1:1 pra exercises.id nesta versão do produto — não existe uma unidade de
-- vocabulário/gramática separada do exercício ainda. Cada exercício é reforçado via SM-2
-- independentemente (Sec. 9 do CLAUDE.md exige TDD em srs/, ver internal/srs).
CREATE TABLE user_chunk_progress (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    chunk_id         UUID NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    status           TEXT NOT NULL DEFAULT 'NOVO' CHECK (status IN ('NOVO', 'EM_REFORCO', 'SOLIDO')),
    easiness_factor  DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    interval_days    INTEGER NOT NULL DEFAULT 0,
    repetitions      INTEGER NOT NULL DEFAULT 0,
    next_review_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, chunk_id)
);

CREATE INDEX idx_user_chunk_progress_next_review ON user_chunk_progress (user_id, next_review_at);
