import { test, expect, type Page } from "@playwright/test";

const API_BASE_URL = "http://localhost:9001";

// addInitScript roda em toda navegação da página (inclusive o redirect pra /login que
// o interceptor dispara), então sem essa guarda ele re-semearia os tokens depois do
// logout e mascararia o clearTokens(). sessionStorage sobrevive à navegação (mesma aba),
// diferente de uma variável em memória, então serve de flag "já semeei".
async function seedTokensOnce(page: Page, accessToken: string, refreshToken: string): Promise<void> {
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

const EXERCISE = {
  id: "ex-1",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Cumprimente alguém de manhã",
  expected_transcript: "おはようございます",
  sprint_day_ref: 1,
};

test.describe("interceptor de refresh de token", () => {
  test("token de acesso expirado é renovado automaticamente e a requisição original é repetida sem exigir novo login", async ({
    page,
  }) => {
    let exercisesCalls = 0;
    let refreshCalls = 0;

    await page.route(`${API_BASE_URL}/exercises`, async (route) => {
      const authHeader = route.request().headers()["authorization"];
      exercisesCalls++;
      if (authHeader === "Bearer expired-token") {
        await route.fulfill({ status: 401, json: { error: "token expirado" } });
        return;
      }
      if (authHeader === "Bearer new-token") {
        await route.fulfill({ status: 200, json: [EXERCISE] });
        return;
      }
      await route.fulfill({ status: 401, json: { error: "token inesperado" } });
    });

    await page.route(`${API_BASE_URL}/auth/refresh`, async (route) => {
      refreshCalls++;
      const body = route.request().postDataJSON() as { refresh_token: string };
      if (body.refresh_token !== "valid-refresh-token") {
        await route.fulfill({ status: 401, json: { error: "refresh token inválido" } });
        return;
      }
      await route.fulfill({ status: 200, json: { access_token: "new-token" } });
    });

    await seedTokensOnce(page, "expired-token", "valid-refresh-token");

    await page.goto("/exercises");

    await expect(page.getByText(EXERCISE.prompt_pt)).toBeVisible();
    await expect(page).toHaveURL(/\/exercises$/);
    await expect(page.getByText("Erro ao carregar exercícios")).not.toBeVisible();

    // React StrictMode invoca o efeito de carregamento duas vezes em dev, então cada
    // tentativa (401 + retry) pode ocorrer duas vezes — o que importa é que cada 401
    // tenha exatamente um retry (sem loop infinito) e que o refresh real ocorra uma
    // única vez, mesmo com duas chamadas concorrentes (dedup via refreshInFlight).
    expect(refreshCalls).toBe(1);
    expect(exercisesCalls % 2).toBe(0);
    expect(exercisesCalls).toBeGreaterThanOrEqual(2);
    await expect
      .poll(() => page.evaluate(() => localStorage.getItem("access_token")))
      .toBe("new-token");
  });

  test("desloga e redireciona para /login quando o refresh token também é inválido", async ({ page }) => {
    await page.route(`${API_BASE_URL}/exercises`, async (route) => {
      await route.fulfill({ status: 401, json: { error: "token expirado" } });
    });

    await page.route(`${API_BASE_URL}/auth/refresh`, async (route) => {
      await route.fulfill({ status: 401, json: { error: "refresh token inválido" } });
    });

    await seedTokensOnce(page, "expired-token", "expired-refresh-token");

    await page.goto("/exercises");

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: /Entrar/ })).toBeVisible();

    await expect
      .poll(() => page.evaluate(() => localStorage.getItem("access_token")))
      .toBeNull();
    expect(await page.evaluate(() => localStorage.getItem("refresh_token"))).toBeNull();
  });
});
