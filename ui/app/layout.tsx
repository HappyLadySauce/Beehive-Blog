import type { Metadata } from "next";

import { AuthProvider } from "@/components/auth/AuthProvider";
import { SiteHeader } from "@/components/SiteHeader";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import { ToastProvider } from "@/components/toast/ToastProvider";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"),
  title: {
    default: "Beehive Blog",
    template: "%s | Beehive Blog"
  },
  description: "个人博客、AI 协作创作与面向智能体的个人知识中台。",
  openGraph: {
    title: "Beehive Blog",
    description: "个人博客、AI 协作创作与面向智能体的个人知识中台。",
    type: "website"
  }
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="zh-CN" data-scroll-behavior="smooth" data-theme="light" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeBootstrapScript }} />
      </head>
      <body>
        <ThemeProvider>
          <AuthProvider>
            <ToastProvider>
              <div className="app-shell">
                <SiteHeader />
                {children}
              </div>
            </ToastProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}

const themeBootstrapScript = `
(function () {
  var key = "beehive.theme";
  var preference = "system";
  try {
    var stored = window.localStorage && window.localStorage.getItem(key);
    if (stored === "system" || stored === "light" || stored === "dark") preference = stored;
  } catch (_) {}
  var media = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;
  function resolved() {
    return preference === "system" ? (media && media.matches ? "dark" : "light") : preference;
  }
  function apply() {
    document.documentElement.setAttribute("data-theme", resolved());
  }
  function persist(next) {
    preference = next;
    try {
      if (window.localStorage) window.localStorage.setItem(key, next);
    } catch (_) {}
    apply();
  }
  apply();
  if (media && media.addEventListener) media.addEventListener("change", apply);
  document.addEventListener("click", function (event) {
    var target = event.target && event.target.closest ? event.target.closest("[data-theme-toggle]") : null;
    if (!target) return;
    var next = preference === "system" ? "light" : preference === "light" ? "dark" : "system";
    persist(next);
  }, true);
})();
`;
