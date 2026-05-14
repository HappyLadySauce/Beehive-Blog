import { beforeEach, describe, expect, it, vi } from "vitest";

import { getGithubOAuth2Settings, getSettings, patchGithubOAuth2Settings, patchSettings, testEmailSettings } from "./settings";

describe("settings API client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("loads settings through the BFF route", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ revision: 1, email: { enabled: false } }));
    vi.stubGlobal("fetch", fetchMock);

    await getSettings();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/settings",
      expect.objectContaining({
        method: "GET"
      })
    );
  });

  it("does not add a password when the caller omits it", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ revision: 2, email: { enabled: true } }));
    vi.stubGlobal("fetch", fetchMock);

    await patchSettings({ email: { enabled: true, host: "smtp.example.com" } });

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(init.body))).toEqual({ email: { enabled: true, host: "smtp.example.com" } });
  });

  it("sends SMTP test requests through the BFF route", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ recipient: "reader@example.com" }));
    vi.stubGlobal("fetch", fetchMock);

    await testEmailSettings({ recipient: "reader@example.com" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/settings/email/test",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ recipient: "reader@example.com" })
      })
    );
  });

  it("loads GitHub OAuth2 settings through the BFF route", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ revision: 3, github_oauth2: { enabled: false } }));
    vi.stubGlobal("fetch", fetchMock);

    await getGithubOAuth2Settings();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/settings/github-oauth2",
      expect.objectContaining({
        method: "GET"
      })
    );
  });

  it("patches GitHub OAuth2 settings through the BFF route", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ revision: 4, github_oauth2: { enabled: true } }));
    vi.stubGlobal("fetch", fetchMock);

    await patchGithubOAuth2Settings({ enabled: true, client_id: "client-id" });

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(fetchMock.mock.calls[0][0]).toBe("/api/bff/settings/github-oauth2");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(String(init.body))).toEqual({ enabled: true, client_id: "client-id" });
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
