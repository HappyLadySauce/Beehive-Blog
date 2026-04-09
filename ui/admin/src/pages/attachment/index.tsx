import { useState, useEffect, useCallback } from 'react';
import {
  getAttachments,
  deleteAttachment,
  uploadAttachment,
  getAttachmentFamily,
  listAttachmentGroups,
  createAttachmentGroup,
  Attachment,
  AttachmentFamilyResponse,
  AttachmentGroupItem,
} from '../../api/attachment';
import { toast } from 'sonner';
import {
  Image as ImageIcon,
  Trash2,
  Copy,
  Upload,
  FileImage,
  File,
} from 'lucide-react';
import AdminModal from '../../components/AdminModal';
import StagedFileUploadModal, { mergeStagedFilesDedupe } from '../../components/StagedFileUploadModal';
import ImageAttachmentDetailModal from './ImageAttachmentDetailModal';
import { formatSize } from './attachmentUtils';
import { ADMIN_TABLE_CHECKBOX_CLASS } from '../../app/constants/adminTable';

export default function Attachments() {
  const [items, setItems] = useState<Attachment[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [keywordDraft, setKeywordDraft] = useState('');
  const [keyword, setKeyword] = useState('');
  const [groupId, setGroupId] = useState<number | null>(null);
  const [groups, setGroups] = useState<AttachmentGroupItem[]>([]);

  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [uploadStagedFiles, setUploadStagedFiles] = useState<File[]>([]);

  const [detailOpen, setDetailOpen] = useState(false);
  const [familyData, setFamilyData] = useState<AttachmentFamilyResponse | null>(null);

  const [newGroupOpen, setNewGroupOpen] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [creatingGroup, setCreatingGroup] = useState(false);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [batchBusy, setBatchBusy] = useState(false);

  const fetchGroups = useCallback(async () => {
    try {
      const res = await listAttachmentGroups();
      if (res.code === 200) {
        setGroups(res.data || []);
      }
    } catch {
      /* ignore */
    }
  }, []);

  const fetchAttachments = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getAttachments({
        page,
        pageSize: 20,
        keyword: keyword.trim() || undefined,
        groupId: groupId ?? undefined,
        rootsOnly: true,
      });
      if (res.code === 200 && res.data) {
        setItems(res.data.items || []);
        setTotal(res.data.total);
      } else {
        toast.error(res.message || '获取附件失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '请求附件失败');
    } finally {
      setLoading(false);
    }
  }, [page, keyword, groupId]);

  useEffect(() => {
    void fetchGroups();
  }, [fetchGroups]);

  useEffect(() => {
    void fetchAttachments();
  }, [fetchAttachments]);

  useEffect(() => {
    setSelectedIds([]);
  }, [page, keyword, groupId]);

  const toggleSelectAll = () => {
    if (selectedIds.length === items.length && items.length > 0) {
      setSelectedIds([]);
    } else {
      setSelectedIds(items.map((a) => a.id));
    }
  };

  const handleBatchDelete = () => {
    if (!selectedIds.length) {
      toast.warning('请先选择附件');
      return;
    }
    if (!window.confirm(`确定删除选中的 ${selectedIds.length} 个根附件吗？将同时删除其派生文件。`)) return;
    void (async () => {
      const ids = [...selectedIds];
      setBatchBusy(true);
      let ok = 0;
      let lastErr = '';
      for (const id of ids) {
        try {
          const res = await deleteAttachment(id);
          if (res.code === 200) ok += 1;
          else lastErr = res.message || '删除失败';
        } catch (error: unknown) {
          lastErr =
            error && typeof error === 'object' && 'response' in error
              ? (error as { response?: { data?: { message?: string } } }).response?.data?.message ||
                '请求失败'
              : '请求失败';
        }
      }
      setBatchBusy(false);
      setSelectedIds([]);
      void fetchAttachments();
      if (ok === ids.length) toast.success(`已删除 ${ok} 个附件`);
      else if (ok > 0) toast.warning(`成功删除 ${ok} 个${lastErr ? `，部分失败：${lastErr}` : ''}`);
      else toast.error(lastErr || '批量删除失败');
    })();
  };

  const handleDelete = async (id: number, e?: React.MouseEvent) => {
    e?.stopPropagation();
    if (!window.confirm('删除根附件将同时删除其全部派生文件，确定？')) return;
    try {
      const res = await deleteAttachment(id);
      if (res.code === 200) {
        toast.success('删除成功');
        void fetchAttachments();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '删除请求失败');
    }
  };

  const handleCopyUrl = (url: string, e?: React.MouseEvent) => {
    e?.stopPropagation();
    void navigator.clipboard.writeText(url);
    toast.success('链接已复制');
  };

  const uploadGroupLabel =
    groupId == null ? '未指定（全部视图下上传）' : groups.find((g) => g.id === groupId)?.name ?? `分类 #${groupId}`;

  const confirmAttachmentUpload = async () => {
    if (uploadStagedFiles.length === 0) {
      toast.warning('请先添加要上传的文件');
      return;
    }
    setUploading(true);
    try {
      for (let i = 0; i < uploadStagedFiles.length; i++) {
        const file = uploadStagedFiles[i];
        const res = await uploadAttachment(file, groupId ?? undefined);
        if (res.code === 200) {
          toast.success(`已上传 ${file.name}`);
        } else {
          toast.error(res.message || `${file.name} 失败`);
        }
      }
      setUploadStagedFiles([]);
      setUploadModalOpen(false);
      void fetchAttachments();
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '上传失败');
    } finally {
      setUploading(false);
    }
  };

  const openDetail = async (a: Attachment) => {
    try {
      const res = await getAttachmentFamily(a.id);
      if (res.code !== 200 || !res.data) {
        toast.error(res.message || '加载失败');
        return;
      }
      setFamilyData(res.data);
      setDetailOpen(true);
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '请求失败');
    }
  };

  const closeDetail = () => {
    setDetailOpen(false);
    setFamilyData(null);
  };

  const submitNewGroup = async () => {
    const n = newGroupName.trim();
    if (!n) {
      toast.error('请输入分类名称');
      return;
    }
    setCreatingGroup(true);
    try {
      const res = await createAttachmentGroup({ name: n, sortOrder: 0 });
      if (res.code === 200) {
        toast.success('已创建分类');
        setNewGroupOpen(false);
        setNewGroupName('');
        void fetchGroups();
      } else {
        toast.error(res.message || '创建失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '创建失败');
    } finally {
      setCreatingGroup(false);
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / 20));

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <ImageIcon className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
          <h2 className="text-xl font-semibold tracking-tight text-foreground">附件管理</h2>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="text"
            placeholder="搜索文件名..."
            value={keywordDraft}
            onChange={(e) => setKeywordDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                setKeyword(keywordDraft.trim());
                setPage(1);
              }
            }}
            className="rounded-md border border-border bg-input-background px-3 py-1.5 text-sm text-foreground"
          />
          <button
            type="button"
            onClick={() => {
              setKeyword(keywordDraft.trim());
              setPage(1);
            }}
            className="rounded border border-border bg-background px-3 py-1.5 text-sm hover:bg-accent"
          >
            搜索
          </button>
          <button
            type="button"
            onClick={() => {
              setUploadStagedFiles([]);
              setUploadModalOpen(true);
            }}
            disabled={uploading}
            className="inline-flex items-center gap-1.5 rounded bg-primary px-3 py-1.5 text-sm text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Upload className="h-4 w-4" />
            {uploading ? '上传中...' : '上传附件'}
          </button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <span className="text-base text-muted-foreground">分类：</span>
        <button
          type="button"
          onClick={() => {
            setGroupId(null);
            setPage(1);
          }}
          className={`rounded-full border px-3 py-1.5 text-sm ${groupId === null ? 'border-primary bg-primary/10' : 'border-border'}`}
        >
          全部
        </button>
        {groups.map((g) => (
          <button
            key={g.id}
            type="button"
            onClick={() => {
              setGroupId(g.id);
              setPage(1);
            }}
            className={`rounded-full border px-3 py-1.5 text-sm ${groupId === g.id ? 'border-primary bg-primary/10' : 'border-border'}`}
          >
            {g.name}
          </button>
        ))}
        <button
          type="button"
          onClick={() => setNewGroupOpen(true)}
          className="rounded-full border border-dashed border-border px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent"
        >
          + 新建分类
        </button>
      </div>

      {selectedIds.length > 0 && (
        <div className="mb-2 flex flex-wrap items-center gap-2 rounded border border-border bg-muted/30 px-4 py-2">
          <button
            type="button"
            disabled={batchBusy}
            className="rounded bg-red-600 px-3 py-1.5 text-sm text-white transition-colors hover:bg-red-700 disabled:opacity-50"
            onClick={handleBatchDelete}
          >
            批量删除
          </button>
        </div>
      )}

      <div className="overflow-x-auto rounded border border-border bg-card">
        {loading ? (
          <div className="py-12 text-center text-muted-foreground">加载中...</div>
        ) : items.length === 0 ? (
          <div className="py-12 text-center text-muted-foreground">暂无附件</div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="border-b border-border bg-muted/40">
              <tr>
                <th className="w-10 px-3 py-2">
                  <input
                    type="checkbox"
                    checked={items.length > 0 && selectedIds.length === items.length}
                    onChange={toggleSelectAll}
                    onClick={(e) => e.stopPropagation()}
                    className={ADMIN_TABLE_CHECKBOX_CLASS}
                    aria-label="全选"
                  />
                </th>
                <th className="px-3 py-2 font-medium">类型</th>
                <th className="px-3 py-2 font-medium">文件名</th>
                <th className="px-3 py-2 font-medium">MIME</th>
                <th className="px-3 py-2 font-medium">大小</th>
                <th className="px-3 py-2 font-medium">分类</th>
                <th className="px-3 py-2 font-medium">引用文章数</th>
                <th className="px-3 py-2 font-medium">上传时间</th>
                <th className="px-3 py-2 font-medium text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              {items.map((a) => (
                <tr
                  key={a.id}
                  className="cursor-pointer border-b border-border hover:bg-accent/40"
                  onClick={() => void openDetail(a)}
                >
                  <td
                    className="px-3 py-2 align-middle"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <input
                      type="checkbox"
                      checked={selectedIds.includes(a.id)}
                      onChange={(e) => {
                        e.stopPropagation();
                        setSelectedIds((prev) =>
                          prev.includes(a.id) ? prev.filter((x) => x !== a.id) : [...prev, a.id],
                        );
                      }}
                      onClick={(e) => e.stopPropagation()}
                      className={ADMIN_TABLE_CHECKBOX_CLASS}
                      aria-label={`选择 ${a.originalName || a.name}`}
                    />
                  </td>
                  <td className="px-3 py-2">
                    {a.type === 'image' ? (
                      <FileImage className="h-5 w-5 text-muted-foreground" aria-hidden />
                    ) : (
                      <File className="h-5 w-5 text-muted-foreground" aria-hidden />
                    )}
                  </td>
                  <td className="max-w-[200px] truncate px-3 py-2 font-medium" title={a.originalName}>
                    {a.originalName || a.name}
                  </td>
                  <td className="px-3 py-2 text-muted-foreground">{a.mimeType}</td>
                  <td className="px-3 py-2">{formatSize(a.size)}</td>
                  <td className="px-3 py-2 text-muted-foreground">{a.groupName || '—'}</td>
                  <td className="px-3 py-2">{a.refArticleCount ?? 0}</td>
                  <td className="px-3 py-2 text-muted-foreground">
                    {a.createdAt ? new Date(a.createdAt).toLocaleString() : '—'}
                  </td>
                  <td className="px-3 py-2 text-right">
                    <button
                      type="button"
                      className="mr-2 inline-flex rounded p-1 hover:bg-accent"
                      title="复制 URL"
                      onClick={(e) => handleCopyUrl(a.url, e)}
                    >
                      <Copy className="h-4 w-4" />
                    </button>
                    <button
                      type="button"
                      className="inline-flex rounded p-1 text-destructive hover:bg-destructive/10"
                      title="删除"
                      onClick={(e) => void handleDelete(a.id, e)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            type="button"
            disabled={page <= 1}
            className="rounded border px-2 py-1 text-sm disabled:opacity-50"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            上一页
          </button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            className="rounded border px-2 py-1 text-sm disabled:opacity-50"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          >
            下一页
          </button>
        </div>
      )}

      {detailOpen && familyData && (
        <ImageAttachmentDetailModal
          familyData={familyData}
          groups={groups}
          onClose={closeDetail}
          onFamilyData={setFamilyData}
          onListRefresh={() => void fetchAttachments()}
        />
      )}

      {newGroupOpen && (
        <AdminModal
          title="新建附件分类"
          onClose={() => setNewGroupOpen(false)}
          onConfirm={() => void submitNewGroup()}
          loading={creatingGroup}
          maxWidth="sm"
        >
          <input
            type="text"
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
            placeholder="分类名称"
            className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm"
          />
        </AdminModal>
      )}

      <StagedFileUploadModal
        open={uploadModalOpen}
        title="上传附件"
        description={
          <>
            将文件加入列表后点击「确定」逐个上传。当前附件分类：
            <span className="font-medium text-foreground"> {uploadGroupLabel}</span>
            {groupId != null
              ? '（上传时将归入该分类）'
              : '（未指定分类标签，与在「全部」下列表一致）'}
          </>
        }
        extensionsHint="支持多文件，类型不限"
        loading={uploading}
        disabled={uploading}
        stagedFiles={uploadStagedFiles}
        onStagedFilesChange={setUploadStagedFiles}
        onPickFiles={(picked) => setUploadStagedFiles((prev) => mergeStagedFilesDedupe(picked, prev))}
        onClose={() => {
          setUploadModalOpen(false);
          setUploadStagedFiles([]);
        }}
        onConfirm={() => void confirmAttachmentUpload()}
        confirmLabel="确定"
        confirmDisabled={uploadStagedFiles.length === 0}
      />
    </div>
  );
}
