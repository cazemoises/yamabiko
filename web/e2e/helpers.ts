import type { Page } from "@playwright/test";

export const API_BASE_URL = "http://localhost:9001";

// Sem login/token: a identidade vem dos headers que o Pangolin injetaria
// em produção (Sec. pedida pelo usuário — ver
// web/src/features/auth/AuthContext.tsx), então em e2e o único requisito
// pra "estar autenticado" é GET /users/me responder 200 — mockProfile faz
// exatamente isso. AppearanceProvider (App.tsx) e AuthContext chamam esse
// endpoint em toda página protegida, não só nas telas que mostram o
// perfil — sem mockar isso, testes tomam 401 real do core-api e caem na
// tela de "não foi possível confirmar sua identidade" no meio do teste.
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
