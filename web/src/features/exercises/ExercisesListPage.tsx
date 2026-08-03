import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listExercises, type Exercise } from "./api";

export function ExercisesListPage() {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listExercises()
      .then(setExercises)
      .catch(() => setError("Erro ao carregar exercícios"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p>Carregando...</p>;
  if (error) return <p className="error">{error}</p>;

  const bySprintDay = groupBySprintDay(exercises);

  return (
    <div className="exercises-list">
      <h1>Exercícios</h1>
      {Object.entries(bySprintDay).map(([day, items]) => (
        <section key={day}>
          <h2>Dia {day}</h2>
          <ul>
            {items.map((exercise) => (
              <li key={exercise.id}>
                <Link to={`/exercises/${exercise.id}`}>
                  <span className="category">{exercise.category}</span> {exercise.prompt_pt}
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}

function groupBySprintDay(exercises: Exercise[]): Record<number, Exercise[]> {
  const result: Record<number, Exercise[]> = {};
  for (const exercise of exercises) {
    (result[exercise.sprint_day_ref] ??= []).push(exercise);
  }
  return result;
}
