import type { DiffEntry } from "../features/exercises/api";

// Espelha core-api/internal/comparison/compare.go:normalize() — mesma
// normalização (NFKC + remove espaços + katakana->hiragana) garante que os
// índices de `position` calculados pelo backend continuem válidos aqui.
export function normalizeKana(s: string): string {
  const withoutSpaces = s.normalize("NFKC").replace(/\s+/g, "");
  return katakanaToHiragana(withoutSpaces);
}

function katakanaToHiragana(s: string): string {
  return Array.from(s)
    .map((ch) => {
      const code = ch.codePointAt(0) ?? 0;
      return code >= 0x30a1 && code <= 0x30f6 ? String.fromCodePoint(code - 0x60) : ch;
    })
    .join("");
}

export interface AlignedColumn {
  position: number;
  expectedChar: string | null;
  actualChar: string | null;
  entry: DiffEntry | null;
}

// Reconstrói as duas colunas alinhadas (esperado x transcrito), casadas
// posição a posição, a partir dos textos completos e do `diff` já calculado
// pelo backend — sem reimplementar Levenshtein no cliente. `diff[i].position`
// é o índice dessa coluna na sequência de alinhamento completa (matches
// inclusos, ver compare.go); qualquer posição sem entrada em `diff` é match.
export function alignForDisplay(expectedRaw: string, actualRaw: string, diff: DiffEntry[]): AlignedColumn[] {
  const expected = Array.from(normalizeKana(expectedRaw));
  const actual = Array.from(normalizeKana(actualRaw));
  const diffByPosition = new Map(diff.map((entry) => [entry.position, entry]));

  const columns: AlignedColumn[] = [];
  let expPtr = 0;
  let actPtr = 0;
  let position = 0;

  while (expPtr < expected.length || actPtr < actual.length) {
    const entry = diffByPosition.get(position);
    if (entry) {
      columns.push({
        position,
        expectedChar: entry.expected ?? null,
        actualChar: entry.actual ?? null,
        entry,
      });
      if (entry.op !== "INSERT") expPtr++;
      if (entry.op !== "DELETE") actPtr++;
    } else {
      columns.push({ position, expectedChar: expected[expPtr], actualChar: actual[actPtr], entry: null });
      expPtr++;
      actPtr++;
    }
    position++;
  }

  return columns;
}
