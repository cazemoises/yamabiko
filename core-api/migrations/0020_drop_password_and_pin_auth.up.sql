-- Remoção completa de login por senha/PIN (Sec. pedida pelo usuário: "não
-- quero mais login com pin ou email e senha, quero usar os headers do
-- pangolin") — identidade agora vem inteiramente do Pangolin (SSO por
-- pessoa, header Remote-Email), ver internal/auth/context.go e
-- internal/auth/postgres_repository.go. accent_color NÃO é removida aqui:
-- virou preferência de aparência genuína (migration 0017), não é mais só
-- decoração da tela de seleção de perfil por PIN.
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users DROP COLUMN pin_hash;
ALTER TABLE users DROP COLUMN pin_failed_attempts;
ALTER TABLE users DROP COLUMN pin_locked_until;
