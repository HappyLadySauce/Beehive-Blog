import type { ReactNode } from "react";

import styles from "./Studio.module.css";

type StudioPanelProps = {
  title: string;
  children: ReactNode;
  action?: ReactNode;
};

export function StudioPanel({ action, children, title }: StudioPanelProps) {
  return (
    <section className={styles.panel} aria-labelledby={title.replace(/\s+/g, "-").toLowerCase()}>
      <div className={styles.panelHeader}>
        <h2 id={title.replace(/\s+/g, "-").toLowerCase()}>{title}</h2>
        {action}
      </div>
      {children}
    </section>
  );
}
