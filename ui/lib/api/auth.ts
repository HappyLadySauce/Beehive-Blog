import { apiFetch } from "./client";
import type { ClientSession } from "@/lib/auth/session";
import type { GithubOAuthBeginResponse, LoginRequest, RegisterRequest } from "./types";

export function login(payload: LoginRequest) {
  return apiFetch<ClientSession>("/bff/auth/login", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function register(payload: RegisterRequest) {
  return apiFetch<ClientSession>("/bff/auth/register", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function refresh() {
  return apiFetch<ClientSession>("/bff/auth/refresh", {
    method: "POST"
  });
}

export function logout() {
  return apiFetch<null>("/bff/auth/logout", {
    method: "POST"
  });
}

export function getSession() {
  return apiFetch<ClientSession>("/bff/auth/session", {
    method: "GET"
  });
}

export function beginGithubOAuth() {
  return apiFetch<GithubOAuthBeginResponse>("/auth/github/authorize", {
    method: "GET"
  });
}
