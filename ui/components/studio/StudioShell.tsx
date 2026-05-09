"use client";

import Link from "next/link";
import { BookOpen, FileText, LogOut, Search, ShieldCheck, Sparkles } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { logout } from "@/lib/api/auth";
import { useAuth } from "@/components/auth/AuthProvider";

const metrics = [
  { label: "公开内容", value: "3", icon: BookOpen },
  { label: "草稿队列", value: "0", icon: FileText },
  { label: "待审 AI 摘要", value: "0", icon: Sparkles }
];

export function StudioShell() {
  const router = useRouter();
  const { clearAuth, isAuthenticated, session } = useAuth();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/login?next=/studio");
    }
  }, [isAuthenticated, router]);

  async function onLogout() {
    const accessToken = session?.token.access_token;
    clearAuth();
    if (accessToken) {
      await logout(accessToken).catch(() => undefined);
    }
    router.replace("/");
  }

  if (!isAuthenticated) {
    return <p className="muted">正在检查登录状态...</p>;
  }

  return (
    <section className="studio-layout" aria-label="Studio 工作台">
      <aside className="studio-sidebar">
        <h1>Studio</h1>
        <nav aria-label="Studio 导航">
          <Link href="/studio"><FileText aria-hidden size={18} /> 内容</Link>
          <Link href="/studio"><Search aria-hidden size={18} /> 搜索</Link>
          <Link href="/studio"><ShieldCheck aria-hidden size={18} /> 权限</Link>
        </nav>
        <button className="secondary-button" type="button" onClick={onLogout}>
          <LogOut aria-hidden size={18} />
          登出
        </button>
      </aside>
      <div className="studio-main">
        <div className="studio-heading">
          <div>
            <p className="eyebrow">Owner workspace</p>
            <h2>内容管理与发布闸门</h2>
          </div>
          <button className="primary-button" type="button">
            <FileText aria-hidden size={18} />
            新建内容
          </button>
        </div>
        <div className="metric-grid">
          {metrics.map((metric) => {
            const Icon = metric.icon;
            return (
              <div className="metric-card" key={metric.label}>
                <Icon aria-hidden size={20} />
                <strong>{metric.value}</strong>
                <span>{metric.label}</span>
              </div>
            );
          })}
        </div>
        <div className="work-surface">
          <h3>阶段一工作区</h3>
          <p>
            当前前端先接入身份闭环与公开页 SEO。内容写入、版本、关系和 AI 审阅会在后续 content API
            稳定后接入这里。
          </p>
        </div>
      </div>
    </section>
  );
}
