import { useState } from "react";
import { CheckIcon, CrossIcon } from "./MultipleChoiceExercise";
import { submitAnswer, type AnswerResult, type Exercise, type VerbConjugationData } from "./api";
import { ApiError } from "../../lib/apiClient";

interface VerbConjugationExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "verb_conjugation" (Frame 11) — mesmo núcleo de escolha por
// índice de multiple_choice_translation (reusa CheckIcon/CrossIcon), só o
// prompt muda: sentence_template com um "___" no lugar do verbo, em vez de
// uma pergunta de tradução direta.
export function VerbConjugationExercise({ exercise, onAnswered }: VerbConjugationExerciseProps) {
  const data = exercise.type_data as VerbConjugationData;
  const [selected, setSelected] = useState<number | null>(null);
  const [result, setResult] = useState<AnswerResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [before, after] = data.sentence_template.split("___");
  const chosenWord = selected !== null ? data.options[selected] : "___";

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

  const blankColor = result ? (result.correct ? "var(--pass)" : "var(--fail)") : "var(--accent-base)";

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">Complete a frase</span>
        <span className="exercise-prompt-main">
          {before}
          <span style={{ display: "inline-block", minWidth: 56, borderBottom: `2px solid ${blankColor}`, color: blankColor }}>
            {chosenWord}
          </span>
          {after}
        </span>
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

          return (
            <button
              key={option}
              type="button"
              className={classes.join(" ")}
              disabled={result != null || submitting}
              onClick={() => handleSelect(index)}
            >
              <span>{option}</span>
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
