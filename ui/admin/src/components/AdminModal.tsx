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
  primary: 'bg-blue-600 hover:bg-blue-700 text-white',
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
      <div className={`bg-white rounded-lg shadow-xl w-full ${maxWidthClass[maxWidth]} mx-4`}>
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h3 className="text-base font-medium text-gray-900">{title}</h3>
          <button
            onClick={onClose}
            className="p-1 text-gray-400 hover:text-gray-600 rounded"
            aria-label="关闭"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4">{children}</div>

        {/* Footer */}
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
          {onConfirm && (
            <button
              onClick={onConfirm}
              disabled={loading}
              className={`px-4 py-2 text-sm rounded transition-colors disabled:opacity-50 ${confirmVariantClass[confirmVariant]}`}
            >
              {loading ? '提交中...' : confirmLabel}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
