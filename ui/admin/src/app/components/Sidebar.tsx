import { LayoutDashboard, FileText, FolderOpen, Tag, MessageSquare, Image, Settings, Users, Plus } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';

export default function Sidebar() {
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

  return (
    <aside className="w-56 shrink-0 bg-white border-r border-gray-200 flex flex-col h-screen">
      {/* Logo */}
      <div className="px-6 py-5 border-b border-gray-200">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 bg-blue-600 rounded flex items-center justify-center text-white text-sm font-semibold">
            B
          </div>
          <span className="text-base font-medium text-gray-900">Beehive</span>
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
              onClick={() => navigate(item.path)}
              className={`w-full px-5 py-2 flex items-center justify-between gap-3 text-sm transition-colors ${
                isActive
                  ? 'bg-blue-50 text-blue-600 border-r-2 border-blue-600'
                  : 'text-gray-700 hover:bg-gray-50'
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Icon className="w-4 h-4" />
                <span>{item.label}</span>
              </div>
              {item.shortcut && (
                <span className="text-xs text-gray-400">{item.shortcut}</span>
              )}
            </button>
          );
        })}
      </nav>

      {/* User Info */}
      <div className="p-4 border-t border-gray-200">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-500 rounded-full flex items-center justify-center text-white text-xs font-medium">
            {user?.nickname?.[0] || user?.username?.[0] || 'A'}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-gray-900 truncate">
              {user?.nickname || user?.username || 'Admin'}
            </div>
            <div className="text-xs text-gray-500 truncate">管理员</div>
          </div>
        </div>
      </div>
    </aside>
  );
}
