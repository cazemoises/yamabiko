import { useState } from "react";
import { Link } from "react-router-dom";
import { LanguageToggle } from "../../components/layout/LanguageToggle";
import { useScenariosProgress } from "./useScenariosProgress";
import { iconForScenario } from "./scenarioIcon";

// Frame 3 do design — lista de cenários com progresso real (X de N
// exercícios), sem categorias fixas mockadas (o filtro por category do
// design foi simplificado: o backend não expõe category em Scenario, só em
// Exercise, então o filtro aqui é só por idioma via LanguageToggle).
export function ScenariosListPage() {
  const [language, setLanguage] = useState("ja-JP");
  const { progress, loading, error } = useScenariosProgress(language);

  return (
    <div className="page">
      <div className="page-header">
        <span className="page-title">Cenários</span>
        <LanguageToggle language={language} onChange={setLanguage} />
      </div>

      {loading && <p className="center-message">Carregando...</p>}
      {error && <p className="error">{error}</p>}
      {!loading && !error && progress.length === 0 && <p className="center-message">Nenhum cenário nesse idioma.</p>}

      <ul className="plain-list" style={{ gap: 10 }}>
        {progress.map(({ scenario, completed, total }) => {
          const done = total > 0 && completed === total;
          const Icon = iconForScenario(scenario.title_pt);
          return (
            <li key={scenario.id}>
              <Link to={`/scenarios/${scenario.id}`} className="list-row">
                <div className="list-row-icon">
                  <Icon size={18} strokeWidth={1.75} />
                </div>
                <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 3 }}>
                  <span className="list-row-title">{scenario.title_pt}</span>
                  <span className="list-row-subtitle">
                    {scenario.language === "ja-JP" ? "JA" : "EN"} · {completed}/{total} exercícios
                    {done ? " — concluído" : ""}
                  </span>
                </div>
                {done ? <CheckIcon /> : <ChevronIcon />}
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function ChevronIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14">
      <polygon points="4,2 11,7 4,12" fill="var(--accent-base)" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16">
      <circle cx="8" cy="8" r="7" fill="none" stroke="var(--pass)" strokeWidth="1.4" />
      <line x1="5" y1="8.2" x2="7" y2="10.2" stroke="var(--pass)" strokeWidth="1.6" />
      <line x1="7" y1="10.2" x2="11" y2="5.8" stroke="var(--pass)" strokeWidth="1.6" />
    </svg>
  );
}
