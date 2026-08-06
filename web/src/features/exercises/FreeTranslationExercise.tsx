import { useState } from "react";
import { TextResultView } from "./TextResultView";
import { submitTextAttempt, type Exercise, type TextAttemptResult } from "./api";
import { ApiError } from "../../lib/apiClient";

interface FreeTranslationExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "free_translation" (Frame 13) — traduz livremente pro
// idioma-alvo do exercício, sem opções pré-definidas. "Correto" pro
// rodapé compartilhado é verdict === PASS, igual dictation/audio.
export function FreeTranslationExercise({ exercise, onAnswered }: FreeTranslationExerciseProps) {
  const [translation, setTranslation] = useState("");
  const [result, setResult] = useState<TextAttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(): Promise<void> {
    if (!translation.trim() || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const attemptResult = await submitTextAttempt(exercise.id, translation);
      setResult(attemptResult);
      onAnswered(attemptResult.verdict === "PASS");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar resposta");
    } finally {
      setSubmitting(false);
    }
  }

  if (result) {
    return <TextResultView result={result} language={exercise.language || "ja-JP"} />;
  }

  const targetIsJapanese = exercise.language?.startsWith("ja");

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">Traduza para o {targetIsJapanese ? "japonês" : "inglês"}</span>
        <span className="exercise-prompt-main">"{exercise.prompt_pt}"</span>
      </div>

      <div style={{ padding: "26px 20px", display: "flex", flexDirection: "column", gap: 12 }}>
        <textarea
          className={targetIsJapanese ? "jp text-input-area" : "text-input-area"}
          placeholder={targetIsJapanese ? "Digite em japonês..." : "Digite em inglês..."}
          value={translation}
          onChange={(e) => setTranslation(e.target.value)}
          disabled={submitting}
          rows={4}
        />
        <button type="button" className="btn-primary" onClick={() => void handleSubmit()} disabled={submitting || !translation.trim()}>
          {submitting ? "Enviando..." : "Enviar"}
        </button>
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}
