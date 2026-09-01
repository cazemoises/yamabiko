import type { ReactNode } from "react";
import { useAuth } from "../../features/auth/AuthContext";
import { usePortalShellNav } from "./usePortalShellNav";

// Sidebar e bottom-nav próprios foram removidos — a navegação agora vive
// só na sidebar do Portal, alimentada por usePortalShellNav (ver
// PORTAL_SHELL_PROTOCOL.md no repo do portal). AppShell fica só como o
// gate de autenticação que decide o wrapper de layout do conteúdo.
export function AppShell({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  usePortalShellNav();

  if (status !== "authenticated") {
    return <div className="auth-content">{children}</div>;
  }

  return (
    <div className="app-shell">
      <div className="app-content">{children}</div>
    </div>
  );
}
