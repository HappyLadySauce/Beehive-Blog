import Link from "next/link";
import { Hash } from "lucide-react";

import type { PublicContent } from "@/lib/api/types";
import { coverStyle } from "./cover";

export function AnzhiyuPostCard({ content }: { content: PublicContent }) {
  return (
    <article className="anz-post-card">
      <Link className="anz-post-card__cover" href={content.href} style={coverStyle(content)} aria-label={content.title}>
        <span className="anz-post-card__glyph">{content.category?.name?.slice(0, 2) || content.typeLabel}</span>
      </Link>
      <div className="anz-post-card__body">
        <div className="anz-post-card__tips">
          <span>{content.category?.name || content.typeLabel}</span>
          <span>{content.readingMinutes} 分钟</span>
        </div>
        <h2>
          <Link href={content.href}>{content.title}</Link>
        </h2>
        <p>{content.description}</p>
        <div className="anz-post-card__meta">
          {content.tags.slice(0, 3).map((tag) => (
            <span key={tag}>
              <Hash aria-hidden size={12} />
              {tag}
            </span>
          ))}
          <time dateTime={content.publishedAt}>{formatRelativeMonth(content.publishedAt)}</time>
        </div>
      </div>
    </article>
  );
}

function formatRelativeMonth(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("zh-CN", { month: "long", year: "numeric" }).format(date);
}
