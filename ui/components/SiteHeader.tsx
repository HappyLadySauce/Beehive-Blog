"use client";

import Link from "next/link";
import { Grid2X2, LogOut, PenLine, Search, User, Wrench } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { AnzhiyuRightTools } from "@/components/anzhiyu/AnzhiyuRightTools";
import { AnzhiyuSearch } from "@/components/anzhiyu/AnzhiyuSearch";
import { AnzhiyuShortcutPanel } from "@/components/anzhiyu/AnzhiyuShortcutPanel";
import { useAuth } from "@/components/auth/AuthProvider";
import { ThemeToggleButton } from "@/components/theme/ThemeProvider";
import { isStudioContentEditorPath } from "@/lib/studio/routes";

export function SiteHeader() {
  const pathname = usePathname();
  const router = useRouter();
  const { claims, clearAuth, isAdmin, isAuthenticated, loading } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const avatarText = claims?.role === "admin" ? "A" : claims?.uid ? String(claims.uid).slice(0, 1) : "";

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    function onPointerDown(event: PointerEvent) {
      if (!menuRef.current?.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    }

    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [menuOpen]);

  async function onLogout() {
    setMenuOpen(false);
    await clearAuth();
    router.replace("/");
  }

  if (isStudioContentEditorPath(pathname)) {
    return null;
  }

  return (
    <>
      <header className="site-header anzhiyu-header">
        <Link className="brand" href="/" aria-label="Beehive Blog 首页">
          <Grid2X2 aria-hidden size={22} />
          <span>Beehive</span>
        </Link>
        <div className="anz-header-actions">
          <button className="anz-header-icon" type="button" aria-label="站内搜索" onClick={() => setSearchOpen(true)}>
            <Search aria-hidden size={19} />
          </button>
          <ThemeToggleButton className="anz-header-icon" />
          {loading ? null : isAuthenticated ? (
            <div className="user-menu" ref={menuRef}>
              <button
                aria-expanded={menuOpen}
                aria-label="打开用户菜单"
                className="avatar-button"
                type="button"
                onClick={() => setMenuOpen((open) => !open)}
              >
                {avatarText ? <span>{avatarText}</span> : <User aria-hidden size={18} />}
              </button>
              {menuOpen ? (
                <div className="user-dropdown" role="menu">
                  <div className="user-dropdown__meta">
                    <strong>{isAdmin ? "管理员" : "普通用户"}</strong>
                    <span>UID {claims?.uid ?? "-"}</span>
                  </div>
                  {isAdmin ? (
                    <Link href="/studio" role="link" onClick={() => setMenuOpen(false)}>
                      <Wrench aria-hidden size={16} />
                      Studio
                    </Link>
                  ) : null}
                  <button role="menuitem" type="button" onClick={onLogout}>
                    <LogOut aria-hidden size={16} />
                    登出
                  </button>
                </div>
              ) : null}
            </div>
          ) : (
            <Link className="nav-action" href="/login">
              <PenLine aria-hidden size={16} />
              登录
            </Link>
          )}
        </div>
      </header>
      <AnzhiyuSearch open={searchOpen} onClose={() => setSearchOpen(false)} />
      <AnzhiyuShortcutPanel onSearch={() => setSearchOpen(true)} />
      <AnzhiyuRightTools onSearch={() => setSearchOpen(true)} />
    </>
  );
}
