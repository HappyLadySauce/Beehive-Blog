import { ShieldCheck } from "lucide-react";

import { StudioPlaceholderPage } from "@/components/studio/StudioPlaceholderPage";

export default function StudioPermissionsPage() {
  return (
    <StudioPlaceholderPage
      description="集中展示发布状态、人类可见性、AI 可读性与管理员访问边界。"
      emptyMessage="权限矩阵等待 content RBAC 接入"
      icon={ShieldCheck}
      title="权限"
    />
  );
}
