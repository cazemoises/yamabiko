import { test, expect } from "@playwright/test";
import { API_BASE_URL, mockProfile, seedTokensOnce } from "./helpers";

const JA_EXERCISE = {
  id: "ex-ja",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Cumprimente alguém de manhã",
  expected_transcript: "おはよう",
  sprint_day_ref: 1,
  language: "ja-JP",
};

const EN_EXERCISE = {
  id: "ex-en",
  category: "saudacoes",
  difficulty: 1,
  prompt_pt: "Cumprimente alguém casualmente",
  expected_transcript: "hi, how are you?",
  sprint_day_ref: 1,
  language: "en-US",
};

test("toggle JA/EN filtra a lista por idioma e reflete no botão de pronúncia e no diff", async ({ page }) => {
  await page.route(`${API_BASE_URL}/exercises*`, async (route) => {
    const url = new URL(route.request().url());
    const language = url.searchParams.get("language");
    if (language === "en-US") {
      await route.fulfill({ status: 200, json: [EN_EXERCISE] });
      return;
    }
    await route.fulfill({ status: 200, json: [JA_EXERCISE] });
  });

  await mockProfile(page);
  await seedTokensOnce(page, "valid-token", "valid-refresh-token");

  await page.goto("/exercises");

  // Default é japonês.
  await expect(page.getByText(JA_EXERCISE.prompt_pt)).toBeVisible();
  await expect(page.getByText(EN_EXERCISE.prompt_pt)).not.toBeVisible();

  await page.getByRole("button", { name: "EN" }).click();

  await expect(page.getByText(EN_EXERCISE.prompt_pt)).toBeVisible();
  await expect(page.getByText(JA_EXERCISE.prompt_pt)).not.toBeVisible();
});

test("exercício em inglês mostra o botão de pronúncia por áudio real e não duplica romaji no diff", async ({ page }) => {
  await page.route(`${API_BASE_URL}/exercises/${EN_EXERCISE.id}`, async (route) => {
    await route.fulfill({ status: 200, json: EN_EXERCISE });
  });

  await page.route(`${API_BASE_URL}/exercises/${EN_EXERCISE.id}/attempts`, async (route) => {
    await route.fulfill({
      status: 200,
      json: {
        transcript: "hi, how are u?",
        score: 0.8,
        verdict: "PARTIAL",
        diff: [
          { op: "SUBSTITUTE", position: 8, expected: "y", actual: "u", pattern: "OUTRO" },
        ],
        xp_gained: 5,
      },
    });
  });

  await mockProfile(page);
  await seedTokensOnce(page, "valid-token", "valid-refresh-token");

  await page.goto(`/exercises/${EN_EXERCISE.id}`);
  await expect(page.getByText(EN_EXERCISE.prompt_pt)).toBeVisible();
  await expect(page.locator(".speak-button").first()).toBeVisible();

  await page.getByRole("button", { name: /Gravar/ }).click();
  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await page.getByRole("button", { name: "Enviar" }).click();

  const expectedRow = page.getByTestId("diff-row-expected");
  await expect(expectedRow).toBeVisible();

  // Sem romaji embaixo dos caracteres em inglês (não faz sentido pra Latin script).
  await expect(page.locator(".diff-char-romaji")).toHaveCount(0);
});
