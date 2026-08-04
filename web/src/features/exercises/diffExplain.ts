import { toRomaji } from "../../lib/romaji";
import type { DiffEntry } from "./api";

function isJapanese(language: string): boolean {
  return language.toLowerCase().startsWith("ja");
}

// Traduz o diff técnico do backend (Op + Pattern) numa frase em português
// voltada ao aluno, no lugar de expor rótulos como "SUBSTITUTE"/"OUTRO".
export function explainDiff(entry: DiffEntry, language: string = "ja-JP"): string {
  const expected = entry.expected;
  const actual = entry.actual;
  const japanese = isJapanese(language);
  // Romaji só ajuda leitura de kana — em inglês o "expected"/"actual" já é
  // Latin, então usamos o texto cru em vez de duplicá-lo via toRomaji (que é
  // passthrough pra caracteres fora da tabela hiragana).
  const expectedText = japanese && expected ? toRomaji(expected) : expected;
  const actualText = japanese && actual ? toRomaji(actual) : actual;

  switch (entry.pattern) {
    case "H_ASPIRADO_OMITIDO":
      return `Você não pronunciou o "${expectedText}" (${expected}) — o H aspirado é mudo em português, mas existe em japonês.`;
    case "VOGAL_ENGOLIDA":
      if (entry.op === "DELETE") {
        return `Você engoliu a vogal "${expectedText}" (${expected}).`;
      }
      return `Você disse "${actualText}" (${actual}) onde devia ser "${expectedText}" (${expected}) — confusão comum entre essas vogais.`;
    case "R_L_T_CONFUSAO":
      return `Você disse "${actualText}" (${actual}) no lugar de "${expectedText}" (${expected}) — confusão comum entre R, L e T em japonês.`;
    case "SONORIZACAO_CONFUSA":
      return `Você disse "${actualText}" (${actual}) no lugar de "${expectedText}" (${expected}) — confusão de sonorização (dakuten) comum em japonês.`;
    case "TH_SUBSTITUICAO":
      return `O som "TH" (${expected}) não existe em português — você trocou por outro som ("${actual}"). Tente colocar a língua entre os dentes.`;
    case "VOGAL_FINAL_ADICIONADA":
      return `Você adicionou uma vogal extra ("${actual}") depois da consoante final — em inglês essa consoante fica sozinha, sem vogal de apoio.`;
    case "VOGAL_REDUZIDA_IGNORADA":
      return `Você disse "${actual}" onde devia ser "${expected}" — em inglês essa vogal costuma reduzir (schwa), não é pronunciada por completo.`;
    case "R_AMERICANO_TROCADO":
      return `O R do inglês americano ("${expected}") é diferente do R do português — evite o R gutural/vibrado.`;
    case "CONSOANTE_FINAL_OMITIDA":
      return `Você não pronunciou a consoante final "${expected}" — em inglês ela costuma ser audível, mesmo no fim da palavra.`;
    default:
      if (entry.op === "INSERT") {
        return `Foi ouvido um som extra "${actualText}" (${actual}) que não deveria estar aí.`;
      }
      if (entry.op === "DELETE") {
        return `Você não pronunciou o som "${expectedText}" (${expected}).`;
      }
      return `Você disse "${actualText}" (${actual}) onde devia ser "${expectedText}" (${expected}) — confusão comum entre esses sons.`;
  }
}
