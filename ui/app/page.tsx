import Link from "next/link";
import type { Metadata } from "next";

import { AnzhiyuFooter } from "@/components/anzhiyu/AnzhiyuFooter";
import { AnzhiyuHomeTop } from "@/components/anzhiyu/AnzhiyuHomeTop";
import { AnzhiyuPostCard } from "@/components/anzhiyu/AnzhiyuPostCard";
import { AnzhiyuSidebar } from "@/components/anzhiyu/AnzhiyuSidebar";
import { getPublicSiteOverview } from "@/lib/api/content";

export async function generateMetadata(): Promise<Metadata> {
  const overview = await getPublicSiteOverview();
  const siteName = overview.site?.name || "Beehive";
  const description = overview.site?.description || "个人博客、AI 协作创作与面向智能体的个人知识中台。";
  const keywords = overview.site?.keywords
    ? overview.site.keywords
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean)
    : undefined;
  return {
    title: siteName,
    description,
    keywords,
    openGraph: {
      title: siteName,
      description,
      type: "website",
      url: overview.site?.url || undefined
    }
  };
}

export default async function HomePage() {
  const overview = await getPublicSiteOverview();
  const posts = overview.latest.length > 0 ? overview.latest : overview.featured;

  return (
    <>
      <main className="page anzhiyu-page">
        <AnzhiyuHomeTop featured={overview.featured} />
        <div className="anz-layout">
          <section className="anz-feed" aria-labelledby="latest-posts">
            <div className="anz-tabs" id="latest-posts">
              <Link className="is-active" href="/">
                首页
              </Link>
              {overview.categories.map((category) => (
                <Link href="/posts" key={category.slug}>
                  {category.name}
                </Link>
              ))}
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
      <AnzhiyuFooter author={overview.author} site={overview.site} stats={overview.stats} />
    </>
  );
}
