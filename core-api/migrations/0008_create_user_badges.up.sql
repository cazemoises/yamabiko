-- Catálogo de badges vive em código (internal/gamification), não em tabela — são poucos e
-- estáticos por ora. Esta tabela só registra quais badges cada usuário já conquistou.
CREATE TABLE user_badges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    badge_code  TEXT NOT NULL,
    earned_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, badge_code)
);
