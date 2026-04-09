/**
 * 为 Hexo 默认 figure.highlight 注入 mac 风格工具栏（语言 + 复制），行为参考 anzhiyu 主题逻辑，实现自写。
 */
(function () {
  "use strict";

  function inferLang(fig) {
    const cls = fig.getAttribute("class") || "";
    const parts = cls.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      let name = parts[1];
      if (name === "plain" || !name) return "CODE";
      return name.toUpperCase();
    }
    return "CODE";
  }

  function getCodePre(fig) {
    return fig.querySelector("table td.code pre, td.code pre");
  }

  function getCodeText(fig) {
    const pre = getCodePre(fig);
    return pre ? (pre.innerText || "") : "";
  }

  async function copyText(text) {
    if (navigator.clipboard && window.isSecureContext && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try {
      return document.execCommand("copy");
    } finally {
      document.body.removeChild(ta);
    }
  }

  function flashNotice(bar, message) {
    const n = bar.querySelector(".copy-notice");
    if (!n) return;
    n.textContent = message;
    n.classList.add("is-visible");
    window.clearTimeout(flashNotice._t);
    flashNotice._t = window.setTimeout(function () {
      n.classList.remove("is-visible");
    }, 1800);
  }

  function injectToolbar(fig) {
    if (fig.querySelector(".highlight-tools")) return;

    const lang = inferLang(fig);
    const tools = document.createElement("div");
    tools.className = "highlight-tools";

    const notice = document.createElement("span");
    notice.className = "copy-notice";
    notice.setAttribute("aria-live", "polite");

    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "copy-button";
    btn.setAttribute("aria-label", "复制代码");
    btn.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>';

    const langEl = document.createElement("div");
    langEl.className = "code-lang";
    langEl.textContent = lang;

    tools.appendChild(notice);
    tools.appendChild(btn);
    tools.appendChild(langEl);

    const table = fig.querySelector("table");
    if (table) {
      fig.insertBefore(tools, table);
    } else {
      fig.insertBefore(tools, fig.firstChild);
    }

    btn.addEventListener("click", function () {
      const text = getCodeText(fig);
      if (!text) {
        flashNotice(tools, "无可复制内容");
        return;
      }
      copyText(text)
        .then(function (ok) {
          flashNotice(tools, ok ? "已复制" : "复制失败");
        })
        .catch(function () {
          flashNotice(tools, "复制失败");
        });
    });
  }

  function run() {
    document.querySelectorAll("figure.highlight").forEach(injectToolbar);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", run);
  } else {
    run();
  }
})();
