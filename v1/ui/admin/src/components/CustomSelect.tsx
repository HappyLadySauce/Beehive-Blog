import React, {
  KeyboardEvent,
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { createPortal } from 'react-dom';
import { ChevronDown } from 'lucide-react';

interface SelectOption {
  value: string;
  label: string;
}

interface CustomSelectProps {
  value: string;
  options: SelectOption[];
  onChange: (value: string) => void;
  placeholder?: string;
  size?: 'sm' | 'md';
  disabled?: boolean;
  className?: string;
  triggerClassName?: string;
  menuClassName?: string;
  ariaLabel?: string;
}

const triggerSizeClass: Record<'sm' | 'md', string> = {
  sm: 'h-[var(--admin-control-sm)] text-[clamp(0.83rem,0.12vw+0.79rem,0.92rem)] px-2.5',
  md: 'h-[var(--admin-control-md)] text-[clamp(0.86rem,0.14vw+0.82rem,0.98rem)] px-3',
};

const MENU_MAX_PX = 256; // tailwind max-h-64

interface MenuPosition {
  top: number;
  left: number;
  width: number;
}

/**
 * 自定义下拉选择器。
 * 菜单通过 Portal 挂到 document.body + fixed 定位，避免被祖先 overflow 裁剪。
 */
export default function CustomSelect({
  value,
  options,
  onChange,
  placeholder = '请选择',
  size = 'md',
  disabled = false,
  className = '',
  triggerClassName = '',
  menuClassName = '',
  ariaLabel,
}: CustomSelectProps) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const [menuPos, setMenuPos] = useState<MenuPosition | null>(null);

  const rootRef = useRef<HTMLDivElement | null>(null);
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const menuRef = useRef<HTMLUListElement | null>(null);
  const listId = useId();

  const selected = useMemo(
    () => options.find((opt) => opt.value === value),
    [options, value],
  );

  const updateMenuPosition = useCallback(() => {
    const btn = buttonRef.current;
    if (!btn) return;
    const rect = btn.getBoundingClientRect();
    const gap = 4;
    let menuHeight = Math.min(options.length * 40 + 8, MENU_MAX_PX);
    if (menuRef.current) {
      menuHeight = Math.min(menuRef.current.scrollHeight, MENU_MAX_PX);
    }
    let top = rect.bottom + gap;
    if (top + menuHeight > window.innerHeight - 8 && rect.top - gap - menuHeight > 0) {
      top = rect.top - gap - menuHeight;
    }
    setMenuPos({
      top,
      left: rect.left,
      width: rect.width,
    });
  }, [options.length]);

  useLayoutEffect(() => {
    if (!open) {
      setMenuPos(null);
      return;
    }
    updateMenuPosition();
    const raf = requestAnimationFrame(() => updateMenuPosition());
    return () => cancelAnimationFrame(raf);
  }, [open, updateMenuPosition]);

  useEffect(() => {
    if (!open) return;
    const onScrollOrResize = () => updateMenuPosition();
    window.addEventListener('resize', onScrollOrResize);
    window.addEventListener('scroll', onScrollOrResize, true);
    return () => {
      window.removeEventListener('resize', onScrollOrResize);
      window.removeEventListener('scroll', onScrollOrResize, true);
    };
  }, [open, updateMenuPosition]);

  useEffect(() => {
    const onMouseDown = (event: MouseEvent) => {
      const t = event.target as Node;
      if (rootRef.current?.contains(t)) return;
      if (menuRef.current?.contains(t)) return;
      setOpen(false);
      setActiveIndex(-1);
    };
    document.addEventListener('mousedown', onMouseDown);
    return () => document.removeEventListener('mousedown', onMouseDown);
  }, []);

  useEffect(() => {
    if (!open) return;
    const selectedIndex = options.findIndex((opt) => opt.value === value);
    setActiveIndex(selectedIndex >= 0 ? selectedIndex : 0);
  }, [open, options, value]);

  const commitAt = (index: number) => {
    if (index < 0 || index >= options.length) return;
    const next = options[index];
    onChange(next.value);
    setOpen(false);
    setActiveIndex(index);
    buttonRef.current?.focus();
  };

  const onTriggerKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (disabled) return;
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      if (!open) {
        setOpen(true);
        return;
      }
      setActiveIndex((prev) => {
        const delta = event.key === 'ArrowDown' ? 1 : -1;
        const base = prev < 0 ? 0 : prev;
        return (base + delta + options.length) % options.length;
      });
      return;
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      setOpen((v) => !v);
      return;
    }
    if (event.key === 'Escape') {
      event.preventDefault();
      setOpen(false);
      setActiveIndex(-1);
    }
  };

  const onMenuKeyDown = (event: KeyboardEvent<HTMLUListElement>) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      setOpen(false);
      buttonRef.current?.focus();
      return;
    }
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      setActiveIndex((prev) => {
        const delta = event.key === 'ArrowDown' ? 1 : -1;
        const base = prev < 0 ? 0 : prev;
        return (base + delta + options.length) % options.length;
      });
      return;
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      commitAt(activeIndex < 0 ? 0 : activeIndex);
    }
  };

  const menuContent =
    open && menuPos ? (
      <ul
        ref={menuRef}
        id={listId}
        role="listbox"
        tabIndex={-1}
        aria-activedescendant={
          activeIndex >= 0 ? `${listId}-option-${activeIndex}` : undefined
        }
        onKeyDown={onMenuKeyDown}
        style={{
          position: 'fixed',
          top: menuPos.top,
          left: menuPos.left,
          width: menuPos.width,
          zIndex: 9999,
          maxHeight: MENU_MAX_PX,
        }}
        className={`overflow-auto rounded-md border border-border bg-popover text-popover-foreground shadow-lg py-1 text-[clamp(0.84rem,0.12vw+0.8rem,0.95rem)] ${menuClassName}`}
      >
        {options.map((opt, index) => {
          const isSelected = opt.value === value;
          const isActive = index === activeIndex;
          return (
            <li key={opt.value} id={`${listId}-option-${index}`} role="option" aria-selected={isSelected}>
              <button
                type="button"
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => commitAt(index)}
                className={`w-full text-left px-3 py-2 transition-colors ${
                  isSelected
                    ? 'bg-primary/15 text-primary'
                    : isActive
                      ? 'bg-muted text-foreground'
                      : 'text-foreground'
                }`}
              >
                {opt.label}
              </button>
            </li>
          );
        })}
      </ul>
    ) : null;

  return (
    <div className={className} ref={rootRef}>
      <button
        ref={buttonRef}
        type="button"
        disabled={disabled}
        aria-label={ariaLabel}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        onClick={() => !disabled && setOpen((v) => !v)}
        onKeyDown={onTriggerKeyDown}
        className={`w-full inline-flex items-center justify-between rounded-md border border-border bg-input-background text-foreground ${triggerSizeClass[size]} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:border-transparent disabled:bg-muted disabled:text-muted-foreground disabled:cursor-not-allowed transition-colors ${triggerClassName}`}
      >
        <span className="truncate">{selected?.label || placeholder}</span>
        <ChevronDown
          className={`w-4 h-4 text-muted-foreground transition-transform shrink-0 ${open ? 'rotate-180' : ''}`}
        />
      </button>

      {typeof document !== 'undefined' && menuContent
        ? createPortal(menuContent, document.body)
        : null}
    </div>
  );
}
