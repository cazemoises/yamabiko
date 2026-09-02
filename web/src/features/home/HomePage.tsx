import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { LanguageToggle } from "../../components/layout/LanguageToggle";
import { getProfile, type Profile } from "../users/api";
import { useScenariosProgress } from "../scenarios/useScenariosProgress";
import { iconForScenario } from "../scenarios/scenarioIcon";

// Home (Frame 1/2 do design) — saudação, cartão "continue de onde parou"
// (o 1º cenário com progresso parcial) e uma tira "Em destaque" com os
// próximos cenários. Sem dado mockado: os contadores de progresso vêm de
// useScenariosProgress, que cruza GET /scenarios + GET /scenarios/{id} +
// GET /exercises/{id}/attempts (ver limitação documentada nesse hook).
export function HomePage() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [language, setLanguage] = useState("ja-JP");
  const { progress, loading, error } = useScenariosProgress(language);

  useEffect(() => {
    getProfile()
      .then(setProfile)
      .catch(() => null);
  }, []);

  const inProgress = progress.find((p) => p.completed > 0 && p.completed < p.total);
  const continueTarget = inProgress ?? progress.find((p) => p.completed === 0);
  const featured = progress.filter((p) => p.scenario.id !== continueTarget?.scenario.id).slice(0, 4);

  const firstName = profile?.name?.split(" ")[0] ?? "";

  return (
    <div className="page">
      <div className="page-header">
        <span className="page-title">yamabiko</span>
        <LanguageToggle language={language} onChange={setLanguage} />
      </div>

      <div>
        <h1 style={{ fontSize: 20 }}>{firstName ? `Bom dia, ${firstName}` : "Bom dia"}</h1>
        <p className="page-subtitle">Continue de onde parou</p>
      </div>

      {loading && <p className="center-message">Carregando...</p>}
      {error && <p className="error">{error}</p>}

      {continueTarget && (
        <Link
          to={`/scenarios/${continueTarget.scenario.id}`}
          className="card"
          style={{ background: "var(--accent-soft)", textDecoration: "none", display: "flex", flexDirection: "column", gap: 12 }}
        >
          <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
            <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: "0.06em", textTransform: "uppercase", color: "var(--accent-text)" }}>
              Cenário
            </span>
            <span style={{ fontSize: 12, opacity: 0.7 }}>
              {continueTarget.completed} de {continueTarget.total}
            </span>
          </div>
          <span style={{ fontSize: 17, fontWeight: 700 }}>{continueTarget.scenario.title_pt}</span>
          <div className="progress-track">
            <div
              className="progress-fill"
              style={{ width: `${continueTarget.total ? (continueTarget.completed / continueTarget.total) * 100 : 0}%` }}
            />
          </div>
          <span className="btn-primary" style={{ alignSelf: "flex-start", width: "auto" }}>
            continuar
          </span>
        </Link>
      )}

      {!loading && !error && progress.length === 0 && (
        <p className="center-message">Nenhum cenário disponível ainda nesse idioma.</p>
      )}

      {featured.length > 0 && (
        <>
          <div className="page-header">
            <span className="section-title">em destaque</span>
            <Link to="/scenarios" className="page-subtitle">
              ver todos
            </Link>
          </div>
          <div style={{ display: "flex", gap: 12, overflowX: "auto", paddingBottom: 4 }}>
            {featured.map(({ scenario, completed, total }) => {
              const Icon = iconForScenario(scenario.title_pt);
              return (
              <Link
                key={scenario.id}
                to={`/scenarios/${scenario.id}`}
                className="card"
                style={{ minWidth: 168, textDecoration: "none", display: "flex", flexDirection: "column", gap: 10 }}
              >
                <div className="list-row-icon">
                  <Icon size={18} strokeWidth={1.75} />
                </div>
                <span className="list-row-title" style={{ lineHeight: 1.3 }}>
                  {scenario.title_pt}
                </span>
                <span className="page-subtitle">
                  {scenario.language === "ja-JP" ? "JA" : "EN"} · {completed}/{total}
                </span>
              </Link>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}
