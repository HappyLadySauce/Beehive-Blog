/**
 * 文章页目录：滚动高亮、点击平滑滚动、阅读百分比、窄屏浮层开关。
 * 行为参考 hexo-theme-anzhiyu main.js scrollFnToDo、utils throttle（trailing）/ scrollToDest/getEleTop。
 */
(function () {
  'use strict';

  var MOBILE_MAX = 1024;
  var HEADING_OFFSET = 80;
  var CLICK_SCROLL_OFFSET = 80;
  var THROTTLE_MS = 100;
  /** 与 HEADING_OFFSET 一致，正文 h* 的 scroll-margin 由 article-prose.css 设置 */
  var TOC_INVIEW_PADDING = 12;

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

  /**
   * 与 hexo-theme-anzhiyu utils.throttle 同语义：burst 结束后 trailing 再执行一次，避免快速滚动后 TOC 停在陈旧位置。
   */
  function throttle(fn, wait, options) {
    options = options || {};
    var timeout = null;
    var context;
    var args;
    var previous = 0;

    function later() {
      previous = options.leading === false ? 0 : Date.now();
      timeout = null;
      fn.apply(context, args);
      if (!timeout) {
        context = null;
        args = null;
      }
    }

    return function () {
      var now = Date.now();
      if (!previous && options.leading === false) previous = now;
      var remaining = wait - (now - previous);
      context = this;
      args = arguments;
      if (remaining <= 0 || remaining > wait) {
        if (timeout) {
          clearTimeout(timeout);
          timeout = null;
        }
        previous = now;
        fn.apply(context, args);
        if (!timeout) {
          context = null;
          args = null;
        }
      } else if (!timeout && options.trailing !== false) {
        timeout = setTimeout(later, remaining);
      }
    };
  }

  function findTocLinkForId(tocRoot, id) {
    if (!id) return null;
    var hash = '#' + id;
    var encHash = '';
    try {
      encHash = '#' + encodeURIComponent(id);
    } catch (e) {
      encHash = hash;
    }
    var links = tocRoot.querySelectorAll('.toc-list-link');
    for (var i = 0; i < links.length; i++) {
      var h = links[i].getAttribute('href') || '';
      if (h === hash || h === encHash) return links[i];
      try {
        if (h.charAt(0) === '#' && decodeURIComponent(h.slice(1)) === id) return links[i];
      } catch (e2) {}
    }
    return null;
  }

  /**
   * 正文当前标题若在目录中无对应项（如 max_depth 截断、h4+），沿标题链向上回退到最近有目录锚点的标题。
   * 与 hexo-theme-anzhiyu 的「标题与目录一一对应」思路等价，但兼容目录为子集的情况。
   */
  function resolveActiveTocLink(tocRoot, filteredHeadings, currentIndex) {
    if (!tocRoot || !filteredHeadings.length || currentIndex < 0) return null;
    for (var j = currentIndex; j >= 0; j--) {
      var hid = filteredHeadings[j].id;
      var link = findTocLinkForId(tocRoot, hid);
      if (link) return link;
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

  /**
   * 将当前高亮链接滚入 .toc-content 可视区（相对视口比较容器与链接矩形，避免 ±150px 一步不到位）。
   */
  function autoScrollTocIntoView(container, link) {
    if (!container || !link || !container.contains(link)) return;
    var cr = container.getBoundingClientRect();
    var lr = link.getBoundingClientRect();
    var pad = TOC_INVIEW_PADDING;
    if (lr.top < cr.top + pad) {
      container.scrollTop -= cr.top + pad - lr.top;
    } else if (lr.bottom > cr.bottom - pad) {
      container.scrollTop += lr.bottom - (cr.bottom - pad);
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

    /** 仅最后一次高亮更新后的 rAF 会执行侧栏滚动，避免 setTimeout(0) 乱序。 */
    var tocLayoutGen = 0;

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

    /** 以最终高亮链接的 href 为键；避免「目录无匹配 id 时先写 lastActiveId 导致后续无法重试」的锁死。 */
    var lastHighlightHref = '';

    function findHeadPosition(scrollTop) {
      var filtered = collectHeadings();
      var currentIndex = -1;
      for (var i = 0; i < filtered.length; i++) {
        if (scrollTop >= getEleTop(filtered[i]) - HEADING_OFFSET) {
          currentIndex = i;
        }
      }

      var activeLink = null;
      if (currentIndex >= 0) {
        activeLink = resolveActiveTocLink(tocContent, filtered, currentIndex);
      } else if (filtered.length > 0) {
        /* 尚未滚过第一个标题：展开并高亮目录首项，避免折叠态下只剩顶层编号 */
        activeLink = tocContent.querySelector('.toc-list-link');
      }

      var hrefKey = activeLink ? activeLink.getAttribute('href') || '' : '';
      if (hrefKey === lastHighlightHref) return;
      lastHighlightHref = hrefKey;

      clearActive();
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

      tocLayoutGen += 1;
      var gen = tocLayoutGen;
      requestAnimationFrame(function () {
        if (gen !== tocLayoutGen) return;
        if (!activeLink.classList.contains('active')) return;
        autoScrollTocIntoView(tocContent, activeLink);
      });
    }

    function applyScrollState() {
      var scrollTop = window.scrollY || document.documentElement.scrollTop;
      findHeadPosition(scrollTop);
      updateReadingPercent(scrollTop, percentEl);
    }

    var onScrollThrottled = throttle(applyScrollState, THROTTLE_MS);

    window.addEventListener('scroll', onScrollThrottled, { passive: true });
    /* 快速滚动结束后补同步（Chrome 114+ / Firefox / Safari 18+；旧浏览器不触发即可） */
    window.addEventListener('scrollend', applyScrollState, { passive: true });

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

    applyScrollState();

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
