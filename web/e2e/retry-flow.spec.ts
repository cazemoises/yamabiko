import { test, expect } from "@playwright/test";
import { API_BASE_URL, mockProfile, seedTokensOnce } from "./helpers";

const EXERCISE = {
  id: "retry-ex-1",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Cumprimente dizendo bom dia.",
  expected_transcript: "おはよう",
  sprint_day_ref: 1,
  language: "ja-JP",
};

const FAIL_RESULT = { transcript: "ちがう", score: 0.2, verdict: "FAIL", diff: [], xp_gained: 1 };
const PASS_RESULT = { transcript: "おはよう", score: 1.0, verdict: "PASS", diff: [], xp_gained: 10 };

test("erro -> retry -> acerto: 1 clique só entre ver o erro e voltar a gravar, sem re-navegação", async ({
  page,
}) => {
  let exerciseFetchCalls = 0;
  let attemptCalls = 0;

  await page.route(`${API_BASE_URL}/exercises/${EXERCISE.id}`, async (route) => {
    exerciseFetchCalls++;
    await route.fulfill({ status: 200, json: EXERCISE });
  });
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE.id}/attempts`, async (route) => {
    attemptCalls++;
    await route.fulfill({ status: 200, json: attemptCalls === 1 ? FAIL_RESULT : PASS_RESULT });
  });

  await mockProfile(page);
  await seedTokensOnce(page, "valid-token", "valid-refresh-token");
  await page.goto(`/exercises/${EXERCISE.id}`);
  await expect(page.getByText(EXERCISE.prompt_pt)).toBeVisible();

  // 1ª tentativa: grava, envia, erra.
  await page.getByRole("button", { name: /Gravar/ }).click();
  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await page.getByRole("button", { name: "Enviar" }).click();
  await expect(page.locator(".verdict-pill-fail")).toBeVisible();

  const urlAfterFail = page.url();
  // React StrictMode invoca o efeito de carregamento 2x em dev (mesmo padrão
  // documentado em token-refresh.spec.ts) — o que importa não é o valor
  // absoluto, é que o retry não dispare NENHUM fetch novo do exercício.
  const exerciseFetchCallsAfterFail = exerciseFetchCalls;

  // O momento medido: da tela de erro até estar gravando de novo — 1 clique.
  await page.getByRole("button", { name: "Tentar de novo" }).click();
  await expect(page.getByRole("button", { name: /Parar gravação/ })).toBeVisible();

  // Sem re-navegação (mesma URL) e sem re-fetch do exercício (mesma instância
  // de página, sem re-render pesado disparando a busca de novo).
  expect(page.url()).toBe(urlAfterFail);
  expect(exerciseFetchCalls).toBe(exerciseFetchCallsAfterFail);

  // 2ª tentativa: acerta.
  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await page.getByRole("button", { name: "Enviar" }).click();
  await expect(page.locator(".verdict-pill-pass")).toBeVisible();

  expect(attemptCalls).toBe(2);
});
