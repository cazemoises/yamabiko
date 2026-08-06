-- Login por PIN numérico (tela de seleção de perfil, pedido do usuário) —
-- pin_hash NULL = usuário ainda não configurou PIN, não aparece em
-- GET /auth/profiles (ver internal/auth/postgres_repository.go). Setado só
-- via POST /auth/pin-setup, autenticado pelo login por senha existente (que
-- continua sendo a porta de entrada única pra configurar/resetar PIN).
--
-- pin_failed_attempts/pin_locked_until implementam lockout (5 tentativas,
-- 15min) igual em espírito ao rate limiting que falta em /auth/login (ver
-- débito técnico #4 do BUILD_STATE.md) — aqui é tratado desde já porque o
-- PIN de 6 dígitos tem espaço de busca muito menor que uma senha.
ALTER TABLE users ADD COLUMN pin_hash TEXT;
ALTER TABLE users ADD COLUMN pin_failed_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN pin_locked_until TIMESTAMPTZ;
