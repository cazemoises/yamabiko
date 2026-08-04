import { alignForDisplay } from "../../lib/kanaAlign";
import { toRomaji } from "../../lib/romaji";
import { explainDiff } from "./diffExplain";
import { SpeakButton } from "../../components/audio/SpeakButton";
import type { DiffEntry } from "./api";

interface DiffComparisonProps {
  expected: string;
  actual: string;
  diff: DiffEntry[];
}

export function DiffComparison({ expected, actual, diff }: DiffComparisonProps) {
  const columns = alignForDisplay(expected, actual, diff);

  return (
    <div className="diff-comparison">
      <div className="diff-row" data-testid="diff-row-expected">
        <span className="diff-row-label">Esperado</span>
        <div className="diff-chars">
          {columns.map((col) => (
            <DiffChar key={col.position} char={col.expectedChar} mismatch={col.entry !== null} />
          ))}
        </div>
        <SpeakButton text={expected} className="speak-button-inline" />
      </div>
      <div className="diff-row" data-testid="diff-row-actual">
        <span className="diff-row-label">Você disse</span>
        <div className="diff-chars">
          {columns.map((col) => (
            <DiffChar key={col.position} char={col.actualChar} mismatch={col.entry !== null} />
          ))}
        </div>
      </div>

      {diff.length > 0 && (
        <ul className="diff-explanations" data-testid="diff-explanations">
          {diff.map((entry, index) => (
            <li key={index}>{explainDiff(entry)}</li>
          ))}
        </ul>
      )}
    </div>
  );
}

function DiffChar({ char, mismatch }: { char: string | null; mismatch: boolean }) {
  if (char === null) {
    return (
      <span className="diff-char diff-char-gap" aria-hidden="true">
        <span className="diff-char-kana">–</span>
      </span>
    );
  }
  const romaji = toRomaji(char);
  return (
    <span className={mismatch ? "diff-char diff-char-mismatch" : "diff-char"}>
      <span className="diff-char-kana">{char}</span>
      {romaji && (
        <span className={mismatch ? "diff-char-romaji diff-char-romaji-mismatch" : "diff-char-romaji"}>
          {romaji}
        </span>
      )}
    </span>
  );
}
