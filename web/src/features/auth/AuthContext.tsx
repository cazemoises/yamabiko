import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api } from "../../lib/apiClient";

// Sem login algum no app: a identidade vem inteiramente dos headers
// Remote-Email/Remote-Name que o Pangolin injeta antes de encaminhar a
// requisição pro core-api (Sec. pedida pelo usuário — "não quero mais
// login com pin ou email e senha, quero usar os headers do pangolin"). O
// browser não vê esses headers (só existem no trecho Pangolin -> core-api),
// então o único jeito de saber se a sessão é válida é perguntar pro
// backend — daí este contexto só chamar GET /users/me uma vez ao montar.
type AuthStatus = "checking" | "authenticated" | "unauthenticated";

interface AuthContextValue {
  status: AuthStatus;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>("checking");

  useEffect(() => {
    api
      .get("/users/me")
      .then(() => setStatus("authenticated"))
      .catch(() => setStatus("unauthenticated"));
  }, []);

  return <AuthContext.Provider value={{ status }}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth deve ser usado dentro de AuthProvider");
  }
  return ctx;
}
