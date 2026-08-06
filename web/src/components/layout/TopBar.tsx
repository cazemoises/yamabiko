import { useNavigate } from "react-router-dom";

interface TopBarProps {
  /** "Exercício N de M" + barra de progresso — frames de exercício/resultado dentro de cenário. */
  progress?: { current: number; total: number };
  backTo?: string;
}

// Cabeçalho compartilhado pelas telas de exercício/resultado (Frames
// 4/5/9-19 do design): seta de voltar + barra de progresso do cenário
// quando aplicável. O design original ora combina os dois na mesma linha
// (frames 16-19), ora separa em 2 linhas (frames 4/9-15) — normalizado
// aqui pra sempre a mesma linha, mais simples e sem perda visual.
export function TopBar({ progress, backTo }: TopBarProps) {
  const navigate = useNavigate();

  return (
    <div className="topbar">
      <button
        type="button"
        className="topbar-back"
        aria-label="Voltar"
        onClick={() => (backTo ? navigate(backTo) : navigate(-1))}
      >
        <svg width="18" height="18" viewBox="0 0 18 18">
          <polygon points="12,3 6,9 12,15" fill="none" stroke="currentColor" strokeWidth="1.6" />
        </svg>
      </button>

      {progress && (
        <div className="topbar-progress">
          <span className="topbar-progress-label">
            Exercício {progress.current} de {progress.total}
          </span>
          <div className="progress-track">
            <div className="progress-fill" style={{ width: `${(progress.current / progress.total) * 100}%` }} />
          </div>
        </div>
      )}
    </div>
  );
}
