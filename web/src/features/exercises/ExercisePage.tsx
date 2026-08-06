import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { getExercise, getScenario, type Exercise, type ScenarioDetail } from "./api";
import { SpeakButton } from "../../components/audio/SpeakButton";
import { AudioExercise } from "./AudioExercise";
import { MultipleChoiceExercise } from "./MultipleChoiceExercise";
import { WordOrderExercise } from "./WordOrderExercise";
import { VerbConjugationExercise } from "./VerbConjugationExercise";
import { TrueFalseExercise } from "./TrueFalseExercise";
import { MatchingPairsExercise } from "./MatchingPairsExercise";
import { DictationExercise } from "./DictationExercise";
import { FreeTranslationExercise } from "./FreeTranslationExercise";
import { TopBar } from "../../components/layout/TopBar";

// Shell compartilhado por todo exercise_type (Frames 4-19): topbar,
// contexto de cenário, e o rodapé de navegação (Próximo/Tentar de
// novo/cenário concluído) — o que muda por tipo é só o miolo (pergunta +
// resposta + feedback inline), delegado pro componente de cada tipo via
// onAnswered(correct). "Tentar de novo" remonta o miolo inteiro
// (key={resetKey}), limpando qualquer seleção/estado do tipo específico.
export function ExercisePage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [scenario, setScenario] = useState<ScenarioDetail | null>(null);
  const [answered, setAnswered] = useState<boolean | null>(null);
  const [resetKey, setResetKey] = useState(0);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setAnswered(null);
    setResetKey(0);
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

  function handleAnswered(correct: boolean): void {
    setAnswered(correct);
  }

  function handleRetry(): void {
    setAnswered(null);
    setResetKey((n) => n + 1);
  }

  if (loadError) return <p className="error center-message">{loadError}</p>;
  if (!exercise) return <p className="center-message">Carregando...</p>;

  const scenarioExercises = scenario?.exercises ?? [];
  const currentIndex = scenario ? scenarioExercises.findIndex((e) => e.id === exercise.id) : -1;
  const inScenario = scenario !== null && currentIndex >= 0;
  const nextExercise = inScenario ? scenarioExercises[currentIndex + 1] : undefined;
  const passedInScenario = inScenario && answered === true;

  function handleNext(): void {
    if (!nextExercise) return;
    navigate(`/exercises/${nextExercise.id}`);
  }

  return (
    <div>
      <TopBar
        progress={inScenario ? { current: currentIndex + 1, total: scenarioExercises.length } : undefined}
        backTo={inScenario ? undefined : "/exercises"}
      />

      {inScenario && answered === null && (
        <div className="context-banner">
          <span className="context-banner-body">{scenario.context_description_pt}</span>
        </div>
      )}

      {/* key inclui exercise.id: sem isso, navegar pro próximo exercício do
          cenário (mesmo resetKey=0) não remontaria ExerciseBody, e o estado
          interno do tipo anterior (ex: AudioExercise.result) vazaria pro
          exercício novo. */}
      <ExerciseBody key={`${exercise.id}-${resetKey}`} exercise={exercise} resetKey={resetKey} onAnswered={handleAnswered} />

      {answered !== null && (
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
          ) : !answered ? (
            <button type="button" className="btn-primary" onClick={handleRetry}>
              Tentar de novo
            </button>
          ) : null}

          {exercise.exercise_type === "audio_pronunciation" && (
            <SpeakButton exerciseId={exercise.id} label="Ouvir pronúncia esperada" />
          )}
        </div>
      )}
    </div>
  );
}

function ExerciseBody({
  exercise,
  resetKey,
  onAnswered,
}: {
  exercise: Exercise;
  resetKey: number;
  onAnswered: (correct: boolean) => void;
}) {
  switch (exercise.exercise_type) {
    case "multiple_choice_translation":
      return <MultipleChoiceExercise exercise={exercise} onAnswered={onAnswered} />;
    case "word_order":
      return <WordOrderExercise exercise={exercise} onAnswered={onAnswered} />;
    case "verb_conjugation":
      return <VerbConjugationExercise exercise={exercise} onAnswered={onAnswered} />;
    case "true_false":
      return <TrueFalseExercise exercise={exercise} onAnswered={onAnswered} />;
    case "matching_pairs":
      return <MatchingPairsExercise exercise={exercise} onAnswered={onAnswered} />;
    case "dictation":
      return <DictationExercise exercise={exercise} onAnswered={onAnswered} />;
    case "free_translation":
      return <FreeTranslationExercise exercise={exercise} onAnswered={onAnswered} />;
    case "audio_pronunciation":
    case undefined:
      return (
        <AudioExercise
          exercise={exercise}
          autoStart={resetKey > 0}
          onAnswered={(correct) => onAnswered(correct)}
        />
      );
    default:
      return (
        <p className="center-message">
          Tipo de exercício "{exercise.exercise_type}" ainda não tem tela própria — chega em breve.
        </p>
      );
  }
}
