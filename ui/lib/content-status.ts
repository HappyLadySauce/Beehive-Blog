import type { ContentStatus } from "./api/types";

/** Allowed single-step content status transitions (mirrors backend). */
/** 允许的单步内容状态流转（与后端一致）。 */
const ALLOWED_TRANSITIONS: Record<string, ContentStatus[]> = {
  draft: ["review", "archived"],
  review: ["published", "draft"],
  published: ["archived"],
  archived: ["draft"]
};

/**
 * Plans the shortest sequence of statuses from `from` to `to`.
 * Returns an empty array when already at the target or no path exists.
 * planStatusTransition 规划从 from 到 to 的最短状态序列；已在目标或无路径时返回空数组。
 */
export function planStatusTransition(from: string, to: string): ContentStatus[] {
  if (from === to) {
    return [];
  }

  const queue: { status: string; path: ContentStatus[] }[] = [{ status: from, path: [] }];
  const visited = new Set<string>([from]);

  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) {
      break;
    }

    for (const next of ALLOWED_TRANSITIONS[current.status] ?? []) {
      const path = [...current.path, next];
      if (next === to) {
        return path;
      }
      if (!visited.has(next)) {
        visited.add(next);
        queue.push({ status: next, path });
      }
    }
  }

  return [];
}
