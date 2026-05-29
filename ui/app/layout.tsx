import type { Metadata } from "next";

import { AuthProvider } from "@/components/auth/AuthProvider";
import { SiteHeader } from "@/components/SiteHeader";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import { ToastProvider } from "@/components/toast/ToastProvider";
import { themeBootstrapScript } from "@/lib/theme/bootstrap";
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
        <meta name="color-scheme" content="light dark" />
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
