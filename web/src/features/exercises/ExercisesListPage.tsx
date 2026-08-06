import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { LanguageToggle } from "../../components/layout/LanguageToggle";
import { listExercises, type Exercise } from "./api";

// Frame 6 do design ("Exercícios avulsos") — só as frases fora de um
// cenário (scenario_id nulo). O backend não tem um filtro "sem cenário"
// pronto (Filter.ScenarioID só filtra POR um cenário específico), então o
// corte é client-side sobre a lista já carregada — mesmo custo de rede de
// antes, só muda o que aparece na tela.
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

  const standalone = exercises.filter((exercise) => !exercise.scenario_id);

  return (
    <div className="page">
      <div className="page-header">
        <span className="page-title">Exercícios</span>
        <LanguageToggle language={language} onChange={setLanguage} />
      </div>
      <p className="page-subtitle">Frases avulsas, fora de um cenário</p>

      {loading && <p className="center-message">Carregando...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && standalone.length === 0 && (
        <p className="center-message">Nenhum exercício avulso nesse idioma.</p>
      )}

      <ul className="plain-list">
        {standalone.map((exercise) => (
          <li key={exercise.id}>
            <Link to={`/exercises/${exercise.id}`} className="plain-list-row">
              <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 2 }}>
                <span className={exercise.language?.startsWith("ja") ? "jp" : undefined} style={{ fontSize: 15, fontWeight: 700, color: "var(--text)" }}>
                  {exercise.expected_transcript || exercise.prompt_pt}
                </span>
                <span className="list-row-subtitle">{exercise.prompt_pt}</span>
              </div>
              <span className="chip-dot" style={{ background: "var(--accent-base)" }} />
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
