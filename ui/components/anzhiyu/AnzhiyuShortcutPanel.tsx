"use client";

import { Keyboard } from "lucide-react";
import { useEffect, useState } from "react";

const shortcuts = [
  ["Shift", "K", "开启/关闭快捷键功能"],
  ["Shift", "S", "站内搜索"],
  ["Shift", "H", "返回首页"],
  ["Shift", "P", "关于本站"],
  ["Shift", "T", "回到顶部"]
];

export function AnzhiyuShortcutPanel({ onSearch }: { onSearch?: () => void }) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Shift") setVisible(true);
      if (event.shiftKey && event.key.toLowerCase() === "s") {
        event.preventDefault();
        onSearch?.();
      }
      if (event.shiftKey && event.key.toLowerCase() === "h") {
        window.location.href = "/";
      }
      if (event.shiftKey && event.key.toLowerCase() === "t") {
        window.scrollTo({ top: 0, behavior: "smooth" });
      }
    }
    function onKeyUp(event: KeyboardEvent) {
      if (event.key === "Shift") setVisible(false);
    }
    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("keyup", onKeyUp);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("keyup", onKeyUp);
    };
  }, [onSearch]);

  if (!visible) return null;

  return (
    <div className="anz-shortcut-panel" role="status">
      <h2>
        <Keyboard aria-hidden size={16} />
        博客快捷键
      </h2>
      <p>按住 Shift 键查看可用快捷键</p>
      {shortcuts.map(([mod, key, label]) => (
        <div className="anz-shortcut-panel__row" key={`${mod}-${key}`}>
          <span>
            <kbd>{mod}</kbd>
            <kbd>{key}</kbd>
          </span>
          <strong>{label}</strong>
        </div>
      ))}
      <small>松开 Shift 键或点击外部区域关闭</small>
    </div>
  );
}
