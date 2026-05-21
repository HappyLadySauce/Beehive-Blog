import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import type { PublicContent } from "@/lib/api/types";

export function PublicContentDetail({ content, eyebrow }: { content: PublicContent; eyebrow: string }) {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": content.type === "project" ? "CreativeWork" : "BlogPosting",
    headline: content.title,
    description: content.description,
    datePublished: content.publishedAt,
    keywords: content.tags.join(", ")
  };

  return (
    <main className="page">
      <article className="article">
        <p className="eyebrow">{eyebrow}</p>
        <h1>{content.title}</h1>
        <div className="article-meta">
          <time dateTime={content.publishedAt}>{new Intl.DateTimeFormat("zh-CN").format(new Date(content.publishedAt))}</time>
          <span>{content.readingMinutes} 分钟阅读</span>
        </div>
        <div className="tag-row" aria-label="标签">
          {content.tags.map((tag) => (
            <span key={tag}>{tag}</span>
          ))}
        </div>
        <div className="article-body">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{content.body}</ReactMarkdown>
        </div>
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      </article>
    </main>
  );
}
