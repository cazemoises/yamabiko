import { useEffect, useState } from "react";
import { listScenarios } from "./api";
import { getScenario, getAttemptHistory, type Scenario } from "../exercises/api";

export interface ScenarioProgress {
  scenario: Scenario;
  total: number;
  completed: number;
}

// Progresso de cenário é derivado de attempts (audio_pronunciation) — os 7
// tipos novos de exercício (Fase A) não persistem tentativa (decisão de
// escopo documentada em BUILD_STATE.md: POST /answer e /text-attempt são
// stateless), então um exercício desses tipos dentro de um cenário nunca
// conta como "completo" aqui até essa decisão ser revisitada.
export function useScenariosProgress(language: string): {
  progress: ScenarioProgress[];
  loading: boolean;
  error: string | null;
} {
  const [progress, setProgress] = useState<ScenarioProgress[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    async function load(): Promise<void> {
      try {
        const scenarios = await listScenarios(language);
        const withProgress = await Promise.all(
          scenarios.map(async (scenario): Promise<ScenarioProgress> => {
            const detail = await getScenario(scenario.id);
            const histories = await Promise.all(detail.exercises.map((exercise) => getAttemptHistory(exercise.id)));
            const completed = histories.filter((history) => history[0]?.verdict === "PASS").length;
            return { scenario, total: detail.exercises.length, completed };
          }),
        );
        if (!cancelled) setProgress(withProgress);
      } catch {
        if (!cancelled) setError("Erro ao carregar cenários");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [language]);

  return { progress, loading, error };
}
