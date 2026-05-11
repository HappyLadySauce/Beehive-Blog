import { apiFetch } from "./client";
import type { SettingsPatchRequest, SettingsResponse } from "./types";

export function getSettings() {
  return apiFetch<SettingsResponse>("/bff/settings", {
    method: "GET"
  });
}

export function patchSettings(payload: SettingsPatchRequest) {
  return apiFetch<SettingsResponse>("/bff/settings", {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}
