import type { Metadata } from "next";

import { PublicContentCard } from "@/components/PublicContentCard";
import { listPublicContents } from "@/lib/api/content";

export const metadata: Metadata = {
  title: "笔记",
  description: "Beehive Blog 的公开笔记，沉淀轻量记录、阅读摘录与知识原材料。"
};

export default async function NotesPage() {
  const notes = await listPublicContents("note");

  return (
    <main className="page">
      <section className="page-title">
        <p className="eyebrow">Notes</p>
        <h1>公开笔记</h1>
        <p>轻量记录与知识原材料，保留思考过程，也为后续文章和项目提供上下文。</p>
      </section>
      <section className="post-grid" aria-label="笔记列表">
        {notes.length > 0 ? (
          notes.map((note) => <PublicContentCard key={note.slug} content={note} />)
        ) : (
          <div className="public-empty">暂无公开笔记。发布第一条公开笔记后会显示在这里。</div>
        )}
      </section>
    </main>
  );
}
