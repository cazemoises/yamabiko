import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listExercises, getAttemptHistory, type Attempt, type Exercise } from "../exercises/api";

interface ProgressRow {
  exercise: Exercise;
  latestAttempt: Attempt | null;
}

export function DashboardPage() {
  const [rows, setRows] = useState<ProgressRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load(): Promise<void> {
      try {
        const exercises = await listExercises();
        const withHistory = await Promise.all(
          exercises.map(async (exercise): Promise<ProgressRow> => {
            const history = await getAttemptHistory(exercise.id);
            return { exercise, latestAttempt: history[0] ?? null };
          }),
        );
        setRows(withHistory);
      } catch {
        setError("Erro ao carregar progresso");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) return <p>Carregando progresso...</p>;
  if (error) return <p className="error">{error}</p>;

  const attempted = rows.filter((row) => row.latestAttempt !== null).length;

  return (
    <div className="dashboard-page">
      <h1>Meu progresso</h1>
      <p>
        {attempted} de {rows.length} exercícios tentados
      </p>
      <table>
        <thead>
          <tr>
            <th>Exercício</th>
            <th>Último veredito</th>
            <th>Score</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(({ exercise, latestAttempt }) => (
            <tr key={exercise.id}>
              <td>
                <Link to={`/exercises/${exercise.id}`}>{exercise.prompt_pt}</Link>
              </td>
              <td>{latestAttempt ? latestAttempt.verdict : "—"}</td>
              <td>{latestAttempt ? `${Math.round(latestAttempt.similarity_score * 100)}%` : "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
