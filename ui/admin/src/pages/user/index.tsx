import { useEffect, useMemo, useState } from 'react';
import { Eye, Lock, Search, Trash2, UserCog } from 'lucide-react';
import { toast } from 'sonner';
import {
  AdminUserItem,
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  AdminUserListQuery,
  UserRole,
  UserStatus,
  createAdminUser,
  deleteAdminUser,
  getAdminUserDetail,
  getAdminUsers,
  resetAdminUserPassword,
  updateAdminUser,
  updateAdminUserStatus,
} from '../../api/user';
import AdminModal from '../../components/AdminModal';
import {
  TextField,
  PasswordField,
  TextareaField,
  SelectField,
} from '../../components/FormField';
import Pagination from '../../components/Pagination';

// ─── 常量映射 ────────────────────────────────────────────────────────────────

type ModalType = 'detail' | 'form' | 'password' | null;
type FormMode = 'create' | 'edit';

const roleLabelMap: Record<UserRole, string> = {
  guest: '访客',
  user: '用户',
  admin: '管理员',
};

const statusLabelMap: Record<UserStatus, string> = {
  active: '正常',
  inactive: '未激活',
  disabled: '禁用',
  deleted: '已删除',
};

const statusColorMap: Record<UserStatus, string> = {
  active: 'bg-green-100 text-green-800',
  inactive: 'bg-yellow-100 text-yellow-800',
  disabled: 'bg-red-100 text-red-800',
  deleted: 'bg-gray-100 text-gray-800',
};

const roleOptions = [
  { value: 'guest', label: '访客' },
  { value: 'user', label: '用户' },
  { value: 'admin', label: '管理员' },
];

const statusOptions = [
  { value: 'active', label: '正常' },
  { value: 'inactive', label: '未激活' },
  { value: 'disabled', label: '禁用' },
];

// ─── 表单数据结构 ────────────────────────────────────────────────────────────

interface UserFormData {
  /** 仅创建模式使用 */
  username: string;
  /** 仅创建模式使用 */
  password: string;
  nickname: string;
  email: string;
  avatar: string;
  role: UserRole;
  /** 仅创建模式使用 */
  status: Exclude<UserStatus, 'deleted'>;
}

const defaultFormData: UserFormData = {
  username: '',
  password: '',
  nickname: '',
  email: '',
  avatar: '',
  role: 'user',
  status: 'active',
};

// ─── UserFormModal ────────────────────────────────────────────────────────────

/**
 * 新建/编辑用户的统一表单弹窗。
 * - create 模式：额外渲染 username / password / status 字段
 * - edit 模式：只渲染 nickname / email / avatar / role 字段
 */
interface UserFormModalProps {
  mode: FormMode;
  formData: UserFormData;
  onChange: (patch: Partial<UserFormData>) => void;
  onClose: () => void;
  onSubmit: () => void;
  loading: boolean;
}

function UserFormModal({
  mode,
  formData,
  onChange,
  onClose,
  onSubmit,
  loading,
}: UserFormModalProps) {
  const isCreate = mode === 'create';

  return (
    <AdminModal
      title={isCreate ? '新建用户' : '编辑用户'}
      onClose={onClose}
      onConfirm={onSubmit}
      confirmLabel={isCreate ? '创建用户' : '保存'}
      loading={loading}
    >
      <div className="space-y-3">
        {isCreate && (
          <>
            <TextField
              label="用户名"
              value={formData.username}
              onChange={(v) => onChange({ username: v })}
              placeholder="3-20 位字母或数字"
              required
            />
            <PasswordField
              label="初始密码"
              value={formData.password}
              onChange={(v) => onChange({ password: v })}
              placeholder="6-20 位"
              required
            />
          </>
        )}
        <TextField
          label="昵称"
          value={formData.nickname}
          onChange={(v) => onChange({ nickname: v })}
        />
        <TextField
          label="邮箱"
          type="email"
          value={formData.email}
          onChange={(v) => onChange({ email: v })}
          required
        />
        <TextField
          label="头像 URL"
          type="url"
          value={formData.avatar}
          onChange={(v) => onChange({ avatar: v })}
          placeholder="https://..."
        />
        {isCreate ? (
          <div className="grid grid-cols-2 gap-3">
            <SelectField
              label="角色"
              value={formData.role}
              onChange={(v) => onChange({ role: v as UserRole })}
              options={roleOptions}
            />
            <SelectField
              label="状态"
              value={formData.status}
              onChange={(v) => onChange({ status: v as Exclude<UserStatus, 'deleted'> })}
              options={statusOptions}
            />
          </div>
        ) : (
          <SelectField
            label="角色"
            value={formData.role}
            onChange={(v) => onChange({ role: v as UserRole })}
            options={roleOptions}
          />
        )}
      </div>
    </AdminModal>
  );
}

