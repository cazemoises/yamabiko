# BUILD STATE

## Fase atual: 6 / 6 — CONCLUÍDA. Web validado em browser real. MVP end-to-end completo.

## Última ação
Corrigidos dois bugs reais encontrados na primeira verificação de browser de verdade do `web`
(o item #1 do débito técnico anterior — nesta sessão consegui instalar Playwright via `npx` no
scratchpad e rodar um browser Chromium de verdade, então a ressalva "não testado em browser" da Fase 5
**caiu**):

1. **CORS bloqueando o frontend** — `core-api` não tinha nenhum middleware de CORS; preflight `OPTIONS`
   pra `/auth/register` vindo de `http://localhost:5173` voltava 405, então nenhuma chamada do `web`
   funcionava. Adicionado `github.com/go-chi/cors` no router, origins permitidas configuráveis via
   `CORS_ALLOWED_ORIGINS` (env var, default `http://localhost:5173` se não setada — produção seta a env
   var com o domínio real). Testado via curl simulando preflight (`OPTIONS` com `Origin` +
   `Access-Control-Request-Method`) antes e depois: 405→200 com os headers corretos; origin não
   permitida não recebe `Access-Control-Allow-Origin` (bloqueio continua funcionando pro que não é
   `localhost:5173`). Confirmado depois com login real no browser (Playwright) — zero erros de CORS no
   console.
2. **Telas do `web` não centralizadas** — `#root`/`main` não tinham `display:flex`+centralização; o
   conteúdo (login, lista, exercício, dashboard) ficava colado no canto superior esquerdo da viewport.
   Corrigido em `index.css`: `#root` virou flex-column com `min-height:100svh`; `main` (que envolve
   todas as rotas em `App.tsx`) ganhou `flex:1; display:flex; align-items:center; justify-content:center`,
   centralizando qualquer página tanto na horizontal quanto na vertical; cada container de página
   (`.auth-page`, `.exercise-page`, `.exercises-list`, `.dashboard-page`) ganhou `width:100%` +
   `max-width` (360px pro card de auth, 600px pras telas com mais conteúdo) pra não esticar full-bleed
   num monitor grande.
3. **Validação real de ponta a ponta, não só descrição visual**: subi `docker compose up` (stt-service +
   postgres + core-api) e `npm run dev` no `web`, instalei Playwright/Chromium via `npx` no diretório de
   scratchpad (não virou dependência do projeto), e rodei um script Node fazendo login de verdade
   (`moises@example.com`) e navegando por `/login`, `/exercises`, `/exercises/{id}`, `/dashboard`, em
   desktop (1280×800) e mobile (390×844). Resultado: **zero erros de console**, todas as 4 telas
   centralizadas corretamente nos dois tamanhos, dashboard mostrando dados reais e coerentes com o que
   eu já tinha validado via curl na Fase 6 (`2 de 10 exercícios tentados`, `FAIL 46%` no exercício de
   auto-apresentação — mesmo score da correção de katakana/hiragana). Screenshots ficaram no diretório de
   scratchpad da sessão (não commitados — são artefato de verificação, não parte do produto).
4. Suite Go completa continua passando (`go test ./...`), `tsc -b && vite build` e `oxlint` continuam
   limpos no `web`.

## Última ação (Fase 6, histórico)
Fase 6 (gamificação + SRS) completa e verificada via curl com **áudio real** contra o stack completo
(`docker compose up stt-service postgres core-api`):
- `internal/srs`: SM-2 puro (easiness factor, intervalo, repetições), **TDD literal** (RED confirmado
  antes da implementação, igual ao `comparison` na Fase 3).
- `internal/gamification`: XP por veredito, streak diário (consecutivo/mesmo-dia/quebra-de-sequência),
  catálogo de 5 badges, com testes.
- `internal/phonetics`: agrega `phonetic_diff` de cada attempt em `phonetic_error_patterns` por usuário.
- Tudo isso plugado em `attempts.Service.Submit` — cada tentativa agora atualiza `user_chunk_progress`
  (SM-2, chunk = exercise_id), XP/streak/badges em `users`, e `phonetic_error_patterns`, **de verdade**,
  não mais o XP hardcoded da Fase 4.
- Endpoints novos: `GET /users/me` (perfil completo: XP, streak, badges), `GET /users/me/progress`
  (chunks NOVO/EM_REFORCO/SOLIDO + heatmap fonético), `GET /dashboard/heatmap`.
- Verificado via curl real: 1ª tentativa → `xp_total: 1`, `streak: 1`, badge `PRIMEIRA_TENTATIVA`; 2ª
  tentativa no mesmo dia (outro exercício) → XP soma, streak **não** muda (comportamento correto);
  `GET /users/me/progress` → `chunks_em_reforco` subiu de 1→2 conforme as revisões SM-2 se acumulavam;
  heatmap agregando `OUTRO` corretamente a partir dos diffs reais.
