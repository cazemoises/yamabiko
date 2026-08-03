# BUILD STATE

## Fase atual: 6 / 6 — CONCLUÍDA. Todas as fases do MVP entregues (com 1 ressalva, ver abaixo).

## Última ação
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
  `localStorage`. **Não testado interativamente em browser real nesta sessão** (sem ferramenta de
  browser disponível) — ver débito técnico abaixo.

## Débito técnico conhecido (ordenado por prioridade)
1. **Fase 5 (web) não foi testada interativamente num browser real.** `tsc -b && vite build` e `oxlint`
   passam limpo, dev server serve os módulos, mas cliques reais, `MediaRecorder` com microfone de
   verdade, e as telas renderizando dados reais do `core-api` não foram verificados por mim (sem
   ferramenta de browser nesta sessão). Rodar manualmente `cd web && npm run dev`, abrir
   `http://localhost:5173`, com `docker compose up` no ar, e percorrer registro → lista → gravação →
   resultado → dashboard antes de confiar na tela de exercício em uso real.
2. **Efeitos colaterais de `attempts.Submit` (SM-2/gamificação/phonetics) não são transacionais** — ver
   decisão técnica acima. Se isso incomodar, próximo passo é envolver os 4 passos numa transação
   Postgres ou introduzir padrão outbox.
3. **Refresh token não é revogável nem rotacionado** — sem blacklist/allowlist em Redis (que ainda não
   entrou no projeto — nada até agora precisou dele de verdade).
4. **`JWT_SECRET` do docker-compose é um valor fixo de desenvolvimento** — trocar antes de qualquer
   deploy real.
5. **Sem rate limiting em `/auth/login`, `/auth/register`** — vulnerável a brute-force/enumeração.
6. **`GET /exercises/{id}/attempts` sem paginação** — ok pro volume atual (dezenas por usuário/exercício
   em 60 dias), vira problema se isso crescer muito.
7. **`react-router-dom@7.18.2` tem 1 advisory de severidade alta em aberto** (RSC Mode CSRF,
   GHSA-qwww-vcr4-c8h2) — não aplicável: SPA cliente puro, sem RSC/Server Actions. Reavaliar se o
   projeto adotar SSR algum dia.
8. Nenhum HF_TOKEN configurado no stt-service — download do modelo usa rate limit anônimo do HF Hub.
9. `stt-service` não tem teste de integração automatizado com o modelo real, só smoke test manual via
   curl.
10. `.gitignore` ignora `*.wav`/`*.mp3` globalmente — arquivos de laboratório (`mic_input.wav`,
    `ttsMP3.com_*.mp3`) e os scripts originais (`mic_test.py`, `transcribe.py`) continuam soltos no
    working tree, nunca commitados — decisão deliberada (são só prova de conceito de laboratório, Sec. 0
    do CLAUDE.md), não esquecimento.

## Próximo passo imediato
Todas as 6 fases do MVP end-to-end estão implementadas e verificadas via curl/testes automatizados
(exceto a verificação manual de browser da Fase 5, item 1 do débito técnico acima — maior prioridade).
Se reabrir uma sessão sem instrução nova do usuário: (a) rodar a suite completa (`go test ./...` em
`core-api`, `npm run build` em `web`) pra confirmar que nada quebrou; (b) se houver acesso a browser,
fechar o item 1 do débito técnico; (c) senão, atacar o item 2 (transação nos efeitos colaterais de
`attempts.Submit`) ou o item 3 (Redis + revogação de refresh token), que são os próximos de maior
impacto técnico real.

---

## Histórico — Fase 1 (stt-service), concluída
`docker compose up stt-service`, depois `curl -X POST http://localhost:8001/transcribe -F
"file=@mic_input.wav"` retornou transcrição real. Modelo carregado no lifespan do FastAPI; volume
nomeado pro cache HF; testes mockam o engine (a prova real foi o curl, não a suíte automatizada).
