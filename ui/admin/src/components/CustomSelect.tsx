import React, { KeyboardEvent, useEffect, useId, useMemo, useRef, useState } from 'react';
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
  sm: 'h-8 text-sm px-2.5',
  md: 'h-10 text-sm px-3',
};

/**
 * 自定义下拉选择器。
 * 支持键盘导航、Esc 关闭、点击外部关闭、ARIA 语义。
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
  const rootRef = useRef<HTMLDivElement | null>(null);
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const listId = useId();

  const selected = useMemo(
    () => options.find((opt) => opt.value === value),
    [options, value],
  );

  useEffect(() => {
    const onMouseDown = (event: MouseEvent) => {
      if (!rootRef.current) return;
      if (!rootRef.current.contains(event.target as Node)) {
        setOpen(false);
        setActiveIndex(-1);
      }
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

  return (
    <div className={`relative ${className}`} ref={rootRef}>
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
        className={`w-full inline-flex items-center justify-between rounded-md border border-gray-300 bg-white text-gray-900 ${triggerSizeClass[size]} focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:border-transparent disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed transition-colors ${triggerClassName}`}
      >
        <span className="truncate">{selected?.label || placeholder}</span>
        <ChevronDown
          className={`w-4 h-4 text-gray-500 transition-transform ${open ? 'rotate-180' : ''}`}
        />
      </button>

      {open && (
        <ul
          id={listId}
          role="listbox"
          tabIndex={-1}
          aria-activedescendant={
            activeIndex >= 0 ? `${listId}-option-${activeIndex}` : undefined
          }
          onKeyDown={onMenuKeyDown}
          className={`absolute z-50 mt-1 max-h-64 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg py-1 ${menuClassName}`}
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
                  className={`w-full text-left px-3 py-2 text-sm transition-colors ${
                    isSelected
                      ? 'bg-blue-50 text-blue-700'
                      : isActive
                        ? 'bg-gray-50 text-gray-900'
                        : 'text-gray-700'
                  }`}
                >
                  {opt.label}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
