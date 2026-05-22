"use client";

import { ArrowUp, Moon, Search, SquareMenu } from "lucide-react";

export function AnzhiyuRightTools({ onSearch }: { onSearch: () => void }) {
  return (
    <div className="anz-right-tools" aria-label="页面工具">
      <button aria-label="站内搜索" type="button" onClick={onSearch}>
        <Search aria-hidden size={18} />
      </button>
      <button aria-label="显示模式" type="button">
        <Moon aria-hidden size={18} />
      </button>
      <button aria-label="页面菜单" type="button">
        <SquareMenu aria-hidden size={18} />
      </button>
      <button aria-label="回到顶部" type="button" onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}>
        <ArrowUp aria-hidden size={18} />
      </button>
    </div>
  );
}