- **Bônus fora do escopo estrito da Fase 6, mas era o item de maior prioridade do débito técnico**:
  corrigido `comparison.normalize` pra converter katakana→hiragana (TDD, RED→GREEN) — o Whisper
  transcreve às vezes num script diferente do gabarito pra sons idênticos, o que inflava a distância de
  edição artificialmente. Revalidado com o mesmo áudio real da Fase 4: score subiu de 0.25 (18 diffs,
  todos espúrios por causa do script) pra 0.46 (13 diffs, agora refletindo divergência de conteúdo real,
  não mais de script).
- Suite Go completa (`go test ./...`) passando: **54 testes** no total, build limpo, `go vet` limpo.

## Achado de ambiente (não é bug do produto)
Durante a validação via curl nesta sessão, `curl -d '{"name":"Moisés"}'` rodado através do Bash tool no
Git Bash do Windows corrompeu o "é" no banco (virou `Mois�s`, confirmado via `psql` direto, não é só
exibição de terminal). Testado e confirmado: `curl --data-binary @arquivo.json` com o JSON escrito via
ferramenta de arquivo preserva UTF-8 corretamente. **Não é um bug no `core-api`** — é uma armadilha do
ambiente de teste (encoding do Git Bash/Windows ao passar string inline com acento pra um processo
filho). Registrar como nota de processo pra quem for testar manualmente via terminal Windows: prefira
`--data-binary @arquivo.json` a `-d '...com acento...'` inline.

## Decisões técnicas — CORS + layout (pós-Fase 6)
- **`CORS_ALLOWED_ORIGINS` é uma lista separada por vírgula, default `http://localhost:5173`** —
  hardcoded como fallback pra não exigir configuração extra em dev local; produção deve sempre setar a
  env var explicitamente com o(s) domínio(s) real(is) do `web` (a origin do stt-service/core-api interno
  não precisa entrar aqui, CORS só importa pra chamadas partindo de um browser).
- **`AllowCredentials: true` no CORS** — o `web` não usa cookies hoje (tokens vão em `Authorization:
  Bearer`, guardados em `localStorage`), mas deixei habilitado porque é inofensivo com a whitelist de
  origin explícita (CORS não permite `AllowCredentials:true` + `AllowedOrigins:["*"]` juntos por
  design do browser, e aqui já não uso wildcard) e evita ter que mexer nisso de novo se algo migrar pra
  cookie no futuro (ex: refresh token em cookie httpOnly).
- **Centralização via `main` (flex, `align-items:center; justify-content:center`), não uma classe
  aplicada individualmente em cada página** — garante que qualquer tela nova adicionada no futuro já
  nasce centralizada por herdar o layout do `App.tsx`, sem precisar lembrar de replicar CSS por página.

## Decisões técnicas — Fase 6
- **"Chunk" (Sec. 8 do CLAUDE.md) mapeia 1:1 pra `exercises.id`** — não existe uma unidade de
  vocabulário/gramática separada do exercício no produto ainda. Cada exercício é reforçado via SM-2
  independentemente. Documentado na migration `0006_create_user_chunk_progress`.
  Se o produto evoluir pra ter "chunks" de fato distintos de exercícios (ex: um chunk gramatical
  reaparecendo em vários exercícios), essa migration precisa mudar.
- **XP/streak vivem como colunas em `users`** (`xp_total`, `current_streak_days`,
  `longest_streak_days`, `last_attempt_date`), não numa tabela separada de gamificação — o modelo de
  dados da Sec. 2 do CLAUDE.md não previa uma tabela própria pra isso, e são poucos campos, sempre lidos
  junto do perfil.
- **Badges: catálogo estático em código (`internal/gamification`), só o "já conquistou" persiste**
  (`user_badges`, migration `0008`) — 5 badges por ora (primeira tentativa, primeiro PASS, streak 3/7
  dias, 100 XP). Adicionar badge novo = adicionar constante + regra em `RecordAttempt`, sem migration.
- **Efeitos colaterais de gamificação (SM-2, XP/streak, phonetic patterns) rodam sequencialmente dentro
  de `attempts.Service.Submit`, depois do `attempt` já persistido** — se um desses passos falhar (ex:
  erro de rede no Postgres no meio do fluxo), o attempt já foi salvo mas o usuário recebe 502. Não há
  transação cobrindo os 4 passos (attempt + srs + gamification + phonetics) nem outbox/retry. Aceitável
  pro estágio atual do produto (uso pessoal, não multi-tenant crítico); documentado como débito abaixo.
- **`QualityFromScore` (srs) usa os mesmos limiares 0.85/0.6 do `comparison.Verdict`** — mantém as duas
  escalas conceitualmente alinhadas (quem passa também "lembrou bem" pro SM-2) sem acoplar os pacotes
  (comparison não importa srs, nem vice-versa; quem faz a ponte é `attempts.Service`).
- **`GET /users/me/progress` calcula `chunks_novos` como `total_exercícios - solido - em_reforco`**,
  não como contagem de linhas com status NOVO — porque uma linha só existe em `user_chunk_progress`
  depois da 1ª revisão; um exercício nunca tentado não tem linha nenhuma lá.

