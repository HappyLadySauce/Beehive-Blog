import Link from "next/link";

import type { PublicPost } from "@/lib/api/types";

export function PostCard({ post }: { post: PublicPost }) {
  return (
    <article className="post-card">
      <div className="post-card__meta">
        <time dateTime={post.publishedAt}>{new Intl.DateTimeFormat("zh-CN").format(new Date(post.publishedAt))}</time>
        <span>{post.readingMinutes} 分钟阅读</span>
      </div>
      <h2>
        <Link href={`/posts/${post.slug}`}>{post.title}</Link>
      </h2>
      <p>{post.description}</p>
      <div className="tag-row" aria-label="标签">
        {post.tags.map((tag) => (
          <span key={tag}>{tag}</span>
        ))}
      </div>
    </article>
  );
}
