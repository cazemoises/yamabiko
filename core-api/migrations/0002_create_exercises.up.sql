CREATE TABLE exercises (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category              TEXT NOT NULL,
    difficulty            SMALLINT NOT NULL CHECK (difficulty BETWEEN 1 AND 5),
    prompt_pt             TEXT NOT NULL,
    expected_transcript   TEXT NOT NULL,
    expected_romaji       TEXT,
    sprint_day_ref        INTEGER NOT NULL
);

CREATE INDEX idx_exercises_sprint_day_ref ON exercises (sprint_day_ref);
CREATE INDEX idx_exercises_category ON exercises (category);
