import type { Metadata } from "next";

import { AuthProvider } from "@/components/auth/AuthProvider";
import { SiteHeader } from "@/components/SiteHeader";
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
    <html lang="zh-CN">
      <body>
        <AuthProvider>
          <div className="app-shell">
            <SiteHeader />
            {children}
          </div>
        </AuthProvider>
      </body>
    </html>
  );
}