// ─── 主页面 ──────────────────────────────────────────────────────────────────

export default function Users() {
  const [items, setItems] = useState<AdminUserItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const [keyword, setKeyword] = useState('');
  const [role, setRole] = useState<string>('');
  const [status, setStatus] = useState<string>('');

  const [currentUser, setCurrentUser] = useState<AdminUserItem | null>(null);
  const [modalType, setModalType] = useState<ModalType>(null);
  const [formMode, setFormMode] = useState<FormMode>('create');
  const [submitting, setSubmitting] = useState(false);

  const [formData, setFormData] = useState<UserFormData>(defaultFormData);
  const [newPassword, setNewPassword] = useState('');

  const query = useMemo<AdminUserListQuery>(
    () => ({
      page,
      pageSize,
      keyword: keyword || undefined,
      role: (role || undefined) as UserRole | undefined,
      status: (status || undefined) as UserStatus | undefined,
    }),
    [page, keyword, role, status],
  );

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await getAdminUsers(query);
      if (res.code === 200) {
        setItems(res.data.items || []);
        setTotal(res.data.total || 0);
      } else {
        toast.error(res.message || '获取用户列表失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [query]);

  const patchForm = (patch: Partial<UserFormData>) =>
    setFormData((f) => ({ ...f, ...patch }));

  const openDetail = async (id: number) => {
    try {
      const res = await getAdminUserDetail(id);
      if (res.code !== 200) {
        toast.error(res.message || '获取用户详情失败');
        return;
      }
      setCurrentUser(res.data);
      setModalType('detail');
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求用户详情失败');
    }
  };

  const openCreate = () => {
    setCurrentUser(null);
    setFormData(defaultFormData);
    setFormMode('create');
    setModalType('form');
  };

  const openEdit = (user: AdminUserItem) => {
    setCurrentUser(user);
    setFormData({
      username: user.username,
      password: '',
      nickname: user.nickname || '',
      email: user.email || '',
      avatar: user.avatar || '',
      role: user.role,
      status: (user.status === 'deleted' ? 'disabled' : user.status) as Exclude<UserStatus, 'deleted'>,
    });
    setFormMode('edit');
    setModalType('form');
  };

  const openPassword = (user: AdminUserItem) => {
    setCurrentUser(user);
    setNewPassword('');
    setModalType('password');
  };

  const closeModal = () => {
    setModalType(null);
    setCurrentUser(null);
    setNewPassword('');
  };

  const handleFormSubmit = async () => {
    if (formMode === 'create') {
      if (!formData.username.trim()) { toast.error('用户名不能为空'); return; }
      if (!formData.email.trim()) { toast.error('邮箱不能为空'); return; }
      if (formData.password.length < 6 || formData.password.length > 20) {
        toast.error('密码长度需在 6-20 位');
        return;
      }
      setSubmitting(true);
      try {
        const payload: AdminCreateUserRequest = {
          username: formData.username.trim(),
          password: formData.password,
          nickname: formData.nickname.trim() || undefined,
          email: formData.email.trim(),
          avatar: formData.avatar.trim() || undefined,
          role: formData.role,
          status: formData.status,
        };
        const res = await createAdminUser(payload);
        if (res.code === 200) {
          toast.success('用户创建成功');
          closeModal();
          setPage(1);
          fetchUsers();
        } else {
          toast.error(res.message || '创建用户失败');
        }
      } catch (error: any) {
        toast.error(error.response?.data?.message || '创建用户请求失败');
      } finally {
        setSubmitting(false);
      }
    } else {
      if (!currentUser) return;
      if (!formData.nickname.trim()) { toast.error('昵称不能为空'); return; }
      if (!formData.email.trim()) { toast.error('邮箱不能为空'); return; }
      setSubmitting(true);
      try {
        const payload: AdminUpdateUserRequest = {
          nickname: formData.nickname.trim(),
          email: formData.email.trim(),
          avatar: formData.avatar.trim(),
          role: formData.role,
        };
        const res = await updateAdminUser(currentUser.id, payload);
        if (res.code === 200) {
          toast.success('用户信息更新成功');
          closeModal();
          fetchUsers();
        } else {
          toast.error(res.message || '更新失败');
        }
      } catch (error: any) {
        toast.error(error.response?.data?.message || '更新请求失败');
      } finally {
        setSubmitting(false);
      }
    }
  };

  const handleStatusChange = async (
    user: AdminUserItem,
    nextStatus: Exclude<UserStatus, 'deleted'>,
  ) => {
    if (
      !window.confirm(
        `确认将用户 ${user.username} 状态设置为 ${statusLabelMap[nextStatus]} 吗？`,
      )
    )
      return;
    try {
      const res = await updateAdminUserStatus(user.id, { status: nextStatus });
      if (res.code === 200) {
        toast.success('状态更新成功');
        fetchUsers();
      } else {
        toast.error(res.message || '状态更新失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '状态更新请求失败');
    }
  };

  const handleResetPassword = async () => {
    if (!currentUser) return;
    if (newPassword.length < 6 || newPassword.length > 20) {
      toast.error('新密码长度需在 6-20 位');
      return;
    }
    setSubmitting(true);
    try {
      const res = await resetAdminUserPassword(currentUser.id, { newPassword });
      if (res.code === 200) {
        toast.success('密码重置成功');
        closeModal();
      } else {
        toast.error(res.message || '密码重置失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '密码重置请求失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (user: AdminUserItem) => {
    if (!window.confirm(`确认软删除用户 ${user.username} 吗？该操作不可直接恢复。`)) return;
    try {
      const res = await deleteAdminUser(user.id);
      if (res.code === 200) {
        toast.success('用户已删除');
        fetchUsers();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '删除请求失败');
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium text-gray-900">用户管理</h2>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          新建用户
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded">
        {/* 搜索/筛选栏 */}
        <div className="p-4 border-b border-gray-200 flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-52">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              value={keyword}
              onChange={(e) => {
                setKeyword(e.target.value);
                setPage(1);
              }}
              placeholder="搜索用户名/昵称/邮箱"
              className="w-full pl-9 pr-4 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <select
            value={role}
            onChange={(e) => {
              setRole(e.target.value);
              setPage(1);
            }}
            className="px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">全部角色</option>
            {roleOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            className="px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">全部状态</option>
            {statusOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>

        {/* 表格 */}
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-sm font-medium text-gray-600">ID</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">用户名</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">昵称</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">邮箱</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">角色</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">状态</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">最后登录</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-gray-500">
                    加载中...
                  </td>
                </tr>
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-gray-500">
                    暂无用户
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr key={item.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-600">{item.id}</td>
                    <td className="px-4 py-3 text-sm text-gray-900">{item.username}</td>
                    <td className="px-4 py-3 text-sm text-gray-900">{item.nickname || '-'}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{item.email}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {roleLabelMap[item.role]}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span
                        className={`px-2 py-1 rounded text-xs ${statusColorMap[item.status] || 'bg-gray-100 text-gray-800'}`}
                      >
                        {statusLabelMap[item.status] || item.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {item.lastLoginAt
                        ? new Date(item.lastLoginAt).toLocaleString()
                        : '-'}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          className="p-1.5 text-gray-600 hover:bg-gray-100 rounded"
                          title="详情"
                          onClick={() => openDetail(item.id)}
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        <button
                          className="p-1.5 text-blue-600 hover:bg-blue-50 rounded"
                          title="编辑"
                          onClick={() => openEdit(item)}
                        >
                          <UserCog className="w-4 h-4" />
                        </button>
                        <button
                          className="p-1.5 text-orange-600 hover:bg-orange-50 rounded"
                          title="重置密码"
                          onClick={() => openPassword(item)}
                        >
                          <Lock className="w-4 h-4" />
                        </button>
                        <button
                          className="p-1.5 text-red-600 hover:bg-red-50 rounded"
                          title="删除"
                          onClick={() => handleDelete(item)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                      <div className="mt-2">
                        <select
                          value={item.status}
                          onChange={(e) =>
                            handleStatusChange(
                              item,
                              e.target.value as Exclude<UserStatus, 'deleted'>,
                            )
                          }
                          className="px-2 py-1 text-xs border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        >
                          {statusOptions.map((o) => (
                            <option key={o.value} value={o.value}>
                              {o.label}
                            </option>
                          ))}
                        </select>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <Pagination
          total={total}
          page={page}
          pageSize={pageSize}
          onPageChange={setPage}
          unit="位用户"
        />
      </div>

      {/* 新建/编辑用户弹窗 */}
      {modalType === 'form' && (
        <UserFormModal
          mode={formMode}
          formData={formData}
          onChange={patchForm}
          onClose={closeModal}
          onSubmit={handleFormSubmit}
          loading={submitting}
        />
      )}

      {/* 详情弹窗 */}
      {modalType === 'detail' && currentUser && (
        <AdminModal title="用户详情" onClose={closeModal} maxWidth="md">
          <div className="space-y-2 text-sm">
            <p>
              <span className="text-gray-500">ID：</span>
              {currentUser.id}
            </p>
            <p>
              <span className="text-gray-500">用户名：</span>
              {currentUser.username}
            </p>
            <p>
              <span className="text-gray-500">昵称：</span>
              {currentUser.nickname || '-'}
            </p>
            <p>
              <span className="text-gray-500">邮箱：</span>
              {currentUser.email}
            </p>
            <p>
              <span className="text-gray-500">角色：</span>
              {roleLabelMap[currentUser.role]}
            </p>
            <p>
              <span className="text-gray-500">状态：</span>
              {statusLabelMap[currentUser.status]}
            </p>
            <p>
              <span className="text-gray-500">创建时间：</span>
              {new Date(currentUser.createdAt).toLocaleString()}
            </p>
            <p>
              <span className="text-gray-500">最后登录：</span>
              {currentUser.lastLoginAt
                ? new Date(currentUser.lastLoginAt).toLocaleString()
                : '-'}
            </p>
          </div>
        </AdminModal>
      )}

      {/* 重置密码弹窗 */}
      {modalType === 'password' && currentUser && (
        <AdminModal
          title="重置密码"
          onClose={closeModal}
          onConfirm={handleResetPassword}
          confirmLabel="重置密码"
          confirmVariant="warning"
          loading={submitting}
          maxWidth="md"
        >
          <div className="space-y-3">
            <p className="text-sm text-gray-600">
              为用户 <span className="font-medium">{currentUser.username}</span> 设置新密码：
            </p>
            <PasswordField
              label="新密码"
              value={newPassword}
              onChange={setNewPassword}
              placeholder="6-20 位新密码"
              required
            />
          </div>
        </AdminModal>
      )}
    </div>
  );
}
