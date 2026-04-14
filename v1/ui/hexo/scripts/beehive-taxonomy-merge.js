/**
 * 将 Beehive 同步生成的 source/_data/beehive_taxonomy.json 中的 tag_map、category_map
 * 合并进 hexo.config，使标签/分类展示名与 URL slug（Beehive）一致。
 * 须使用顶层 hexo 形参（站点脚本加载约定），勿使用 module.exports = function (hexo) { ... }。
 */
'use strict';

const fs = require('node:fs');
const path = require('node:path');

hexo.extend.filter.register('after_init', () => {
  const f = path.join(hexo.source_dir, '_data', 'beehive_taxonomy.json');
  if (!fs.existsSync(f)) {
    return;
  }
  let doc;
  try {
    doc = JSON.parse(fs.readFileSync(f, 'utf8'));
  } catch (e) {
    hexo.log.warn('beehive_taxonomy.json: %s', e.message);
    return;
  }
  if (doc.tag_map && typeof doc.tag_map === 'object') {
    hexo.config.tag_map = Object.assign({}, hexo.config.tag_map || {}, doc.tag_map);
  }
  if (doc.category_map && typeof doc.category_map === 'object') {
    hexo.config.category_map = Object.assign({}, hexo.config.category_map || {}, doc.category_map);
  }
});
