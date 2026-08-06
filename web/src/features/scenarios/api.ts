import { api } from "../../lib/apiClient";
import type { Scenario } from "../exercises/api";

export function listScenarios(language: string): Promise<Scenario[]> {
  return api.get<Scenario[]>(`/scenarios?language=${encodeURIComponent(language)}`);
}
