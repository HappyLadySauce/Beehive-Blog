/**
 * 评论区预留：文章由 Hexo 静态生成，评论列表在运行时从 Beehive API 拉取。
 * 需在 post 模板中提供 #beehive-comments-root 与 data-beehive-id。
 */
(function () {
  'use strict';

  function apiBaseFromRoot(root) {
    var fromDom = root.getAttribute('data-api-base') || '';
    if (fromDom) return fromDom.replace(/\/$/, '');
    if (typeof window.__BEEHIVE_API_BASE__ === 'string' && window.__BEEHIVE_API_BASE__) {
      return window.__BEEHIVE_API_BASE__.replace(/\/$/, '');
    }
    return '';
  }

  async function loadComments(articleId, apiBase) {
    var url = apiBase + '/api/v1/articles/' + encodeURIComponent(articleId) + '/comments';
    var res = await fetch(url, { credentials: 'omit' });
    if (!res.ok) return null;
    var json = await res.json();
    if (json && json.code === 200) return json.data;
    return null;
  }

  function renderPlaceholder(root, message) {
    root.innerHTML = '<p class="beehive-comments-root__placeholder">' + message + '</p>';
  }

  document.addEventListener('DOMContentLoaded', function () {
    var root = document.getElementById('beehive-comments-root');
    if (!root) return;
    var id = root.getAttribute('data-beehive-id');
    if (!id) return;
    var base = apiBaseFromRoot(root);
    if (!base) {
      renderPlaceholder(root, '未配置 API 地址，无法加载评论。');
      return;
    }
    loadComments(id, base).then(function (data) {
      if (data == null) {
        renderPlaceholder(root, '评论接口暂不可用（将随后端实现自动启用）。');
        return;
      }
      renderPlaceholder(root, '评论加载成功（渲染逻辑可随产品细化）。');
    }).catch(function () {
      renderPlaceholder(root, '评论加载失败。');
    });
  });
})();
