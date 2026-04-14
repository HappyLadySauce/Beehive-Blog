import React, { useEffect, useRef, useState } from 'react';
import { Upload, X } from 'lucide-react';
import AdminModal from './AdminModal';

export function defaultStagedFileKey(f: File): string {
  return `${f.name}:${f.size}:${f.lastModified}`;
}

export function formatStagedFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** 按 name+size+lastModified 去重后追加，常用于附件等多文件上传。 */
export function mergeStagedFilesDedupe(incoming: File[], prev: File[]): File[] {
  const key = defaultStagedFileKey;
  const seen = new Set(prev.map(key));
  const next = [...prev];
  for (const f of incoming) {
    const k = key(f);
    if (!seen.has(k)) {
      seen.add(k);
      next.push(f);
    }
  }
  return next;
}

export type StagedFileUploadModalProps = {
  open: boolean;
  title: string;
  description: React.ReactNode;
  /** 拖放区下方列出的扩展名或类型说明 */
  extensionsHint?: string;
  /** 传给原生 input；不传表示不限制 */
  accept?: string;
  multiple?: boolean;
  loading?: boolean;
  /** 与 loading 一并用于禁用拖放区与移除按钮 */
  disabled?: boolean;
  confirmLabel?: string;
  confirmDisabled?: boolean;
  maxWidth?: React.ComponentProps<typeof AdminModal>['maxWidth'];
  stagedFiles: File[];
  onStagedFilesChange: (files: File[]) => void;
  /** 用户通过选择或拖放添加文件时调用，由父组件决定如何合并/过滤列表 */
  onPickFiles: (files: File[]) => void;
  onClose: () => void;
  onConfirm: () => void;
  fileKey?: (f: File) => string;
  formatFileSize?: (bytes: number) => string;
};

/**
 * 管理后台通用「待上传文件列表 + 虚线拖放区」弹窗，与文章导入 / 附件上传等场景复用。
 */
export default function StagedFileUploadModal({
  open,
  title,
  description,
  extensionsHint,
  accept,
  multiple = true,
  loading = false,
  disabled = false,
  confirmLabel = '确定',
  confirmDisabled = false,
  maxWidth = '2xl',
  stagedFiles,
  onStagedFilesChange,
  onPickFiles,
  onClose,
  onConfirm,
  fileKey = defaultStagedFileKey,
  formatFileSize = formatStagedFileSize,
}: StagedFileUploadModalProps) {
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const busy = loading || disabled;

  useEffect(() => {
    if (!open) {
      setDragActive(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }, [open]);

  if (!open) return null;

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    e.target.value = '';
    if (!files?.length) return;
    onPickFiles(Array.from(files));
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (busy) return;
    onPickFiles(Array.from(e.dataTransfer.files || []));
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!e.currentTarget.contains(e.relatedTarget as Node)) {
      setDragActive(false);
    }
  };

  return (
    <AdminModal
      title={title}
      maxWidth={maxWidth}
      onClose={onClose}
      onConfirm={onConfirm}
      confirmLabel={confirmLabel}
      loading={loading}
      confirmDisabled={confirmDisabled}
    >
      <div className="max-h-[min(70vh,560px)] space-y-4 overflow-y-auto pr-1">
        <div className="text-sm text-muted-foreground">{description}</div>
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          accept={accept}
          multiple={multiple}
          onChange={handleFileInputChange}
        />
        <div
          role="button"
          tabIndex={0}
          onKeyDown={(ev) => {
            if (ev.key === 'Enter' || ev.key === ' ') {
              ev.preventDefault();
              if (!busy) fileInputRef.current?.click();
            }
          }}
          onClick={() => {
            if (!busy) fileInputRef.current?.click();
          }}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          className={`flex min-h-[220px] cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed px-4 py-10 text-center transition-colors ${
            dragActive ? 'border-primary bg-primary/10' : 'border-border bg-muted/30'
          } ${busy ? 'pointer-events-none opacity-60' : ''}`}
        >
          <Upload className="h-10 w-10 text-muted-foreground" aria-hidden />
          <span className="text-sm font-medium text-foreground">点击选择文件或拖放到此处</span>
          {extensionsHint ? (
            <span className="text-xs text-muted-foreground">{extensionsHint}</span>
          ) : null}
        </div>

        <div className="space-y-2">
          <span className="text-sm font-medium text-foreground">待上传文件 ({stagedFiles.length})</span>
          {stagedFiles.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无文件，请先添加。</p>
          ) : (
            <ul className="max-h-48 space-y-1.5 overflow-y-auto rounded border border-border bg-muted/20 p-2">
              {stagedFiles.map((f) => {
                const k = fileKey(f);
                return (
                  <li
                    key={k}
                    className="flex items-center justify-between gap-2 rounded px-2 py-1.5 text-sm"
                  >
                    <span className="min-w-0 truncate text-foreground" title={f.name}>
                      {f.name}
                      <span className="ml-2 text-muted-foreground">({formatFileSize(f.size)})</span>
                    </span>
                    <button
                      type="button"
                      className="shrink-0 rounded p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
                      disabled={busy}
                      aria-label={`移除 ${f.name}`}
                      onClick={(ev) => {
                        ev.stopPropagation();
                        onStagedFilesChange(stagedFiles.filter((x) => fileKey(x) !== k));
                      }}
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </AdminModal>
  );
}
