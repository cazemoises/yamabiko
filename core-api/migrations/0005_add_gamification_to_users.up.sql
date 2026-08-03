ALTER TABLE users
    ADD COLUMN xp_total INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN current_streak_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN longest_streak_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_attempt_date DATE;
