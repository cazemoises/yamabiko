import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  retries: 0,
  reporter: "list",
  use: {
    baseURL: "http://localhost:5173",
    permissions: ["microphone"],
    launchOptions: {
      // fake device alimenta um tom sintético em vez de exigir microfone real;
      // fake-ui autoaceita o prompt de permissão sem interação manual.
      args: ["--use-fake-device-for-media-stream", "--use-fake-ui-for-media-stream"],
    },
  },
  webServer: {
    command: "npm run dev -- --port 5173 --strictPort",
    // Força HTTP neste processo mesmo que web/.env tenha VITE_TLS_CERT/
    // VITE_TLS_KEY setados (fluxo normal de dev local com HTTPS via
    // Tailscale, ver vite.config.ts) — process.env já setado aqui vence
    // sobre o valor lido do .env por loadEnv(), então o dev server que o
    // Playwright sobe sempre serve HTTP puro em localhost, batendo com a
    // baseURL/url abaixo.
    env: { VITE_TLS_CERT: "", VITE_TLS_KEY: "" },
    url: "http://localhost:5173",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
