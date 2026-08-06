import { alignForDisplay } from "../../lib/kanaAlign";
import { toRomaji } from "../../lib/romaji";
import type { DiffEntry } from "./api";

interface DiffComparisonProps {
  expected: string;
  actual: string;
  diff: DiffEntry[];
  language?: string;
  /** Rótulo da 2ª linha — "Você disse" (áudio) ou "Você digitou" (texto/ditado). */
  actualLabel?: string;
}

// Romaji só faz sentido como apoio de leitura pra kana — em inglês seria só
// duplicar a mesma letra embaixo dela mesma, então fica desligado fora de ja-JP.
function isJapanese(language: string): boolean {
  return language.toLowerCase().startsWith("ja");
}

// Card "Esperado / Você disse" do Frame 5 — o botão de ouvir pronúncia saiu
// daqui (não faz mais parte da linha "Esperado") e virou um botão próprio no
// rodapé da tela de resultado (AudioResultView), igual ao design.
export function DiffComparison({ expected, actual, diff, language = "ja-JP", actualLabel = "Você disse" }: DiffComparisonProps) {
  const columns = alignForDisplay(expected, actual, diff, language);
  const showRomaji = isJapanese(language);

  return (
    <div className="card diff-comparison">
      <div className="diff-row" data-testid="diff-row-expected">
        <span className="diff-row-label">Esperado</span>
        <div className="diff-chars">
          {columns.map((col) => (
            <DiffChar key={col.position} char={col.expectedChar} mismatch={col.entry !== null} showRomaji={showRomaji} />
          ))}
        </div>
      </div>
      <div className="diff-row" data-testid="diff-row-actual">
        <span className="diff-row-label">{actualLabel}</span>
        <div className="diff-chars">
          {columns.map((col) => (
            <DiffChar key={col.position} char={col.actualChar} mismatch={col.entry !== null} showRomaji={showRomaji} />
          ))}
        </div>
      </div>
    </div>
  );
}

function DiffChar({
  char,
  mismatch,
  showRomaji,
}: {
  char: string | null;
  mismatch: boolean;
  showRomaji: boolean;
}) {
  if (char === null) {
    return (
      <span className="diff-char diff-char-gap" aria-hidden="true">
        <span className="diff-char-kana">–</span>
      </span>
    );
  }
  const romaji = showRomaji ? toRomaji(char) : null;
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
