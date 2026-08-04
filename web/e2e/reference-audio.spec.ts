import { test, expect } from "@playwright/test";
import { API_BASE_URL, seedTokensOnce } from "./helpers";

const JA_EXERCISE = {
  id: "ex-ja-ref-audio",
  category: "saudacao",
  difficulty: 1,
  prompt_pt: "Cumprimente alguém de manhã",
  expected_transcript: "おはよう",
  sprint_day_ref: 1,
  language: "ja-JP",
};

// WAV mínimo (mono, 8-bit PCM, 8kHz) com silêncio — precisa ser um arquivo de
// verdade (não bytes arbitrários) pra <audio>.play() não rejeitar por erro de
// decodificação no Chromium headless.
function silentWav(sampleCount = 400): Buffer {
  const dataSize = sampleCount;
  const buffer = Buffer.alloc(44 + dataSize);
  buffer.write("RIFF", 0, "ascii");
  buffer.writeUInt32LE(36 + dataSize, 4);
  buffer.write("WAVE", 8, "ascii");
  buffer.write("fmt ", 12, "ascii");
  buffer.writeUInt32LE(16, 16); // Subchunk1Size (PCM)
  buffer.writeUInt16LE(1, 20); // AudioFormat = PCM
  buffer.writeUInt16LE(1, 22); // NumChannels
  buffer.writeUInt32LE(8000, 24); // SampleRate
  buffer.writeUInt32LE(8000, 28); // ByteRate
  buffer.writeUInt16LE(1, 32); // BlockAlign
  buffer.writeUInt16LE(8, 34); // BitsPerSample
  buffer.write("data", 36, "ascii");
  buffer.writeUInt32LE(dataSize, 40);
  buffer.fill(128, 44); // silêncio (PCM 8-bit unsigned, 128 = zero)
  return buffer;
}

test("exercício ja-JP toca o áudio de referência real do VOICEVOX via <audio>, testável em headless", async ({
  page,
}) => {
  let referenceAudioCalls = 0;

  await page.route(`${API_BASE_URL}/exercises/${JA_EXERCISE.id}`, async (route) => {
    await route.fulfill({ status: 200, json: JA_EXERCISE });
  });

  await page.route(`${API_BASE_URL}/exercises/${JA_EXERCISE.id}/reference-audio`, async (route) => {
    referenceAudioCalls++;
    await route.fulfill({ status: 200, body: silentWav(), headers: { "Content-Type": "audio/wav" } });
  });

  await seedTokensOnce(page, "valid-token", "valid-refresh-token");

  await page.goto(`/exercises/${JA_EXERCISE.id}`);
  await expect(page.getByText(JA_EXERCISE.prompt_pt)).toBeVisible();

  const speakButton = page.getByRole("button", { name: /Ouvir pronúncia esperada/ });
  await expect(speakButton).toBeEnabled();
  await speakButton.click();

  await expect.poll(() => referenceAudioCalls).toBe(1);

  const audioEl = page.getByTestId("reference-audio");
  await expect(audioEl).toHaveAttribute("src", /^blob:/);

  // Diferente da Web Speech API (não roda em Chromium headless, BUILD_STATE.md),
  // este é um <audio> real com um WAV real — dá pra confirmar que play() não
  // ficou pausado/rejeitado, sem depender de vozes instaladas no SO.
  await expect
    .poll(() => audioEl.evaluate((el: HTMLAudioElement) => el.paused))
    .toBe(false);
});

test("exercício en-US não chama reference-audio — continua usando Web Speech API", async ({ page }) => {
  const enExercise = { ...JA_EXERCISE, id: "ex-en-ref-audio", language: "en-US", expected_transcript: "hi" };
  let referenceAudioCalls = 0;

  await page.route(`${API_BASE_URL}/exercises/${enExercise.id}`, async (route) => {
    await route.fulfill({ status: 200, json: enExercise });
  });
  await page.route(`${API_BASE_URL}/exercises/${enExercise.id}/reference-audio`, async (route) => {
    referenceAudioCalls++;
    await route.fulfill({ status: 404, json: { error: "não deveria ser chamado" } });
  });

  await seedTokensOnce(page, "valid-token", "valid-refresh-token");

  await page.goto(`/exercises/${enExercise.id}`);
  await expect(page.getByText(enExercise.prompt_pt)).toBeVisible();

  await expect(page.getByTestId("reference-audio")).toHaveCount(0);
  expect(referenceAudioCalls).toBe(0);
});
