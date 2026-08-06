import { DiffComparison } from "./DiffComparison";
import { explainDiff } from "./diffExplain";
import type { AttemptResult, Exercise } from "./api";

const VERDICT_LABEL: Record<string, string> = { PASS: "PASS", PARTIAL: "PARTIAL", FAIL: "FAIL" };
const VERDICT_CLASS: Record<string, string> = {
  PASS: "verdict-pill-pass",
  PARTIAL: "verdict-pill-partial",
  FAIL: "verdict-pill-fail",
};
const VERDICT_DOT: Record<string, string> = { PASS: "var(--pass)", PARTIAL: "var(--partial)", FAIL: "var(--fail)" };

// Corpo do Frame 5 (Resultado da Tentativa — áudio): chip de veredito + score
// + card "Esperado/Você disse" (DiffComparison) + notas explicando cada
// divergência. Botões de ação (Tentar de novo/Próximo/Ouvir pronúncia)
// ficam em ExercisePage, que já orquestra a navegação de cenário.
export function AudioResultView({ exercise, result }: { exercise: Exercise; result: AttemptResult }) {
  return (
    <div className="result-body">
      <div className="result-header-row">
        <span className={`chip ${VERDICT_CLASS[result.verdict] ?? "verdict-pill-fail"}`}>
          <span className="chip-dot" style={{ background: VERDICT_DOT[result.verdict] }} />
          {VERDICT_LABEL[result.verdict] ?? result.verdict}
        </span>
        <span className="result-score">{Math.round(result.score * 100)}%</span>
      </div>

      <DiffComparison
        expected={exercise.expected_transcript}
        actual={result.transcript}
        diff={result.diff}
        language={exercise.language || "ja-JP"}
      />

      <div className="result-what-happened">
        <span className="result-what-happened-title">O que aconteceu</span>
        {result.diff.length === 0 ? (
          <div className="result-note">
            <span className="result-note-dot" style={{ background: "var(--pass)" }} />
            <span className="result-note-text">Perfeito — pronúncia bateu exatamente com o esperado.</span>
          </div>
        ) : (
          <ul className="diff-explanations" data-testid="diff-explanations">
            {result.diff.map((entry, index) => (
              <li key={index} className="result-note">
                <span className="result-note-dot" style={{ background: "var(--fail)" }} />
                <span className="result-note-text">{explainDiff(entry, exercise.language || "ja-JP")}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
