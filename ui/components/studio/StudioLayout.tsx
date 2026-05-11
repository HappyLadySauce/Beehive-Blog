import type { ReactNode } from "react";

import styles from "./Studio.module.css";
import { StudioSidebar } from "./StudioSidebar";

export function StudioLayout({ children }: { children: ReactNode }) {
  return (
    <main className={styles.layout}>
      <StudioSidebar />
      <div className={styles.main}>{children}</div>
    </main>
  );
}
