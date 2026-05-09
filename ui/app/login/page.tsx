import type { Metadata } from "next";
import { Suspense } from "react";

import { LoginForm } from "@/components/auth/LoginForm";

export const metadata: Metadata = {
  title: "登录",
  robots: {
    index: false,
    follow: false
  }
};

export default function LoginPage() {
  return (
    <main className="page auth-page">
      <section>
        <p className="eyebrow">Identity</p>
        <h1>登录 Studio</h1>
        <p className="muted">支持用户名或邮箱登录，也可以通过 GitHub OAuth 完成授权。</p>
      </section>
      <section className="auth-panel" aria-label="登录表单">
        <Suspense fallback={<p className="muted">正在加载登录表单...</p>}>
          <LoginForm />
        </Suspense>
      </section>
    </main>
  );
}
