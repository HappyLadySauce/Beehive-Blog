"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";

import { isStudioContentEditorPath } from "@/lib/studio/routes";
import styles from "./Studio.module.css";
import { StudioSidebar } from "./StudioSidebar";

export function StudioLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();

  if (isStudioContentEditorPath(pathname)) {
    return <main className={styles.editorFullscreenLayout}>{children}</main>;
  }

  return (
    <main className={styles.layout}>
      <StudioSidebar />
      <div className={styles.main}>{children}</div>
    </main>
  );
}
