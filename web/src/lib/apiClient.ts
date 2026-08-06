import { clearTokens, getAccessToken, getRefreshToken, setAccessToken } from "./auth";

// Porta do core-api em dev — fixa porque é a mesma em qualquer rede (mapeada
// pelo docker-compose, "9001:8080"); só o HOST varia (localhost, IP da LAN,
// Tailscale). Ver CORE_API_DEV_PORT.
const CORE_API_DEV_PORT = "9001";

// "localhost" hardcoded aqui quebrava o acesso via celular/rede local: o
// dispositivo cliente resolve "localhost" pra ELE MESMO, nunca pro servidor
// (Sec. pedida pelo usuário). Sem VITE_API_BASE_URL setado explicitamente
// (build de produção deve sempre setar), o default agora é relativo ao HOST
// com que a própria página foi carregada — abrir o app em
// http://192.168.0.106:5173 já faz as chamadas de API irem pra
// http://192.168.0.106:9001 sozinho, sem configuração manual por
// dispositivo. Funciona igual pra localhost, IP de LAN ou Tailscale, porque
// os 3 casos só diferem no hostname que o browser já usou pra carregar a
// página.
function defaultApiBaseUrl(): string {
  return `${window.location.protocol}//${window.location.hostname}:${CORE_API_DEV_PORT}`;
}

const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? defaultApiBaseUrl();

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

// Requisições concorrentes que tomam 401 ao mesmo tempo devem compartilhar a
// mesma chamada de refresh, não disparar uma pra cada uma.
let refreshInFlight: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    throw new ApiError(401, "sem refresh token");
  }

  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok) {
    throw new ApiError(response.status, "refresh token inválido");
  }

  const body = (await response.json()) as { access_token: string };
  setAccessToken(body.access_token);
  return body.access_token;
}

function redirectToLogin(): void {
  clearTokens();
  // window.location, ao contrário de <Link>/navigate() do React Router,
  // ignora o `basename` do BrowserRouter — precisa montar o caminho com
  // import.meta.env.BASE_URL manualmente (ver main.tsx) pra funcionar sob
  // /yamabiko/ em produção, não só na raiz.
  window.location.assign(`${import.meta.env.BASE_URL}login`);
}

// requestRaw concentra a lógica de auth + retry de refresh compartilhada por
// request() (respostas JSON) e requestBlob() (respostas binárias, ex: WAV do
// endpoint de áudio de referência) — só o parsing do corpo de sucesso difere
// entre as duas.
async function requestRaw(path: string, options: RequestInit = {}, isRetry = false): Promise<Response> {
  const token = getAccessToken();
  const headers = new Headers(options.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (options.body && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });

  if (response.status === 401 && !isRetry && !path.startsWith("/auth/")) {
    try {
      refreshInFlight ??= refreshAccessToken().finally(() => {
        refreshInFlight = null;
      });
      await refreshInFlight;
    } catch {
      redirectToLogin();
      throw new ApiError(401, "sessão expirada");
    }
    return requestRaw(path, options, true);
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({}) as { error?: string });
    throw new ApiError(response.status, body.error ?? `Erro ${response.status}`);
  }
  return response;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await requestRaw(path, options);
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

async function requestBlob(path: string, options: RequestInit = {}): Promise<Blob> {
  const response = await requestRaw(path, options);
  return response.blob();
}

export const api = {
  get: <T,>(path: string): Promise<T> => request<T>(path),
  post: <T,>(path: string, body?: unknown): Promise<T> =>
    request<T>(path, {
      method: "POST",
      body: body instanceof FormData ? body : body !== undefined ? JSON.stringify(body) : undefined,
    }),
  patch: <T,>(path: string, body?: unknown): Promise<T> =>
    request<T>(path, { method: "PATCH", body: body !== undefined ? JSON.stringify(body) : undefined }),
  getBlob: (path: string): Promise<Blob> => requestBlob(path),
};
