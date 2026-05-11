import type { LucideIcon } from "lucide-react";

import styles from "./Studio.module.css";

type StudioMetricCardProps = {
  detail: string;
  icon: LucideIcon;
  label: string;
  value: string;
};

export function StudioMetricCard({ detail, icon: Icon, label, value }: StudioMetricCardProps) {
  return (
    <article className={styles.metricCard}>
      <span className={styles.metricIcon}>
        <Icon aria-hidden size={20} />
      </span>
      <strong>{value}</strong>
      <span>{label}</span>
      <small className={styles.muted}>{detail}</small>
    </article>
  );
}
