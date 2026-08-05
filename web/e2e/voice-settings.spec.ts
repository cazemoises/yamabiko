import { test, expect } from "@playwright/test";
import { API_BASE_URL, seedTokensOnce } from "./helpers";

// Mesmo WAV mínimo sintético usado em reference-audio.spec.ts — precisa ser
// um arquivo de verdade (não bytes arbitrários) pra <audio>.play() não
// rejeitar por erro de decodificação no Chromium headless.
function silentWav(sampleCount = 400): Buffer {
  const dataSize = sampleCount;
  const buffer = Buffer.alloc(44 + dataSize);
  buffer.write("RIFF", 0, "ascii");
  buffer.writeUInt32LE(36 + dataSize, 4);
  buffer.write("WAVE", 8, "ascii");
  buffer.write("fmt ", 12, "ascii");
  buffer.writeUInt32LE(16, 16);
  buffer.writeUInt16LE(1, 20);
  buffer.writeUInt16LE(1, 22);
  buffer.writeUInt32LE(8000, 24);
  buffer.writeUInt32LE(8000, 28);
  buffer.writeUInt16LE(1, 32);
  buffer.writeUInt16LE(8, 34);
  buffer.write("data", 36, "ascii");
  buffer.writeUInt32LE(dataSize, 40);
  buffer.fill(128, 44);
  return buffer;
}

const JA_VOICES = [
  { id: "ja-announcer-neutral", name: "Locutor(a) Neutro(a)", language: "ja-JP" },
  { id: "ja-female-natural", name: "Voz Feminina Natural", language: "ja-JP" },
];

const EN_VOICES = [
  { id: "en-lessac", name: "Lessac (voz masculina neutra)", language: "en-US" },
  { id: "en-amy", name: "Amy (voz feminina)", language: "en-US" },
];

test("seletor de voz: mostra a preferência salva, toca preview, e salva nova seleção", async ({ page }) => {
  let patchedBody: unknown = null;
  let previewCalls = 0;

  await page.route(`${API_BASE_URL}/users/me`, async (route) => {
    await route.fulfill({
      status: 200,
      json: {
        id: "user-1",
        email: "aluno@example.com",
        name: "Aluno",
        created_at: "2026-01-01T00:00:00Z",
        current_sprint_day: 1,
        xp_total: 0,
        current_streak_days: 0,
        longest_streak_days: 0,
        badges: [],
        preferred_voice_ja: "ja-announcer-neutral",
      },
    });
  });

  await page.route(`${API_BASE_URL}/tts/voices?language=ja-JP`, async (route) => {
    await route.fulfill({ status: 200, json: JA_VOICES });
  });
  await page.route(`${API_BASE_URL}/tts/voices?language=en-US`, async (route) => {
    await route.fulfill({ status: 200, json: EN_VOICES });
  });

  await page.route(`${API_BASE_URL}/tts/voices/ja-female-natural/preview`, async (route) => {
    previewCalls++;
    await route.fulfill({ status: 200, body: silentWav(), headers: { "Content-Type": "audio/wav" } });
  });

  await page.route(`${API_BASE_URL}/users/me/voice-preference`, async (route) => {
    patchedBody = route.request().postDataJSON();
    await route.fulfill({ status: 204 });
  });

  await seedTokensOnce(page, "valid-token", "valid-refresh-token");
  await page.goto("/settings/voice");

  await expect(page.getByText("Locutor(a) Neutro(a)")).toBeVisible();
  await expect(page.getByText("Voz Feminina Natural")).toBeVisible();

  const neutralRow = page.locator(".voice-row", { hasText: "Locutor(a) Neutro(a)" });
  const naturalRow = page.locator(".voice-row", { hasText: "Voz Feminina Natural" });

  // A voz salva no perfil (preferred_voice_ja) já vem marcada como selecionada.
  await expect(neutralRow.getByRole("button", { name: "✓ Selecionada" })).toBeVisible();
  await expect(naturalRow.getByRole("button", { name: "Selecionar" })).toBeVisible();

  // Ouvir o preview de uma voz ainda não selecionada toca via <audio> real.
  await naturalRow.getByRole("button", { name: "▶ Ouvir" }).click();
  await expect.poll(() => previewCalls).toBe(1);
  const audioEl = page.getByTestId("voice-preview-audio");
  await expect(audioEl).toHaveAttribute("src", /^blob:/);
  await expect.poll(() => audioEl.evaluate((el: HTMLAudioElement) => el.paused)).toBe(false);

  // Selecionar a nova voz dispara o PATCH com language+voice_id corretos e
  // atualiza a marcação de "selecionada" sem precisar recarregar a página.
  await naturalRow.getByRole("button", { name: "Selecionar" }).click();
  await expect.poll(() => patchedBody).toEqual({ language: "ja-JP", voice_id: "ja-female-natural" });
  await expect(naturalRow.getByRole("button", { name: "✓ Selecionada" })).toBeVisible();
  await expect(neutralRow.getByRole("button", { name: "Selecionar" })).toBeVisible();

  // Trocar de idioma troca a lista pro catálogo en-US.
  await page.getByRole("button", { name: "🇺🇸 Inglês" }).click();
  await expect(page.getByText("Lessac (voz masculina neutra)")).toBeVisible();
  await expect(page.getByText("Amy (voz feminina)")).toBeVisible();
  await expect(page.getByText("Locutor(a) Neutro(a)")).not.toBeVisible();
});
