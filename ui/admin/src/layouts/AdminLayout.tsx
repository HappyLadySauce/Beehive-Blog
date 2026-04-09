import { Outlet, Navigate, useLocation } from 'react-router-dom';
import { Menu } from 'lucide-react';
import { useState } from 'react';
import Sidebar from '../app/components/Sidebar';
import { useAuthStore } from '../store/authStore';
import { Drawer, DrawerContent } from '../app/components/ui/drawer';
import ThemeToggle from '../components/ThemeToggle';

export default function AdminLayout() {
  const { token } = useAuthStore();
  const location = useLocation();
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  const path = location.pathname;
  let pageTitle = '管理后台';
  if (path === '/') {
    pageTitle = '概览';
  } else if (path.startsWith('/articles/edit/')) {
    pageTitle = '编辑文章';
  } else if (path.startsWith('/articles/categories')) {
    pageTitle = '文章分类';
  } else if (path.startsWith('/articles/tags')) {
    pageTitle = '文章标签';
  } else if (path.startsWith('/articles/trash')) {
    pageTitle = '文章回收站';
  } else if (path.startsWith('/articles/create')) {
    pageTitle = '新建文章';
  } else if (path.startsWith('/articles')) {
    pageTitle = '文章';
  } else if (path.startsWith('/pages/trash')) {
    pageTitle = '页面回收站';
  } else if (path.startsWith('/pages/create')) {
    pageTitle = '新建页面';
  } else if (path.startsWith('/pages/edit/')) {
    pageTitle = '编辑页面';
  } else if (path.startsWith('/pages')) {
    pageTitle = '独立页面';
  } else if (path.startsWith('/comments')) {
    pageTitle = '评论';
  } else if (path.startsWith('/attachments')) {
    pageTitle = '附件';
  } else if (path.startsWith('/users')) {
    pageTitle = '用户';
  } else if (path.startsWith('/settings')) {
    pageTitle = '设置';
  }

  return (
    <div className="admin-shell flex h-screen overflow-hidden">
      <div className="hidden lg:block">
        <Sidebar mode="desktop" />
      </div>

      <Drawer direction="left" open={mobileNavOpen} onOpenChange={setMobileNavOpen}>
        <DrawerContent className="admin-shell p-0 data-[vaul-drawer-direction=left]:w-[var(--admin-sidebar-width-drawer)] data-[vaul-drawer-direction=left]:max-w-none">
          <Sidebar
            mode="drawer"
            onNavigate={() => setMobileNavOpen(false)}
            onClose={() => setMobileNavOpen(false)}
          />
        </DrawerContent>
      </Drawer>

      <main className="min-w-0 flex-1 overflow-y-auto">
        <header className="lg:hidden sticky top-0 z-30 flex h-[var(--admin-mobile-header-height)] items-center gap-2 border-b border-border bg-card/95 px-3 backdrop-blur">
          <button
            type="button"
            onClick={() => setMobileNavOpen(true)}
            className="admin-control-md inline-flex w-10 shrink-0 items-center justify-center rounded border border-border bg-card text-foreground hover:bg-accent"
            aria-label="打开导航菜单"
          >
            <Menu className="h-5 w-5" />
          </button>
          <h1 className="min-w-0 flex-1 truncate text-base font-medium text-foreground">
            {pageTitle}
          </h1>
          <ThemeToggle />
        </header>
        <div className="p-4 md:p-6 lg:p-8 xl:p-10">
        <Outlet />
        </div>
      </main>
    </div>
  );
}
