import { test, expect } from "@playwright/test";
import { API_BASE_URL, mockProfile } from "./helpers";

const SCENARIO_ID = "scn-1";

const EXERCISE_1 = {
  id: "scn-1-ex-1",
  category: "cenario_trabalho_manha",
  difficulty: 1,
  prompt_pt: "Cumprimente dizendo bom dia.",
  expected_transcript: "おはようございます",
  sprint_day_ref: 51,
  language: "ja-JP",
  scenario_id: SCENARIO_ID,
  order_in_scenario: 1,
};

const EXERCISE_2 = {
  id: "scn-1-ex-2",
  category: "cenario_trabalho_manha",
  difficulty: 2,
  prompt_pt: "Pergunte de volta como ele está.",
  expected_transcript: "おげんきですか",
  sprint_day_ref: 51,
  language: "ja-JP",
  scenario_id: SCENARIO_ID,
  order_in_scenario: 2,
};

const SCENARIO_DETAIL = {
  id: SCENARIO_ID,
  language: "ja-JP",
  title_pt: "Cumprimentar um colega no trabalho de manhã",
  context_description_pt: "Você chega no escritório e cruza com um colega. É de manhã, ambiente informal mas profissional.",
  order_index: 1,
  exercises: [EXERCISE_1, EXERCISE_2],
};

function passResult(transcript: string) {
  return { transcript, score: 1.0, verdict: "PASS", diff: [], xp_gained: 10 };
}

async function recordAndSubmit(page: import("@playwright/test").Page): Promise<void> {
  await page.getByRole("button", { name: /Gravar/ }).click();
  await page.getByRole("button", { name: /Parar gravação/ }).click();
  await page.getByRole("button", { name: "Enviar" }).click();
}

test("percorre um cenário completo do início ao fim sem voltar pra lista entre um exercício e outro", async ({
  page,
}) => {
  await page.route(`${API_BASE_URL}/scenarios/${SCENARIO_ID}`, async (route) => {
    await route.fulfill({ status: 200, json: SCENARIO_DETAIL });
  });
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE_1.id}`, async (route) => {
    await route.fulfill({ status: 200, json: EXERCISE_1 });
  });
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE_2.id}`, async (route) => {
    await route.fulfill({ status: 200, json: EXERCISE_2 });
  });
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE_1.id}/attempts`, async (route) => {
    await route.fulfill({ status: 200, json: passResult(EXERCISE_1.expected_transcript) });
  });
  await page.route(`${API_BASE_URL}/exercises/${EXERCISE_2.id}/attempts`, async (route) => {
    await route.fulfill({ status: 200, json: passResult(EXERCISE_2.expected_transcript) });
  });

  await mockProfile(page);

  await page.goto(`/exercises/${EXERCISE_1.id}`);

  // Contexto do cenário e progresso "1 de 2" no 1º exercício.
  await expect(page.getByText(SCENARIO_DETAIL.context_description_pt)).toBeVisible();
  await expect(page.getByText("1 de 2")).toBeVisible();
  await expect(page.getByText(EXERCISE_1.prompt_pt)).toBeVisible();

  await recordAndSubmit(page);

  // PASS no 1º: botão principal vira "Próximo" — sem link de volta pra lista.
  const nextButton = page.getByRole("button", { name: /Próximo/ });
  await expect(nextButton).toBeVisible();
  await expect(page.getByText("🎉 Cenário concluído!")).not.toBeVisible();

  await nextButton.click();

  // Navegou pro 2º exercício do cenário (client-side, sem passar pela lista):
  // contexto continua visível, progresso avançou, prompt trocou.
  await expect(page).toHaveURL(new RegExp(`/exercises/${EXERCISE_2.id}$`));
  await expect(page.getByText(SCENARIO_DETAIL.context_description_pt)).toBeVisible();
  await expect(page.getByText("2 de 2")).toBeVisible();
  await expect(page.getByText(EXERCISE_2.prompt_pt)).toBeVisible();

  await recordAndSubmit(page);

  // PASS no último exercício: cenário concluído, sem botão "Próximo".
  await expect(page.getByText("🎉 Cenário concluído!")).toBeVisible();
  await expect(page.getByRole("button", { name: /Próximo/ })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Voltar aos exercícios" })).toBeVisible();
});
