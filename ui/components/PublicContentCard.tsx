import Link from "next/link";

import type { PublicContent } from "@/lib/api/types";

export function PublicContentCard({ content }: { content: PublicContent }) {
  return (
    <article className="post-card">
      <div className="post-card__meta">
        <span>{content.typeLabel}</span>
        <time dateTime={content.publishedAt}>{new Intl.DateTimeFormat("zh-CN").format(new Date(content.publishedAt))}</time>
        <span>{content.readingMinutes} 分钟阅读</span>
      </div>
      <h2>
        <Link href={content.href}>{content.title}</Link>
      </h2>
      <p>{content.description}</p>
      <div className="tag-row" aria-label="标签">
        {content.tags.map((tag) => (
          <span key={tag}>{tag}</span>
        ))}
      </div>
    </article>
  );
}
