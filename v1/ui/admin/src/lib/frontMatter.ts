import YAML from 'yaml';

/**
 * 与 Go 端 `pkg/markdownfrontmatter.SplitFrontMatter` / `archives.splitFrontMatter` 行为一致。
 */
export function splitFrontMatter(raw: string): { yamlPart: string; body: string } | null {
  let s = raw;
  if (s.charCodeAt(0) === 0xfeff) {
    s = s.slice(1);
  }
  s = s.trimStart();
  if (!s.startsWith('---')) {
    return null;
  }
  s = s.slice(3);
  while (s.length > 0 && (s[0] === '\r' || s[0] === '\n')) {
    s = s.slice(1);
  }
  const idx = s.indexOf('\n---');
  if (idx < 0) {
    return null;
  }
  const yamlPart = s.slice(0, idx).trim();
  let body = s.slice(idx + 4);
  while (body.length > 0 && (body[0] === '\r' || body[0] === '\n')) {
    body = body.slice(1);
  }
  if (!yamlPart) {
    return null;
  }
  return { yamlPart, body };
}

export type ArticleMatterInput = {
  title: string;
  slug: string;
  summary: string;
  status: string;
  /** 分类显示名，无则为空数组 */
  categoryNames: string[];
  /** 标签显示名 */
  tagNames: string[];
  /** RFC3339，可选写入 date */
  publishedAt?: string;
};

/**
 * 由当前表单状态生成与导入/导出语义一致的 YAML 文档（不含外层 ---）。
 */
export function buildMatterDocument(input: ArticleMatterInput): Record<string, unknown> {
  const doc: Record<string, unknown> = {
    title: input.title.trim(),
    status: input.status,
  };
  const slug = input.slug.trim();
  if (slug) {
    doc.slug = slug;
  }
  const summary = input.summary.trim();
  if (summary) {
    doc.summary = summary;
  }
  if (input.categoryNames.length > 0) {
    doc.categories = input.categoryNames;
  }
  if (input.tagNames.length > 0) {
    doc.tags = input.tagNames;
  }
  if (input.publishedAt?.trim()) {
    const d = new Date(input.publishedAt.trim());
    if (!Number.isNaN(d.getTime())) {
      doc.date = d.toISOString();
    }
  }
  return doc;
}

/**
 * 组装写入 API 的完整 `articles.content`：单段 Front Matter + 正文。
 */
export function buildArticleContent(matter: Record<string, unknown>, body: string): string {
  const yml = YAML.stringify(matter).trimEnd();
  return `---\n${yml}\n---\n\n${body}`;
}
