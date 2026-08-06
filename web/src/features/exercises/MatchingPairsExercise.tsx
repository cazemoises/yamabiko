import { useMemo, useState } from "react";
import { submitAnswer, type AnswerResult, type Exercise, type MatchingPairsData } from "./api";
import { ApiError } from "../../lib/apiClient";

// Embaralha com uma seed derivada do exercise.id (não Math.random — precisa
// ser estável entre re-renders sem precisar de mais um useState) — o
// backend manda pairs já alinhados 1:1 (left[i] combina com right[i]), sem
// isso a coluna direita ficaria na mesma ordem da esquerda e "combinar"
// seria só clicar na mesma linha dos dois lados.
function shuffledWithSeed<T>(items: T[], seed: string): T[] {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) hash = (hash * 31 + seed.charCodeAt(i)) >>> 0;
  const result = [...items];
  for (let i = result.length - 1; i > 0; i--) {
    hash = (hash * 1103515245 + 12345) >>> 0;
    const j = hash % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

interface MatchingPairsExerciseProps {
  exercise: Exercise;
  onAnswered: (correct: boolean) => void;
}

// exercise_type "matching_pairs" (Frame 14) — toca 1 item da esquerda (PT)
// + 1 da direita (idioma-alvo) pra formar um par; par formado fica
// "matched" e sai da seleção. Confere no backend (POST /answer) só quando
// todos os pares da esquerda foram usados — é tudo ou nada (Fase A:
// validateMatchingPairs não dá crédito parcial), então não faz sentido
// mandar menos que o total.
export function MatchingPairsExercise({ exercise, onAnswered }: MatchingPairsExerciseProps) {
  const data = exercise.type_data as MatchingPairsData;
  const [matches, setMatches] = useState<Record<string, string>>({}); // left -> right
  const [selectedLeft, setSelectedLeft] = useState<string | null>(null);
  const [result, setResult] = useState<AnswerResult | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shuffledRights = useMemo(
    () => shuffledWithSeed(data.pairs.map((p) => p.right), exercise.id),
    [data.pairs, exercise.id],
  );

  const matchedRights = new Set(Object.values(matches));

  function selectLeft(left: string): void {
    if (result || submitting || left in matches) return;
    setSelectedLeft(left);
  }

  async function selectRight(right: string): Promise<void> {
    if (result || submitting || !selectedLeft || matchedRights.has(right)) return;
    const nextMatches = { ...matches, [selectedLeft]: right };
    setMatches(nextMatches);
    setSelectedLeft(null);

    if (Object.keys(nextMatches).length === data.pairs.length) {
      setSubmitting(true);
      setError(null);
      try {
        const submittedPairs = Object.entries(nextMatches).map(([left, r]) => ({ left, right: r }));
        const answerResult = await submitAnswer(exercise.id, { submitted_pairs: submittedPairs });
        setResult(answerResult);
        onAnswered(answerResult.correct);
      } catch (err) {
        setError(err instanceof ApiError ? err.message : "Erro ao enviar resposta");
      } finally {
        setSubmitting(false);
      }
    }
  }

  const correctByLeft = new Map((result?.correct_pairs ?? []).map((p) => [p.left, p.right]));

  function leftClass(left: string): string {
    const classes = ["matching-item"];
    if (result) {
      if (correctByLeft.get(left) === matches[left]) classes.push("matched");
    } else if (left === selectedLeft) classes.push("selected");
    else if (left in matches) classes.push("matched");
    return classes.join(" ");
  }

  function rightClass(right: string): string {
    const classes = ["matching-item"];
    const matchedLeft = Object.entries(matches).find(([, r]) => r === right)?.[0];
    if (result && matchedLeft) {
      if (correctByLeft.get(matchedLeft) === right) classes.push("matched");
      else classes.push("incorrect");
    } else if (matchedRights.has(right)) classes.push("matched");
    return classes.join(" ");
  }

  return (
    <>
      <div className="exercise-prompt">
        <span className="exercise-prompt-hint">{exercise.prompt_pt}</span>
      </div>
      <p style={{ textAlign: "center", fontSize: 13, color: "var(--text-secondary)", padding: "0 20px" }}>
        Toque uma palavra de cada lado para formar o par
      </p>

      <div className="matching-columns">
        <div className="matching-column">
          {data.pairs.map((pair) => (
            <button
              key={pair.left}
              type="button"
              className={leftClass(pair.left)}
              disabled={result != null || submitting || pair.left in matches}
              onClick={() => selectLeft(pair.left)}
            >
              {pair.left}
            </button>
          ))}
        </div>
        <div className="matching-column">
          {shuffledRights.map((right) => (
            <button
              key={right}
              type="button"
              className={`jp ${rightClass(right)}`}
              disabled={result != null || submitting || matchedRights.has(right)}
              onClick={() => void selectRight(right)}
            >
              {right}
            </button>
          ))}
        </div>
      </div>

      {error && <p className="error">{error}</p>}
    </>
  );
}
