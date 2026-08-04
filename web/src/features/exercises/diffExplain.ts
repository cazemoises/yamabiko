import { toRomaji } from "../../lib/romaji";
import type { DiffEntry } from "./api";

// Traduz o diff técnico do backend (Op + Pattern) numa frase em português
// voltada ao aluno, no lugar de expor rótulos como "SUBSTITUTE"/"OUTRO".
export function explainDiff(entry: DiffEntry): string {
  const expected = entry.expected;
  const actual = entry.actual;
  const expectedRomaji = expected ? toRomaji(expected) : null;
  const actualRomaji = actual ? toRomaji(actual) : null;

  switch (entry.pattern) {
    case "H_ASPIRADO_OMITIDO":
      return `Você não pronunciou o "${expectedRomaji}" (${expected}) — o H aspirado é mudo em português, mas existe em japonês.`;
    case "VOGAL_ENGOLIDA":
      if (entry.op === "DELETE") {
        return `Você engoliu a vogal "${expectedRomaji}" (${expected}).`;
      }
      return `Você disse "${actualRomaji}" (${actual}) onde devia ser "${expectedRomaji}" (${expected}) — confusão comum entre essas vogais.`;
    case "R_L_T_CONFUSAO":
      return `Você disse "${actualRomaji}" (${actual}) no lugar de "${expectedRomaji}" (${expected}) — confusão comum entre R, L e T em japonês.`;
    default:
      if (entry.op === "INSERT") {
        return `Foi ouvido um som extra "${actualRomaji}" (${actual}) que não deveria estar aí.`;
      }
      if (entry.op === "DELETE") {
        return `Você não pronunciou o som "${expectedRomaji}" (${expected}).`;
      }
      return `Você disse "${actualRomaji}" (${actual}) onde devia ser "${expectedRomaji}" (${expected}) — confusão comum entre esses sons.`;
  }
}