## Decisões técnicas — Fases 1-5 (resumo)
- Portas locais pra não colidir com o projeto irmão `ascend`: stt-service 8001, postgres 5433, core-api
  9001, web (vite dev) 5173.
- `whisper_engine.py` carrega o modelo no lifespan do FastAPI; volume nomeado pro cache HF; extensões
  aceitas .wav/.webm/.mp3/.m4a/.ogg.
- Migrations do `core-api` rodam automaticamente no boot via `golang-migrate/migrate`.
- JWT com dois tipos (access/refresh) discriminados por claim; refresh não rotaciona (só emite novo
  access token, literal ao contrato da Sec. 4).
- Arquitetura hexagonal leve em todo o `core-api`: handler → service → repositório (interface), fakes em
  memória nos testes de serviço, nunca batendo no Postgres real.
- `comparison`: Levenshtein a nível de rune com backtrace de operações; verdict PASS≥0.85/PARTIAL≥0.6/
  FAIL<0.6; padrões H_ASPIRADO_OMITIDO/VOGAL_ENGOLIDA/R_L_T_CONFUSAO/OUTRO — TDD literal.
- `web`: Vite+React+TS, `MediaRecorder` nativo, roteamento protegido, cliente HTTP fino com JWT em
  `localStorage`. **Testado interativamente em browser real (Chromium via Playwright), mais tarde
  na mesma sessão** — ver "Última ação" no topo do arquivo. CORS e centralização de layout eram os dois
  bugs reais que essa verificação revelou; ambos corrigidos.

## Débito técnico conhecido (ordenado por prioridade)
1. **Efeitos colaterais de `attempts.Submit` (SM-2/gamificação/phonetics) não são transacionais** — ver
   decisão técnica da Fase 6. Se isso incomodar, próximo passo é envolver os 4 passos numa transação
   Postgres ou introduzir padrão outbox.
2. **Refresh token não é revogável nem rotacionado** — sem blacklist/allowlist em Redis (que ainda não
   entrou no projeto — nada até agora precisou dele de verdade).
3. **`JWT_SECRET` do docker-compose é um valor fixo de desenvolvimento** — trocar antes de qualquer
   deploy real.
4. **Sem rate limiting em `/auth/login`, `/auth/register`** — vulnerável a brute-force/enumeração.
5. **`GET /exercises/{id}/attempts` sem paginação** — ok pro volume atual (dezenas por usuário/exercício
   em 60 dias), vira problema se isso crescer muito.
6. **`react-router-dom@7.18.2` tem 1 advisory de severidade alta em aberto** (RSC Mode CSRF,
   GHSA-qwww-vcr4-c8h2) — não aplicável: SPA cliente puro, sem RSC/Server Actions. Reavaliar se o
   projeto adotar SSR algum dia.
7. **`GET /dashboard/heatmap` ainda não foi exercitado pela UI do `web`** — o dashboard atual só usa
   `GET /exercises` + `GET /exercises/{id}/attempts` (N+1) pra montar a tabela de progresso; o endpoint
   agregado (`GET /users/me/progress`, `GET /dashboard/heatmap`) existe no backend desde a Fase 6 mas o
   frontend não foi atualizado pra consumi-lo. Trocar o N+1 do dashboard por esses endpoints é uma
   melhoria de performance pendente, não um bug.
8. Nenhum HF_TOKEN configurado no stt-service — download do modelo usa rate limit anônimo do HF Hub.
9. `stt-service` não tem teste de integração automatizado com o modelo real, só smoke test manual via
   curl.
10. `.gitignore` ignora `*.wav`/`*.mp3` globalmente — arquivos de laboratório (`mic_input.wav`,
    `ttsMP3.com_*.mp3`) e os scripts originais (`mic_test.py`, `transcribe.py`) continuam soltos no
    working tree, nunca commitados — decisão deliberada (são só prova de conceito de laboratório, Sec. 0
    do CLAUDE.md), não esquecimento.

## Próximo passo imediato
Todas as 6 fases do MVP end-to-end estão implementadas e verificadas — incluindo o `web` num browser
real, não só build/tsc. Não há mais nenhuma fase pendente do CLAUDE.md original. Se reabrir uma sessão
sem instrução nova do usuário: (a) rodar a suite completa (`go test ./...` em `core-api`, `npm run
build` em `web`) pra confirmar que nada quebrou; (b) atacar o item 1 do débito técnico (transação nos
efeitos colaterais de `attempts.Submit`) ou o item 2 (Redis + revogação de refresh token), que são os
próximos de maior impacto técnico real; (c) considerar trocar o N+1 do dashboard (item 7) pelos
endpoints agregados que já existem no backend.

---

## Histórico — Fase 1 (stt-service), concluída
`docker compose up stt-service`, depois `curl -X POST http://localhost:8001/transcribe -F
"file=@mic_input.wav"` retornou transcrição real. Modelo carregado no lifespan do FastAPI; volume
nomeado pro cache HF; testes mockam o engine (a prova real foi o curl, não a suíte automatizada).
