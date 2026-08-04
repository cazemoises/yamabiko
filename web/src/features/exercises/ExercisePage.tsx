import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getExercise, getScenario, submitAttempt, type AttemptResult, type Exercise, type ScenarioDetail } from "./api";
import { AudioRecorder } from "../../components/audio/AudioRecorder";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { DiffComparison } from "./DiffComparison";
import { ApiError } from "../../lib/apiClient";

export function ExercisePage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [scenario, setScenario] = useState<ScenarioDetail | null>(null);
  const [result, setResult] = useState<AttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setResult(null);
    setError(null);
    getExercise(id)
      .then(setExercise)
      .catch(() => setLoadError("Exercício não encontrado"));
  }, [id]);

  // Só refaz a busca do cenário quando o exercício muda de cenário (não a
  // cada exercício dentro do mesmo cenário) — os detalhes (contexto, lista
  // ordenada) já foram carregados na 1ª vez, reaproveitados pro resto da
  // sequência sem outra requisição, o que mantém o "Próximo" leve.
  useEffect(() => {
    if (!exercise?.scenario_id) {
      setScenario(null);
      return;
    }
    getScenario(exercise.scenario_id).catch(() => null).then((s) => s && setScenario(s));
  }, [exercise?.scenario_id]);

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

  const scenarioExercises = scenario?.exercises ?? [];
  const currentIndex = scenario ? scenarioExercises.findIndex((e) => e.id === exercise.id) : -1;
  const inScenario = scenario !== null && currentIndex >= 0;
  const nextExercise = inScenario ? scenarioExercises[currentIndex + 1] : undefined;
  const passedInScenario = inScenario && result?.verdict === "PASS";

  function handleNext(): void {
    if (!nextExercise) return;
    navigate(`/exercises/${nextExercise.id}`);
  }

  return (
    <div className="exercise-page">
      <Link to="/exercises">&larr; Voltar</Link>

      {inScenario && (
        <div className="scenario-banner">
          <p className="scenario-context">{scenario.context_description_pt}</p>
          <div
            className="scenario-progress"
            role="progressbar"
            aria-label={`Progresso do cenário ${scenario.title_pt}`}
            aria-valuemin={1}
            aria-valuemax={scenarioExercises.length}
            aria-valuenow={currentIndex + 1}
          >
            <div className="scenario-progress-bar">
              <div
                className="scenario-progress-fill"
                style={{ width: `${((currentIndex + 1) / scenarioExercises.length) * 100}%` }}
              />
            </div>
            <span className="scenario-progress-label">
              {currentIndex + 1} de {scenarioExercises.length}
            </span>
          </div>
        </div>
      )}

      <h1>{exercise.prompt_pt}</h1>
      <p className="expected-transcript">{exercise.expected_transcript}</p>
      {exercise.expected_romaji && <p className="expected-romaji">{exercise.expected_romaji}</p>}
      <SpeakButton exerciseId={exercise.id} />

      {!passedInScenario && <AudioRecorder key={exercise.id} onRecorded={handleRecorded} disabled={submitting} />}

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
              language={exercise.language || "ja-JP"}
              exerciseId={exercise.id}
            />
          )}

          {passedInScenario &&
            (nextExercise ? (
              <button type="button" className="scenario-next-button" onClick={handleNext}>
                Próximo &rarr;
              </button>
            ) : (
              <div className="scenario-complete">
                <p>🎉 Cenário concluído!</p>
                <Link to="/exercises">Voltar aos exercícios</Link>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
