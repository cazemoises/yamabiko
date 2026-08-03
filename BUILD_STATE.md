# BUILD STATE

## Fase atual: 4 / 6 — CONCLUÍDA. Iniciando Fase 5.

## Última ação
Fase 4 (attempts end-to-end) completa e verificada via curl com **áudio real** (`mic_input.wav`) contra
o `core-api` completo (`docker compose up stt-service postgres core-api`):
- `GET /exercises` sem token → 401; com token → 200, lista filtrável por `sprint_day`/`category`/`difficulty`.
- `POST /exercises/{id}/attempts` sem arquivo → 400; com `audio=@mic_input.wav` → 201, integrando
  stt-service (transcrição real) → `comparison.Compare` → persistência em `attempts` → resposta
  `{transcript, score, verdict, diff, xp_gained}`.
- `GET /exercises/{id}/attempts` → 200, histórico do usuário retornado em `snake_case`.
- Suite Go completa (`go test ./...`) passando: 34 testes no total (auth/middleware/comparison + 4 novos
  de `sttclient` com servidor HTTP fake + 5 novos de `attempts.Service` com repo/transcriber/exercise
  finder fakes).

## Achado real do teste com áudio (não hipotético)
O `mic_input.wav` de laboratório contém uma auto-apresentação em japonês, mas o Whisper transcreveu em
**katakana** (`ホタシノナマワ...`) um enunciado que corresponderia, no gabarito do exercício, a
**hiragana** (`わたしのなまえは...`). `comparison.normalize` só aplica NFKC + remoção de espaço — não
converte katakana↔hiragana — então várias moras viraram `SUBSTITUTE`/`OUTRO` mesmo sendo foneticamente
equivalentes (た/タ, し/シ, の/ノ são a mesma sílaba em scripts diferentes). Isso infla artificialmente a
distância de edição e derruba o score. Registrado como débito técnico abaixo — não corrigido nesta fase
pra não expandir escopo do commit, mas é a próxima melhoria de precisão mais importante da engine.

## Decisões técnicas — Fase 4
- **Campo do multipart em `POST /exercises/{id}/attempts` é `audio`**, diferente do `file` usado
  internamente pelo `stt-service`. A Sec. 4 do CLAUDE.md só diz "multipart: audio file" sem nomear o
  campo; escolhido `audio` por ser mais descritivo na API pública; `sttclient` internamente sempre manda
  `file` pro stt-service, então a escolha é só da fronteira pública do core-api.
- **XP mínimo por veredito (PASS=10, PARTIAL=5, FAIL=1), hardcoded em `attempts.xpFor`** — só fecha o
  contrato de resposta (`xp_gained`) da Sec. 4. Gamificação completa (streak, multiplicadores, badges) é
  escopo da Fase 6 e vai substituir essa fórmula.
- **`attempts.Service` depende de interfaces (`Transcriber`, `ExerciseFinder`), não dos tipos concretos
  `sttclient.Client`/`exercises.PostgresRepository`** — permite testar a lógica de orquestração
  (transcrever → comparar → persistir → calcular XP) com fakes em memória, sem subir stt-service nem
  Postgres nos testes unitários.
- **Todas as rotas exceto `/auth/*` exigem Bearer token**, incluindo `GET /exercises` — leitura literal
  da Sec. 4 do CLAUDE.md ("Todas as rotas exceto /auth/* exigem..."), mesmo sendo dado não-sensível.
- **`GET /exercises/{id}/attempts` retorna todo o histórico do usuário pra aquele exercício sem
  paginação** — volume esperado é baixo (dezenas de tentativas por usuário/exercício no curso de 60
  dias); paginação vira débito técnico só se isso mudar.

## Decisões técnicas — Fase 3
- **Padrões fonéticos implementados como classificação heurística por conjunto de moras, não ML/NLP.**
  `H_ASPIRADO_OMITIDO`: deleção de mora do は行 (は/ひ/ふ/へ/ほ). `VOGAL_ENGOLIDA`: deleção de vogal pura
  (あ/い/う/え/お). `R_L_T_CONFUSAO`: substituição entre ら行 e た/だ行 (flap japonês percebido/pronunciado
  como L/T/D por falante de PT-BR). Qualquer coisa fora desses três grupos (incluindo toda inserção, já
  que não há um padrão fonético conhecido pra "mora extra alucinada") cai em `OUTRO`, conforme Sec. 3.
