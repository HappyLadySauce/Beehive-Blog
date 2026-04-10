/**
 * 标签链接：Beehive 字色（内联 color）；字号由 CSS .tag-cloud-link 控制。
 * tag_colors 以 slug 为键；若 Hexo 的 tag.slug 与库不一致，则按 tag_map[name]、name 等回退查找。
 */
'use strict';

/**
 * @param {object} tag - Hexo Tag
 * @param {object} [tax] - site.data.beehive_taxonomy
 * @returns {string} #RRGGBB 或 ''
 */
hexo.extend.helper.register('beehiveTagColorHex', (tag, tax) => {
  if (!tag) return '';
  const colors = (tax && tax.tag_colors) || {};
  const tmap = (tax && tax.tag_map) || {};
  const pick = (key) => {
    if (key == null || key === '') return '';
    const s = String(key).trim();
    if (!s) return '';
    if (colors[s]) return String(colors[s]).trim();
    const low = s.toLowerCase();
    if (colors[low]) return String(colors[low]).trim();
    return '';
  };
  const slug = tag.slug != null ? String(tag.slug).trim() : '';
  const name = tag.name != null ? String(tag.name).trim() : '';
  let c = pick(slug);
  if (c) return c;
  c = pick(name);
  if (c) return c;
  if (name && tmap[name] != null) {
    c = pick(String(tmap[name]).trim());
  }
  return c || '';
});

hexo.extend.helper.register('beehiveTagCloudStyle', (colorHex) => {
  const c = colorHex != null && String(colorHex).trim() !== '' ? String(colorHex).trim() : '';
  return c ? `color:${c}` : '';
});
