import { api } from "../../lib/apiClient";

export interface PatternCount {
  pattern: string;
  occurrences: number;
}

export function getHeatmap(): Promise<PatternCount[]> {
  return api.get<PatternCount[]>("/dashboard/heatmap");
}
