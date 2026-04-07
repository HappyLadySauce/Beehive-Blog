import { Outlet, Navigate, useLocation } from 'react-router-dom';
import { Menu } from 'lucide-react';
import { useState } from 'react';
import Sidebar from '../app/components/Sidebar';
import { useAuthStore } from '../store/authStore';
import { Drawer, DrawerContent } from '../app/components/ui/drawer';

export default function AdminLayout() {
  const { token } = useAuthStore();
  const location = useLocation();
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  const pageTitleMap: Record<string, string> = {
    '/': '概览',
    '/articles': '文章',
    '/comments': '评论',
    '/attachments': '附件',
    '/categories': '分类',
    '/tags': '标签',
    '/users': '用户',
    '/settings': '设置',
  };
  const pageTitle =
    Object.entries(pageTitleMap).find(([path]) =>
      path === '/'
        ? location.pathname === '/'
        : location.pathname.startsWith(path),
    )?.[1] ?? '管理后台';

  return (
    <div className="admin-shell flex h-screen overflow-hidden bg-gray-50">
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
        <header className="lg:hidden sticky top-0 z-30 flex h-[var(--admin-mobile-header-height)] items-center border-b border-gray-200 bg-white/95 px-3 backdrop-blur">
          <button
            type="button"
            onClick={() => setMobileNavOpen(true)}
            className="admin-control-md inline-flex w-10 items-center justify-center rounded border border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
            aria-label="打开导航菜单"
          >
            <Menu className="h-5 w-5" />
          </button>
          <h1 className="ml-3 text-base font-medium text-gray-900">{pageTitle}</h1>
        </header>
        <div className="p-4 md:p-6 lg:p-8 xl:p-10">
        <Outlet />
        </div>
      </main>
    </div>
  );
}
