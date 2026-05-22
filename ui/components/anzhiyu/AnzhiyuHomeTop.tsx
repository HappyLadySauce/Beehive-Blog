import Link from "next/link";
import { ArrowUpRight, Flame, Sparkles, Tags } from "lucide-react";

import type { PublicContent, PublicTaxonomyItem } from "@/lib/api/types";
import { coverStyle } from "./cover";

export function AnzhiyuHomeTop({
  categories,
  featured
}: {
  categories: PublicTaxonomyItem[];
  featured: PublicContent[];
}) {
  const main = featured[0];
  const secondary = featured.slice(1, 4);
  const categoryTiles = normalizedCategories(categories);

  return (
    <section className="anz-home-top" aria-label="首页推荐">
      <div className="anz-home-top__left">
        <Link className="anz-today-card" href={main?.href ?? "/posts"} style={main ? coverStyle(main) : undefined}>
          <div>
            <span>生活明朗</span>
            <strong>{main?.title ?? "生活明朗，万物可爱。"}</strong>
            <small>{main?.authorUsername || "ANHEYU.COM"}</small>
          </div>
          <span className="anz-today-card__button">
            <ArrowUpRight aria-hidden size={18} />
            更多推荐
          </span>
        </Link>
        <div className="anz-category-row">
          {categoryTiles.map((category, index) => (
            <Link className={`anz-category-tile anz-category-tile--${index + 1}`} href="/posts" key={`${category.slug}-${index}`}>
              <span>{category.name}</span>
              <Tags aria-hidden size={76} />
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
        <div className="anz-mini-feature-grid">
          {secondary.map((item) => (
            <Link href={item.href} key={item.id} style={coverStyle(item)}>
              <Flame aria-hidden size={18} />
              <span>{item.title}</span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
}

function normalizedCategories(categories: PublicTaxonomyItem[]) {
  const fallback = [
    { id: 1, name: "前端", slug: "frontend" },
    { id: 2, name: "生活", slug: "life" },
    { id: 3, name: "安和鱼", slug: "anzhiyu" }
  ];
  return [...categories, ...fallback].slice(0, 3);
}
