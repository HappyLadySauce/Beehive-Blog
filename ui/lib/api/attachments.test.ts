import { beforeEach, describe, expect, it, vi } from "vitest";

import { listAttachments, uploadLocalAttachment } from "./attachments";

describe("attachments API client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("serializes list filters into the BFF query", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 200, message: "success", data: { items: [] } }), {
        status: 200,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await listAttachments({ category_id: 7, limit: 20, purpose: "content", status: "active" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/bff/attachments?purpose=content&status=active&category_id=7&limit=20",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("does not force a JSON content type for multipart uploads", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ code: 200, message: "success", data: { id: 1 } }), {
        status: 200,
        headers: { "content-type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    const formData = new FormData();
    formData.set("file", new File(["hello"], "hello.txt", { type: "text/plain" }));
    await uploadLocalAttachment(formData);

    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(init.body).toBe(formData);
    expect(init.headers).not.toMatchObject({ "content-type": "application/json" });
  });
});
