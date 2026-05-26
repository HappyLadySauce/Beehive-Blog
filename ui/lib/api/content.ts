import type {
  BaseResponse,
  CreateCommentRequest,
  CreateCommentResponse,
  PublicContent,
  PublicContentKind,
  PublicPost,
  PublicReaderContext,
  PublicSiteOverview
} from "./types";

type PublicContentListResponse = {
  items: PublicContentItem[];
  total: number;
  page: number;
  page_size: number;
};

type PublicContentDetailResponse = PublicContentItem & {
  body?: string | null;
};

type PublicContentItem = {
  id: number;
  type: string;
  title: string;
  slug: string;
  excerpt?: string | null;
  published_at?: string | null;
  cover_url?: string;
  view_count?: number;
  author_username?: string;
  word_count?: number;
  reading_time_minutes?: number;
  tags?: Array<{ id?: number; name: string; slug?: string; color?: string | null; content_count?: number }>;
  categories?: Array<{ id?: number; name: string; slug?: string; content_count?: number }>;
  category?: { id?: number; name: string; slug?: string; content_count?: number } | null;
  created_at: string;
  updated_at: string;
};

class PublicContentApiError extends Error {
  readonly status: number;
  readonly code: number;

  constructor(message: string, status: number, code = status) {
    super(message);
    this.name = "PublicContentApiError";
    this.status = status;
    this.code = code;
  }
}

const fallbackPosts: PublicPost[] = [
  {
    id: 1,
    slug: "ai-assisted-writing-loop",
    title: "AI 协作写作回路",
    description: "从零散笔记到可审阅文章，让 AI 参与整理，但把发布权保留给创作者。",
    body:
      "Beehive Blog 的核心不是把内容交给模型代写，而是把长期积累的文章、笔记和项目组织成可追溯的知识材料。AI 可以帮助摘要、改写和发现关系，但每一次公开发布都需要人的审阅。这样的流程让创作者保留判断力，也让公开内容保持稳定可信。",
    publishedAt: "2026-05-09T00:00:00.000Z",
    tags: ["AI", "Writing", "Workflow"],
    categories: [{ id: 1, name: "AI", slug: "ai", content_count: 1 }],
    category: { id: 1, name: "AI", slug: "ai", content_count: 1 },
    viewCount: 2805,
    readingMinutes: 4
  },
  {
    id: 2,
    slug: "public-web-and-studio",
    title: "Public Web 与 Studio 的产品边界",
    description: "读者侧追求清晰阅读和 SEO，创作者侧追求高效管理、版本和权限。",
    body:
      "Public Web 面向访客，应该尽快呈现文章、笔记、项目和搜索入口。Studio 面向创作者，应该承担草稿、关系、版本、附件和发布闸门。两侧共享同一套内容真相源，但用状态、可见性和 AI 访问策略区分不同使用场景。",
    publishedAt: "2026-05-08T00:00:00.000Z",
    tags: ["Product", "Studio", "SEO"],
    categories: [{ id: 2, name: "项目", slug: "project", content_count: 1 }],
    category: { id: 2, name: "项目", slug: "project", content_count: 1 },
    viewCount: 1540,
    readingMinutes: 3
  },
  {
    id: 3,
    slug: "content-as-knowledge-source",
    title: "把内容服务作为知识真相源",
    description: "统一内容模型让文章、笔记、项目和关系不再散落在孤岛里。",
    body:
      "统一内容抽象让博客不只是一组文章列表。article、note、project 三种类型都可以拥有版本、标签、关系和发布状态。检索索引、向量摘要和 AI 草稿只是派生数据，真正的内容事实仍然由主数据服务负责。",
    publishedAt: "2026-05-07T00:00:00.000Z",
    tags: ["Content", "Architecture"],
    categories: [{ id: 3, name: "前端", slug: "frontend", content_count: 1 }],
    category: { id: 3, name: "前端", slug: "frontend", content_count: 1 },
    viewCount: 1320,
    readingMinutes: 5
  }
];

export const publicContentConfig: Record<PublicContentKind, { label: string; listPath: string; detailPath: string }> = {
  article: { label: "文章", listPath: "/posts", detailPath: "/posts" },
  note: { label: "笔记", listPath: "/notes", detailPath: "/notes" },
  project: { label: "项目", listPath: "/projects", detailPath: "/projects" }
};

const publicContentTags = {
  list: (type: PublicContentKind) => `public-${type}s`,
  detail: (type: PublicContentKind, slug: string) => `public-${type}:${slug}`
};

export async function listPublicPosts(options: { pageSize?: number } = {}): Promise<PublicPost[]> {
  const contents = await listPublicContents("article", options);
  return contents.map(publicContentToPost);
}

export async function getPublicPost(slug: string): Promise<PublicPost | null> {
  const content = await getPublicContent("article", slug);
  if (!content) return null;
  return publicContentToPost(content);
}

