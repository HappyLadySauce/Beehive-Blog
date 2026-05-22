"use client";

import Link from "next/link";
import { Search, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

type SearchItem = {
  href: string;
  title: string;
  description: string;
  typeLabel: string;
};

export function AnzhiyuSearch({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [items, setItems] = useState<SearchItem[]>([]);
  const [query, setQuery] = useState("");

  useEffect(() => {
    if (!open || items.length > 0) return;
    let cancelled = false;
    Promise.all([loadType("article", "文章", "/posts"), loadType("note", "笔记", "/notes"), loadType("project", "项目", "/projects")])
      .then((groups) => {
        if (!cancelled) setItems(groups.flat());
      })
      .catch(() => {
        if (!cancelled) setItems([]);
      });
    return () => {
      cancelled = true;
    };
  }, [items.length, open]);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }
    if (open) window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose, open]);

  const results = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    if (!keyword) return items.slice(0, 8);
    return items.filter((item) => `${item.title} ${item.description}`.toLowerCase().includes(keyword)).slice(0, 12);
  }, [items, query]);

  if (!open) return null;

  return (
    <div className="anz-search-overlay" role="dialog" aria-modal="true" aria-label="站内搜索">
      <div className="anz-search-overlay__panel">
        <div className="anz-search-overlay__bar">
          <Search aria-hidden size={20} />
          <input autoFocus value={query} placeholder="搜索公开文章、笔记与项目" onChange={(event) => setQuery(event.target.value)} />
          <button type="button" aria-label="关闭搜索" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <div className="anz-search-overlay__results">
          {results.length > 0 ? (
            results.map((item) => (
              <Link href={item.href} key={item.href} onClick={onClose}>
                <span>{item.typeLabel}</span>
                <strong>{item.title}</strong>
                <p>{item.description}</p>
              </Link>
            ))
          ) : (
            <p className="anz-empty-inline">没有匹配内容</p>
          )}
        </div>
      </div>
    </div>
  );
}

async function loadType(type: string, typeLabel: string, basePath: string): Promise<SearchItem[]> {
  const response = await fetch(`/api/v1/contents?page=1&page_size=20&type=${type}`, { headers: { accept: "application/json" } });
  if (!response.ok) return [];
  const envelope = (await response.json()) as {
    data?: { items?: Array<{ slug: string; title: string; excerpt?: string | null }> };
  };
  return (envelope.data?.items ?? []).map((item) => ({
    href: `${basePath}/${item.slug}`,
    title: item.title,
    description: item.excerpt ?? item.title,
    typeLabel
  }));
}
