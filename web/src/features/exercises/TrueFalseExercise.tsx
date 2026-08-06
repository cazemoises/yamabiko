import { useState } from "react";
import { submitAnswer, type AnswerResult, type Exercise, type TrueFalseData } from "./api";
import { ApiError } from "../../lib/apiClient";

interface TrueFalseExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "true_false" (Frame 15) — 2 botões grandes, o escolhido
// fica preenchido de acento até a resposta voltar; depois vira
// verde/vermelho conforme certo/errado.
export function TrueFalseExercise({ exercise, onAnswered }: TrueFalseExerciseProps) {
  const data = exercise.type_data as TrueFalseData;
  const [selected, setSelected] = useState<boolean | null>(null);
  const [result, setResult] = useState<AnswerResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSelect(answer: boolean): Promise<void> {
    if (result || submitting) return;
    setSelected(answer);
    setSubmitting(true);
    setError(null);
    try {
      const answerResult = await submitAnswer(exercise.id, { answer });
      setResult(answerResult);
      onAnswered(answerResult.correct);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar resposta");
      setSelected(null);
    } finally {
      setSubmitting(false);
    }
  }

  function buttonClass(value: boolean): string {
    if (!result) return selected === value ? "true-false-button chosen-true" : "true-false-button";
    if (value === result.correct_answer) return "true-false-button correct";
    if (value === selected) return "true-false-button incorrect";
    return "true-false-button dimmed";
  }

  return (
    <>
      <div className="true-false-statement">
        <p>{data.statement}</p>
      </div>

      <div className="true-false-buttons">
        <button type="button" className={buttonClass(true)} disabled={result != null || submitting} onClick={() => handleSelect(true)}>
          Verdadeiro
        </button>
        <button type="button" className={buttonClass(false)} disabled={result != null || submitting} onClick={() => handleSelect(false)}>
          Falso
        </button>
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}
