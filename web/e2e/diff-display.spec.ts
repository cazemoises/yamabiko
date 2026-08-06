import { test, expect } from "@playwright/test";
import { API_BASE_URL, mockProfile, seedTokensOnce } from "./helpers";

const EXERCISE = {
  id: "ex-1",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Diga que o dia está bonito",
  expected_transcript: "わた",
  sprint_day_ref: 1,
};

// Par observado em produção (BUILD_STATE.md): esperado「わ」, transcrito「お」 —
// não bate em nenhum padrão fonético conhecido, então o backend classifica como
// OUTRO. É o pior caso pra UX: sem highlight/romaji/explicação, sobra só um rótulo
// técnico ilegível pra quem não lê japonês fluente. た casa em ambos (match), pra
// cobrir que o romaji dos caracteres corretos também aparece, só que discreto.
const ATTEMPT_RESULT = {
  transcript: "おた",
  score: 0.5,
  verdict: "PARTIAL",
  diff: [{ op: "SUBSTITUTE", position: 0, expected: "わ", actual: "お", pattern: "OUTRO" }],
  xp_gained: 0,
};

test("resultado do desafio destaca a divergência, mostra romaji e explica em português em vez de rótulos técnicos", async ({
  page,
}) => {
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE.id}`, async (route) => {
    await route.fulfill({ status: 200, json: EXERCISE });
  });

  await page.route(`${API_BASE_URL}/exercises/${EXERCISE.id}/attempts`, async (route) => {
    await route.fulfill({ status: 200, json: ATTEMPT_RESULT });
  });

  await mockProfile(page);
  await seedTokensOnce(page, "valid-token", "valid-refresh-token");

  await page.goto(`/exercises/${EXERCISE.id}`);
  await expect(page.getByText(EXERCISE.prompt_pt)).toBeVisible();

  await page.getByRole("button", { name: /Gravar/ }).click();
  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await page.getByRole("button", { name: "Enviar" }).click();

  const expectedRow = page.getByTestId("diff-row-expected");
  const actualRow = page.getByTestId("diff-row-actual");
  await expect(expectedRow).toBeVisible();
  await expect(actualRow).toBeVisible();

  // Highlight visual no caractere divergente, dos dois lados.
  await expect(expectedRow.locator(".diff-char-mismatch")).toHaveText(/わ/);
  await expect(actualRow.locator(".diff-char-mismatch")).toHaveText(/お/);

  // Romaji do trecho divergente, destacado.
  await expect(expectedRow.locator(".diff-char-romaji-mismatch")).toHaveText("wa");
  await expect(actualRow.locator(".diff-char-romaji-mismatch")).toHaveText("o");

  // Romaji também aparece nos caracteres que bateram certo (た em ambas as
  // linhas), só que discreto — sem a classe de destaque dos divergentes.
  const expectedNeutralRomaji = expectedRow.locator(".diff-char-romaji:not(.diff-char-romaji-mismatch)");
  const actualNeutralRomaji = actualRow.locator(".diff-char-romaji:not(.diff-char-romaji-mismatch)");
  await expect(expectedNeutralRomaji).toHaveText("ta");
  await expect(actualNeutralRomaji).toHaveText("ta");

  // Explicação em português voltada ao aluno, não o rótulo técnico cru.
  const explanations = page.getByTestId("diff-explanations");
  await expect(explanations).toContainText('Você disse "o"');
  await expect(explanations).toContainText('"wa"');
  await expect(explanations).toContainText("confusão comum entre esses sons");
  await expect(page.getByText("SUBSTITUTE", { exact: false })).toHaveCount(0);
  await expect(page.getByText(/\bOUTRO\b/)).toHaveCount(0);
});
