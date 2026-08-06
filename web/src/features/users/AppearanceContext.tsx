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
  { id: "terracota", label: "Terracota", value: "#C1662F", swatch: "#C1662F" },
  { id: "ambar", label: "Âmbar", value: "#B98A2E", swatch: "#B98A2E" },
];

interface AppearanceContextValue {
  theme: string; // "" = segue o sistema, "light", "dark"
  accentColor: string; // "" = default (terracota), "mono", ou "#RRGGBB"
  setTheme: (theme: string) => void;
  setAccentColor: (accentColor: string) => void;
}

const AppearanceContext = createContext<AppearanceContextValue | undefined>(undefined);

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
  const { isAuthenticated } = useAuth();
  const [theme, setThemeState] = useState("");
  const [accentColor, setAccentColorState] = useState("");

  useEffect(() => {
    if (!isAuthenticated) return;
    getProfile()
      .then((profile) => {
        const t = profile.theme ?? "";
        const a = profile.accent_color ?? "";
        setThemeState(t);
        setAccentColorState(a);
        applyToDocument(t, a);
      })
      .catch(() => {});
  }, [isAuthenticated]);

  function setTheme(next: string): void {
    setThemeState(next);
    applyToDocument(next, accentColor);
    void patchAppearance({ theme: next });
  }

  function setAccentColor(next: string): void {
    setAccentColorState(next);
    applyToDocument(theme, next);
    void patchAppearance({ accent_color: next });
  }

  return (
    <AppearanceContext.Provider value={{ theme, accentColor, setTheme, setAccentColor }}>
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
