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
  /** 禁用确定按钮（例如无待提交数据时） */
  confirmDisabled?: boolean;
  /** 弹窗最大宽度，默认 max-w-lg */
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '6xl' | '7xl';
  /**
   * 为 true 时限制弹窗高度，仅中间内容区滚动（适合长表单或大布局），
   * 遮罩使用与后台主布局相近的阶梯 padding。
   */
  scrollableBody?: boolean;
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
  xl: 'max-w-3xl',
  '2xl': 'max-w-4xl',
  '6xl': 'max-w-6xl',
  '7xl': 'max-w-7xl',
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
  confirmDisabled = false,
  maxWidth = 'lg',
  scrollableBody = false,
  children,
}: AdminModalProps) {
  const headerFooterClass =
    'flex items-center justify-between border-b border-border px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.4vw+0.65rem,1rem)]';

  const bodyClassName = `space-y-4 px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.45vw+0.65rem,1rem)]${
    scrollableBody ? ' min-h-0 flex-1 overflow-y-auto' : ''
  }`;

  return (
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center bg-black/40${
        scrollableBody ? ' p-2 sm:p-4' : ''
      }`}
    >
      <div
        className={`bg-popover text-popover-foreground rounded-lg border border-border shadow-xl w-full ${maxWidthClass[maxWidth]} admin-card ${
          scrollableBody
            ? 'flex max-h-[min(90dvh,100vh)] min-h-0 flex-col overflow-hidden'
            : 'mx-4'
        }`}
      >
        {/* Header */}
        <div className={`${headerFooterClass} shrink-0`}>
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
        <div className={bodyClassName}>{children}</div>

        {/* Footer */}
        <div
          className={`flex shrink-0 justify-end gap-3 border-t border-border px-[clamp(1rem,0.6vw+0.7rem,1.5rem)] py-[clamp(0.8rem,0.45vw+0.65rem,1rem)]`}
        >
          <button
            onClick={onClose}
            className="admin-control-md rounded border border-border bg-background px-4 text-[clamp(0.86rem,0.13vw+0.82rem,0.96rem)] hover:bg-accent transition-colors"
          >
            取消
          </button>
          {onConfirm && (
            <button
              onClick={onConfirm}
              disabled={loading || confirmDisabled}
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
