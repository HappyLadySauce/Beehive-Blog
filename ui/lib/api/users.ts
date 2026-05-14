import { apiFetch } from "./client";
import type {
  CreateUserRequest,
  CreateUserResponse,
  DeleteUserResponse,
  ListUsersRequest,
  ListUsersResponse,
  UpdateUserRequest,
  UserDetailResponse
} from "./types";

export function listUsers(params?: ListUsersRequest) {
  const sp = new URLSearchParams();
  if (params?.page) sp.set("page", String(params.page));
  if (params?.page_size) sp.set("page_size", String(params.page_size));
  if (params?.status) sp.set("status", params.status);
  if (params?.role) sp.set("role", params.role);
  if (params?.search) sp.set("search", params.search);
  const qs = sp.toString();
  return apiFetch<ListUsersResponse>(`/bff/users${qs ? `?${qs}` : ""}`, { method: "GET" });
}

export function getUser(id: number) {
  return apiFetch<UserDetailResponse>(`/bff/users/${id}`, { method: "GET" });
}

export function createUser(payload: CreateUserRequest) {
  return apiFetch<CreateUserResponse>("/bff/users", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateUser(id: number, payload: UpdateUserRequest) {
  return apiFetch<UserDetailResponse>(`/bff/users/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteUser(id: number) {
  return apiFetch<DeleteUserResponse>(`/bff/users/${id}`, { method: "DELETE" });
}
