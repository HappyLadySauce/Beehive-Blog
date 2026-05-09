import type { Metadata } from "next";
import { Suspense } from "react";

import { GithubCallback } from "@/components/auth/GithubCallback";

export const metadata: Metadata = {
  title: "GitHub 登录回调",
  robots: {
    index: false,
    follow: false
  }
};

export default function GithubCallbackPage() {
  return (
    <main className="page auth-page">
      <section>
        <p className="eyebrow">OAuth callback</p>
        <h1>GitHub 授权处理中</h1>
        <p className="muted">页面会用 GitHub 返回的 code 和 state 向后端换取登录令牌。</p>
      </section>
      <Suspense fallback={<p className="muted">正在读取 GitHub 回调参数...</p>}>
        <GithubCallback />
      </Suspense>
    </main>
  );
}
