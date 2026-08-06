import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // host: true == '0.0.0.0' — o dev server escuta em todas as interfaces
    // de rede, não só localhost, pra dar pra abrir o app de outro
    // dispositivo (ex: celular) na mesma rede Wi-Fi via
    // http://<IP-da-máquina>:5173. Sem isso o Vite só aceita conexão vinda
    // da própria máquina. Precisa também do core-api com
    // CORS_ALLOW_LOCAL_NETWORK=true (ver docker-compose.yml) pra aceitar
    // esse origin.
    host: true,
  },
})
