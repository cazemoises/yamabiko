-- Reversão só recria as colunas (dados de senha/PIN antigos são perdidos —
-- irreversível de fato, mas o schema volta a compilar contra o código
-- anterior à Sec. do Pangolin se for preciso).
ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE users ALTER COLUMN password_hash DROP DEFAULT;
ALTER TABLE users ADD COLUMN pin_hash TEXT;
ALTER TABLE users ADD COLUMN pin_failed_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN pin_locked_until TIMESTAMPTZ;
