"use client";

import { slugifyFromName, type SlugMode } from "@/lib/slug";
import styles from "./Studio.module.css";

type StudioSlugFieldProps = {
  slugMode: SlugMode;
  onSlugModeChange: (mode: SlugMode) => void;
  slug: string;
  onSlugChange: (slug: string) => void;
  sourceValue: string;
  sourceLabel: string;
  maxLength: number;
  disabled?: boolean;
  pathPreview?: string | null;
};

export function StudioSlugField({
  slugMode,
  onSlugModeChange,
  slug,
  onSlugChange,
  sourceValue,
  sourceLabel,
  maxLength,
  disabled = false,
  pathPreview = null
}: StudioSlugFieldProps) {
  const isAuto = slugMode === "auto";

  function switchToAuto() {
    if (disabled || isAuto) return;
    onSlugModeChange("auto");
    onSlugChange(slugifyFromName(sourceValue, { maxLength }));
  }

  function switchToManual() {
    if (disabled || !isAuto) return;
    onSlugModeChange("manual");
  }

  return (
    <div className={styles.field}>
      <div className={styles.slugFieldHeader}>
        <span>Slug</span>
        <div aria-label="Slug 生成方式" className={styles.editorModeTabs} role="group">
          <button
            aria-pressed={isAuto}
            className={isAuto ? styles.editorModeTabActive : styles.editorModeTab}
            disabled={disabled}
            type="button"
            onClick={switchToAuto}
          >
            自动
          </button>
          <button
            aria-pressed={!isAuto}
            className={!isAuto ? styles.editorModeTabActive : styles.editorModeTab}
            disabled={disabled}
            type="button"
            onClick={switchToManual}
          >
            自定义
          </button>
        </div>
      </div>
      <input
        aria-label="Slug"
        readOnly={isAuto}
        value={slug}
        onChange={(event) => onSlugChange(event.target.value)}
      />
      <p className={styles.slugFieldHint}>
        {isAuto
          ? `根据${sourceLabel}自动生成（拼音 + 英文），可切换为自定义。`
          : `自定义 Slug；修改${sourceLabel}不会更新 Slug。`}
      </p>
      {pathPreview ? <p className={styles.slugFieldHint}>公开路径：{pathPreview}</p> : null}
    </div>
  );
}
