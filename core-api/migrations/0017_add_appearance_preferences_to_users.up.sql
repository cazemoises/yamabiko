-- Persistência de tema/accent color (Sec. pedida pelo usuário — "vire
-- preferência de usuário real, persistida", mesmo padrão de
-- preferred_voice_ja/en da migration 0015).
--
-- theme: NULL = segue o SO (prefers-color-scheme), 'light'/'dark' = força
-- explicitamente (ver web/src/index.css, :root[data-theme]).
--
-- accent_color: NULL = default (terracota, #C1662F). Guarda o literal
-- 'mono' (acento neutro relativo ao tema — var(--text) — ou um hex
-- #RRGGBB (os outros 3 presets do design e o hex customizado usam o mesmo
-- formato, não há diferença de representação entre "preset" e "custom").
ALTER TABLE users ADD COLUMN theme TEXT CHECK (theme IS NULL OR theme IN ('light', 'dark'));
ALTER TABLE users ADD COLUMN accent_color TEXT;
