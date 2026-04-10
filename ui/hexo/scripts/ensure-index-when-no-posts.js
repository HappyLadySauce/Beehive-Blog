/**
 * hexo-pagination 在 per_page > 0 且文章数为 0 时 total=0，不产出任何路由，根路径无 index.html。
 * 仅在「过滤后文章数为 0」时补一条与 hexo-generator-index 同构的首页，有文章时返回空交由官方 generator。
 *
 * 站点脚本由 Hexo 以 (exports, require, module, __filename, __dirname, hexo) 包装执行；
 * 必须使用形参 hexo 注册扩展，勿使用 module.exports = function (hexo)（该函数不会被调用）。
 */
'use strict';

hexo.extend.generator.register('ensure_index_when_no_posts', function (locals) {
  const config = this.config;
  const ig = config.index_generator || {};
  const posts = locals.posts.filter(function (post) {
    return !post.hidden;
  }).sort(ig.order_by || '-date');

  posts.data.sort(function (a, b) {
    return (b.sticky || 0) - (a.sticky || 0);
  });

  if (posts.length > 0) {
    return [];
  }

  const perPage = ig.per_page != null ? ig.per_page : config.per_page;
  const layout = ig.layout || ['index', 'archive'];
  const base = ig.path || '';
  const normalizedBase = base && !base.endsWith('/') ? base + '/' : base;
  const path = normalizedBase + 'index.html';

  const pageSlice = perPage ? posts.slice(0, perPage) : posts;

  return [
    {
      path: path,
      layout: layout,
      data: {
        __index: true,
        page: {
          base: normalizedBase,
          total: 1,
          current: 1,
          current_url: normalizedBase || '/',
          posts: pageSlice,
          prev: 0,
          prev_link: '',
          next: 0,
          next_link: ''
        }
      }
    }
  ];
});
