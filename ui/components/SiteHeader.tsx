import Link from "next/link";
import { Hexagon, PenLine } from "lucide-react";

export function SiteHeader() {
  return (
    <header className="site-header">
      <Link className="brand" href="/" aria-label="Beehive Blog 首页">
        <Hexagon aria-hidden size={24} />
        <span>Beehive Blog</span>
      </Link>
      <nav className="main-nav" aria-label="主导航">
        <Link href="/posts">文章</Link>
        <Link href="/studio">Studio</Link>
        <Link className="nav-action" href="/login">
          <PenLine aria-hidden size={16} />
          登录
        </Link>
      </nav>
    </header>
  );
}
