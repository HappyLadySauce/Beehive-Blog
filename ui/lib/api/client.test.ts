import { describe, expect, it } from "vitest";

import { ApiError, parseBaseResponse } from "./client";

describe("parseBaseResponse", () => {
  it("returns data from a successful Go API envelope", async () => {
    const response = new Response(
      JSON.stringify({
        code: 200,
        message: "success",
        data: { token: { access_token: "access", token_type: "Bearer", expires_in: 900 } }
      }),
      { status: 200, headers: { "content-type": "application/json" } }
    );

    await expect(parseBaseResponse<{ token: { access_token: string } }>(response)).resolves.toEqual({
      token: { access_token: "access", token_type: "Bearer", expires_in: 900 }
    });
  });

  it("throws ApiError with status and public message for failed envelopes", async () => {
    const response = new Response(
      JSON.stringify({ code: 429, message: "rate limited", data: null }),
      { status: 429, headers: { "content-type": "application/json" } }
    );

    await expect(parseBaseResponse(response)).rejects.toMatchObject({
      name: "ApiError",
      status: 429,
      code: 429,
      message: "rate limited"
    } satisfies Partial<ApiError>);
  });

  it("throws ApiError when the server returns non-json content", async () => {
    const response = new Response("upstream unavailable", {
      status: 502,
      headers: { "content-type": "text/plain" }
    });

    await expect(parseBaseResponse(response)).rejects.toMatchObject({
      status: 502,
      code: 502,
      message: "API returned an invalid response"
    } satisfies Partial<ApiError>);
  });
});
