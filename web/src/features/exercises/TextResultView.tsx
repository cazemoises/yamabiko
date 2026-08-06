import { DiffComparison } from "./DiffComparison";
import { explainDiff } from "./diffExplain";
import type { TextAttemptResult } from "./api";

const VERDICT_LABEL: Record<string, string> = { PASS: "PASS", PARTIAL: "PARTIAL", FAIL: "FAIL" };
const VERDICT_CLASS: Record<string, string> = {
  PASS: "verdict-pill-pass",
  PARTIAL: "verdict-pill-partial",
  FAIL: "verdict-pill-fail",
};
const VERDICT_DOT: Record<string, string> = { PASS: "var(--pass)", PARTIAL: "var(--partial)", FAIL: "var(--fail)" };

// Frames 18/19 (Resultado texto livre) — corpo compartilhado por dictation
// e free_translation, mesmo padrão de AudioResultView só que "Você disse"
// vira "Você digitou" e o texto esperado vem de result.expected (não de
// exercise.expected_transcript — free_translation pode ter comparado
// contra qualquer uma das acceptable_answers, ver TextAttemptResult).
export function TextResultView({ result, language }: { result: TextAttemptResult; language: string }) {
  return (
    <div className="result-body">
      <div className="result-header-row">
        <span className={`chip ${VERDICT_CLASS[result.verdict] ?? "verdict-pill-fail"}`}>
          <span className="chip-dot" style={{ background: VERDICT_DOT[result.verdict] }} />
          {VERDICT_LABEL[result.verdict] ?? result.verdict} · {Math.round(result.score * 100)}%
        </span>
      </div>

      <DiffComparison
        expected={result.expected}
        actual={result.transcript}
        diff={result.diff}
        language={language}
        actualLabel="Você digitou"
      />

      <div className="result-what-happened">
        <span className="result-what-happened-title">O que aconteceu</span>
        {result.diff.length === 0 ? (
          <div className="result-note">
            <span className="result-note-dot" style={{ background: "var(--pass)" }} />
            <span className="result-note-text">Perfeito — bateu exatamente com o esperado.</span>
          </div>
        ) : (
          <ul className="diff-explanations" data-testid="diff-explanations">
            {result.diff.map((entry, index) => (
              <li key={index} className="result-note">
                <span className="result-note-dot" style={{ background: "var(--fail)" }} />
                <span className="result-note-text">{explainDiff(entry, language)}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
