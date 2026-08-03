CREATE TABLE attempts (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    exercise_id        UUID NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    audio_transcript   TEXT NOT NULL,
    similarity_score    DOUBLE PRECISION NOT NULL,
    verdict            TEXT NOT NULL CHECK (verdict IN ('PASS', 'PARTIAL', 'FAIL')),
    phonetic_diff       JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_attempts_user_id ON attempts (user_id);
CREATE INDEX idx_attempts_exercise_id ON attempts (exercise_id);
