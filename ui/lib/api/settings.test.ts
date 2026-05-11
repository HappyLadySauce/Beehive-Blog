import { beforeEach, describe, expect, it, vi } from "vitest";

import { getSettings, patchSettings } from "./settings";

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
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
