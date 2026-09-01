import { test, expect, type Page } from "@playwright/test";
import { API_BASE_URL, mockProfile } from "./helpers";

// Cobre pelo menos 1 exercício de cada exercise_type ponta a ponta
// (pergunta -> resposta -> resultado real), rodando o MESMO conjunto de
// casos em mobile (390x844) e desktop (1440x900) — layout responsivo do
// conteúdo muda entre as duas larguras, a lógica de pergunta/resposta/
// resultado não. Navegação própria (sidebar/bottom nav) foi removida:
// vive na sidebar do Portal agora (ver components/layout/AppShell.tsx e
// usePortalShellNav.ts). Todos os exercícios de teste são standalone (sem
// scenario_id): o fluxo de cenário (contexto + "Próximo") já tem
// cobertura própria em scenario-flow.spec.ts, aqui o foco é validar cada
// tipo por si.

const MOBILE_VIEWPORT = { width: 390, height: 844 };
const DESKTOP_VIEWPORT = { width: 1440, height: 900 };

interface TypeCase {
  name: string;
  exercise: Record<string, unknown>;
  endpoint: "attempts" | "answer" | "text-attempt";
  response: Record<string, unknown>;
  interact: (page: Page) => Promise<void>;
  assertResult: (page: Page) => Promise<void>;
  /** Texto pra confirmar que a tela carregou antes de interagir — default é
   * prompt_pt, mas alguns tipos (ex: true_false) não renderizam prompt_pt
   * na tela (Frame 15 só mostra a statement), então usam outro texto. */
  initialText?: string;
}

const CASES: TypeCase[] = [
  {
    name: "audio_pronunciation",
    exercise: {
      id: "type-audio-1",
      category: "saudacao",
      difficulty: 1,
      prompt_pt: "Cumprimente dizendo bom dia.",
      expected_transcript: "おはよう",
      sprint_day_ref: 1,
      language: "ja-JP",
      exercise_type: "audio_pronunciation",
    },
    endpoint: "attempts",
    response: { transcript: "おはよう", score: 1, verdict: "PASS", diff: [], xp_gained: 10 },
    interact: async (page) => {
      await page.getByRole("button", { name: /Gravar/ }).click();
      await page.getByRole("button", { name: /Parar gravação/ }).click();
      await page.getByRole("button", { name: "Enviar" }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".verdict-pill-pass")).toBeVisible();
    },
  },
  {
    name: "multiple_choice_translation",
    exercise: {
      id: "type-mc-1",
      category: "compras",
      difficulty: 1,
      prompt_pt: "Qual frase significa: \"Bom dia\"?",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "multiple_choice_translation",
      type_data: { options: ["Good morning", "Good night", "Good afternoon"], correct_index: 0 },
    },
    endpoint: "answer",
    response: { correct: true, correct_index: 0 },
    interact: async (page) => {
      await page.getByRole("button", { name: /^Good morning/ }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".option-button.correct")).toBeVisible();
    },
  },
  {
    name: "word_order",
    exercise: {
      id: "type-wo-1",
      category: "trabalho",
      difficulty: 1,
      prompt_pt: "Organize a frase",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "word_order",
      type_data: { shuffled_words: ["is", "This"], correct_order: ["This", "is"] },
    },
    endpoint: "answer",
    response: { correct: true, correct_order: ["This", "is"] },
    interact: async (page) => {
      await page.getByRole("button", { name: "This", exact: true }).click();
      await page.getByRole("button", { name: "is", exact: true }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".word-chip-correct").first()).toBeVisible();
    },
  },
  {
    name: "verb_conjugation",
    exercise: {
      id: "type-vc-1",
      category: "trabalho",
      difficulty: 1,
      prompt_pt: "Complete a frase",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "verb_conjugation",
      type_data: {
        sentence_template: "I ___ happy.",
        verb_infinitive: "be",
        options: ["am", "is", "are"],
        correct_index: 0,
      },
    },
    endpoint: "answer",
    response: { correct: true, correct_index: 0 },
    interact: async (page) => {
      await page.getByRole("button", { name: "am", exact: true }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".option-button.correct")).toBeVisible();
    },
  },
  {
    name: "dictation",
    exercise: {
      id: "type-dt-1",
      category: "saudacao",
      difficulty: 1,
      prompt_pt: "Repita o cumprimento que ouvir",
      expected_transcript: "おはよう",
      sprint_day_ref: 1,
      language: "ja-JP",
      exercise_type: "dictation",
    },
    endpoint: "text-attempt",
    response: { transcript: "おはよう", expected: "おはよう", score: 1, verdict: "PASS", diff: [] },
    interact: async (page) => {
      await page.locator("textarea").fill("おはよう");
      await page.getByRole("button", { name: "Enviar" }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".verdict-pill-pass")).toBeVisible();
    },
  },
  {
    name: "free_translation",
    exercise: {
      id: "type-ft-1",
      category: "compras",
      difficulty: 1,
      prompt_pt: "Como você está?",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "free_translation",
      type_data: { acceptable_answers: ["I am fine"] },
    },
    endpoint: "text-attempt",
    response: { transcript: "I am fine", expected: "I am fine", score: 1, verdict: "PASS", diff: [] },
    interact: async (page) => {
      await page.locator("textarea").fill("I am fine");
      await page.getByRole("button", { name: "Enviar" }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".verdict-pill-pass")).toBeVisible();
    },
  },
  {
    name: "matching_pairs",
    exercise: {
      id: "type-mp-1",
      category: "saudacao",
      difficulty: 1,
      prompt_pt: "Combine os pares",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "matching_pairs",
      type_data: {
        pairs: [
          { left: "Bom dia", right: "Good morning" },
          { left: "Boa noite", right: "Good night" },
        ],
      },
    },
    endpoint: "answer",
    response: {
      correct: true,
      correct_pairs: [
        { left: "Bom dia", right: "Good morning" },
        { left: "Boa noite", right: "Good night" },
      ],
    },
    interact: async (page) => {
      await page.getByText("Bom dia", { exact: true }).click();
      await page.getByText("Good morning", { exact: true }).click();
      await page.getByText("Boa noite", { exact: true }).click();
      await page.getByText("Good night", { exact: true }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".matching-item.matched")).toHaveCount(4);
    },
  },
  {
    name: "true_false",
    exercise: {
      id: "type-tf-1",
      category: "trabalho",
      difficulty: 1,
      prompt_pt: "Verdadeiro ou falso",
      expected_transcript: "",
      sprint_day_ref: 1,
      language: "en-US",
      exercise_type: "true_false",
      type_data: { statement: "2 + 2 = 4", correct_answer: true },
    },
    endpoint: "answer",
    response: { correct: true, correct_answer: true },
    interact: async (page) => {
      await page.getByRole("button", { name: "Verdadeiro" }).click();
    },
    assertResult: async (page) => {
      await expect(page.locator(".true-false-button.correct")).toBeVisible();
    },
    initialText: "2 + 2 = 4",
  },
];

