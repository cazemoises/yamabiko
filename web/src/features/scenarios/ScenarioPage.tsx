import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getScenario, getAttemptHistory, type ScenarioDetail } from "../exercises/api";

interface ExerciseRow {
  id: string;
  promptPt: string;
  passed: boolean;
}

// Overview de um cenário: contexto + lista dos exercícios em ordem
// (order_in_scenario), cada um com seu status real (última tentativa
// PASS) — clicar entra em ExercisePage, que já sabe navegar "Próximo"
// dentro do cenário (fluxo existente, reaproveitado sem mudança).
export function ScenarioPage() {
  const { id } = useParams<{ id: string }>();
  const [scenario, setScenario] = useState<ScenarioDetail | null>(null);
  const [rows, setRows] = useState<ExerciseRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    getScenario(id)
      .then(async (detail) => {
        setScenario(detail);
        const withStatus = await Promise.all(
          detail.exercises.map(async (exercise): Promise<ExerciseRow> => {
            const history = await getAttemptHistory(exercise.id);
            return { id: exercise.id, promptPt: exercise.prompt_pt, passed: history[0]?.verdict === "PASS" };
          }),
        );
        setRows(withStatus);
      })
      .catch(() => setError("Erro ao carregar cenário"))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <p className="center-message">Carregando...</p>;
  if (error || !scenario) return <p className="error">{error ?? "Cenário não encontrado"}</p>;

  const firstIncomplete = rows.find((row) => !row.passed);

  return (
    <div className="page">
      <div className="page-header">
        <span className="page-title">{scenario.title_pt}</span>
      </div>
      <div className="context-banner" style={{ margin: 0 }}>
        <p className="context-banner-body">{scenario.context_description_pt}</p>
      </div>

      {firstIncomplete && (
        <Link to={`/exercises/${firstIncomplete.id}`} className="btn-primary" style={{ textDecoration: "none" }}>
          continuar
        </Link>
      )}

      <ul className="plain-list">
        {rows.map((row, index) => (
          <li key={row.id}>
            <Link to={`/exercises/${row.id}`} className="plain-list-row">
              <span
                className="chip-dot"
                style={{ background: row.passed ? "var(--pass)" : "var(--border)", flexShrink: 0 }}
              />
              <span style={{ flex: 1, color: "var(--text)" }}>
                {index + 1}. {row.promptPt}
              </span>
              {row.passed && <span style={{ color: "var(--pass)", fontSize: 12, fontWeight: 700 }}>✓</span>}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
