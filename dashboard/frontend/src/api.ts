import type { ChainStatus, Deployment, TeleporterInfo, Validator } from "./types";

const API_BASE_URL: string =
  (import.meta.env.VITE_API_BASE_URL as string | undefined) ??
  "http://localhost:8080";

class ApiRequestError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ApiRequestError";
  }
}

async function request<T>(path: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`);
  } catch {
    throw new ApiRequestError("Backend'e ulaşılamıyor");
  }

  if (!response.ok) {
    throw new ApiRequestError(`İstek başarısız (${response.status})`);
  }

  try {
    return (await response.json()) as T;
  } catch {
    throw new ApiRequestError("Yanıt ayrıştırılamadı");
  }
}

export async function fetchHealth(): Promise<{ status: string }> {
  return request<{ status: string }>("/api/health");
}

export async function fetchDeployments(): Promise<Deployment[]> {
  return request<Deployment[]>("/api/deployments");
}

export async function fetchChainStatus(
  chainName: string
): Promise<ChainStatus> {
  return request<ChainStatus>(
    `/api/status/${encodeURIComponent(chainName)}`
  );
}

export async function fetchValidators(
  chainName: string
): Promise<Validator[]> {
  return request<Validator[]>(
    `/api/validators/${encodeURIComponent(chainName)}`
  );
}

export async function fetchTeleporterInfo(
  chainName: string
): Promise<TeleporterInfo> {
  return request<TeleporterInfo>(
    `/api/teleporter/${encodeURIComponent(chainName)}`
  );
}

export { ApiRequestError };