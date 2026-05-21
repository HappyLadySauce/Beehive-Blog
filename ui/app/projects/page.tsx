import type { Metadata } from "next";

import { PublicContentCard } from "@/components/PublicContentCard";
import { listPublicContents } from "@/lib/api/content";

export const metadata: Metadata = {
  title: "项目",
  description: "Beehive Blog 的公开项目，记录背景、目标、技术选择和阶段性结果。"
};

export default async function ProjectsPage() {
  const projects = await listPublicContents("project");

  return (
    <main className="page">
      <section className="page-title">
        <p className="eyebrow">Projects</p>
        <h1>公开项目</h1>
        <p>结构化记录项目背景、目标、技术路径和结果，让公开内容不只停留在文章列表。</p>
      </section>
      <section className="post-grid" aria-label="项目列表">
        {projects.length > 0 ? (
          projects.map((project) => <PublicContentCard key={project.slug} content={project} />)
        ) : (
          <div className="public-empty">暂无公开项目。发布第一个公开项目后会显示在这里。</div>
        )}
      </section>
    </main>
  );
}
