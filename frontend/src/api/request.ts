import { env } from '../lib/env';

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly statusText: string,
  ) {
    super(`Request failed: ${String(status)} ${statusText}`.trim());
    this.name = 'ApiError';
  }
}

// Orval supplies the response type from the OpenAPI contract.
export async function request<T>(
  path: string,
  options: RequestInit,
): Promise<T> {
  const response = await fetch(
    `${env.apiBaseUrl.replace(/\/$/, '')}${path}`,
    options,
  );

  if (!response.ok) {
    throw new ApiError(response.status, response.statusText);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}
