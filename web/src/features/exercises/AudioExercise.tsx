import { useState } from "react";
import { AudioRecorder } from "../../components/audio/AudioRecorder";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { AudioResultView } from "./AudioResultView";
import { submitAttempt, type AttemptResult, type Exercise } from "./api";
import { ApiError } from "../../lib/apiClient";

interface AudioExerciseProps {
  exercise: Exercise;
  autoStart: boolean;
  onAnswered: (correct: boolean, xpGained: number) => void;
}

// exercise_type "audio_pronunciation" (Frames 4/5) — único tipo que grava
// áudio de verdade (stt-service via /attempts). "Correto" pro rodapé
// compartilhado de ExercisePage é verdict === PASS (PARTIAL/FAIL contam
// como errado, mostram "Tentar de novo").
export function AudioExercise({ exercise, autoStart, onAnswered }: AudioExerciseProps) {
  const [result, setResult] = useState<AttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleRecorded(blob: Blob): Promise<void> {
    setSubmitting(true);
    setError(null);
    try {
      const attemptResult = await submitAttempt(exercise.id, blob);
      setResult(attemptResult);
      onAnswered(attemptResult.verdict === "PASS", attemptResult.xp_gained);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar tentativa");
    } finally {
      setSubmitting(false);
    }
  }

  if (result) {
    return <AudioResultView exercise={exercise} result={result} />;
  }

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">{exercise.prompt_pt}</span>
        <span className={exercise.language?.startsWith("ja") ? "jp exercise-prompt-main jp-lg" : "exercise-prompt-main"}>
          {exercise.expected_transcript}
        </span>
        {exercise.expected_romaji && <span className="exercise-prompt-sub">{exercise.expected_romaji}</span>}
      </div>

      <div className="record-stage">
        <SpeakButton exerciseId={exercise.id} />
        <AudioRecorder autoStart={autoStart} onRecorded={handleRecorded} disabled={submitting} />
      </div>

      {submitting && <p className="center-message">Enviando e transcrevendo...</p>}
      {error && <p className="error">{error}</p>}
    </>
  );
}
