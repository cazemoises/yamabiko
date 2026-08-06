import { createContext, useContext, useState, type ReactNode } from "react";
import { api } from "../../lib/apiClient";
import { saveTokens, clearTokens, isAuthenticated as checkIsAuthenticated, type TokenPair } from "../../lib/auth";

interface AuthContextValue {
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  pinLogin: (userId: string, pin: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [authenticated, setAuthenticated] = useState(checkIsAuthenticated());

  async function login(email: string, password: string): Promise<void> {
    const tokens = await api.post<TokenPair>("/auth/login", { email, password });
    saveTokens(tokens);
    setAuthenticated(true);
  }

  async function register(email: string, password: string, name: string): Promise<void> {
    const tokens = await api.post<TokenPair>("/auth/register", { email, password, name });
    saveTokens(tokens);
    setAuthenticated(true);
  }

  async function pinLogin(userId: string, pin: string): Promise<void> {
    const tokens = await api.post<TokenPair>("/auth/pin-login", { user_id: userId, pin });
    saveTokens(tokens);
    setAuthenticated(true);
  }

  function logout(): void {
    clearTokens();
    setAuthenticated(false);
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated: authenticated, login, register, pinLogin, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth deve ser usado dentro de AuthProvider");
  }
  return ctx;
}
