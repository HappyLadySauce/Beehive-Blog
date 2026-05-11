import type { LucideIcon } from "lucide-react";

import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioTopbar } from "./StudioTopbar";

type StudioPlaceholderPageProps = {
  description: string;
  emptyMessage: string;
  icon: LucideIcon;
  title: string;
};

export function StudioPlaceholderPage({ description, emptyMessage, icon: Icon, title }: StudioPlaceholderPageProps) {
  return (
    <>
      <StudioTopbar
        actions={
          <button className="secondary-button" disabled type="button">
            <Icon aria-hidden size={18} />
            暂不可用
          </button>
        }
        description={description}
        eyebrow="Studio module"
        title={title}
      />
      <StudioPanel title={title}>
        <div className={styles.emptyState}>
          <Icon aria-hidden size={28} />
          <strong>{emptyMessage}</strong>
          <span>Content API 稳定后接入真实数据与变更操作。</span>
        </div>
      </StudioPanel>
    </>
  );
}
