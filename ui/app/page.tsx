import Link from "next/link";

import { AnzhiyuHomeTop } from "@/components/anzhiyu/AnzhiyuHomeTop";
import { AnzhiyuPostCard } from "@/components/anzhiyu/AnzhiyuPostCard";
import { AnzhiyuSidebar } from "@/components/anzhiyu/AnzhiyuSidebar";
import { getPublicSiteOverview } from "@/lib/api/content";

export default async function HomePage() {
  const overview = await getPublicSiteOverview();
  const posts = overview.latest.length > 0 ? overview.latest : overview.featured;

  return (
    <main className="page anzhiyu-page">
      <div className="anz-flash">
        <strong>即刻</strong>
        <span>快写一首情歌，雅俗共赏 ~</span>
        <Link href="/posts">→</Link>
      </div>
      <AnzhiyuHomeTop categories={overview.categories} featured={overview.featured} />
      <div className="anz-layout">
        <section className="anz-feed" aria-labelledby="latest-posts">
          <div className="anz-tabs" id="latest-posts">
            <Link className="is-active" href="/">
              首页
            </Link>
            <Link href="/posts">AI</Link>
            <Link href="/projects">项目</Link>
            <Link href="/notes">生活</Link>
            <Link href="/posts">前端</Link>
            <Link href="/posts">更多</Link>
          </div>
          {posts.length > 0 ? (
            <div className="anz-post-grid">
              {posts.map((post) => (
                <AnzhiyuPostCard content={post} key={`${post.type}-${post.slug}`} />
              ))}
            </div>
          ) : (
            <div className="public-empty">暂无公开文章。发布第一篇公开文章后会显示在这里。</div>
          )}
        </section>
        <AnzhiyuSidebar
          archives={overview.archives}
          author={overview.author}
          recent={overview.recent}
          stats={overview.stats}
          tags={overview.tags}
        />
      </div>
    </main>
  );
}
