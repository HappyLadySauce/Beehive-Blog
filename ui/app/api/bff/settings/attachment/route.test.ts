import { beforeEach, describe, expect, it, vi } from "vitest";

const cookieValues = vi.hoisted(() => ({
  access: undefined as string | undefined,
  refresh: undefined as string | undefined
}));

vi.mock("next/headers", () => ({
  cookies: vi.fn(async () => ({
    get(name: string) {
      if (name === "beehive.access" && cookieValues.access) return { value: cookieValues.access };
      if (name === "beehive.refresh" && cookieValues.refresh) return { value: cookieValues.refresh };
      return undefined;
    }
  }))
}));

import { GET, PATCH } from "./route";

const attachmentPayload = {
  revision: 11,
  attachment: {
    default_storage: "local",
    local_root: "data/attachments",
    max_bytes: 10485760,
    allowed_mime_prefixes: ["image/", "application/pdf"],
    presign_ttl_seconds: 900,
    s3: { bucket: "", upload_base_url: "", download_base_url: "" },
    oss: { bucket: "", upload_base_url: "", download_base_url: "" }
  }
};

describe("BFF settings attachment route", () => {
  beforeEach(() => {
    cookieValues.access = "access-token";
    cookieValues.refresh = undefined;
    vi.restoreAllMocks();
  });

  it("forwards GET attachment settings with the HttpOnly access cookie", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(attachmentPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await GET();
    const body = await response.json();

    expect(response.status).toBe(200);
    expect(body.data).toEqual(attachmentPayload);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/settings/attachment",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ authorization: "Bearer access-token" })
      })
    );
  });

  it("forwards PATCH attachment settings body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(attachmentPayload));
    vi.stubGlobal("fetch", fetchMock);

    const response = await PATCH(
      new Request("http://localhost/api/bff/settings/attachment", {
        method: "PATCH",
        body: JSON.stringify({ default_storage: "s3", max_bytes: 20971520 })
      })
    );

    expect(response.status).toBe(200);
    const forwarded = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(forwarded.body))).toEqual({ default_storage: "s3", max_bytes: 20971520 });
  });
});

function jsonResponse(data: unknown) {
  return new Response(JSON.stringify({ code: 200, message: "success", data }), {
    status: 200,
    headers: { "content-type": "application/json" }
  });
}
