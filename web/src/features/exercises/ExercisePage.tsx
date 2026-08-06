import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getExercise, getScenario, submitAttempt, type AttemptResult, type Exercise, type ScenarioDetail } from "./api";
import { AudioRecorder } from "../../components/audio/AudioRecorder";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { AudioResultView } from "./AudioResultView";
import { TopBar } from "../../components/layout/TopBar";
import { ApiError } from "../../lib/apiClient";

// Dispatcher por exercise_type (Frames 4/5 = audio_pronunciation, Frames
// 9-19 = os 7 tipos novos da Fase A) — só audio_pronunciation está
// implementado nesta tela por enquanto; os outros entram em commits
// próprios seguintes, um tipo por vez.
export function ExercisePage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [scenario, setScenario] = useState<ScenarioDetail | null>(null);
  const [result, setResult] = useState<AttemptResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [attemptNumber, setAttemptNumber] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setResult(null);
    setError(null);
    setAttemptNumber(0);
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

  function handleRetry(): void {
    setResult(null);
    setError(null);
    setAttemptNumber((n) => n + 1);
  }

  if (loadError) return <p className="error center-message">{loadError}</p>;
  if (!exercise) return <p className="center-message">Carregando...</p>;

  const scenarioExercises = scenario?.exercises ?? [];
  const currentIndex = scenario ? scenarioExercises.findIndex((e) => e.id === exercise.id) : -1;
  const inScenario = scenario !== null && currentIndex >= 0;
  const nextExercise = inScenario ? scenarioExercises[currentIndex + 1] : undefined;
  const passedInScenario = inScenario && result?.verdict === "PASS";

  function handleNext(): void {
    if (!nextExercise) return;
    navigate(`/exercises/${nextExercise.id}`);
  }

  if (exercise.exercise_type && exercise.exercise_type !== "audio_pronunciation") {
    return (
      <div>
        <TopBar backTo="/exercises" />
        <p className="center-message">
          Tipo de exercício "{exercise.exercise_type}" ainda não tem tela própria — chega em breve.
        </p>
      </div>
    );
  }

  return (
    <div>
      <TopBar
        progress={inScenario ? { current: currentIndex + 1, total: scenarioExercises.length } : undefined}
        backTo={inScenario ? undefined : "/exercises"}
      />

      {inScenario && !result && (
        <div className="context-banner">
          <span className="context-banner-body">{scenario.context_description_pt}</span>
        </div>
      )}

      {!result && (
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
            <AudioRecorder key={attemptNumber} autoStart={attemptNumber > 0} onRecorded={handleRecorded} disabled={submitting} />
          </div>
        </>
      )}

      {submitting && <p className="center-message">Enviando e transcrevendo...</p>}
      {error && <p className="error">{error}</p>}

      {result && (
        <>
          <AudioResultView exercise={exercise} result={result} />
          <div className="btn-block-group">
            {passedInScenario ? (
              nextExercise ? (
                <button type="button" className="btn-primary" onClick={handleNext}>
                  Próximo →
                </button>
              ) : (
                <div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 6 }}>
                  <p style={{ fontWeight: 700, color: "var(--text)" }}>🎉 Cenário concluído!</p>
                  <Link to="/exercises" style={{ fontSize: 13 }}>
                    Voltar aos exercícios
                  </Link>
                </div>
              )
            ) : (
              <button type="button" className="btn-primary" onClick={handleRetry}>
                Tentar de novo
              </button>
            )}
            <SpeakButton exerciseId={exercise.id} label="Ouvir pronúncia esperada" />
          </div>
        </>
      )}
    </div>
  );
}
