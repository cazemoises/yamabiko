import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listExercises, type Exercise } from "./api";

const LANGUAGES = [
  { value: "ja-JP", label: "🇯🇵 Japonês" },
  { value: "en-US", label: "🇺🇸 Inglês" },
];

export function ExercisesListPage() {
  const [language, setLanguage] = useState("ja-JP");
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    listExercises(undefined, language)
      .then(setExercises)
      .catch(() => setError("Erro ao carregar exercícios"))
      .finally(() => setLoading(false));
  }, [language]);

  const bySprintDay = groupBySprintDay(exercises);

  return (
    <div className="exercises-list">
      <h1>Exercícios</h1>
      <div className="language-toggle" role="group" aria-label="Idioma dos exercícios">
        {LANGUAGES.map((lang) => (
          <button
            key={lang.value}
            type="button"
            className={lang.value === language ? "language-toggle-button active" : "language-toggle-button"}
            aria-pressed={lang.value === language}
            onClick={() => setLanguage(lang.value)}
          >
            {lang.label}
          </button>
        ))}
      </div>
      {loading && <p>Carregando...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && exercises.length === 0 && <p>Nenhum exercício encontrado nesse idioma.</p>}
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
