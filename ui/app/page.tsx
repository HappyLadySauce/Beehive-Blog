import Link from "next/link";
import { ArrowRight, BookOpen } from "lucide-react";

import { PublicContentCard } from "@/components/PublicContentCard";
import { PostCard } from "@/components/PostCard";
import { listPublicContents, listPublicPosts } from "@/lib/api/content";

export default async function HomePage() {
  const [posts, notes, projects] = await Promise.all([
    listPublicPosts(),
    listPublicContents("note", { pageSize: 3 }),
    listPublicContents("project", { pageSize: 3 })
  ]);

  return (
    <main className="page">
      <section className="hero">
        <div>
          <p className="eyebrow">Public Web</p>
          <h1>把长期写作、项目和 AI 协作整理成一个可发布的知识蜂巢。</h1>
          <p>
            Beehive Blog 面向读者呈现公开文章、项目与知识脉络。公开页面强调 SSR 与 SEO，让内容稳定可读、可索引。
          </p>
          <div className="hero-actions">
            <Link className="primary-button" href="/posts">
              <BookOpen aria-hidden size={18} />
              阅读文章
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
          {posts.length > 0 ? (
            posts.slice(0, 3).map((post) => <PostCard key={post.slug} post={post} />)
          ) : (
            <div className="public-empty">暂无公开文章。发布第一篇公开文章后会显示在这里。</div>
          )}
        </div>
      </section>

      <section aria-labelledby="latest-notes">
        <div className="section-heading">
          <h2 id="latest-notes">最新笔记</h2>
          <Link className="text-link" href="/notes">
            全部笔记 <ArrowRight aria-hidden size={16} />
          </Link>
        </div>
        <div className="post-grid">
          {notes.length > 0 ? (
            notes.map((note) => <PublicContentCard key={note.slug} content={note} />)
          ) : (
            <div className="public-empty">暂无公开笔记。发布第一条公开笔记后会显示在这里。</div>
          )}
        </div>
      </section>

      <section aria-labelledby="latest-projects">
        <div className="section-heading">
          <h2 id="latest-projects">项目</h2>
          <Link className="text-link" href="/projects">
            全部项目 <ArrowRight aria-hidden size={16} />
          </Link>
        </div>
        <div className="post-grid">
          {projects.length > 0 ? (
            projects.map((project) => <PublicContentCard key={project.slug} content={project} />)
          ) : (
            <div className="public-empty">暂无公开项目。发布第一个公开项目后会显示在这里。</div>
          )}
        </div>
      </section>
    </main>
  );
}