export async function listPublicContents(
  type: PublicContentKind,
  options: { pageSize?: number } = {}
): Promise<PublicContent[]> {
  try {
    const result = await fetchPublicContentList({ type, pageSize: options.pageSize ?? 20 });
    return result.items.map((item) => publicItemToContent(type, item));
  } catch {
    return fallbackPublicContents(type);
  }
}

export async function getPublicContent(type: PublicContentKind, slug: string): Promise<PublicContent | null> {
  try {
    const result = await fetchPublicContentList({ type, slug, pageSize: 1 });
    const item = result.items[0];
    if (!item) {
      if (isProductionRuntime()) {
        return null;
      }
      return (await fallbackPublicContents(type)).find((fallbackItem) => fallbackItem.slug === slug) ?? null;
    }

    const detail = await fetchPublicContent<PublicContentDetailResponse>(`/contents/${item.id}`, {
      next: { revalidate: 60, tags: [publicContentTags.list(type), publicContentTags.detail(type, slug)] }
    });
    return publicItemToContent(type, detail);
  } catch (error) {
    if (error instanceof PublicContentApiError && error.status === 404) {
      return null;
    }
    if (isProductionRuntime()) {
      return null;
    }
    return fallbackPublicContents(type).then((items) => items.find((item) => item.slug === slug) ?? null);
  }
}

export async function getPublicSiteOverview(): Promise<PublicSiteOverview> {
  try {
    const result = await fetchPublicContent<PublicSiteOverviewResponse>("/public/site-overview", {
      next: { revalidate: 60, tags: ["public-site-overview"] }
    });
    return {
      ...result,
      latest: result.latest.map(publicItemToTypedContent),
      featured: result.featured.map(publicItemToTypedContent),
      recent: result.recent.map(publicItemToTypedContent)
    };
  } catch {
    const latest = await fallbackPublicContents("article");
    return fallbackOverview(latest);
  }
}

export async function getPublicReaderContext(contentID: number): Promise<PublicReaderContext> {
  try {
    const result = await fetchPublicContent<PublicReaderContextResponse>(`/contents/${contentID}/reader-context`, {
      next: { revalidate: 60, tags: [`public-reader-context:${contentID}`] }
    });
    return {
      ...result,
      previous: result.previous ? publicItemToTypedContent(result.previous) : undefined,
      next: result.next ? publicItemToTypedContent(result.next) : undefined,
      related: result.related.map(publicItemToTypedContent),
      recent: result.recent.map(publicItemToTypedContent)
    };
  } catch {
    const related = await fallbackPublicContents("article");
    return {
      related: related.filter((item) => item.id !== contentID).slice(0, 2),
      recent: related.filter((item) => item.id !== contentID).slice(0, 5),
      comments: { items: [], total: 0, page: 1, page_size: 10 }
    };
  }
}

export async function createPublicComment(contentID: number, request: CreateCommentRequest): Promise<CreateCommentResponse> {
  const response = await fetch(`/api/v1/contents/${contentID}/comments`, {
    method: "POST",
    headers: {
      accept: "application/json",
      "content-type": "application/json"
    },
    body: JSON.stringify(request)
  });
  const envelope = (await response.json()) as BaseResponse<CreateCommentResponse>;
  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    throw new Error(envelope.message || "Comment request failed");
  }
  return envelope.data;
}

async function fetchPublicContentList(params: { type: PublicContentKind; slug?: string; pageSize: number }) {
  const searchParams = new URLSearchParams({
    page: "1",
    page_size: String(params.pageSize),
    type: params.type
  });
  if (params.slug) searchParams.set("slug", params.slug);

  return fetchPublicContent<PublicContentListResponse>(`/contents?${searchParams.toString()}`, {
    next: { revalidate: 60, tags: [publicContentTags.list(params.type)] }
  });
}

async function fetchPublicContent<T>(path: string, init: RequestInit & { next?: { revalidate?: number; tags?: string[] } } = {}) {
  const response = await fetch(goApiUrl(path), {
    ...init,
    method: init.method ?? "GET",
    headers: {
      accept: "application/json",
      ...init.headers
    }
  });

  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    throw new PublicContentApiError("API returned an invalid response", response.status, response.status);
  }

  let envelope: BaseResponse<T>;
  try {
    envelope = (await response.json()) as BaseResponse<T>;
  } catch {
    throw new PublicContentApiError("API returned an invalid response", response.status, response.status);
  }

  if (!response.ok || envelope.code < 200 || envelope.code >= 300) {
    throw new PublicContentApiError(envelope.message || "Request failed", response.status, envelope.code);
  }

  return envelope.data;
}

function goApiUrl(path: string) {
  const baseUrl = process.env.BEEHIVE_API_BASE_URL ?? "http://localhost:8080";
  return `${baseUrl}/api/v1${path}`;
}

