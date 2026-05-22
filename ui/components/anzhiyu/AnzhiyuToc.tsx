import { List } from "lucide-react";

export type TocItem = {
  id: string;
  text: string;
  level: number;
};

export function AnzhiyuToc({ items }: { items: TocItem[] }) {
  return (
    <section className="anz-widget anz-toc" aria-label="文章目录">
      <h2>
        <List aria-hidden size={18} />
        目录
        <span>{items.length}</span>
      </h2>
      <nav>
        {items.length > 0 ? (
          items.map((item) => (
            <a className={`anz-toc__level-${item.level}`} href={`#${item.id}`} key={item.id}>
              {item.text}
            </a>
          ))
        ) : (
          <span className="anz-empty-inline">暂无目录</span>
        )}
      </nav>
    </section>
  );
}
