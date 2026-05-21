import { pinyin } from "pinyin-pro";

export type SlugifyOptions = {
  maxLength?: number;
};

const DEFAULT_MAX_LENGTH = 64;
const CONTENT_MAX_LENGTH = 200;

const CJK_SEGMENT = /^[\u4e00-\u9fff]+$/;
const SEGMENT_PATTERN = /[\u4e00-\u9fff]+|[^\u4e00-\u9fff]+/g;

/**
 * Builds a URL-safe ASCII slug from a display name (title or tag/category name).
 * Chinese segments are converted with pinyin-pro; Latin segments become kebab-case.
 * 从展示名称生成 URL 安全的 ASCII slug；中文转拼音，拉丁字符转为 kebab-case。
 */
export function slugifyFromName(name: string, options?: SlugifyOptions): string {
  const maxLength = options?.maxLength ?? DEFAULT_MAX_LENGTH;
  const trimmed = name.trim();
  if (!trimmed) {
    return truncateSlug(fallbackSlug(), maxLength);
  }

  const parts: string[] = [];
  const segments = trimmed.match(SEGMENT_PATTERN) ?? [];

  for (const segment of segments) {
    if (CJK_SEGMENT.test(segment)) {
      parts.push(...pinyin(segment, { toneType: "none", type: "array" }));
      continue;
    }
    const normalized = segment
      .normalize("NFKD")
      .replace(/[\u0300-\u036f]/g, "")
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-");
    if (normalized) {
      parts.push(...normalized.split("-").filter(Boolean));
    }
  }

  let slug = parts.join("-").replace(/-+/g, "-").replace(/^-+|-+$/g, "");
  if (!slug) {
    slug = fallbackSlug();
  }
  return truncateSlug(slug, maxLength);
}

/** Slug helper tuned for content titles (longer limit). / 内容标题专用 slug（较长上限）。 */
export function slugifyContentTitle(title: string): string {
  return slugifyFromName(title, { maxLength: CONTENT_MAX_LENGTH });
}

/** Slug helper for tags and content categories. / 标签与内容分类专用 slug。 */
export function slugifyTaxonomyName(name: string): string {
  return slugifyFromName(name, { maxLength: DEFAULT_MAX_LENGTH });
}

function fallbackSlug(): string {
  return `draft-${Date.now().toString(36)}`;
}

function truncateSlug(slug: string, maxLength: number): string {
  if (slug.length <= maxLength) {
    return slug;
  }
  const cut = slug.slice(0, maxLength).replace(/-+$/g, "");
  return cut || fallbackSlug().slice(0, maxLength);
}

export type SlugMode = "auto" | "manual";
