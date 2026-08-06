import { useState } from "react";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { TextResultView } from "./TextResultView";
import { submitTextAttempt, type Exercise, type TextAttemptResult } from "./api";
import { ApiError } from "../../lib/apiClient";

interface DictationExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "dictation" (Frame 12) — ouve o áudio de referência
// (mesmo endpoint/áudio de audio_pronunciation) e digita o que ouviu, sem
// gravar nada. "Correto" pro rodapé compartilhado é verdict === PASS,
// igual audio_pronunciation.
export function DictationExercise({ exercise, onAnswered }: DictationExerciseProps) {
  const [transcript, setTranscript] = useState("");
  const [result, setResult] = useState<TextAttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(): Promise<void> {
    if (!transcript.trim() || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      const attemptResult = await submitTextAttempt(exercise.id, transcript);
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

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">{exercise.prompt_pt}</span>
      </div>

      <div className="record-stage">
        <p style={{ fontSize: 14, color: "var(--text-secondary)", textAlign: "center" }}>Ouça e digite o que você ouvir</p>
        <SpeakButtonAsDictationPlay exerciseId={exercise.id} />
        <textarea
          className="text-input-area"
          style={{ maxWidth: 300 }}
          placeholder="Digite o que você ouviu..."
          value={transcript}
          onChange={(e) => setTranscript(e.target.value)}
          disabled={submitting}
          rows={2}
        />
        <button type="button" className="btn-primary" style={{ maxWidth: 300 }} onClick={() => void handleSubmit()} disabled={submitting || !transcript.trim()}>
          {submitting ? "Enviando..." : "Enviar"}
        </button>
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}

// SpeakButton reaproveitado (mesmo GET /exercises/{id}/reference-audio),
// só o visual muda pro círculo grande do Frame 12 em vez do botão
// retangular padrão (baseClassName troca a classe base inteira).
function SpeakButtonAsDictationPlay({ exerciseId }: { exerciseId: string }) {
  return <SpeakButton exerciseId={exerciseId} label="▶" baseClassName="dictation-play-button" />;
}
