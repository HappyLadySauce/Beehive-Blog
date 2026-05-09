import type { Metadata } from "next";

import { RegisterForm } from "@/components/auth/RegisterForm";

export const metadata: Metadata = {
  title: "注册",
  robots: {
    index: false,
    follow: false
  }
};

export default function RegisterPage() {
  return (
    <main className="page auth-page">
      <section>
        <p className="eyebrow">Create account</p>
        <h1>创建 Beehive 账号</h1>
        <p className="muted">注册成功后会自动登录，并进入 Studio 工作台。</p>
      </section>
      <section className="auth-panel" aria-label="注册表单">
        <RegisterForm />
      </section>
    </main>
  );
}
