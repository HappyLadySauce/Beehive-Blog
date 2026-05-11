import { FileText } from "lucide-react";

import { StudioPlaceholderPage } from "@/components/studio/StudioPlaceholderPage";

export default function StudioContentPage() {
  return (
    <StudioPlaceholderPage
      description="管理文章、笔记、项目与发布状态。当前仅提供安全的只读骨架。"
      emptyMessage="内容工作台等待 content API 接入"
      icon={FileText}
      title="内容"
    />
  );
}
