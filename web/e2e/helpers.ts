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
