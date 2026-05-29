import Link from "next/link";
import { ArrowUpRight, BookOpen, FolderKanban, NotebookTabs, Sparkles } from "lucide-react";

import type { PublicContent } from "@/lib/api/types";
import { coverStyle } from "./cover";

export function AnzhiyuHomeTop({
  featured
}: {
  featured: PublicContent[];
}) {
  const main = featured[0];
  const entryTiles = homeEntries;

  return (
    <section className="anz-home-top" aria-label="首页推荐">
      <div className="anz-home-top__left">
        <Link className="anz-today-card" href={main?.href ?? "/posts"} style={main ? coverStyle(main) : undefined}>
          <div>
            <span>Beehive</span>
            <strong>{main?.title ?? "个人博客与知识中台"}</strong>
            <small>{main?.authorUsername || "Beehive Blog"}</small>
          </div>
          <span className="anz-today-card__button">
            <ArrowUpRight aria-hidden size={18} />
            更多推荐
          </span>
        </Link>
        <div className="anz-category-row">
          {entryTiles.map((entry, index) => (
            <Link className={`anz-category-tile anz-category-tile--${index + 1}`} href={entry.href} key={entry.href}>
              <span>{entry.name}</span>
              <entry.icon aria-hidden size={76} />
            </Link>
          ))}
        </div>
      </div>
      <div className="anz-home-top__right">
        <div className="anz-theme-card">
          <span>新品框架</span>
          <strong>Theme-AnZhiYu</strong>
          <Sparkles aria-hidden size={88} />
          <Link href="/posts">
            <ArrowUpRight aria-hidden size={16} />
            更多推荐
          </Link>
        </div>
      </div>
    </section>
  );
}

const homeEntries = [
  { name: "文库", href: "/posts", icon: BookOpen },
  { name: "笔记", href: "/notes", icon: NotebookTabs },
  { name: "项目", href: "/projects", icon: FolderKanban }
];
