import { LayoutDashboard, FileText, FolderOpen, Tag, MessageSquare, Image, Settings, Users, X } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import ThemeToggle from '../../components/ThemeToggle';

interface SidebarProps {
  mode?: 'desktop' | 'drawer';
  onNavigate?: () => void;
  onClose?: () => void;
}

export default function Sidebar({
  mode = 'desktop',
  onNavigate,
  onClose,
}: SidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuthStore();

  const menuItems = [
    { id: 'dashboard', path: '/', label: '概览', icon: LayoutDashboard, shortcut: 'Cmd+K' },
    { id: 'articles', path: '/articles', label: '文章', icon: FileText },
    { id: 'comments', path: '/comments', label: '评论', icon: MessageSquare },
    { id: 'attachments', path: '/attachments', label: '附件', icon: Image },
    { id: 'categories', path: '/categories', label: '分类', icon: FolderOpen },
    { id: 'tags', path: '/tags', label: '标签', icon: Tag },
    { id: 'users', path: '/users', label: '用户', icon: Users },
    { id: 'settings', path: '/settings', label: '设置', icon: Settings },
  ];

  const isDrawer = mode === 'drawer';

  return (
    <aside
      className={`shrink-0 border-r border-border bg-card/90 flex flex-col backdrop-blur-md ${
        isDrawer
          ? 'h-full w-[var(--admin-sidebar-width-drawer)]'
          : 'h-screen w-[var(--admin-sidebar-width-desktop)]'
      }`}
    >
      {/* Logo */}
      <div className="border-b border-border px-5 py-4 lg:px-6 lg:py-5">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded bg-primary text-primary-foreground admin-sidebar-label font-semibold">
            B
          </div>
          <span className="admin-sidebar-label font-medium text-foreground">Beehive</span>
          {isDrawer && (
            <button
              type="button"
              onClick={onClose}
              aria-label="关闭导航菜单"
              className="ml-auto flex h-9 w-9 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground"
            >
              <X className="h-5 w-5" />
            </button>
          )}
        </div>
      </div>

      {/* Menu Items */}
      <nav className="flex-1 overflow-y-auto py-2">
        {menuItems.map((item) => {
          const Icon = item.icon;
          const isActive = location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path));

          return (
            <button
              key={item.id}
              onClick={() => {
                navigate(item.path);
                onNavigate?.();
              }}
              className={`admin-sidebar-item w-full px-4 lg:px-5 py-2.5 flex items-center justify-between gap-3 admin-sidebar-label transition-colors focus-visible:ring-2 focus-visible:ring-ring ${
                isActive
                  ? 'bg-primary/10 text-primary border-r-2 border-primary'
                  : 'text-foreground hover:bg-accent'
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Icon className="admin-sidebar-icon" />
                <span className="truncate">{item.label}</span>
              </div>
              {item.shortcut && (
                <span className="admin-sidebar-meta text-muted-foreground">{item.shortcut}</span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Theme + User */}
      <div className="border-t border-border p-4 space-y-3">
        <ThemeToggle variant="sidebar" />
        <div className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-primary to-purple-600 text-xs font-medium text-primary-foreground">
            {user?.nickname?.[0] || user?.username?.[0] || 'A'}
          </div>
          <div className="min-w-0 flex-1">
            <div className="admin-sidebar-label font-medium text-foreground truncate">
              {user?.nickname || user?.username || 'Admin'}
            </div>
            <div className="admin-sidebar-meta text-muted-foreground truncate">管理员</div>
          </div>
        </div>
      </div>
    </aside>
  );
}
