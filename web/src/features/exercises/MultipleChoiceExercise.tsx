import { useState } from "react";
import { submitAnswer, type AnswerResult, type Exercise, type MultipleChoiceTranslationData } from "./api";
import { ApiError } from "../../lib/apiClient";

interface MultipleChoiceExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "multiple_choice_translation" (Frame 9, resultado nos
// Frames 16/17) — a lista de opções não troca de tela ao responder, só
// ganha cor (certa=verde, escolhida errada=vermelha, resto apagado), igual
// o design mostra.
export function MultipleChoiceExercise({ exercise, onAnswered }: MultipleChoiceExerciseProps) {
  const data = exercise.type_data as MultipleChoiceTranslationData;
  const [selected, setSelected] = useState<number | null>(null);
  const [result, setResult] = useState<AnswerResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSelect(index: number): Promise<void> {
    if (result || submitting) return;
    setSelected(index);
    setSubmitting(true);
    setError(null);
    try {
      const answerResult = await submitAnswer(exercise.id, { selected_index: index });
      setResult(answerResult);
      onAnswered(answerResult.correct);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar resposta");
      setSelected(null);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">{exercise.prompt_pt}</span>
      </div>

      <div className="option-list">
        {data.options.map((option, index) => {
          const isCorrectOption = result != null && index === result.correct_index;
          const isWrongSelected = result != null && !result.correct && index === selected;
          const classes = ["option-button"];
          if (isCorrectOption) classes.push("correct");
          else if (isWrongSelected) classes.push("incorrect");
          else if (result) classes.push("dimmed");
          else if (index === selected) classes.push("selected");

          let suffix = "";
          if (result && !result.correct) {
            if (isCorrectOption) suffix = " — correta";
            else if (isWrongSelected) suffix = " — sua resposta";
          }

          return (
            <button
              key={option}
              type="button"
              className={classes.join(" ")}
              disabled={result != null || submitting}
              onClick={() => handleSelect(index)}
            >
              <span>
                {option}
                {suffix}
              </span>
              {isCorrectOption && <CheckIcon />}
              {isWrongSelected && <CrossIcon />}
            </button>
          );
        })}
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}

export function CheckIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" style={{ flexShrink: 0 }}>
      <circle cx="8" cy="8" r="7" fill="var(--pass)" />
      <line x1="5" y1="8.2" x2="7" y2="10.2" stroke="var(--bg-card)" strokeWidth="1.6" />
      <line x1="7" y1="10.2" x2="11" y2="5.8" stroke="var(--bg-card)" strokeWidth="1.6" />
    </svg>
  );
}

export function CrossIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" style={{ flexShrink: 0 }}>
      <circle cx="8" cy="8" r="7" fill="var(--fail)" />
      <line x1="5.5" y1="5.5" x2="10.5" y2="10.5" stroke="var(--bg-card)" strokeWidth="1.6" />
      <line x1="10.5" y1="5.5" x2="5.5" y2="10.5" stroke="var(--bg-card)" strokeWidth="1.6" />
    </svg>
  );
}
