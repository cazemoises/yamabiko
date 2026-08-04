-- O produto é ja-JP-only por enquanto (Sec. 1 do CLAUDE.md), mas o seed
-- curricular importado nesta sessão referencia language='ja-JP' explicitamente
-- por exercício — coluna adicionada como metadado explícito em vez de deixar
-- implícito, sem mudar o escopo do produto (ainda não há suporte de verdade a
-- múltiplos idiomas, é só rotulagem).
ALTER TABLE exercises ADD COLUMN language TEXT NOT NULL DEFAULT 'ja-JP';
