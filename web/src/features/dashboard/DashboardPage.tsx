import { useEffect, useState } from "react";
import { listExercises, getAttemptHistory, type Attempt, type Exercise } from "../exercises/api";
import { getHeatmap, type PatternCount } from "./api";
import { patternLabel } from "./patternLabels";

interface AttemptWithExercise extends Attempt {
  exercisePromptOrTranscript: string;
}

// Frame 7 (Progresso) — estatísticas reais: total de tentativas + acerto
// médio (sobre TODAS as tentativas de TODOS os exercícios, não só a mais
// recente de cada um, ao contrário da tabela antiga desta página),
// padrões de erro mais comuns (GET /dashboard/heatmap) e as tentativas
// recentes. Mesma limitação de performance que a versão anterior desta
// página já tinha (N+1: 1 requisição de histórico por exercício) — não é
// regressão desta mudança, só não foi "consertada" agora porque não fazia
// parte do pedido.
export function DashboardPage() {
  const [attempts, setAttempts] = useState<AttemptWithExercise[]>([]);
  const [totalExercises, setTotalExercises] = useState(0);
  const [patterns, setPatterns] = useState<PatternCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load(): Promise<void> {
      try {
        const [exercises, heatmap] = await Promise.all([listExercises(), getHeatmap()]);
        setTotalExercises(exercises.length);
        setPatterns(heatmap);

        const byExercise = await Promise.all(
          exercises.map(async (exercise: Exercise) => {
            const history = await getAttemptHistory(exercise.id);
            return history.map((attempt) => ({
              ...attempt,
              exercisePromptOrTranscript: exercise.expected_transcript || exercise.prompt_pt,
            }));
          }),
        );
        setAttempts(byExercise.flat());
      } catch {
        setError("Erro ao carregar progresso");
      } finally {
        setLoading(false);
      }
    }
    void load();
  }, []);

  if (loading) return <p className="center-message">Carregando progresso...</p>;
  if (error) return <p className="error">{error}</p>;

  const averageScore = attempts.length > 0 ? attempts.reduce((sum, a) => sum + a.similarity_score, 0) / attempts.length : 0;
  const maxOccurrences = Math.max(1, ...patterns.map((p) => p.occurrences));
  const topPatterns = [...patterns].sort((a, b) => b.occurrences - a.occurrences).slice(0, 5);
  const recent = [...attempts].sort((a, b) => b.created_at.localeCompare(a.created_at)).slice(0, 8);

  return (
    <div className="page">
      <div className="page-header">
        <span className="page-title">Progresso</span>
      </div>

      <div className="stat-row">
        <div className="stat-card">
          <span className="stat-card-value">{attempts.length}</span>
          <span className="stat-card-label">tentativas totais</span>
        </div>
        <div className="stat-card">
          <span className="stat-card-value">{Math.round(averageScore * 100)}%</span>
          <span className="stat-card-label">taxa de acerto média</span>
        </div>
      </div>

      {totalExercises > 0 && (
        <p className="page-subtitle">
          {new Set(attempts.map((a) => a.exercise_id)).size} de {totalExercises} exercícios tentados
        </p>
      )}

      {topPatterns.length > 0 && (
        <>
          <span className="section-title">Padrões de erro mais comuns</span>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {topPatterns.map((p) => (
              <div key={p.pattern} className="error-bar-row">
                <span className="error-bar-label">{patternLabel(p.pattern)}</span>
                <div className="error-bar-track">
                  <div className="error-bar-fill" style={{ width: `${(p.occurrences / maxOccurrences) * 100}%` }} />
                </div>
                <span className="error-bar-count">{p.occurrences}x</span>
              </div>
            ))}
          </div>
        </>
      )}

      {recent.length > 0 && (
        <>
          <span className="section-title">Tentativas recentes</span>
          <div>
            {recent.map((attempt) => (
              <div key={attempt.id} className="recent-attempt-row">
                <span
                  className="chip-dot"
                  style={{ background: verdictColor(attempt.verdict), flexShrink: 0 }}
                />
                <span className="recent-attempt-text">{attempt.exercisePromptOrTranscript}</span>
                <span className="recent-attempt-when">{relativeDay(attempt.created_at)}</span>
              </div>
            ))}
          </div>
        </>
      )}

      {attempts.length === 0 && <p className="center-message">Nenhuma tentativa ainda — comece um exercício!</p>}
    </div>
  );
}

function verdictColor(verdict: string): string {
  if (verdict === "PASS") return "var(--pass)";
  if (verdict === "PARTIAL") return "var(--partial)";
  return "var(--fail)";
}

function relativeDay(isoDate: string): string {
  const date = new Date(isoDate);
  const now = new Date();
  const diffDays = Math.floor((now.setHours(0, 0, 0, 0) - new Date(date).setHours(0, 0, 0, 0)) / 86_400_000);
  if (diffDays <= 0) return "hoje";
  if (diffDays === 1) return "ontem";
  return `${diffDays}d atrás`;
}
