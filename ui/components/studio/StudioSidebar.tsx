"use client";

import Link from "next/link";
import { Menu } from "lucide-react";
import { usePathname } from "next/navigation";
import { useState } from "react";

import styles from "./Studio.module.css";
import { studioNavItems } from "./studio-data";

export function StudioSidebar() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  return (
    <aside className={styles.sidebar} aria-label="Studio 侧边导航">
      <div className={styles.brand}>
        <span className={styles.brandMark}>BH</span>
        <div className={styles.brandText}>
          <strong>Studio</strong>
          <small>Admin workspace</small>
        </div>
      </div>
      <button
        aria-controls="studio-navigation"
        aria-expanded={open}
        className={`secondary-button ${styles.mobileToggle}`}
        type="button"
        onClick={() => setOpen((value) => !value)}
      >
        <Menu aria-hidden size={18} />
        导航
      </button>
      <nav
        aria-label="Studio 导航"
        className={`${styles.nav} ${open ? "" : styles.navCollapsed}`}
        id="studio-navigation"
      >
        {studioNavItems.map((item) => {
          const Icon = item.icon;
          const active = item.href === "/studio" ? pathname === item.href : pathname.startsWith(item.href);
          return (
            <Link
              aria-current={active ? "page" : undefined}
              className={`${styles.navLink} ${active ? styles.navLinkActive : ""}`}
              href={item.href}
              key={item.href}
              prefetch={false}
              onClick={() => setOpen(false)}
            >
              <Icon aria-hidden size={18} />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
