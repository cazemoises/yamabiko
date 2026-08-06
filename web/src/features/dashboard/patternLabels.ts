// Rótulo curto pra barra do heatmap (Frame 7) — diferente de
// features/exercises/diffExplain.ts, que explica UMA ocorrência específica
// numa frase completa, isto aqui é só o nome do padrão agregado.
const LABELS: Record<string, string> = {
  H_ASPIRADO_OMITIDO: "H mudo",
  VOGAL_ENGOLIDA: "Vogal engolida",
  R_L_T_CONFUSAO: "R / L / T",
  SONORIZACAO_CONFUSA: "Sonorização",
  TH_SUBSTITUICAO: "TH",
  VOGAL_FINAL_ADICIONADA: "Vogal final extra",
  VOGAL_REDUZIDA_IGNORADA: "Vogal reduzida",
  R_AMERICANO_TROCADO: "R americano",
  CONSOANTE_FINAL_OMITIDA: "Consoante final",
  OUTRO: "Outro",
};

export function patternLabel(pattern: string): string {
  return LABELS[pattern] ?? pattern;
}
