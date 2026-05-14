"use client";

import { KeyboardEvent, useEffect, useId, useMemo, useRef, useState } from "react";
import { Check, ChevronDown } from "lucide-react";

import styles from "./Studio.module.css";

export type StudioSelectOption = {
  value: string;
  label: string;
};

type StudioSelectProps = {
  ariaLabel: string;
  disabled?: boolean;
  options: readonly StudioSelectOption[];
  value: string;
  onChange: (value: string) => void;
};

export function StudioSelect({ ariaLabel, disabled = false, options, value, onChange }: StudioSelectProps) {
  const id = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const selected = useMemo(() => options.find((option) => option.value === value) ?? options[0], [options, value]);

  useEffect(() => {
    if (!open) return;

    function closeOnOutsideClick(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", closeOnOutsideClick);
    return () => document.removeEventListener("mousedown", closeOnOutsideClick);
  }, [open]);

  function selectValue(nextValue: string) {
    onChange(nextValue);
    setOpen(false);
  }

  function onTriggerKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key === "Escape") {
      setOpen(false);
      return;
    }
    if (event.key === "Enter" || event.key === " " || event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      setOpen(true);
    }
  }

  function onOptionKeyDown(event: KeyboardEvent<HTMLButtonElement>, optionValue: string) {
    if (event.key === "Escape") {
      setOpen(false);
      return;
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      selectValue(optionValue);
    }
  }

  return (
    <div className={styles.selectRoot} ref={rootRef}>
      <button
        aria-controls={`${id}-listbox`}
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-label={ariaLabel}
        className={styles.selectTrigger}
        disabled={disabled}
        role="combobox"
        type="button"
        onClick={() => setOpen((current) => !current)}
        onKeyDown={onTriggerKeyDown}
      >
        <span>{selected?.label ?? ""}</span>
        <ChevronDown aria-hidden className={open ? styles.selectIconOpen : styles.selectIcon} size={18} />
      </button>

      {open ? (
        <div className={styles.selectPopover} id={`${id}-listbox`} role="listbox">
          {options.map((option) => {
            const active = option.value === value;
            return (
              <button
                aria-selected={active}
                className={active ? styles.selectOptionActive : styles.selectOption}
                key={option.value}
                role="option"
                tabIndex={0}
                type="button"
                onClick={() => selectValue(option.value)}
                onKeyDown={(event) => onOptionKeyDown(event, option.value)}
              >
                <span>{option.label}</span>
                {active ? <Check aria-hidden size={16} /> : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}
