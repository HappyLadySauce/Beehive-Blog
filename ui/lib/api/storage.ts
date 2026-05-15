import { apiFetch } from "./client";
import type {
  DeleteStorageMountResponse,
  DriverListResponse,
  DriverResponse,
  StorageMountCheckResponse,
  StorageMountCreateRequest,
  StorageMountListResponse,
  StorageMountPatchRequest,
  StorageMountResponse
} from "./types";

export function listFileDrivers() {
  return apiFetch<DriverListResponse>("/bff/file-drivers", { method: "GET" });
}

export function getFileDriver(name: string) {
  return apiFetch<DriverResponse>(`/bff/file-drivers/${encodeURIComponent(name)}`, { method: "GET" });
}

export function listStorageMounts() {
  return apiFetch<StorageMountListResponse>("/bff/storage-mounts", { method: "GET" });
}

export function createStorageMount(payload: StorageMountCreateRequest) {
  return apiFetch<StorageMountResponse>("/bff/storage-mounts", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateStorageMount(id: number, payload: StorageMountPatchRequest) {
  return apiFetch<StorageMountResponse>(`/bff/storage-mounts/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function enableStorageMount(id: number) {
  return apiFetch<StorageMountResponse>(`/bff/storage-mounts/${id}/enable`, { method: "POST" });
}

export function disableStorageMount(id: number) {
  return apiFetch<StorageMountResponse>(`/bff/storage-mounts/${id}/disable`, { method: "POST" });
}

export function checkStorageMount(id: number) {
  return apiFetch<StorageMountCheckResponse>(`/bff/storage-mounts/${id}/check`, { method: "POST" });
}

export function deleteStorageMount(id: number) {
  return apiFetch<DeleteStorageMountResponse>(`/bff/storage-mounts/${id}`, { method: "DELETE" });
}
