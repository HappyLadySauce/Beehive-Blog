import type { ReactNode } from "react";

import { ThemeToggleButton } from "@/components/theme/ThemeProvider";
import styles from "./Studio.module.css";

type StudioTopbarProps = {
  eyebrow: string;
  title: string;
  description: string;
  actions?: ReactNode;
};

export function StudioTopbar({ actions, description, eyebrow, title }: StudioTopbarProps) {
  return (
    <div className={styles.topbar}>
      <div>
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      <div className={styles.actions}>
        <ThemeToggleButton className="icon-button" />
        {actions}
      </div>
    </div>
  );
}
