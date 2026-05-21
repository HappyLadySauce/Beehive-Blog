import { BookOpen, Database, FileArchive, FileText, Search, Settings, ShieldCheck, Sparkles, Tags, Users } from "lucide-react";

export const studioNavItems = [
  { href: "/studio", label: "总览", icon: BookOpen },
  { href: "/studio/content", label: "内容", icon: FileText },
  { href: "/studio/search", label: "搜索", icon: Search },
  { href: "/studio/tags", label: "标签", icon: Tags },
  { href: "/studio/attachments", label: "附件", icon: FileArchive },
  { href: "/studio/users", label: "用户", icon: Users },
  { href: "/studio/storage", label: "存储", icon: Database },
  { href: "/studio/permissions", label: "权限", icon: ShieldCheck },
  { href: "/studio/settings", label: "设置", icon: Settings }
];

export const studioMetrics = [
  { label: "公开内容", value: "3", detail: "SSR public entries", icon: BookOpen, tone: "green" },
  { label: "草稿队列", value: "0", detail: "Content API pending", icon: FileText, tone: "amber" },
  { label: "待审 AI 摘要", value: "0", detail: "Review gate ready", icon: Sparkles, tone: "blue" }
];

export const releaseChecks = [
  { label: "公开页 SSR 正常", state: "ready" },
  { label: "Studio 仅管理员可见", state: "ready" },
  { label: "BFF Cookie 会话已接管", state: "ready" },
  { label: "后端 RBAC 待 content API 接入", state: "pending" }
];
