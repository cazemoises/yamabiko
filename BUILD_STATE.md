# BUILD STATE

## Fase atual: 1 / 6 — CONCLUÍDA. Iniciando Fase 2.

## Última ação
Fase 1 (stt-service) completa e verificada end-to-end: `docker compose up stt-service`, depois
`curl -X POST http://localhost:8001/transcribe -F "file=@mic_input.wav"` retornou transcrição real
(`{"transcript":"エジメマシュティ、ホタシノナマワ、モイゼスです。","language":"ja","confidence":1}`).

## Decisões técnicas tomadas (e por quê)
- **Porta do stt-service: 8001 (host), não 8000.** O projeto irmão `ascend` já usa 8000 (web), 9000 (api),
  5432 (postgres) e 6379 (redis) no Docker local. Ao chegar em `core-api` (Fase 2), postgres/redis do
  yamabiko também vão precisar de portas de host alternativas (ex: 5433, 6380) pra não colidir com o Ascend.
- **whisper_engine.py carrega o modelo no lifespan do FastAPI (startup), não lazy na primeira request** —
  evita que o primeiro curl de verificação pague o custo de load do modelo (~40s) misturado com o de
  inferência.
- **Volume nomeado `whisper-models` montado em `/root/.cache/huggingface`** — evita re-download do
  large-v3-turbo (~1.6GB) a cada rebuild da imagem.
- **Extensões aceitas em `/transcribe`: .wav, .webm, .mp3, .m4a, .ogg** — cobre gravação via
  MediaRecorder do browser (webm) e os formatos de teste manual já usados em mic_test.py/transcribe.py.
- **Testes do stt-service mockam `get_model`/`transcribe`** (`tests/test_main.py`) — não fazem download
  nem carregam o modelo real, só validam contrato HTTP (rota, status, shape). A prova real de que a
  transcrição funciona é o teste curl documentado acima, não a suíte automatizada.
- **requirements.txt (prod) separado de requirements-dev.txt** — imagem Docker não carrega pytest/httpx.

## Débito técnico conhecido
- Nenhum HF_TOKEN configurado — download do modelo usa rate limit anônimo do Hugging Face Hub (mais lento,
  mas funcional). Se virar gargalo, considerar variável de ambiente com token.
- `stt-service` ainda não tem teste de integração real (upload de áudio de verdade rodando o modelo) —
  só o smoke test manual via curl, feito uma vez, documentado aqui. Se quebrar futuramente, não há teste
  automatizado que pegue.
- `.gitignore` ignora `*.wav`/`*.mp3` globalmente — os arquivos de laboratório (`mic_input.wav`,
  `ttsMP3.com_*.mp3`) continuam soltos no working tree, não commitados, usados só pra teste manual local.

## Próximo passo imediato
Fase 2 — Auth no `core-api` (Go 1.22 + chi + PostgreSQL + Redis): schema de `users`, registro, login,
JWT (access+refresh), bcrypt, middleware de autenticação. Testes de auth antes de seguir pra Fase 3.
