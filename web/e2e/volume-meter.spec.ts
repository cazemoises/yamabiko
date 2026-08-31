import { test, expect } from "@playwright/test";
import { API_BASE_URL, mockProfile } from "./helpers";

const EXERCISE = {
  id: "ex-1",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Diga que o dia está bonito",
  expected_transcript: "わ",
  sprint_day_ref: 1,
};

test("mostra um medidor de volume reagindo ao áudio captado durante a gravação", async ({ page }) => {
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE.id}`, async (route) => {
    await route.fulfill({ status: 200, json: EXERCISE });
  });

  await mockProfile(page);
  await page.goto(`/exercises/${EXERCISE.id}`);

  const meter = page.getByTestId("volume-meter");
  await expect(meter).toHaveCount(0);

  await page.getByRole("button", { name: /Gravar/ }).click();

  await expect(meter).toBeVisible();
  // Chromium com --use-fake-device-for-media-stream alimenta um tom sintético
  // (não silêncio), então o AnalyserNode deve reportar volume > 0 em tempo real.
  await expect
    .poll(async () => Number(await meter.getAttribute("data-volume")), { timeout: 5000 })
    .toBeGreaterThan(0);

  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await expect(meter).toHaveCount(0);
});
