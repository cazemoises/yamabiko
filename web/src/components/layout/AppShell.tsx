import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { useAuth } from "../../features/auth/AuthContext";

const NAV_ITEMS = [
  { to: "/", label: "Home", icon: HomeIcon, end: true },
  { to: "/scenarios", label: "Cenários", icon: ScenariosIcon, end: false },
  { to: "/exercises", label: "Exercícios", icon: ExercisesIcon, end: false },
  { to: "/dashboard", label: "Progresso", icon: ProgressIcon, end: false },
  { to: "/settings/voice", label: "Voz", icon: VoiceIcon, end: false },
];

// AppShell substitui a antiga <nav className="navbar">: mesmo componente,
// mesma lista de itens (NAV_ITEMS) e mesma lógica de navegação em mobile e
// desktop — só a MARCAÇÃO do container muda (bottom nav vs. sidebar) e
// isso é resolvido por CSS (breakpoint ~1024px em index.css), não por 2
// implementações React separadas. Os dois containers ficam sempre no DOM;
// display:none esconde o que não é da faixa de largura atual.
export function AppShell({ children }: { children: ReactNode }) {
  const { isAuthenticated, logout } = useAuth();

  if (!isAuthenticated) {
    return <div className="auth-content">{children}</div>;
  }

  return (
    <div className="app-shell">
      <nav className="sidebar-nav" aria-label="Navegação principal">
        <span className="sidebar-brand">やまびこ</span>
        <div className="sidebar-nav-items">
          {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) => (isActive ? "sidebar-nav-item active" : "sidebar-nav-item")}
            >
              <Icon />
              <span>{label}</span>
            </NavLink>
          ))}
        </div>
        <button type="button" className="sidebar-nav-item sidebar-nav-exit" onClick={logout}>
          <ExitIcon />
          <span>Sair</span>
        </button>
      </nav>

      <div className="app-content">{children}</div>

      <nav className="bottom-nav" aria-label="Navegação principal">
        {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) => (isActive ? "bottom-nav-item active" : "bottom-nav-item")}
          >
            <Icon />
            <span>{label}</span>
          </NavLink>
        ))}
        <button type="button" className="bottom-nav-item" onClick={logout}>
          <ExitIcon />
          <span>Sair</span>
        </button>
      </nav>
    </div>
  );
}

function HomeIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <rect x="3" y="3" width="12" height="12" rx="2.5" fill="none" strokeWidth="1.5" />
    </svg>
  );
}

function ScenariosIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <rect x="2" y="2" width="6" height="6" rx="1.4" fill="none" strokeWidth="1.4" />
      <rect x="10" y="2" width="6" height="6" rx="1.4" fill="none" strokeWidth="1.4" />
      <rect x="2" y="10" width="6" height="6" rx="1.4" fill="none" strokeWidth="1.4" />
      <rect x="10" y="10" width="6" height="6" rx="1.4" fill="none" strokeWidth="1.4" />
    </svg>
  );
}

function ExercisesIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <circle cx="9" cy="7" r="4" fill="none" strokeWidth="1.4" />
      <line x1="9" y1="11" x2="9" y2="15" strokeWidth="1.4" />
    </svg>
  );
}

function ProgressIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <rect x="2" y="8" width="3" height="8" className="nav-icon-fill" stroke="none" />
      <rect x="7.5" y="4" width="3" height="12" className="nav-icon-fill" stroke="none" />
      <rect x="13" y="10" width="3" height="6" className="nav-icon-fill" stroke="none" />
    </svg>
  );
}

function VoiceIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <circle cx="9" cy="9" r="3" fill="none" strokeWidth="1.4" />
      <circle cx="9" cy="9" r="7" fill="none" strokeWidth="1" />
    </svg>
  );
}

function ExitIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18">
      <path d="M7 3H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3" fill="none" strokeWidth="1.4" />
      <path d="M11 6l4 3-4 3M15 9H6" fill="none" strokeWidth="1.4" />
    </svg>
  );
}
