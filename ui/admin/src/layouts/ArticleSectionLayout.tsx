import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { FileText, FolderOpen, Tags, Trash2, type LucideIcon } from 'lucide-react';

const sectionTitle = (pathname: string): string => {
  if (pathname === '/articles' || pathname === '/articles/') return '文章';
  if (pathname.startsWith('/articles/categories')) return '文章分类';
  if (pathname.startsWith('/articles/tags')) return '文章标签';
  if (pathname.startsWith('/articles/trash')) return '文章回收站';
  return '文章';
};

const sectionIcon = (pathname: string): LucideIcon => {
  if (pathname.startsWith('/articles/categories')) return FolderOpen;
  if (pathname.startsWith('/articles/tags')) return Tags;
  if (pathname.startsWith('/articles/trash')) return Trash2;
  return FileText;
};

export default function ArticleSectionLayout() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const title = sectionTitle(pathname);
  const TitleIcon = sectionIcon(pathname);

  const navBtn = (target: string, label: string, active: boolean) => (
    <button
      type="button"
      onClick={() => navigate(target)}
      className={`rounded-md border px-3 py-1.5 text-sm transition-colors ${
        active
          ? 'border-primary bg-primary/10 text-primary'
          : 'border-border bg-background text-foreground hover:bg-accent'
      }`}
    >
      {label}
    </button>
  );

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <TitleIcon className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
          <h1 className="text-xl font-semibold tracking-tight text-foreground">{title}</h1>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {navBtn('/articles/categories', '分类', pathname.startsWith('/articles/categories'))}
          {navBtn('/articles/tags', '标签', pathname.startsWith('/articles/tags'))}
          {navBtn('/articles/trash', '回收站', pathname.startsWith('/articles/trash'))}
          <button
            type="button"
            onClick={() => navigate('/articles/create')}
            className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
          >
            新建
          </button>
        </div>
      </header>
      <Outlet />
    </div>
  );
}
