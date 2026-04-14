import { prepare, layout } from '/vendor/pretext/layout.js';

const TITLE_SELECTOR = '[data-pretext-title]';
const EXCERPT_SELECTOR = '[data-pretext-excerpt]';
const BALANCE_SELECTOR = '[data-pretext-balance]';

const preparedCache = new WeakMap();

function toFontShorthand(style) {
  if (style.font && style.font !== '') {
    return style.font;
  }

  const fontStyle = style.fontStyle || 'normal';
  const fontVariant = style.fontVariant || 'normal';
  const fontWeight = style.fontWeight || '400';
  const fontSize = style.fontSize || '16px';
  const fontFamily = style.fontFamily || 'sans-serif';
  return `${fontStyle} ${fontVariant} ${fontWeight} ${fontSize} ${fontFamily}`;
}

function toLineHeightPx(style) {
  const raw = style.lineHeight;
  if (!raw || raw === 'normal') {
    const fontSize = parseFloat(style.fontSize);
    return Number.isFinite(fontSize) ? Math.round(fontSize * 1.35) : 20;
  }
  const asPx = parseFloat(raw);
  if (Number.isFinite(asPx)) {
    if (String(raw).endsWith('px')) return Math.round(asPx);
    const fs = parseFloat(style.fontSize);
    if (String(raw).endsWith('rem') && Number.isFinite(fs)) {
      return Math.round(asPx * 16);
    }
    if (Number.isFinite(fs)) {
      return Math.round(asPx * fs);
    }
    return Math.round(asPx * 16);
  }
  return 20;
}

function measureElement(element, extraPadding = 0) {
  const width = element.clientWidth;
  if (!width) return;

  const style = window.getComputedStyle(element);
  const text = element.textContent ? element.textContent.trim() : '';
  if (!text) return;

  const fontKey = toFontShorthand(style);
  const lineHeightPx = toLineHeightPx(style);

  try {
    let entry = preparedCache.get(element);
    if (!entry || entry.text !== text || entry.fontKey !== fontKey) {
      entry = {
        text,
        fontKey,
        prepared: prepare(text, fontKey, { whiteSpace: 'normal' }),
      };
      preparedCache.set(element, entry);
    }

    const result = layout(entry.prepared, width, lineHeightPx);
    element.style.minHeight = `${Math.ceil(result.height + extraPadding)}px`;
    element.setAttribute('data-pretext-lines', String(result.lineCount));
  } catch (error) {
    /* 测量失败时保留 CSS 流式布局 */
  }
}

function refreshPretextLayout() {
  document.querySelectorAll(TITLE_SELECTOR).forEach((element) => measureElement(element, 2));
  document.querySelectorAll(EXCERPT_SELECTOR).forEach((element) => measureElement(element, 2));
  document.querySelectorAll(BALANCE_SELECTOR).forEach((element) => measureElement(element, 4));
}

let resizeTimer = 0;

function scheduleRefresh() {
  window.clearTimeout(resizeTimer);
  resizeTimer = window.setTimeout(refreshPretextLayout, 80);
}

async function boot() {
  if (document.fonts && document.fonts.ready) {
    try {
      await document.fonts.ready;
    } catch (error) {}
  }

  refreshPretextLayout();
  window.addEventListener('resize', scheduleRefresh, { passive: true });
  document.addEventListener('hls:theme-change', scheduleRefresh);
}

boot();
