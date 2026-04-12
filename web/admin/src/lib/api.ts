import type { AuthSession, CreateLinkInput, Link, LinkAnalytics, UpdateLinkInput } from "../types";

const ADMIN_API_PREFIX = "/admin/api/v1";

type ApiEnvelope<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${ADMIN_API_PREFIX}${path}`, {
    credentials: "include",
    ...init,
    headers: {
      Accept: "application/json",
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });

  const body = (await response.json().catch(() => ({
    success: false,
    error: "invalid_response",
  }))) as ApiEnvelope<T>;

  if (!response.ok || body.success === false) {
    throw new ApiError(body.error ?? "request_failed", response.status);
  }

  return body.data as T;
}

export function getSession() {
  return request<AuthSession>("/auth/session");
}

export function login(username: string, password: string) {
  return request<AuthSession>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return request<{ ok: boolean }>("/auth/logout", {
    method: "POST",
  });
}

export function listLinks() {
  return request<Link[]>("/links");
}

export function createLink(payload: CreateLinkInput) {
  return request<Link>("/links", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateLink(id: number, payload: UpdateLinkInput) {
  return request<Link>(`/links/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function getLinkAnalytics(id: number, days: number) {
  return request<LinkAnalytics>(`/links/${id}/analytics?days=${days}`);
}
