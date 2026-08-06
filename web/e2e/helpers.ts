import type { Page } from "@playwright/test";

export const API_BASE_URL = "http://localhost:9001";

// addInitScript roda em toda navegação da página (inclusive um redirect que o
// interceptor de auth dispare no meio do teste), então sem essa guarda ele
// re-semearia os tokens depois de um logout e mascararia um clearTokens() real.
// sessionStorage sobrevive à navegação (mesma aba), diferente de uma variável em
// memória, então serve de flag "já semeei".
export async function seedTokensOnce(page: Page, accessToken: string, refreshToken: string): Promise<void> {
  await page.addInitScript(
    ({ accessToken, refreshToken }) => {
      if (sessionStorage.getItem("e2e-seeded")) return;
      sessionStorage.setItem("e2e-seeded", "1");
      localStorage.setItem("access_token", accessToken);
      localStorage.setItem("refresh_token", refreshToken);
    },
    { accessToken, refreshToken },
  );
}

// AppearanceProvider (App.tsx) busca GET /users/me em TODA página
// autenticada, não só nas telas que mostram o perfil — sem mockar isso,
// testes com token falso tomam 401 real do core-api, o interceptor de
// refresh tenta renovar (também falha) e redireciona pra /login no meio do
// teste, quebrando qualquer página que não seja sobre perfil/voz. Specs que
// já mockam /users/me com um perfil específico (ex: voice-settings,
// token-refresh) não precisam disso — só as que não têm nenhum interesse
// no conteúdo do perfil, só querem que a chamada não derrube a sessão.
export async function mockProfile(page: Page): Promise<void> {
  await page.route(`${API_BASE_URL}/users/me`, async (route) => {
    await route.fulfill({
      status: 200,
      json: {
        id: "e2e-user",
        email: "e2e@example.com",
        name: "E2E",
        created_at: "2026-01-01T00:00:00Z",
        current_sprint_day: 1,
        xp_total: 0,
        current_streak_days: 0,
        longest_streak_days: 0,
        badges: [],
      },
    });
  });
}
