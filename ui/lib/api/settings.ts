import { apiFetch } from "./client";
import type {
  AttachmentPatch,
  GithubOAuth2Patch,
  SettingsEmailTestRequest,
  SettingsEmailTestResponse,
  SettingsPatchRequest,
  SettingsResponse
} from "./types";

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

export function getGithubOAuth2Settings() {
  return apiFetch<SettingsResponse>("/bff/settings/github-oauth2", {
    method: "GET"
  });
}

export function patchGithubOAuth2Settings(payload: GithubOAuth2Patch) {
  return apiFetch<SettingsResponse>("/bff/settings/github-oauth2", {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function getAttachmentSettings() {
  return apiFetch<SettingsResponse>("/bff/settings/attachment", {
    method: "GET"
  });
}

export function patchAttachmentSettings(payload: AttachmentPatch) {
  return apiFetch<SettingsResponse>("/bff/settings/attachment", {
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
