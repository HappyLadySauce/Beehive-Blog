import { prepare, layout } from '/vendor/pretext/layout.js';

const TITLE_SELECTOR = '[data-pretext-title]';
const EXCERPT_SELECTOR = '[data-pretext-excerpt]';
const BALANCE_SELECTOR = '[data-pretext-balance]';

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

function toLineHeight(style) {
  const numeric = parseFloat(style.lineHeight);
  if (Number.isFinite(numeric)) return numeric;
  const fontSize = parseFloat(style.fontSize);
  return Number.isFinite(fontSize) ? Math.round(fontSize * 1.5) : 24;
}

function measureElement(element, extraPadding = 0) {
  const width = element.clientWidth;
  if (!width) return;

  const style = window.getComputedStyle(element);
  const text = element.textContent ? element.textContent.trim() : '';
  if (!text) return;

  try {
    const prepared = prepare(text, toFontShorthand(style));
    const result = layout(prepared, width, toLineHeight(style));
    element.style.minHeight = `${Math.ceil(result.height + extraPadding)}px`;
    element.setAttribute('data-pretext-lines', String(result.lineCount));
  } catch (error) {
    // 发生异常时保留 CSS 自然流布局，不阻塞主内容渲染。
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
