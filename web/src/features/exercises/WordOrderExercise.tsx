import { useState } from "react";
import { submitAnswer, type AnswerResult, type Exercise, type WordOrderData } from "./api";
import { ApiError } from "../../lib/apiClient";

interface WordOrderExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "word_order" (Frame 10) — toca uma palavra do banco pra
// colocar na área de resposta, toca uma palavra já colocada pra devolver
// ao banco. Confere no backend (POST /answer) só quando todas as palavras
// do banco foram usadas.
export function WordOrderExercise({ exercise, onAnswered }: WordOrderExerciseProps) {
  const data = exercise.type_data as WordOrderData;
  const [placed, setPlaced] = useState<number[]>([]); // índices em data.shuffled_words, na ordem colocada
  const [result, setResult] = useState<AnswerResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const bankIndices = data.shuffled_words.map((_, i) => i).filter((i) => !placed.includes(i));

  function placeWord(index: number): void {
    if (result || submitting) return;
    setPlaced((prev) => [...prev, index]);
  }

  function unplaceWord(position: number): void {
    if (result || submitting) return;
    setPlaced((prev) => prev.filter((_, i) => i !== position));
  }

  async function handleSubmit(finalPlaced: number[]): Promise<void> {
    setSubmitting(true);
    setError(null);
    try {
      const submittedOrder = finalPlaced.map((i) => data.shuffled_words[i]);
      const answerResult = await submitAnswer(exercise.id, { submitted_order: submittedOrder });
      setResult(answerResult);
      onAnswered(answerResult.correct);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Erro ao enviar resposta");
    } finally {
      setSubmitting(false);
    }
  }

  function handlePlaceWord(index: number): void {
    placeWord(index);
    const nextPlaced = [...placed, index];
    if (nextPlaced.length === data.shuffled_words.length) {
      void handleSubmit(nextPlaced);
    }
  }

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">{exercise.prompt_pt}</span>
      </div>
      <p style={{ textAlign: "center", fontSize: 13, color: "var(--text-secondary)", padding: "0 20px" }}>
        Organize as palavras para formar a frase certa
      </p>

      <div className="word-order-answer">
        {placed.map((index, position) => (
          <button
            key={`${index}-${position}`}
            type="button"
            className="word-chip word-chip-placed"
            disabled={result != null || submitting}
            onClick={() => unplaceWord(position)}
          >
            {data.shuffled_words[index]}
          </button>
        ))}
      </div>

      {result && !result.correct && result.correct_order && (
        <p style={{ textAlign: "center", fontSize: 13, color: "var(--pass)", padding: "0 20px" }}>
          Correta: {result.correct_order.join(" ")}
        </p>
      )}

      <div className="word-bank">
        {bankIndices.map((index) => (
          <button
            key={index}
            type="button"
            className="word-chip word-chip-bank"
            disabled={result != null || submitting}
            onClick={() => handlePlaceWord(index)}
          >
            {data.shuffled_words[index]}
          </button>
        ))}
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}