function registerCases(viewportLabel: string): void {
  for (const testCase of CASES) {
    test(`${testCase.name}: pergunta -> resposta -> resultado real (${viewportLabel})`, async ({ page }) => {
      const exerciseId = testCase.exercise.id as string;

      await page.route(`${API_BASE_URL}/exercises/${exerciseId}`, async (route) => {
        await route.fulfill({ status: 200, json: testCase.exercise });
      });
      await page.route(`${API_BASE_URL}/exercises/${exerciseId}/${testCase.endpoint}`, async (route) => {
        await route.fulfill({ status: 200, json: testCase.response });
      });

      await mockProfile(page);

      await page.goto(`/exercises/${exerciseId}`);
      await expect(page.getByText(testCase.initialText ?? (testCase.exercise.prompt_pt as string))).toBeVisible();

      await testCase.interact(page);
      await testCase.assertResult(page);
    });
  }
}

test.describe("mobile (390x844)", () => {
  test.use({ viewport: MOBILE_VIEWPORT });

  test("não renderiza navegação própria — vive só na sidebar do Portal", async ({ page }) => {
    await mockProfile(page);
    await page.goto("/");
    await expect(page.locator(".bottom-nav")).toHaveCount(0);
    await expect(page.locator(".sidebar-nav")).toHaveCount(0);
  });

  registerCases("mobile");
});

test.describe("desktop (1440x900)", () => {
  test.use({ viewport: DESKTOP_VIEWPORT });

  test("não renderiza navegação própria — vive só na sidebar do Portal", async ({ page }) => {
    await mockProfile(page);
    await page.goto("/");
    await expect(page.locator(".bottom-nav")).toHaveCount(0);
    await expect(page.locator(".sidebar-nav")).toHaveCount(0);
  });

  registerCases("desktop");
});
