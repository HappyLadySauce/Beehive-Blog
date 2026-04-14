import { useAuthStore } from '../store/authStore';
import type { ArticleDetailResponse, UpdateArticleRequest } from '../api/article';

/** 与后端 cmd/app/routes/ws 对齐 */
export type ArticleAutosaveWsResult = {
  code: number;
  message?: string;
  data?: ArticleDetailResponse;
};

type Pending = {
  resolve: (v: ArticleAutosaveWsResult) => void;
  reject: (e: Error) => void;
};

const pending = new Map<string, Pending>();

let socket: WebSocket | null = null;
let connectPromise: Promise<WebSocket> | null = null;
let pingTimer: ReturnType<typeof setInterval> | null = null;

/**
 * 与 axios（request.ts）一致：有 VITE_API_BASE_URL 则直连 API；否则用当前页面 origin（依赖 Vite proxy 的 ws: true 转发到后端）。
 */
function buildWsOrigin(): string {
  const base = import.meta.env.VITE_API_BASE_URL || '';
  if (base) {
    try {
      const u = new URL(base, typeof window !== 'undefined' ? window.location.href : undefined);
      const wsProto = u.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${wsProto}//${u.host}`;
    } catch {
      /* fall through */
    }
  }
  if (typeof window === 'undefined') {
    return 'ws://localhost';
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}`;
}

function wsURL(): string {
  const token = useAuthStore.getState().token;
  if (!token) {
    throw new Error('未登录，无法建立 WebSocket');
  }
  return `${buildWsOrigin()}/api/v1/ws?token=${encodeURIComponent(token)}`;
}

function stopPing() {
  if (pingTimer != null) {
    clearInterval(pingTimer);
    pingTimer = null;
  }
}

function startPing(ws: WebSocket) {
  stopPing();
  pingTimer = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify({ type: 'ping' }));
      } catch {
        /* ignore */
      }
    }
  }, 60_000);
}

function attachHandlers(ws: WebSocket) {
  ws.onmessage = (ev) => {
    let msg: {
      type?: string;
      requestId?: string;
      code?: number;
      message?: string;
      data?: ArticleDetailResponse;
    };
    try {
      msg = JSON.parse(String(ev.data)) as typeof msg;
    } catch {
      return;
    }
    const t = msg.type;
    if (t === 'pong') {
      return;
    }
    if (t === 'error') {
      const id = msg.requestId;
      if (id && pending.has(id)) {
        const p = pending.get(id)!;
        pending.delete(id);
        p.resolve({
          code: msg.code ?? 400,
          message: msg.message,
        });
      }
      return;
    }
    if (t === 'article.autosave.result') {
      const id = msg.requestId;
      if (id && pending.has(id)) {
        const p = pending.get(id)!;
        pending.delete(id);
        p.resolve({
          code: msg.code ?? 500,
          message: msg.message,
          data: msg.data,
        });
      }
    }
  };

  ws.onclose = () => {
    socket = null;
    connectPromise = null;
    stopPing();
    for (const [, p] of pending) {
      p.reject(new Error('WebSocket 已断开'));
    }
    pending.clear();
  };
}

/**
 * 确保管理端 WebSocket 已连接（带 access token），供自动保存等高频操作使用。
 */
export async function ensureAdminWebSocket(): Promise<WebSocket> {
  if (socket?.readyState === WebSocket.OPEN) {
    return socket;
  }
  if (connectPromise) {
    return connectPromise;
  }
  connectPromise = new Promise<WebSocket>((resolve, reject) => {
    let ws: WebSocket;
    try {
      ws = new WebSocket(wsURL());
    } catch (e) {
      connectPromise = null;
      reject(e instanceof Error ? e : new Error(String(e)));
      return;
    }
    ws.onopen = () => {
      socket = ws;
      connectPromise = null;
      attachHandlers(ws);
      startPing(ws);
      resolve(ws);
    };
    ws.onerror = () => {
      connectPromise = null;
      reject(new Error('WebSocket 连接失败'));
    };
  });
  return connectPromise;
}

/**
 * 通过 WebSocket 发送文章自动保存请求，语义对齐 PUT /api/v1/admin/articles/:id（含 autoSave）。
 * 连接意外断开时关闭并重连后重试一次，便于 token 刷新后恢复。
 */
export async function sendArticleAutosave(
  articleId: number,
  payload: UpdateArticleRequest,
  requestId: string,
): Promise<ArticleAutosaveWsResult> {
  const once = (): Promise<ArticleAutosaveWsResult> =>
    ensureAdminWebSocket().then(
      (ws) =>
        new Promise<ArticleAutosaveWsResult>((resolve, reject) => {
          pending.set(requestId, { resolve, reject });
          try {
            ws.send(
              JSON.stringify({
                type: 'article.autosave',
                requestId,
                articleId,
                payload,
              }),
            );
          } catch (e) {
            pending.delete(requestId);
            reject(e instanceof Error ? e : new Error(String(e)));
          }
        }),
    );

  try {
    return await once();
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    if (/WebSocket|disconnect|连接|not authenticated|未登录/i.test(msg)) {
      closeAdminWebSocket();
      return once();
    }
    throw e;
  }
}

/** 编辑页卸载时可调用，关闭连接释放服务端配额。 */
export function closeAdminWebSocket() {
  stopPing();
  if (socket) {
    try {
      socket.close();
    } catch {
      /* ignore */
    }
    socket = null;
  }
  connectPromise = null;
}
