import { apiFetch } from "./client";
import type { SettingsEmailTestRequest, SettingsEmailTestResponse, SettingsPatchRequest, SettingsResponse } from "./types";

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

export function testEmailSettings(payload: SettingsEmailTestRequest) {
  return apiFetch<SettingsEmailTestResponse>("/bff/settings/email/test", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