- **`phonetic_diff` é reconstruído via backtrace da matriz de Levenshtein**, não um diff ingênuo — dá
  alinhamento correto posição-a-posição mesmo com inserção/deleção combinadas.
- **Seed de exercícios entrou como migration versionada (`0004_seed...`)** — roda automaticamente no
  boot do `core-api` junto das migrations de schema, mesma pipeline, sem passo manual extra.
- TDD seguido literalmente no `comparison`: suite escrita primeiro (RED — `go test` falhou por pacote
  inexistente), depois implementação até GREEN.

## Decisões técnicas — Fases 1-2
- **Porta do stt-service: 8001 (host); postgres: 5433; core-api: 9001.** O projeto irmão `ascend` já usa
  8000 (web), 9000 (api), 5432 (postgres) e 6379 (redis) no Docker local.
- **Redis ainda não entrou no docker-compose.yml** — nenhuma feature construída até agora precisa dele.
  Vai entrar quando algo concreto exigir (blacklist de refresh token, rate limiting, cache de sessão).
- **Migrations rodadas automaticamente no boot do `core-api`** via `golang-migrate/migrate`, chamado no
  início de `main()` antes de abrir o pool de conexões.
- **JWT: dois tipos de token (`access`/`refresh`) discriminados por claim `type`, mesma secret HS256.**
- **`POST /auth/refresh` retorna só `{access_token}`, sem rotacionar o refresh token** — literal ao
  contrato da Sec. 4. Rotação fica como débito técnico.
- **`GET /users/me` devolve só `{id}`** — não é a rota completa da Sec. 4 (perfil + sprint + XP +
  streak), que depende de domínios ainda não construídos. Prova só que o middleware de auth funciona.
- **Repositórios com interface + implementação Postgres via `pgx/v5`** — arquitetura hexagonal leve
  (handler → service → repositório); testes de serviço usam fakes em memória, não batem no Postgres.

## Débito técnico conhecido
- **`comparison.normalize` não converte katakana↔hiragana** — ver "Achado real" acima. Causa scores
  artificialmente baixos quando o Whisper transcreve num script diferente do gabarito. Próxima melhoria
  de precisão mais importante da engine.
- Nenhum HF_TOKEN configurado no stt-service — download do modelo usa rate limit anônimo do Hugging Face
  Hub.
- `stt-service` não tem teste de integração automatizado com o modelo real — só smoke test manual via
  curl (documentado no histórico da Fase 1).
- `.gitignore` ignora `*.wav`/`*.mp3` globalmente — arquivos de laboratório continuam soltos no working
  tree, não commitados.
- **Refresh token não é revogável nem rotacionado** — sem blacklist/allowlist em Redis ainda.
- **`JWT_SECRET` do docker-compose é um valor fixo de desenvolvimento** — ok local, precisa virar secret
  de verdade antes de qualquer deploy.
- **Sem rate limiting em `/auth/login`, `/auth/register`** — vulnerável a brute-force/enumeração de
  e-mail por ora.
- **XP hardcoded por verdict** (Fase 4) — será substituído pela gamificação real na Fase 6.
- **`GET /exercises/{id}/attempts` sem paginação.**

## Próximo passo imediato
Fase 5 — Web (React + TypeScript + Vite): telas de login/registro, lista de exercícios, tela de
exercício com gravador (`MediaRecorder` API nativa, sem libs pesadas), dashboard de progresso. Consumir
a API já pronta (`/auth/*`, `/exercises`, `/exercises/{id}/attempts`) rodando em `http://localhost:9001`.

---

## Histórico — Fase 1 (stt-service), concluída
`docker compose up stt-service`, depois `curl -X POST http://localhost:8001/transcribe -F
"file=@mic_input.wav"` retornou transcrição real
(`{"transcript":"エジメマシュティ、ホタシノナマワ、モイゼスです。","language":"ja","confidence":1}`).
Decisões: modelo carregado no lifespan do FastAPI (não lazy); volume nomeado pro cache do Hugging Face;
extensões aceitas .wav/.webm/.mp3/.m4a/.ogg; testes mockam o engine (a prova real é o curl documentado
aqui, não a suíte automatizada); requirements.txt de produção separado de requirements-dev.txt.
