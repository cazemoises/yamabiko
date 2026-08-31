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
  body?: Record<string, unknown>;

  constructor(status: number, message: string, body?: Record<string, unknown>) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

// Sem Authorization/refresh: identidade vem dos headers que o Pangolin
// injeta pro core-api (não pro browser), então o único requisito daqui é
// mandar `credentials: "include"` pra requisição carregar o cookie de
// sessão do Pangolin — sem ele o Pangolin não teria como reconhecer quem
// está logado e não injetaria Remote-Email (ver
// core-api/internal/auth/context.go). 401 aqui significa mesmo "sem
// identidade válida", não "token expirado" — não há retry possível do lado
// do browser.
async function requestRaw(path: string, options: RequestInit = {}): Promise<Response> {
  const headers = new Headers(options.headers);
  if (options.body && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers, credentials: "include" });

  if (!response.ok) {
    const body = await response.json().catch(() => ({}) as { error?: string });
    throw new ApiError(response.status, body.error ?? `Erro ${response.status}`, body);
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
