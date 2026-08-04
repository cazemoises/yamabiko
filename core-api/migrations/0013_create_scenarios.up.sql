-- Cenários (Sec. pedida pelo usuário): agrupam exercícios numa sequência
-- narrativa única (ex: "chegar no escritório de manhã") em vez de exercícios
-- soltos. Retrocompatível de propósito: scenario_id/order_in_scenario em
-- exercises são NULLABLE — todo exercício existente continua funcionando
-- solto exatamente como hoje, sem cenário nenhum.
CREATE TABLE scenarios (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language                 TEXT NOT NULL,
    title_pt                 TEXT NOT NULL,
    context_description_pt  TEXT NOT NULL,
    order_index              INTEGER NOT NULL
);

CREATE INDEX idx_scenarios_language ON scenarios (language);

-- ON DELETE SET NULL (não CASCADE): apagar um cenário não deve apagar os
-- exercícios que pertenciam a ele — eles só voltam a ficar soltos.
ALTER TABLE exercises ADD COLUMN scenario_id UUID REFERENCES scenarios (id) ON DELETE SET NULL;
ALTER TABLE exercises ADD COLUMN order_in_scenario INTEGER;

CREATE INDEX idx_exercises_scenario_id ON exercises (scenario_id);
