import AdminModal from './AdminModal';

interface ConfirmModalProps {
  open: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmLabel?: string;
  confirmVariant?: 'danger' | 'warning' | 'primary';
  loading?: boolean;
}

/**
 * 确认对话框，替代 window.confirm，样式与 AdminModal 一致。
 */
export default function ConfirmModal({
  open,
  title,
  message,
  onConfirm,
  onCancel,
  confirmLabel = '确认',
  confirmVariant = 'primary',
  loading = false,
}: ConfirmModalProps) {
  if (!open) return null;

  return (
    <AdminModal
      title={title}
      onClose={onCancel}
      onConfirm={onConfirm}
      confirmLabel={confirmLabel}
      confirmVariant={confirmVariant}
      loading={loading}
      maxWidth="sm"
    >
      <p className="text-sm text-gray-600">{message}</p>
    </AdminModal>
  );
}
