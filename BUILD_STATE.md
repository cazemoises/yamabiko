# BUILD STATE

## Fase atual: 3 / 6 — CONCLUÍDA. Iniciando Fase 4.

## Última ação
Fase 3 (exercícios + engine de comparação) completa:
- Migrations `exercises` e `attempts` aplicadas (schema_migrations em v4).
- Seed de 10 exercícios (dias 1-3: saudação, apresentação, konbini) confirmado via `psql` dentro do
  container `postgres` — 10 linhas, categorias e sprint_day_ref corretos.
- `internal/comparison`: engine pura (normalização NFKC + remoção de espaços, Levenshtein a nível de
  rune com backtrace de operações, verdict PASS/≥0.85 · PARTIAL/0.6-0.85 · FAIL/<0.6, phonetic_diff
  classificando H_ASPIRADO_OMITIDO, VOGAL_ENGOLIDA, R_L_T_CONFUSAO, OUTRO).
- TDD seguido literalmente: suite escrita primeiro (RED — `go test` falhou por pacote inexistente),
  depois implementação até GREEN. 13 testes cobrindo os 3 padrões fonéticos + fallback OUTRO
  (substituição e inserção não reconhecidas) + limiares exatos de verdict (0.85 e 0.6) + casos de borda
  (ambos vazios, actual vazio) + normalização (espaços, kana de meia-largura).
- Suite Go completa (`go test ./...`) passando: 25 testes no total (12 de auth/middleware + 13 de
  comparison).

## Decisões técnicas tomadas (e por quê)
- **Padrões fonéticos implementados como classificação heurística por conjunto de moras, não ML/NLP.**
  `H_ASPIRADO_OMITIDO`: deleção de mora do は行 (は/ひ/ふ/へ/ほ). `VOGAL_ENGOLIDA`: deleção de vogal pura
  (あ/い/う/え/お). `R_L_T_CONFUSAO`: substituição entre ら行 e た/だ行 (flap japonês percebido/pronunciado
  como L/T/D por falante de PT-BR). Qualquer coisa fora desses três grupos (incluindo toda inserção,
  já que não há um padrão fonético conhecido pra "mora extra alucinada") cai em `OUTRO`, conforme
  Sec. 3 do CLAUDE.md.
- **`phonetic_diff` é reconstruído via backtrace da matriz de Levenshtein**, não um diff ingênuo — dá
  alinhamento correto posição-a-posição mesmo com inserção/deleção combinadas, condição necessária pra
  classificar cada divergência individualmente.
- **Seed de exercícios entrou como migration versionada (`0004_seed...`), não como script Go avulso** —
  roda automaticamente no boot do `core-api` junto das migrations de schema, mesma pipeline, sem passo
  manual extra.
- **`exercises`/`attempts` ainda não têm repositório Go nem endpoints HTTP** — só o schema+seed foram
  entregues nesta fase, conforme escopo da Sec. 5 do CLAUDE.md ("Fase 3: schema de exercises... engine
  de comparação"). Repositório, `GET /exercises`, e `POST /exercises/{id}/attempts` ficam pra Fase 4
  ("Attempts end-to-end"), que também integra o `sttclient` e o `comparison` recém-criado.

## Decisões técnicas tomadas nas fases anteriores (e por quê)
- **Porta do stt-service: 8001 (host), não 8000.** O projeto irmão `ascend` já usa 8000 (web), 9000 (api),
  5432 (postgres) e 6379 (redis) no Docker local.
- **postgres do yamabiko: porta 5433 (host); core-api: porta 9001 (host)** — mesma razão: evitar colisão
  com os containers do Ascend (`ascend-postgres-1` em 5432, `ascend-api-1` em 9000).
- **Redis ainda não entrou no docker-compose.yml** — nada em auth (Fase 2) precisa dele. Vai entrar
  quando alguma feature concreta exigir (ex: blacklist de refresh token, rate limiting, cache de sessão).
  Evita infra ociosa sem uso real.
- **Migrations rodadas automaticamente no boot do `core-api`** via `golang-migrate/migrate` (lib pura Go,
  sem precisar de CLI externa no container) — `internal/db/migrate.go`, chamado no início de `main()`
  antes de abrir o pool de conexões da aplicação.
- **JWT: dois tipos de token (`access`/`refresh`) discriminados por claim `type`, mesma secret HS256.**
  `Parse` exige o tipo esperado — um refresh token não passa como access token no middleware, e
  vice-versa (coberto em `middleware/auth_test.go` e `jwt_test.go`).
- **`POST /auth/refresh` retorna só `{access_token}`, sem rotacionar o refresh token** — segue literal o
  contrato da Sec. 4 do CLAUDE.md (`{refresh_token} -> {access_token}`). Rotação de refresh token (mais
  seguro contra replay) fica como débito técnico abaixo.
- **`GET /users/me` existe já na Fase 2, mas devolve só `{id}`** — não é a rota completa da Sec. 4
  (perfil + sprint + XP + streak), que depende de `users`/`gamification` (Fases 3/6). Serve aqui só de
  prova end-to-end de que o middleware de auth funciona; será expandida quando esses domínios existirem.
- **Repositório de users com interface (`UserRepository`) + implementação Postgres via `pgx/v5`** —
  arquitetura hexagonal leve (handler → service → repositório) conforme Sec. 9 do CLAUDE.md; testes de
  `Service` usam um fake repo em memória, não batem no Postgres.

## Débito técnico conhecido
- Nenhum HF_TOKEN configurado no stt-service — download do modelo usa rate limit anônimo do Hugging Face
  Hub (mais lento, mas funcional).
- `stt-service` não tem teste de integração real com o modelo carregado — só smoke test manual via curl
  (documentado na entrada da Fase 1 abaixo).
- `.gitignore` ignora `*.wav`/`*.mp3` globalmente — arquivos de laboratório continuam soltos no working
  tree, não commitados.
- **Refresh token não é revogável nem rotacionado** — um refresh token vazado continua válido até expirar
  (7 dias). Sem blacklist/allowlist em Redis ainda. Resolver quando Redis entrar no projeto.
- **`JWT_SECRET` do docker-compose é um valor fixo de desenvolvimento (`dev-secret-troque-em-producao`)** —
  ok para ambiente local, precisa virar secret de verdade antes de qualquer deploy.
- **Sem rate limiting em `/auth/login` e `/auth/register`** — vulnerável a brute-force/enumeração de
  e-mail por ora. Endereçar quando houver infra de Redis/rate limit no projeto.

## Próximo passo imediato
Fase 4 — Attempts end-to-end: repositório Go de `exercises` (GET /exercises, GET /exercises/{id}) e de
`attempts`, `sttclient` (HTTP interno pro stt-service), `POST /exercises/{id}/attempts` (multipart:
audio) integrando stt-service → comparison → persistência do attempt, com XP mínimo (fórmula simples
por verdict; gamificação completa com streak/badges é Fase 6). Validar fluxo completo via curl antes de
tocar no frontend.

---

## Histórico — Fase 1 (stt-service), concluída
`docker compose up stt-service`, depois `curl -X POST http://localhost:8001/transcribe -F
"file=@mic_input.wav"` retornou transcrição real
(`{"transcript":"エジメマシュティ、ホタシノナマワ、モイゼスです。","language":"ja","confidence":1}`).
Decisões: modelo carregado no lifespan do FastAPI (não lazy); volume nomeado pro cache do Hugging Face;
extensões aceitas .wav/.webm/.mp3/.m4a/.ogg; testes mockam o engine (a prova real é o curl documentado
aqui, não a suíte automatizada); requirements.txt de produção separado de requirements-dev.txt.
