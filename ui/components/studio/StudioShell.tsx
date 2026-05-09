"use client";

import Link from "next/link";
import { BookOpen, Clock3, FileText, Search, ShieldCheck, Sparkles, Tags } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/components/auth/AuthProvider";

const metrics = [
  { label: "公开内容", value: "3", icon: BookOpen },
  { label: "草稿队列", value: "0", icon: FileText },
  { label: "待审 AI 摘要", value: "0", icon: Sparkles }
];

export function StudioShell() {
  const router = useRouter();
  const { isAdmin, isAuthenticated } = useAuth();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace("/login?next=/studio");
      return;
    }
    if (!isAdmin) {
      router.replace("/");
    }
  }, [isAdmin, isAuthenticated, router]);

  if (!isAuthenticated) {
    return <p className="muted">正在检查登录状态...</p>;
  }

  if (!isAdmin) {
    return <p className="muted">当前账号无权访问 Studio，正在返回首页...</p>;
  }

  return (
    <section className="studio-layout" aria-label="Studio 工作台">
      <aside className="studio-sidebar">
        <div className="studio-sidebar__brand">
          <span>BH</span>
          <div>
            <strong>Studio</strong>
            <small>Admin workspace</small>
          </div>
        </div>
        <nav aria-label="Studio 导航">
          <Link href="/studio"><FileText aria-hidden size={18} /> 内容</Link>
          <Link href="/studio"><Search aria-hidden size={18} /> 搜索</Link>
          <Link href="/studio"><Tags aria-hidden size={18} /> 标签</Link>
          <Link href="/studio"><ShieldCheck aria-hidden size={18} /> 权限</Link>
        </nav>
      </aside>
      <div className="studio-main">
        <div className="studio-heading">
          <div>
            <p className="eyebrow">Owner workspace</p>
            <h2>内容管理与发布闸门</h2>
            <p>管理公开内容、草稿队列、发布状态与后续 AI 审阅入口。</p>
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
        <div className="studio-work-grid">
          <div className="work-surface">
            <div className="work-surface__heading">
              <Clock3 aria-hidden size={20} />
              <h3>待处理</h3>
            </div>
            <p>当前前端先接入身份闭环与公开页 SEO。内容写入、版本、关系和 AI 审阅会在后续 content API 稳定后接入这里。</p>
          </div>
          <div className="work-surface work-surface--compact">
            <h3>发布检查</h3>
            <ul>
              <li>公开页 SSR 正常</li>
              <li>Studio 仅管理员可见</li>
              <li>后端 RBAC 待 content API 接入</li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}
