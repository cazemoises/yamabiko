import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // loadEnv (não import.meta.env, que só existe em código de CLIENTE) é como
  // vite.config.ts lê .env em contexto Node — 3º argumento '' carrega TODA
  // variável do .env, não só as prefixadas VITE_ (não precisamos disso aqui
  // já que VITE_TLS_CERT/VITE_TLS_KEY já têm o prefixo, mas loadEnv com
  // prefixo restrito exigiria repetir isso em outro lugar sem ganho real).
  const env = loadEnv(mode, process.cwd(), '')

  // HTTPS opcional via certificado real (ex: Tailscale, ver BUILD_STATE.md
  // pra como gerar) — sem VITE_TLS_CERT/VITE_TLS_KEY setados, cai pro HTTP
  // normal (fluxo de dev em localhost puro continua igual). Necessário pra
  // getUserMedia (microfone): Safari iOS exige "secure context" (HTTPS) pra
  // acessar o microfone em qualquer origin que não seja localhost — ao
  // contrário do Chrome Android, que tem bypass pra IP de rede privada em
  // HTTP puro. Sem hardcoded: os caminhos vêm inteiramente do .env local
  // (não commitado), o certificado em si nunca entra no código.
  let https: { cert: Buffer; key: Buffer } | undefined
  if (env.VITE_TLS_CERT && env.VITE_TLS_KEY) {
    if (fs.existsSync(env.VITE_TLS_CERT) && fs.existsSync(env.VITE_TLS_KEY)) {
      https = {
        cert: fs.readFileSync(env.VITE_TLS_CERT),
        key: fs.readFileSync(env.VITE_TLS_KEY),
      }
    } else {
      console.warn(
        `[vite] VITE_TLS_CERT/VITE_TLS_KEY setados mas o arquivo não existe (cert: ${env.VITE_TLS_CERT}, key: ${env.VITE_TLS_KEY}) — caindo pra HTTP.`,
      )
    }
  }

  return {
    plugins: [react()],
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
