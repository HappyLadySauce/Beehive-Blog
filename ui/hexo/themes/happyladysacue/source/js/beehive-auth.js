/**
 * Beehive-Blog 账号前端集成：
 * - 导航用户态同步
 * - 登录页和注册页表单挂载
 * - 访问令牌本地缓存与退出登录
 */
(function () {
  'use strict';

  var TOKEN_KEY = 'beehive_token';
  var REFRESH_KEY = 'beehive_refresh_token';

  function getApiBase() {
    var value = typeof window.__BEEHIVE_API_BASE__ === 'string' ? window.__BEEHIVE_API_BASE__ : '';
    return value.replace(/\/$/, '');
  }

  function getToken() {
    try {
      return localStorage.getItem(TOKEN_KEY) || '';
    } catch (err) {
      return '';
    }
  }

  function setTokens(payload) {
    if (!payload || typeof payload !== 'object') return;
    try {
      if (payload.token) localStorage.setItem(TOKEN_KEY, payload.token);
      if (payload.refreshToken) localStorage.setItem(REFRESH_KEY, payload.refreshToken);
    } catch (err) {}
  }

  function clearTokens() {
    try {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(REFRESH_KEY);
    } catch (err) {}
  }

  function fetchJSON(path, options) {
    var opts = options || {};
    var headers = { Accept: 'application/json' };
    if (opts.headers) {
      Object.keys(opts.headers).forEach(function (key) {
        headers[key] = opts.headers[key];
      });
    }
    if (opts.body && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }

    return fetch(getApiBase() + path, {
      method: opts.method || 'GET',
      headers: headers,
      body: opts.body,
      credentials: 'omit',
    }).then(function (response) {
      var contentType = response.headers.get('content-type') || '';
      if (contentType.indexOf('application/json') >= 0) {
        return response.json().then(function (body) {
          return { response: response, body: body };
        });
      }
      return response.text().then(function (body) {
        return { response: response, body: body };
      });
    });
  }

  var defaultAvatar = '';

  function setMessage(element, text, type) {
    if (!element) return;
    element.textContent = text || '';
    element.className = 'beehive-auth-msg' + (type ? ' ' + type : '');
  }

  function closeDropdown() {
    var dropdown = document.getElementById('beehive-account-dropdown');
    var trigger = document.getElementById('beehive-account-trigger');
    if (dropdown) dropdown.classList.add('is-hidden');
    if (trigger) trigger.setAttribute('aria-expanded', 'false');
  }

  function setGuestState() {
    var guest = document.getElementById('beehive-dropdown-guest');
    var user = document.getElementById('beehive-dropdown-user');
    var avatar = document.getElementById('beehive-nav-avatar');
    if (guest) guest.classList.remove('is-hidden');
    if (user) user.classList.add('is-hidden');
    if (avatar && defaultAvatar) avatar.src = defaultAvatar;
  }

  function setUserState(me) {
    var guest = document.getElementById('beehive-dropdown-guest');
    var user = document.getElementById('beehive-dropdown-user');
    var nickname = document.getElementById('beehive-nav-nickname');
    var avatar = document.getElementById('beehive-nav-avatar');
    if (guest) guest.classList.add('is-hidden');
    if (user) user.classList.remove('is-hidden');
    if (nickname) nickname.textContent = (me && (me.nickname || me.username)) || 'Beehive 用户';
    if (avatar) avatar.src = me && me.avatar ? me.avatar : defaultAvatar;
  }

  function refreshNav() {
    if (!document.getElementById('beehive-account-wrap')) return Promise.resolve();
    var token = getToken();
    if (!token || !getApiBase()) {
      setGuestState();
      return Promise.resolve();
    }

    return fetchJSON('/api/v1/user/me', {
      headers: { Authorization: 'Bearer ' + token },
    })
      .then(function (result) {
        if (result.response.ok && result.body && result.body.data) {
          setUserState(result.body.data);
          return;
        }
        clearTokens();
        setGuestState();
      })
      .catch(function () {
        clearTokens();
        setGuestState();
      });
  }

  function bindNav() {
    var trigger = document.getElementById('beehive-account-trigger');
    var dropdown = document.getElementById('beehive-account-dropdown');
    var avatar = document.getElementById('beehive-nav-avatar');
    if (avatar) defaultAvatar = avatar.src;
    if (!trigger || !dropdown) return refreshNav();

    trigger.addEventListener('click', function (event) {
      event.preventDefault();
      event.stopPropagation();
      var hidden = dropdown.classList.toggle('is-hidden');
      trigger.setAttribute('aria-expanded', hidden ? 'false' : 'true');
    });

    dropdown.addEventListener('click', function (event) {
      event.stopPropagation();
    });

    document.addEventListener('click', closeDropdown);
    document.addEventListener('keydown', function (event) {
      if (event.key === 'Escape') closeDropdown();
    });

    var logout = document.getElementById('beehive-nav-logout');
    if (logout) {
      logout.addEventListener('click', function () {
        var token = getToken();
        if (!token || !getApiBase()) {
          clearTokens();
          setGuestState();
          closeDropdown();
          return;
        }

        fetchJSON('/api/v1/user/logout', {
          method: 'POST',
          headers: {
            Authorization: 'Bearer ' + token,
            'Content-Type': 'application/json',
          },
          body: '{}',
        }).finally(function () {
          clearTokens();
          setGuestState();
          closeDropdown();
        });
      });
    }

    return refreshNav();
  }

  function initLoginPage() {
    var root = document.getElementById('beehive-login-root');
    if (!root) return;

    root.className = 'beehive-auth-page';
    root.innerHTML =
      '<h2>登录</h2>' +
      '<p>使用 Beehive 账号进入你的内容控制台。</p>' +
      '<form id="beehive-login-form">' +
      '<label for="bh-login-account">账号</label>' +
      '<input id="bh-login-account" name="account" type="text" autocomplete="username" required />' +
      '<label for="bh-login-password">密码</label>' +
      '<input id="bh-login-password" name="password" type="password" autocomplete="current-password" required />' +
      '<button type="submit">登录</button>' +
      '</form>' +
      '<div id="beehive-login-msg" class="beehive-auth-msg" aria-live="polite"></div>' +
      '<p class="beehive-auth-footer">还没有账号？<a href="/register/">立即注册</a></p>';

    var form = document.getElementById('beehive-login-form');
    var message = document.getElementById('beehive-login-msg');
    var submit = form.querySelector('button[type="submit"]');

    form.addEventListener('submit', function (event) {
      event.preventDefault();
      if (!getApiBase()) {
        setMessage(message, '未配置 API 地址，无法登录。', 'err');
        return;
      }

      var account = (document.getElementById('bh-login-account') || {}).value || '';
      var password = (document.getElementById('bh-login-password') || {}).value || '';
      setMessage(message, '登录中...', '');
      submit.disabled = true;

      fetchJSON('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ account: account, password: password }),
      })
        .then(function (result) {
          if (result.response.ok && result.body && result.body.data) {
            setTokens(result.body.data);
            setMessage(message, '登录成功，正在跳转...', 'ok');
            window.location.href = '/';
            return;
          }

          setMessage(message, (result.body && result.body.message) || '登录失败，请检查账号与密码。', 'err');
        })
        .catch(function (error) {
          setMessage(message, (error && error.message) || '网络错误，请稍后重试。', 'err');
        })
        .finally(function () {
          submit.disabled = false;
        });
    });
  }

  function initRegisterPage() {
    var root = document.getElementById('beehive-register-root');
    if (!root) return;

    root.className = 'beehive-auth-page';
    root.innerHTML =
      '<h2>注册</h2>' +
      '<p>创建新的 Beehive 账号，开始发布与管理你的内容。</p>' +
      '<form id="beehive-register-form">' +
      '<label for="bh-reg-user">用户名</label>' +
      '<input id="bh-reg-user" name="username" type="text" autocomplete="username" required minlength="3" maxlength="20" />' +
      '<label for="bh-reg-email">邮箱</label>' +
      '<input id="bh-reg-email" name="email" type="email" autocomplete="email" required />' +
      '<label for="bh-reg-pass">密码</label>' +
      '<input id="bh-reg-pass" name="password" type="password" autocomplete="new-password" required minlength="6" maxlength="20" />' +
      '<button type="submit">注册</button>' +
      '</form>' +
      '<div id="beehive-register-msg" class="beehive-auth-msg" aria-live="polite"></div>' +
      '<p class="beehive-auth-footer">已有账号？<a href="/login/">去登录</a></p>';

    var form = document.getElementById('beehive-register-form');
    var message = document.getElementById('beehive-register-msg');
    var submit = form.querySelector('button[type="submit"]');

    form.addEventListener('submit', function (event) {
      event.preventDefault();
      if (!getApiBase()) {
        setMessage(message, '未配置 API 地址，无法注册。', 'err');
        return;
      }

      var username = (document.getElementById('bh-reg-user') || {}).value || '';
      var email = (document.getElementById('bh-reg-email') || {}).value || '';
      var password = (document.getElementById('bh-reg-pass') || {}).value || '';
      setMessage(message, '注册中...', '');
      submit.disabled = true;

      fetchJSON('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          username: username,
          email: email,
          password: password,
        }),
      })
        .then(function (result) {
          if (result.response.ok && result.body && result.body.data) {
            setTokens(result.body.data);
            setMessage(message, '注册成功，正在跳转...', 'ok');
            window.location.href = '/';
            return;
          }

          setMessage(message, (result.body && result.body.message) || '注册失败，请稍后再试。', 'err');
        })
        .catch(function (error) {
          setMessage(message, (error && error.message) || '网络错误，请稍后重试。', 'err');
        })
        .finally(function () {
          submit.disabled = false;
        });
    });
  }

  function boot() {
    bindNav();
    initLoginPage();
    initRegisterPage();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})();
