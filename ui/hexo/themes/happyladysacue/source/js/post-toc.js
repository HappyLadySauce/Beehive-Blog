/**
 * 文章页目录：滚动高亮、点击平滑滚动、阅读百分比、窄屏浮层开关。
 * 行为参考 hexo-theme-anzhiyu main.js scrollFnToDo、utils scrollToDest/getEleTop（等价自写）。
 */
(function () {
  'use strict';

  var MOBILE_MAX = 1024;
  var HEADING_OFFSET = 80;
  var CLICK_SCROLL_OFFSET = 80;
  var THROTTLE_MS = 100;

  function isMobileLayout() {
    return window.innerWidth <= MOBILE_MAX;
  }

  function getEleTop(ele) {
    var actualTop = ele.offsetTop;
    var current = ele.offsetParent;
    while (current !== null) {
      actualTop += current.offsetTop;
      current = current.offsetParent;
    }
    return actualTop;
  }

  function scrollToDest(pos, time) {
    time = time === undefined ? 500 : time;
    if ('scrollBehavior' in document.documentElement.style) {
      window.scrollTo({ top: pos, behavior: 'smooth' });
      return;
    }
    var currentPos = window.pageYOffset;
    var start = null;
    pos = +pos;
    window.requestAnimationFrame(function step(currentTime) {
      start = !start ? currentTime : start;
      var progress = currentTime - start;
      if (currentPos < pos) {
        window.scrollTo(0, ((pos - currentPos) * progress) / time + currentPos);
      } else {
        window.scrollTo(0, currentPos - ((currentPos - pos) * progress) / time);
      }
      if (progress < time) {
        window.requestAnimationFrame(step);
      } else {
        window.scrollTo(0, pos);
      }
    });
  }

  function throttle(fn, wait) {
    var last = 0;
    return function () {
      var now = Date.now();
      if (now - last >= wait) {
        last = now;
        fn.apply(this, arguments);
      }
    };
  }

  function findTocLinkForId(tocRoot, id) {
    if (!id) return null;
    var hash = '#' + id;
    var links = tocRoot.querySelectorAll('.toc-list-link');
    for (var i = 0; i < links.length; i++) {
      var h = links[i].getAttribute('href') || '';
      if (h === hash) return links[i];
      try {
        if (h.charAt(0) === '#' && decodeURIComponent(h.slice(1)) === id) return links[i];
      } catch (e) {}
    }
    return null;
  }

  function updateReadingPercent(scrollTop, percentEl) {
    if (!percentEl) return;
    var wrap = document.getElementById('body-wrap');
    var winHeight = document.documentElement.clientHeight;
    var docHeight = wrap ? wrap.clientHeight : document.documentElement.scrollHeight;
    var contentMath =
      docHeight > winHeight ? docHeight - winHeight : document.documentElement.scrollHeight - winHeight;
    if (contentMath <= 0) return;
    var scrollPercent = scrollTop / contentMath;
    var n = Math.round(scrollPercent * 100);
    if (n > 100) n = 100;
    if (n < 0) n = 0;
    percentEl.textContent = n + '%';
    if (window.innerWidth > MOBILE_MAX) {
      percentEl.removeAttribute('aria-hidden');
    }
  }

  function init() {
    var cardTocLayout = document.getElementById('card-toc');
    var article = document.getElementById('article-container');
    if (!cardTocLayout || !article) return;

    var tocContent = cardTocLayout.querySelector('.toc-content');
    if (!tocContent) return;

    var isExpand = tocContent.classList.contains('is-expand');
    var percentEl = cardTocLayout.querySelector('.toc-percentage');
    var mobileBtn = document.querySelector('[data-post-toc-trigger]');

    function collectHeadings() {
      var list = article.querySelectorAll('h1,h2,h3,h4,h5,h6');
      return Array.prototype.slice.call(list).filter(function (heading) {
        return heading.id && heading.id !== 'CrawlerTitle';
      });
    }

    function clearActive() {
      tocContent.querySelectorAll('.active').forEach(function (el) {
        el.classList.remove('active');
      });
    }

    function autoScrollToc(item) {
      if (!item) return;
      var activePosition = item.getBoundingClientRect().top;
      var sidebarScrollTop = tocContent.scrollTop;
      if (activePosition > document.documentElement.clientHeight - 100) {
        tocContent.scrollTop = sidebarScrollTop + 150;
      }
      if (activePosition < 100) {
        tocContent.scrollTop = sidebarScrollTop - 150;
      }
    }

    var lastActiveId = '';

    function findHeadPosition(scrollTop) {
      var filtered = collectHeadings();
      var currentId = '';
      for (var i = 0; i < filtered.length; i++) {
        var ele = filtered[i];
        if (scrollTop >= getEleTop(ele) - HEADING_OFFSET) {
          currentId = ele.id;
        }
      }
      if (lastActiveId === currentId) return;
      lastActiveId = currentId;

      clearActive();
      if (!currentId) return;

      var activeLink = findTocLinkForId(tocContent, currentId);
      if (!activeLink) return;

      activeLink.classList.add('active');

      if (!isExpand) {
        var p = activeLink.parentElement;
        while (p && p !== tocContent) {
          if (p.classList && p.classList.contains('toc-list-item')) {
            p.classList.add('active');
          }
          p = p.parentElement;
        }
      }

      setTimeout(function () {
        autoScrollToc(activeLink);
      }, 0);
    }

    var onScroll = throttle(function () {
      var scrollTop = window.scrollY || document.documentElement.scrollTop;
      findHeadPosition(scrollTop);
      updateReadingPercent(scrollTop, percentEl);
    }, THROTTLE_MS);

    window.addEventListener('scroll', onScroll, { passive: true });

    tocContent.addEventListener('click', function (e) {
      var target = e.target.closest('.toc-list-link');
      if (!target) return;
      e.preventDefault();
      var raw = target.getAttribute('href') || '';
      var id = '';
      try {
        id = raw.charAt(0) === '#' ? decodeURIComponent(raw.slice(1)) : '';
      } catch (err) {
        id = raw.replace(/^#/, '');
      }
      var dest = id ? document.getElementById(id) : null;
      if (dest) {
        scrollToDest(getEleTop(dest) - CLICK_SCROLL_OFFSET, 300);
      } else if (raw.charAt(0) === '#') {
        window.location.hash = raw;
      }
      if (isMobileLayout()) {
        cardTocLayout.classList.remove('open');
        if (mobileBtn) mobileBtn.setAttribute('aria-expanded', 'false');
      }
    });

    if (mobileBtn) {
      mobileBtn.addEventListener('click', function () {
        var open = cardTocLayout.classList.toggle('open');
        mobileBtn.setAttribute('aria-expanded', open ? 'true' : 'false');
        var rect = mobileBtn.getBoundingClientRect();
        cardTocLayout.style.transformOrigin = 'right ' + (rect.top + 17) + 'px';
        cardTocLayout.style.transition = 'transform 0.3s ease-in-out';
        cardTocLayout.addEventListener(
          'transitionend',
          function onTe(ev) {
            if (ev.propertyName !== 'transform') return;
            cardTocLayout.style.transition = '';
            cardTocLayout.style.transformOrigin = '';
            cardTocLayout.removeEventListener('transitionend', onTe);
          },
          { once: true }
        );
      });
    }

    document.addEventListener('click', function (e) {
      if (!cardTocLayout.classList.contains('open')) return;
      if (!isMobileLayout()) return;
      if (cardTocLayout.contains(e.target)) return;
      if (mobileBtn && mobileBtn.contains(e.target)) return;
      cardTocLayout.classList.remove('open');
      if (mobileBtn) mobileBtn.setAttribute('aria-expanded', 'false');
    });

    var initialTop = window.scrollY || document.documentElement.scrollTop;
    findHeadPosition(initialTop);
    updateReadingPercent(initialTop, percentEl);

    window.addEventListener(
      'resize',
      throttle(function () {
        if (!isMobileLayout()) {
          cardTocLayout.classList.remove('open');
          if (mobileBtn) mobileBtn.setAttribute('aria-expanded', 'false');
        }
      }, 200)
    );
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
