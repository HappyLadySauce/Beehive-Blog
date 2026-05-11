import { Search } from "lucide-react";

import { StudioPlaceholderPage } from "@/components/studio/StudioPlaceholderPage";

export default function StudioSearchPage() {
  return (
    <StudioPlaceholderPage
      description="预留公开内容、私有草稿和后续 AI 语义检索入口。"
      emptyMessage="搜索索引与查询能力尚未接入"
      icon={Search}
      title="搜索"
    />
  );
}
