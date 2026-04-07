import { ReactNode } from 'react';
import { X } from 'lucide-react';

type ConfirmVariant = 'primary' | 'danger' | 'warning';

interface AdminModalProps {
  title: string;
  onClose: () => void;
  /** 点击确认按钮的回调；不传则不显示确认按钮 */
  onConfirm?: () => void;
  confirmLabel?: string;
  confirmVariant?: ConfirmVariant;
  loading?: boolean;
  /** 弹窗最大宽度，默认 max-w-lg */
  maxWidth?: 'sm' | 'md' | 'lg';
  children: ReactNode;
}

const confirmVariantClass: Record<ConfirmVariant, string> = {
  primary: 'bg-primary hover:bg-primary/90 text-primary-foreground',
  danger: 'bg-red-600 hover:bg-red-700 text-white',
  warning: 'bg-orange-600 hover:bg-orange-700 text-white',
};

const maxWidthClass: Record<string, string> = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
};

/**
 * 通用管理后台弹窗组件。
 *
 * @example
 * <AdminModal title="新建分类" onClose={closeModal} onConfirm={handleSubmit} loading={submitting}>
 *   <TextField label="名称" value={form.name} onChange={(v) => setForm(f => ({ ...f, name: v }))} required />
 * </AdminModal>
 */
export default function AdminModal({
  title,
  onClose,
  onConfirm,
  confirmLabel = '确认',
  confirmVariant = 'primary',
  loading = false,
  maxWidth = 'lg',
  children,
}: AdminModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className={`bg-popover text-popover-foreground rounded-lg border border-border shadow-xl w-full ${maxWidthClass[maxWidth]} mx-4 admin-card`}>
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.4vw+0.65rem,1rem)]">
          <h3 className="text-[clamp(0.98rem,0.2vw+0.9rem,1.12rem)] font-medium">{title}</h3>
          <button
            onClick={onClose}
            className="admin-control-sm rounded p-1 text-muted-foreground hover:text-foreground"
            aria-label="关闭"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="space-y-4 px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.45vw+0.65rem,1rem)]">{children}</div>

        {/* Footer */}
        <div className="flex justify-end gap-3 border-t border-border px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.45vw+0.65rem,1rem)]">
          <button
            onClick={onClose}
            className="admin-control-md rounded border border-border bg-background px-4 text-[clamp(0.86rem,0.13vw+0.82rem,0.96rem)] hover:bg-accent transition-colors"
          >
            取消
          </button>
          {onConfirm && (
            <button
              onClick={onConfirm}
              disabled={loading}
              className={`admin-control-md rounded px-4 text-[clamp(0.86rem,0.13vw+0.82rem,0.96rem)] transition-colors disabled:opacity-50 ${confirmVariantClass[confirmVariant]}`}
            >
              {loading ? '提交中...' : confirmLabel}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
