import Link from "next/link";
import { ArrowRight, BookOpen, LayoutDashboard } from "lucide-react";

import { PostCard } from "@/components/PostCard";
import { listPublicPosts } from "@/lib/api/content";

export default async function HomePage() {
  const posts = await listPublicPosts();

  return (
    <main className="page">
      <section className="hero">
        <div>
          <p className="eyebrow">Public Web + Studio</p>
          <h1>把长期写作、项目和 AI 协作整理成一个可发布的知识蜂巢。</h1>
          <p>
            Beehive Blog 同时服务读者、创作者和外部智能体。公开页面强调 SSR 与 SEO，Studio
            负责草稿、关系、版本和发布闸门。
          </p>
          <div className="hero-actions">
            <Link className="primary-button" href="/posts">
              <BookOpen aria-hidden size={18} />
              阅读文章
            </Link>
            <Link className="secondary-button" href="/studio">
              <LayoutDashboard aria-hidden size={18} />
              进入 Studio
            </Link>
          </div>
        </div>
        <div className="hero-visual" aria-hidden>
          <div className="honeycomb">
            {Array.from({ length: 12 }, (_, index) => (
              <span key={index} />
            ))}
          </div>
        </div>
      </section>

      <section aria-labelledby="latest-posts">
        <div className="section-heading">
          <h2 id="latest-posts">最新文章</h2>
          <Link className="text-link" href="/posts">
            全部文章 <ArrowRight aria-hidden size={16} />
          </Link>
        </div>
        <div className="post-grid">
          {posts.slice(0, 3).map((post) => (
            <PostCard key={post.slug} post={post} />
          ))}
        </div>
      </section>
    </main>
  );
}
