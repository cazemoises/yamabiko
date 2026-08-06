import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // loadEnv (não import.meta.env, que só existe em código de CLIENTE injetado
  // pelo bundler em build time) é como vite.config.ts lê .env em contexto
  // Node — 3º argumento '' carrega TODA variável do .env, não só as
  // prefixadas VITE_ (não precisamos disso aqui já que VITE_TLS_CERT/
  // VITE_TLS_KEY já têm o prefixo, mas loadEnv com prefixo restrito exigiria
  // repetir isso em outro lugar sem ganho real). loadEnv() carrega
  // `.env` pra QUALQUER mode (além de `.env.[mode]` quando existir) — por
  // isso VITE_TLS_CERT/VITE_TLS_KEY em web/.env ativam HTTPS tanto em
  // `npm run dev` quanto em qualquer outro mode, sem precisar de script
  // dedicado. O webServer do Playwright (e2e/playwright.config.ts) força
  // HTTP de propósito via `env: { VITE_TLS_CERT: '', VITE_TLS_KEY: '' }`
  // nesse processo específico — ver comentário lá.
  const env = loadEnv(mode, process.cwd(), '')

  // HTTPS opcional via certificado real (ex: Tailscale, ver BUILD_STATE.md
  // pra como gerar) — sem VITE_TLS_CERT/VITE_TLS_KEY setados, cai pro HTTP
  // normal (fluxo de dev em localhost puro continua igual). Necessário pra
  // getUserMedia (microfone): Safari iOS exige "secure context" (HTTPS) pra
  // acessar o microfone em qualquer origin que não seja localhost — ao
  // contrário do Chrome Android, que tem bypass pra IP de rede privada em
  // HTTP puro. Sem hardcoded: os caminhos vêm inteiramente do .env local
  // (não commitado), o certificado em si nunca entra no código.
  //
  // Os 3 desfechos abaixo sempre logam algo explícito no terminal — a v1
  // caía pro HTTP em silêncio quando as variáveis não estavam definidas, o
  // que fez uma sessão anterior achar (errado) que HTTPS estava ativo.
  let https: { cert: Buffer; key: Buffer } | undefined
  if (!env.VITE_TLS_CERT && !env.VITE_TLS_KEY) {
    console.log('[vite] HTTPS não ativado: VITE_TLS_CERT/VITE_TLS_KEY não definidos em web/.env — servindo HTTP.')
  } else if (!env.VITE_TLS_CERT || !env.VITE_TLS_KEY) {
    console.warn(
      `[vite] HTTPS não ativado: apenas ${env.VITE_TLS_CERT ? 'VITE_TLS_CERT' : 'VITE_TLS_KEY'} está definido — os 2 precisam estar setados. Servindo HTTP.`,
    )
  } else if (!fs.existsSync(env.VITE_TLS_CERT) || !fs.existsSync(env.VITE_TLS_KEY)) {
    console.warn(
      `[vite] HTTPS não ativado: certificado não encontrado (cert: ${env.VITE_TLS_CERT} existe=${fs.existsSync(env.VITE_TLS_CERT)}, key: ${env.VITE_TLS_KEY} existe=${fs.existsSync(env.VITE_TLS_KEY)}) — servindo HTTP.`,
    )
  } else {
    https = {
      cert: fs.readFileSync(env.VITE_TLS_CERT),
      key: fs.readFileSync(env.VITE_TLS_KEY),
    }
    console.log(`[vite] HTTPS ativado com certificado ${env.VITE_TLS_CERT}`)
  }

  return {
    plugins: [react()],
    // Same-origin deploy no nginx compartilhado do Ascend (ver
    // BUILD_STATE.md): o app fica servido em /yamabiko/, não na raiz, então
    // todo asset gerado pelo build (JS, CSS, ícones) precisa desse prefixo
    // nos <script>/<link> do index.html — é isso que `base` controla. Sem
    // VITE_BASE_PATH setado (dev normal, ou preview local), cai pro '/'
    // de sempre. Passado via build ARG em web/Dockerfile.prod, não hardcoded
    // aqui, porque dev/e2e nunca devem rodar sob esse prefixo.
    base: env.VITE_BASE_PATH || '/',
    server: {
      // host: true == '0.0.0.0' — o dev server escuta em todas as interfaces
      // de rede, não só localhost, pra dar pra abrir o app de outro
      // dispositivo (ex: celular) na mesma rede Wi-Fi via
      // http(s)://<IP-ou-hostname-da-máquina>:5173. Sem isso o Vite só
      // aceita conexão vinda da própria máquina. Precisa também do core-api
      // com CORS_ALLOW_LOCAL_NETWORK=true e/ou o origin certo em
      // CORS_ALLOWED_ORIGINS (ver docker-compose.yml) pra aceitar esse
      // origin.
      host: true,
      https,
    },
  }
})