function publicItemToContent(type: PublicContentKind, item: PublicContentDetailResponse): PublicContent {
  const body = item.body?.trim() ?? "";
  return {
    id: item.id,
    type,
    typeLabel: publicContentConfig[type].label,
    href: `${publicContentConfig[type].detailPath}/${item.slug}`,
    slug: item.slug,
    title: item.title,
    description: item.excerpt?.trim() || excerptFromBody(body) || item.title,
    body,
    publishedAt: item.published_at ?? item.updated_at ?? item.created_at,
    tags: item.tags?.map((tag) => tag.name).filter(Boolean) ?? [],
    categories: item.categories?.map((category) => ({
      id: category.id ?? 0,
      name: category.name,
      slug: category.slug ?? category.name,
      content_count: category.content_count
    })) ?? [],
    category: item.category
      ? {
          id: item.category.id ?? 0,
          name: item.category.name,
          slug: item.category.slug ?? item.category.name,
          content_count: item.category.content_count
        }
      : undefined,
    coverUrl: item.cover_url || undefined,
    authorUsername: item.author_username || undefined,
    viewCount: item.view_count ?? 0,
    readingMinutes: item.reading_time_minutes ?? readingMinutesFromBody(body)
  };
}

type PublicSiteOverviewResponse = Omit<PublicSiteOverview, "latest" | "featured" | "recent"> & {
  latest: PublicContentItem[];
  featured: PublicContentItem[];
  recent: PublicContentItem[];
};

type PublicReaderContextResponse = Omit<PublicReaderContext, "previous" | "next" | "related" | "recent"> & {
  previous?: PublicContentItem;
  next?: PublicContentItem;
  related: PublicContentItem[];
  recent: PublicContentItem[];
};

function publicItemToTypedContent(item: PublicContentItem): PublicContent {
  const type = publicContentKind(item.type);
  return publicItemToContent(type, item);
}

function publicContentKind(type: string): PublicContentKind {
  return type === "note" || type === "project" ? type : "article";
}

function excerptFromBody(body: string) {
  const plainText = body
    .replace(/```[\s\S]*?```/g, "")
    .replace(/[#>*_`~\-[\]()]/g, "")
    .replace(/\s+/g, " ")
    .trim();
  return plainText.slice(0, 120);
}

function readingMinutesFromBody(body: string) {
  const wordCount = body.trim() ? body.trim().split(/\s+/).length : 0;
  if (wordCount === 0) return 0;
  return Math.max(1, Math.ceil(wordCount / 200));
}

async function fallbackPublicContents(type: PublicContentKind): Promise<PublicContent[]> {
  if (isProductionRuntime()) return [];
  if (type !== "article") return [];
  const legacyEndpoint = process.env.PUBLIC_CONTENT_ENDPOINT;
  if (!legacyEndpoint) return fallbackPosts.map((post) => fallbackPostToContent(type, post));

  try {
    const response = await fetch(legacyEndpoint, { next: { revalidate: 60, tags: [publicContentTags.list(type)] } });
    if (!response.ok) return fallbackPosts.map((post) => fallbackPostToContent(type, post));
    const posts = (await response.json()) as PublicPost[];
    return Array.isArray(posts)
      ? posts.map((post) => fallbackPostToContent(type, post))
      : fallbackPosts.map((post) => fallbackPostToContent(type, post));
  } catch {
    return fallbackPosts.map((post) => fallbackPostToContent(type, post));
  }
}

function fallbackPostToContent(type: PublicContentKind, post: PublicPost): PublicContent {
  return {
    ...post,
    type,
    typeLabel: publicContentConfig[type].label,
    href: `${publicContentConfig[type].detailPath}/${post.slug}`
  };
}

function publicContentToPost(content: PublicContent): PublicPost {
  return {
    id: content.id,
    slug: content.slug,
    title: content.title,
    description: content.description,
    body: content.body,
    publishedAt: content.publishedAt,
    tags: content.tags,
    categories: content.categories,
    category: content.category,
    coverUrl: content.coverUrl,
    authorUsername: content.authorUsername,
    viewCount: content.viewCount,
    readingMinutes: content.readingMinutes
  };
}

function fallbackOverview(latest: PublicContent[]): PublicSiteOverview {
  const tags = Array.from(new Set(latest.flatMap((item) => item.tags))).map((name, index) => ({
    id: index + 1,
    name,
    slug: name.toLowerCase(),
    content_count: latest.filter((item) => item.tags.includes(name)).length
  }));
  const categories = latest
    .map((item) => item.category)
    .filter((item): item is NonNullable<typeof item> => Boolean(item));
  return {
    latest,
    featured: latest.slice(0, 4),
    recent: latest,
    categories,
    tags,
    archives: [{ year: 2026, month: 5, label: "2026-05", count: latest.length }],
    stats: {
      articles: latest.length,
      notes: 0,
      projects: 0,
      views: latest.reduce((total, item) => total + item.viewCount, 0),
      tags: tags.length
    },
    author: {
      name: "Beehive",
      description: "个人博客与知识中台"
    },
    generated_at: new Date().toISOString()
  };
}

function isProductionRuntime() {
  return process.env.NODE_ENV === "production";
}
