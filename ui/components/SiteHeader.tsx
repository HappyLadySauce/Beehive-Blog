"use client";

import Link from "next/link";
import { Hexagon, LogOut, PenLine, User, Wrench } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { logout } from "@/lib/api/auth";
import { useAuth } from "@/components/auth/AuthProvider";

export function SiteHeader() {
  const router = useRouter();
  const { claims, clearAuth, isAdmin, isAuthenticated, session } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const avatarText = claims?.role === "admin" ? "A" : claims?.uid ? String(claims.uid).slice(0, 1) : "";

  async function onLogout() {
    const accessToken = session?.token.access_token;
    setMenuOpen(false);
    clearAuth();
    if (accessToken) {
      await logout(accessToken).catch(() => undefined);
    }
    router.replace("/");
  }

  return (
    <header className="site-header">
      <Link className="brand" href="/" aria-label="Beehive Blog 首页">
        <Hexagon aria-hidden size={24} />
        <span>Beehive Blog</span>
      </Link>
      <nav className="main-nav" aria-label="主导航">
        <Link href="/posts">文章</Link>
        {isAuthenticated ? (
          <div className="user-menu">
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
      </nav>
    </header>
  );
}
