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
   * 对 `.post-card`、`.sidebar-card` 等区块应用 IntersectionObserver
   * 触发 CSS `@keyframes fade-up` 入场动效，改善视觉层次感。
   */
  function bindScrollReveal() {
    if (!window.IntersectionObserver) return;

    var targets = document.querySelectorAll('.post-card, .sidebar-card, .taxonomy-card');
    if (!targets.length) return;

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.style.animation = 'fade-up 0.5s ease both';
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.08 });

    targets.forEach(function (el) {
      el.style.opacity = '0';
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
