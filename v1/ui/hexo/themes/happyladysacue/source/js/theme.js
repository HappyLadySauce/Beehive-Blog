(function () {
  'use strict';

  var storageKey = 'hls-theme';

  /* ---------- 主题色处理 ---------- */

  function resolvePreferredTheme() {
    try {
      var saved = localStorage.getItem(storageKey);
      if (saved === 'light' || saved === 'dark') return saved;
    } catch (err) {}

    var defaultTheme = window.__HLS_THEME__ && window.__HLS_THEME__.defaultTheme;
    if (defaultTheme === 'light' || defaultTheme === 'dark') return defaultTheme;

    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light';
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    try {
      localStorage.setItem(storageKey, theme);
    } catch (err) {}
    updateThemeIcon(theme);
  }

  /**
   * 切换 nav 中太阳/月亮图标的可见性
   * @param {string} theme 'light' | 'dark'
   */
  function updateThemeIcon(theme) {
    var sun = document.querySelector('.icon-sun');
    var moon = document.querySelector('.icon-moon');
    if (!sun || !moon) return;
    if (theme === 'dark') {
      sun.style.display = 'block';
      moon.style.display = 'none';
    } else {
      sun.style.display = 'none';
      moon.style.display = 'block';
    }
  }

  function bindThemeToggle() {
    var button = document.querySelector('[data-theme-toggle]');
    if (!button) return;

    button.addEventListener('click', function () {
      var next = document.documentElement.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      applyTheme(next);
      document.dispatchEvent(new CustomEvent('hls:theme-change', { detail: { theme: next } }));
    });
  }

  /* ---------- 移动端导航抽屉 ---------- */

  function bindNavToggle() {
    var button = document.querySelector('[data-nav-toggle]');
    var menu = document.querySelector('[data-nav-menu]');
    if (!button || !menu) return;

    button.addEventListener('click', function () {
      var opened = menu.classList.toggle('is-open');
      button.setAttribute('aria-expanded', opened ? 'true' : 'false');
    });

    // 点击外部关闭
    document.addEventListener('click', function (e) {
      if (!button.contains(e.target) && !menu.contains(e.target)) {
        menu.classList.remove('is-open');
        button.setAttribute('aria-expanded', 'false');
      }
    });
  }

  /* ---------- 滚动感知导航 ---------- */

  /**
   * 监听页面滚动，超过 50px 后给 [data-nav-scroll] 元素
   * 添加 `is-scrolled` class，触发毛玻璃效果切换。
   * 使用 requestAnimationFrame 节流，避免高频回调影响性能。
   */
  function bindScrollNav() {
    var nav = document.querySelector('[data-nav-scroll]');
    if (!nav) return;

    var ticking = false;
    var scrolled = false;
    var THRESHOLD = 50;

    function onScroll() {
      if (!ticking) {
        window.requestAnimationFrame(function () {
          var nowScrolled = window.scrollY > THRESHOLD;
          if (nowScrolled !== scrolled) {
            scrolled = nowScrolled;
            if (scrolled) {
              nav.classList.add('is-scrolled');
            } else {
              nav.classList.remove('is-scrolled');
            }
          }
          ticking = false;
        });
        ticking = true;
      }
    }

    window.addEventListener('scroll', onScroll, { passive: true });
    // 初始化（页面刷新后可能已滚动）
    onScroll();
  }

  /* ---------- 入场动画 ---------- */

  /**
   * 对所有卡片类元素应用 IntersectionObserver 渐进入场动效。
   *
   * 策略：
   *  1. 为每个目标元素添加 CSS class `will-reveal`，隐藏初始状态。
   *  2. 同一父容器内的卡片按 DOM 顺序错开 80ms 延迟（最多 480ms），
   *     通过 CSS 自定义属性 `--reveal-delay` 传递，形成瀑布式出现效果。
   *  3. 进入视口后添加 `is-revealed` class 触发 CSS transition。
   *  4. transition 结束后移除 `will-reveal`，恢复元素正常的 hover 行为
   *     （避免入场 transition 覆盖 hover transform）。
   *  5. 若 IntersectionObserver 不可用则跳过（渐进降级，正常显示）。
   */
  function bindScrollReveal() {
    if (!window.IntersectionObserver) return;

    var CARD_SELECTOR = [
      '.post-card',
      '.sidebar-card',
      '.tag-cloud--page',
      '.taxonomy-cat-card',
      '.taxonomy-card',
      '.page-header-banner__inner',
    ].join(', ');

    var targets = document.querySelectorAll(CARD_SELECTOR);
    if (!targets.length) return;

    // 按父容器分组，计算组内索引用于错开延迟
    var indexMap = new Map();

    targets.forEach(function (el) {
      var parent = el.parentElement;
      var idx = indexMap.has(parent) ? indexMap.get(parent) : 0;
      indexMap.set(parent, idx + 1);

      var delay = Math.min(idx * 80, 480);
      el.style.setProperty('--reveal-delay', delay + 'ms');
      el.classList.add('will-reveal');
    });

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;

        var el = entry.target;
        el.classList.add('is-revealed');
        observer.unobserve(el);

        // transition 结束后清理入场 class，让 hover CSS 完整接管
        el.addEventListener('transitionend', function cleanup(e) {
          if (e.propertyName !== 'opacity') return;
          el.classList.remove('will-reveal', 'is-revealed');
          el.style.removeProperty('--reveal-delay');
          el.removeEventListener('transitionend', cleanup);
        });
      });
    }, { threshold: 0.06, rootMargin: '0px 0px -20px 0px' });

    targets.forEach(function (el) {
      observer.observe(el);
    });
  }

  /* ---------- 启动 ---------- */

  function boot() {
    var theme = resolvePreferredTheme();
    applyTheme(theme);
    bindThemeToggle();
    bindNavToggle();
    bindScrollNav();
    bindScrollReveal();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})();
