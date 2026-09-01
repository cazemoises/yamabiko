import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { useAuth } from "../auth/AuthContext";
import { getProfile, updateAppearance as patchAppearance } from "./api";

// 4 presets do design + hex customizado (Sec. pedida pelo usuário) — "mono"
// é um literal especial (não um hex): index.css tem
// :root[data-accent-preset="mono"] { --accent-base: var(--text) } pra ficar
// relativo ao tema (um hex fixo desapareceria no fundo escuro).
export const ACCENT_PRESETS = [
  { id: "mono", label: "Mono", value: "mono", swatch: "#23201B" },
  { id: "verde-agua", label: "Verde-água", value: "#2F9E8F", swatch: "#2F9E8F" },
  { id: "indigo", label: "Índigo", value: "#4B4A99", swatch: "#4B4A99" },
  { id: "ambar", label: "Âmbar", value: "#B98A2E", swatch: "#B98A2E" },
];

interface AppearanceContextValue {
  theme: string; // "" = segue o sistema, "light", "dark" — só leitura, ver nota abaixo
  accentColor: string; // "" = default (terracota), "mono", ou "#RRGGBB"
  setAccentColor: (accentColor: string) => void;
}

const AppearanceContext = createContext<AppearanceContextValue | undefined>(undefined);

// Chave compartilhada com o portal (mesma origem) e o Ascend — só assume
// 'light'/'dark', nunca 'system', pra manter o contrato simples entre os apps.
const SHARED_THEME_KEY = "theme";

function readSharedTheme(): "light" | "dark" | null {
  try {
    const v = localStorage.getItem(SHARED_THEME_KEY);
    return v === "light" || v === "dark" ? v : null;
  } catch {
    return null;
  }
}

function applyToDocument(theme: string, accentColor: string): void {
  const root = document.documentElement;
  if (theme) root.setAttribute("data-theme", theme);
  else root.removeAttribute("data-theme");

  if (accentColor === "mono") {
    root.setAttribute("data-accent-preset", "mono");
    root.style.removeProperty("--accent-base");
  } else {
    root.removeAttribute("data-accent-preset");
    if (accentColor) root.style.setProperty("--accent-base", accentColor);
    else root.style.removeProperty("--accent-base");
  }
}

// Aplica a preferência salva assim que o perfil carrega (1x por sessão
// autenticada) e otimisticamente a cada mudança do usuário — o PATCH vai
// pro backend em paralelo, sem esperar a resposta pra já trocar a cor na
// tela (mesma UX de outras preferências do app, ex: seleção de voz).
export function AppearanceProvider({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const [theme, setThemeState] = useState("");
  const [accentColor, setAccentColorState] = useState("");

  // Aplica o tema compartilhado (gravado pelo portal ou pelo Ascend) assim
  // que o app monta, antes mesmo do fetch do perfil — evita flash e serve
  // de fonte de verdade quando embutido no portal.
  useEffect(() => {
    const shared = readSharedTheme();
    if (shared) {
      setThemeState(shared);
      applyToDocument(shared, accentColor);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (status !== "authenticated") return;
    getProfile()
      .then((profile) => {
        // O valor compartilhado (se existir) tem prioridade sobre o backend —
        // é a fonte de verdade quando o app está embutido no portal.
        const shared = readSharedTheme();
        const t = shared ?? profile.theme ?? "";
        const a = profile.accent_color ?? "";
        setThemeState(t);
        setAccentColorState(a);
        applyToDocument(t, a);
      })
      .catch(() => {});
  }, [status]);

  // Reage a mudanças da chave compartilhada feitas por outro browsing
  // context de mesma origem (Portal, hoje a única fonte — antes também
  // podia ser o próprio Yamabiko ou o Ascend escrevendo, mas nenhum dos
  // dois tem mais toggle de tema próprio) — nunca dispara na aba que fez
  // a escrita, então não há risco de loop.
  useEffect(() => {
    function handleStorage(e: StorageEvent): void {
      if (e.key !== SHARED_THEME_KEY) return;
      const next = e.newValue === "light" || e.newValue === "dark" ? e.newValue : "";
      setThemeState(next);
      applyToDocument(next, accentColor);
    }
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, [accentColor]);

  function setAccentColor(next: string): void {
    setAccentColorState(next);
    applyToDocument(theme, next);
    void patchAppearance({ accent_color: next });
  }

  return (
    <AppearanceContext.Provider value={{ theme, accentColor, setAccentColor }}>
      {children}
    </AppearanceContext.Provider>
  );
}

export function useAppearance(): AppearanceContextValue {
  const ctx = useContext(AppearanceContext);
  if (!ctx) {
    throw new Error("useAppearance deve ser usado dentro de AppearanceProvider");
  }
  return ctx;
}
