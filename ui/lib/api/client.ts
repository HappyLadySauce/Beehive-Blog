import type { BaseResponse } from "./types";

export class ApiError extends Error {
  readonly status: number;
  readonly code: number;

  constructor(message: string, status: number, code = status) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api/v1";

function resolveApiUrl(path: string) {
  if (path.startsWith("/bff/")) {
    return `/api${path}`;
  }
  return `${apiBaseUrl}${path}`;
}

export async function parseBaseResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    throw new ApiError("API returned an invalid response", response.status, response.status);
  }

  let envelope: BaseResponse<T>;
  try {
    envelope = (await response.json()) as BaseResponse<T>;
  } catch {
    throw new ApiError("API returned an invalid response", response.status, response.status);
  }

  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    throw new ApiError(envelope.message || "Request failed", response.status, envelope.code);
  }

  return envelope.data;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const isFormData = init?.body instanceof FormData;
  const response = await fetch(resolveApiUrl(path), {
    ...init,
    headers: {
      ...(isFormData ? {} : { "content-type": "application/json" }),
      ...init?.headers
    }
  });

  return parseBaseResponse<T>(response);
}

export function humanizeApiError(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return "请求失败，请稍后再试。";
  }

  if (error.status === 401) return "账号或凭证无效，请重新确认。";
  if (error.status === 403) return "当前账号暂不可登录。";
  if (error.status === 409) return error.message || "信息已被占用。";
  if (error.status === 429) return "操作过于频繁，请稍后再试。";
  if (error.status >= 500) return "服务暂时不可用，请稍后再试。";
  return error.message || "请求失败，请检查输入。";
}
