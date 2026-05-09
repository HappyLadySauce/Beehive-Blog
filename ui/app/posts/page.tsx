import type { Metadata } from "next";

import { PostCard } from "@/components/PostCard";
import { listPublicPosts } from "@/lib/api/content";

export const metadata: Metadata = {
  title: "文章",
  description: "Beehive Blog 的公开文章列表，覆盖 AI 协作写作、内容架构与个人知识管理。"
};

export default async function PostsPage() {
  const posts = await listPublicPosts();

  return (
    <main className="page">
      <section className="page-title">
        <p className="eyebrow">Articles</p>
        <h1>公开文章</h1>
        <p>面向读者的正式输出，首屏由服务端渲染，便于搜索引擎索引和弱网阅读。</p>
      </section>
      <section className="post-grid" aria-label="文章列表">
        {posts.map((post) => (
          <PostCard key={post.slug} post={post} />
        ))}
      </section>
    </main>
  );
}
