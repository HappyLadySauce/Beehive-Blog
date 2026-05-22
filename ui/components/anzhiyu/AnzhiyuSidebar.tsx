import Link from "next/link";
import { BookOpen, CalendarDays, Eye, Hash, Rss } from "lucide-react";

import type { ArchiveItem, PublicContent, PublicTaxonomyItem, SiteAuthor, SiteStats } from "@/lib/api/types";

export function AnzhiyuSidebar({
  archives,
  author,
  recent,
  stats,
  tags
}: {
  archives: ArchiveItem[];
  author: SiteAuthor;
  recent: PublicContent[];
  stats: SiteStats;
  tags: PublicTaxonomyItem[];
}) {
  return (
    <aside className="anz-sidebar" aria-label="侧栏">
      <section className="anz-author-card">
        <span className="anz-author-card__hello">欢迎光临</span>
        <div className="anz-author-card__avatar">{author.name.slice(0, 1)}</div>
        <strong>{author.name}</strong>
        <p>{author.description}</p>
      </section>
      <section className="anz-wechat-card">
        <Rss aria-hidden size={28} />
        <div>
          <strong>公众号</strong>
          <span>快人一步获取最新文章</span>
        </div>
      </section>
      <section className="anz-widget">
        <h2>
          <Hash aria-hidden size={18} />
          标签
        </h2>
        <div className="anz-tag-cloud">
          {tags.slice(0, 18).map((tag) => (
            <Link href="/posts" key={tag.slug}>
              {tag.name}
              {tag.content_count ? <sup>{tag.content_count}</sup> : null}
            </Link>
          ))}
        </div>
      </section>
      <section className="anz-widget">
        <h2>
          <CalendarDays aria-hidden size={18} />
          归档
        </h2>
        <div className="anz-archive-grid">
          {archives.map((archive) => (
            <Link href="/posts" key={archive.label}>
              <span>{archive.label}</span>
              <strong>{archive.count} 篇</strong>
            </Link>
          ))}
        </div>
      </section>
      <section className="anz-widget">
        <h2>
          <BookOpen aria-hidden size={18} />
          最近发布
        </h2>
        <div className="anz-recent-list">
          {recent.slice(0, 5).map((item) => (
            <Link href={item.href} key={item.id}>
              <span>{item.title}</span>
              <time dateTime={item.publishedAt}>{new Intl.DateTimeFormat("zh-CN").format(new Date(item.publishedAt))}</time>
            </Link>
          ))}
        </div>
      </section>
      <section className="anz-widget anz-stats">
        <div>
          <BookOpen aria-hidden size={16} />
          <span>文章总数</span>
          <strong>{stats.articles}</strong>
        </div>
        <div>
          <Eye aria-hidden size={16} />
          <span>全站浏览</span>
          <strong>{stats.views}</strong>
        </div>
      </section>
    </aside>
  );
}
