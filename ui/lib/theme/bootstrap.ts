import { THEME_COOKIE_NAME, THEME_STORAGE_KEY } from "./constants";

// themeBootstrapScript runs before React hydration to avoid theme flash.
// themeBootstrapScript 在 React hydration 前执行，避免主题闪烁。
export const themeBootstrapScript = `
(function () {
  var storageKey = ${JSON.stringify(THEME_STORAGE_KEY)};
  var cookieKey = ${JSON.stringify(THEME_COOKIE_NAME)};
  var preference = "system";

  function readCookie(name) {
    var match = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/[.*+?^$\\{}()|[\\]\\\\]/g, "\\\\$&") + "=([^;]*)"));
    return match ? decodeURIComponent(match[1]) : null;
  }

  function readPreference() {
    var fromCookie = readCookie(cookieKey);
    if (fromCookie === "system" || fromCookie === "light" || fromCookie === "dark") return fromCookie;
    try {
      var stored = window.localStorage && window.localStorage.getItem(storageKey);
      if (stored === "system" || stored === "light" || stored === "dark") return stored;
    } catch (_) {}
    return "system";
  }

  preference = readPreference();
  var media = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;

  function resolved() {
    return preference === "system" ? (media && media.matches ? "dark" : "light") : preference;
  }

  function apply() {
    document.documentElement.setAttribute("data-theme", resolved());
  }

  apply();
  if (media && media.addEventListener) media.addEventListener("change", apply);
})();
`;
