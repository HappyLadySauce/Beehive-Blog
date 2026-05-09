import { apiFetch } from "./client";
import type { AuthPayload, GithubOAuthBeginResponse, LoginRequest, RegisterRequest } from "./types";

export function login(payload: LoginRequest) {
  return apiFetch<AuthPayload>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function register(payload: RegisterRequest) {
  return apiFetch<AuthPayload>("/users/register", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function refresh(refreshToken: string) {
  return apiFetch<AuthPayload>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken })
  });
}

export function logout(accessToken: string) {
  return apiFetch<null>("/auth/logout", {
    method: "POST",
    headers: {
      authorization: `Bearer ${accessToken}`
    }
  });
}

export function beginGithubOAuth() {
  return apiFetch<GithubOAuthBeginResponse>("/auth/github/authorize", {
    method: "GET"
  });
}
