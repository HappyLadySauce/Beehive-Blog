import { Tags } from "lucide-react";

import { StudioPlaceholderPage } from "@/components/studio/StudioPlaceholderPage";

export default function StudioTagsPage() {
  return (
    <StudioPlaceholderPage
      description="统一整理标签、专题与内容关系，避免 Public 与 Studio 信息架构分裂。"
      emptyMessage="标签管理等待关系模型接入"
      icon={Tags}
      title="标签"
    />
  );
}
