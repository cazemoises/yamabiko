import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../features/auth/AuthContext";

// Ponte com a sidebar do Portal — ver PORTAL_SHELL_PROTOCOL.md no repo do
// portal pro contrato completo. Substitui a navegação própria do AppShell
// (removida): o Yamabiko não desenha mais sua sidebar/bottom-nav, só
// informa o Portal do que navegar e reage quando o Portal manda navegar.
const NAV_CHANNEL = "portal-shell";

// Origin confiável do bridge: SEMPRE a própria origin do documento — nunca
// uma allowlist dos subdomínios standalone (yamabiko-app.duckdns.org
// etc.). Esses domínios nunca aparecem em event.origin quando o app está
// embutido no Portal: o Caddy serve Portal + apps sob a MESMA origin via
// path (/yamabiko/), nunca via subdomínio, nesse caso. Ver a seção
// "Validação de origin" do protocolo pro porquê disso ser obrigatório.
const TRUSTED_ORIGIN = window.location.origin;

const NAV_ITEMS = [
  {
    id: "home",
    label: "Home",
    path: "/",
    icon: '<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="12" height="12" rx="2.5"/></svg>',
  },
  {
    id: "scenarios",
    label: "Cenários",
    path: "/scenarios",
    icon: '<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="2" y="2" width="6" height="6" rx="1.4"/><rect x="10" y="2" width="6" height="6" rx="1.4"/><rect x="2" y="10" width="6" height="6" rx="1.4"/><rect x="10" y="10" width="6" height="6" rx="1.4"/></svg>',
  },
  {
    id: "exercises",
    label: "Exercícios",
    path: "/exercises",
    icon: '<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.4"><circle cx="9" cy="7" r="4"/><line x1="9" y1="11" x2="9" y2="15"/></svg>',
  },
  {
    id: "dashboard",
    label: "Progresso",
    path: "/dashboard",
    icon: '<svg viewBox="0 0 18 18" fill="currentColor" stroke="none"><rect x="2" y="8" width="3" height="8"/><rect x="7.5" y="4" width="3" height="12"/><rect x="13" y="10" width="3" height="6"/></svg>',
  },
  {
    id: "voice",
    label: "Voz",
    path: "/settings/voice",
    icon: '<svg viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.4"><circle cx="9" cy="9" r="3"/><circle cx="9" cy="9" r="7" stroke-width="1"/></svg>',
  },
] as const;

function activeIdForPath(pathname: string): string {
  if (pathname === "/") return "home";
  // do mais específico pro menos, já que um startsWith ingênuo faria
  // "/exercises/42" bater com qualquer prefixo antes de "/exercises".
  const match = [...NAV_ITEMS].reverse().find((item) => item.path !== "/" && pathname.startsWith(item.path));
  return match?.id ?? "home";
}

export function usePortalShellNav(): void {
  const { status } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    if (window.parent === window) return; // não está embutido, ninguém pra avisar

    // Não esperar status === "authenticated": a navegação contextual é
    // estrutura do app, não identidade — se /users/me demorar, falhar ou
    // nunca resolver, o Portal fica preso em "Carregando navegação…" pra
    // sempre (ver PORTAL_SHELL_PROTOCOL.md). Mesmo padrão do Ascend
    // (useAscendShellNav.ts), que posta incondicionalmente.
    window.parent.postMessage(
      {
        channel: NAV_CHANNEL,
        type: "nav:update",
        items: NAV_ITEMS.map(({ id, label, icon }) => ({ id, label, icon })),
        activeId: activeIdForPath(location.pathname),
      },
      TRUSTED_ORIGIN,
    );
  }, [status, location.pathname]);

  useEffect(() => {
    if (status !== "authenticated") return;

    function handleMessage(event: MessageEvent): void {
      if (event.origin !== TRUSTED_ORIGIN) return;
      if (event.source !== window.parent) return;
      const data = event.data as { channel?: string; type?: string; id?: string } | null;
      if (!data || data.channel !== NAV_CHANNEL || data.type !== "nav:go") return;
      const item = NAV_ITEMS.find((navItem) => navItem.id === data.id);
      if (item) navigate(item.path);
    }

    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [status, navigate]);
}
