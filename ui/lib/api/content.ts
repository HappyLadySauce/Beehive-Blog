import type { PublicPost } from "./types";

const fallbackPosts: PublicPost[] = [
  {
    slug: "ai-assisted-writing-loop",
    title: "AI 协作写作回路",
    description: "从零散笔记到可审阅文章，让 AI 参与整理，但把发布权保留给创作者。",
    body:
      "Beehive Blog 的核心不是把内容交给模型代写，而是把长期积累的笔记、项目和经历组织成可追溯的知识材料。AI 可以帮助摘要、改写和发现关系，但每一次公开发布都需要人的审阅。这样的流程让创作者保留判断力，也让公开内容保持稳定可信。",
    publishedAt: "2026-05-09T00:00:00.000Z",
    tags: ["AI", "Writing", "Workflow"],
    readingMinutes: 4
  },
  {
    slug: "public-web-and-studio",
    title: "Public Web 与 Studio 的产品边界",
    description: "读者侧追求清晰阅读和 SEO，创作者侧追求高效管理、版本和权限。",
    body:
      "Public Web 面向访客，应该尽快呈现文章、项目、经历和搜索入口。Studio 面向创作者，应该承担草稿、关系、版本、附件和发布闸门。两侧共享同一套内容真相源，但用状态、可见性和 AI 访问策略区分不同使用场景。",
    publishedAt: "2026-05-08T00:00:00.000Z",
    tags: ["Product", "Studio", "SEO"],
    readingMinutes: 3
  },
  {
    slug: "content-as-knowledge-source",
    title: "把内容服务作为知识真相源",
    description: "统一内容模型让文章、项目、经历和关系不再散落在孤岛里。",
    body:
      "统一内容抽象让博客不只是一组文章列表。文章、笔记、项目和经历都可以拥有版本、标签、关系和发布状态。检索索引、向量摘要和 AI 草稿只是派生数据，真正的内容事实仍然由主数据服务负责。",
    publishedAt: "2026-05-07T00:00:00.000Z",
    tags: ["Content", "Architecture"],
    readingMinutes: 5
  }
];

export async function listPublicPosts(): Promise<PublicPost[]> {
  const endpoint = process.env.PUBLIC_CONTENT_ENDPOINT;
  if (!endpoint) return fallbackPosts;

  try {
    const response = await fetch(endpoint, { next: { revalidate: 60, tags: ["public-posts"] } });
    if (!response.ok) return fallbackPosts;
    const posts = (await response.json()) as PublicPost[];
    return Array.isArray(posts) ? posts : fallbackPosts;
  } catch {
    return fallbackPosts;
  }
}

export async function getPublicPost(slug: string): Promise<PublicPost | null> {
  const posts = await listPublicPosts();
  return posts.find((post) => post.slug === slug) ?? null;
}
