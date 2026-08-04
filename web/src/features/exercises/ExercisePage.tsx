import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getExercise, submitAttempt, type AttemptResult, type Exercise } from "./api";
import { AudioRecorder } from "../../components/audio/AudioRecorder";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { DiffComparison } from "./DiffComparison";
import { ApiError } from "../../lib/apiClient";

export function ExercisePage() {
  const { id } = useParams<{ id: string }>();
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [result, setResult] = useState<AttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    getExercise(id)
      .then(setExercise)
      .catch(() => setLoadError("Exercício não encontrado"));
  }, [id]);

  async function handleRecorded(blob: Blob): Promise<void> {
    if (!id) return;
    setSubmitting(true);
    setError(null);
    try {
      const attemptResult = await submitAttempt(id, blob);
      setResult(attemptResult);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar tentativa");
    } finally {
      setSubmitting(false);
    }
  }

  if (loadError) return <p className="error">{loadError}</p>;
  if (!exercise) return <p>Carregando...</p>;

  return (
    <div className="exercise-page">
      <Link to="/exercises">&larr; Voltar</Link>
      <h1>{exercise.prompt_pt}</h1>
      <p className="expected-transcript">{exercise.expected_transcript}</p>
      {exercise.expected_romaji && <p className="expected-romaji">{exercise.expected_romaji}</p>}
      <SpeakButton text={exercise.expected_transcript} />

      <AudioRecorder onRecorded={handleRecorded} disabled={submitting} />

      {submitting && <p>Enviando e transcrevendo...</p>}
      {error && <p className="error">{error}</p>}

      {result && (
        <div className={`attempt-result verdict-${result.verdict.toLowerCase()}`}>
          <p>Você disse: {result.transcript || "(nada detectado)"}</p>
          <p>Score: {Math.round(result.score * 100)}%</p>
          <p>Veredito: {result.verdict}</p>
          <p>XP ganho: +{result.xp_gained}</p>
          {result.diff.length > 0 && (
            <DiffComparison
              expected={exercise.expected_transcript}
              actual={result.transcript}
              diff={result.diff}
            />
          )}
        </div>
      )}
    </div>
  );
}
