import { CalendarDays, Clock3, Eye, Hash, MapPin } from "lucide-react";
import Link from "next/link";

import { AnzhiyuCommentBox } from "@/components/anzhiyu/AnzhiyuCommentBox";
import { AnzhiyuMarkdown, extractToc } from "@/components/anzhiyu/AnzhiyuMarkdown";
import { AnzhiyuPostCard } from "@/components/anzhiyu/AnzhiyuPostCard";
import { AnzhiyuSidebar } from "@/components/anzhiyu/AnzhiyuSidebar";
import { AnzhiyuToc } from "@/components/anzhiyu/AnzhiyuToc";
import { coverStyle } from "@/components/anzhiyu/cover";
import { getPublicReaderContext, getPublicSiteOverview } from "@/lib/api/content";
import type { PublicContent } from "@/lib/api/types";

export async function PublicContentDetail({ content, eyebrow }: { content: PublicContent; eyebrow: string }) {
  const [overview, readerContext] = await Promise.all([getPublicSiteOverview(), getPublicReaderContext(content.id)]);
  const toc = extractToc(content.body);
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": content.type === "project" ? "CreativeWork" : "BlogPosting",
    headline: content.title,
    description: content.description,
    datePublished: content.publishedAt,
    keywords: content.tags.join(", ")
  };

  return (
    <>
      <header className="anz-post-hero" style={coverStyle(content)}>
        <div className="anz-post-hero__inner">
          <div className="anz-post-hero__tags">
            <span>原创</span>
            <span>{content.category?.name || eyebrow}</span>
            {content.tags.slice(0, 3).map((tag) => (
              <span key={tag}>
                <Hash aria-hidden size={14} />
                {tag}
              </span>
            ))}
          </div>
          <h1>{content.title}</h1>
          <div className="anz-post-hero__meta">
            <span>
              <Eye aria-hidden size={15} />
              {content.viewCount}
            </span>
            <span>
              <Clock3 aria-hidden size={15} />
              {content.readingMinutes} 分钟
            </span>
            <time dateTime={content.publishedAt}>
              <CalendarDays aria-hidden size={15} />
              {new Intl.DateTimeFormat("zh-CN").format(new Date(content.publishedAt))}
            </time>
            <span>
              <MapPin aria-hidden size={15} />
              广东 深圳
            </span>
          </div>
        </div>
      </header>
      <main className="page anzhiyu-page anz-post-page">
        <div className="anz-layout">
          <article className="article anz-article-card">
            <div className="anz-ai-description">{content.description}</div>
            <div className="article-body">
              <AnzhiyuMarkdown>{content.body}</AnzhiyuMarkdown>
            </div>
            <section className="anz-copyright">
              <div className="anz-copyright__avatar">{overview.author.name.slice(0, 1)}</div>
              <strong>{overview.author.name}</strong>
              <span>{overview.author.description}</span>
              <div>
                <button type="button">打赏作者</button>
                <button type="button">订阅</button>
                <button type="button">分享</button>
              </div>
              <p>本文是原创文章，采用 CC BY-NC-SA 4.0 协议，完整转载请注明来自 {overview.author.name}。</p>
            </section>
            <div className="tag-row" aria-label="标签">
              {content.tags.map((tag) => (
                <span key={tag}># {tag}</span>
              ))}
            </div>
            <section className="anz-related">
              <div className="section-heading">
                <h2>喜欢这篇文章的人也看了</h2>
                <Link className="text-link" href="/posts">
                  随便逛逛
                </Link>
              </div>
              <div className="anz-related-grid">
                {readerContext.related.slice(0, 2).map((item) => (
                  <AnzhiyuPostCard content={item} key={item.id} />
                ))}
              </div>
            </section>
            <AnzhiyuCommentBox comments={readerContext.comments.items} contentID={content.id} />
            <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
          </article>
          <aside className="anz-post-aside">
            <AnzhiyuSidebar
              archives={overview.archives}
              author={overview.author}
              recent={readerContext.recent}
              stats={overview.stats}
              tags={overview.tags}
            />
            <AnzhiyuToc items={toc} />
          </aside>
        </div>
      </main>
    </>
  );
}
