import { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';
import { Monitor, Moon, Sun } from 'lucide-react';
import { ToggleGroup, ToggleGroupItem } from '../app/components/ui/toggle-group';
import { cn } from '../app/components/ui/utils';

export type ThemeToggleVariant = 'icon' | 'sidebar';

function normalizeTheme(t: string | undefined): 'light' | 'dark' | 'system' {
  if (t === 'light' || t === 'dark' || t === 'system') return t;
  return 'system';
}

/**
 * 后台主题：亮色 / 暗色 / 跟随系统（持久化键与 main 中 ThemeProvider 一致）。
 * - sidebar：三档文字分段，适合侧栏。
 * - icon：三档图标分段，适合移动端顶栏。
 */
export default function ThemeToggle({
  className,
  variant = 'icon',
}: {
  className?: string;
  variant?: ThemeToggleVariant;
}) {
  const { theme, setTheme, resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const value = normalizeTheme(theme);

  const onValueChange = (v: string) => {
    if (v === 'light' || v === 'dark' || v === 'system') {
      setTheme(v);
      return;
    }
    // Radix 单选再次点击可能清空，保持一项选中
    setTheme(value);
  };

  if (!mounted) {
    return (
      <div
        className={cn(
          variant === 'sidebar' && 'space-y-2',
          className,
        )}
        aria-hidden
      >
        {variant === 'sidebar' && (
          <span className="admin-sidebar-meta text-muted-foreground">外观与主题</span>
        )}
        <div
          className={cn(
            'h-9 rounded-md border border-border bg-muted/40',
            variant === 'sidebar' ? 'w-full' : 'w-[7.25rem]',
          )}
        />
      </div>
    );
  }

  return (
    <div
      className={cn(variant === 'sidebar' && 'space-y-2', className)}
      role="group"
      aria-label="外观与主题"
    >
      {variant === 'sidebar' && (
        <p className="admin-sidebar-meta text-muted-foreground">外观与主题</p>
      )}

      <ToggleGroup
        type="single"
        value={value}
        onValueChange={onValueChange}
        variant="outline"
        size={variant === 'sidebar' ? 'default' : 'sm'}
        className={cn(
          'border-border bg-card/80',
          variant === 'sidebar' ? 'w-full' : 'w-auto shrink-0',
        )}
      >
        <ToggleGroupItem
          value="light"
          className={cn(
            'flex-1 px-1.5 data-[state=on]:bg-primary/15 data-[state=on]:text-primary',
            variant === 'sidebar' && 'admin-sidebar-meta py-2.5',
          )}
          title="亮色模式"
          aria-label="亮色"
        >
          {variant === 'sidebar' ? (
            <span className="flex flex-col items-center gap-0.5 sm:flex-row sm:gap-1">
              <Sun className="size-3.5 shrink-0 sm:size-4" aria-hidden />
              <span>亮色</span>
            </span>
          ) : (
            <Sun className="size-4" aria-hidden />
          )}
        </ToggleGroupItem>
        <ToggleGroupItem
          value="dark"
          className={cn(
            'flex-1 px-1.5 data-[state=on]:bg-primary/15 data-[state=on]:text-primary',
            variant === 'sidebar' && 'admin-sidebar-meta py-2.5',
          )}
          title="暗色模式"
          aria-label="暗色"
        >
          {variant === 'sidebar' ? (
            <span className="flex flex-col items-center gap-0.5 sm:flex-row sm:gap-1">
              <Moon className="size-3.5 shrink-0 sm:size-4" aria-hidden />
              <span>暗色</span>
            </span>
          ) : (
            <Moon className="size-4" aria-hidden />
          )}
        </ToggleGroupItem>
        <ToggleGroupItem
          value="system"
          className={cn(
            'flex-1 px-1.5 data-[state=on]:bg-primary/15 data-[state=on]:text-primary',
            variant === 'sidebar' && 'admin-sidebar-meta py-2.5',
          )}
          title="跟随系统浅色/深色"
          aria-label="跟随系统"
        >
          {variant === 'sidebar' ? (
            <span className="flex flex-col items-center gap-0.5 sm:flex-row sm:gap-1">
              <Monitor className="size-3.5 shrink-0 sm:size-4" aria-hidden />
              <span>系统</span>
            </span>
          ) : (
            <Monitor className="size-4" aria-hidden />
          )}
        </ToggleGroupItem>
      </ToggleGroup>

      {variant === 'sidebar' && value === 'system' && (
        <p className="admin-sidebar-meta text-muted-foreground">
          跟随系统（当前为{resolvedTheme === 'dark' ? '暗色' : '亮色'}）
        </p>
      )}
    </div>
  );
}
