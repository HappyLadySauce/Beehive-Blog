import { useCallback, useEffect, useMemo, useState } from 'react';
import { Copy, Pencil, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import AdminModal from '../../components/AdminModal';
import CustomSelect from '../../components/CustomSelect';
import {
  updateAttachment,
  deleteAttachment,
  getAttachmentFamily,
  processAttachment,
  replaceAttachmentInArticles,
  type Attachment,
  type AttachmentFamilyResponse,
  type AttachmentGroupItem,
  type ProcessOutputFormat,
} from '../../api/attachment';
import { collectFamilyMembers, findMember, formatSize } from './attachmentUtils';

type ImageAttachmentDetailModalProps = {
  familyData: AttachmentFamilyResponse;
  groups: AttachmentGroupItem[];
  onClose: () => void;
  onFamilyData: (data: AttachmentFamilyResponse) => void;
  onListRefresh: () => void;
};

export default function ImageAttachmentDetailModal({
  familyData,
  groups,
  onClose,
  onFamilyData,
  onListRefresh,
}: ImageAttachmentDetailModalProps) {
  const rootId = familyData.root.id;

  const [selectedMemberId, setSelectedMemberId] = useState(rootId);

  const referencingArticlesForMember = useMemo(() => {
    const entry = familyData.memberReferences?.find((m) => m.attachmentId === selectedMemberId);
    return entry?.articles ?? [];
  }, [familyData.memberReferences, selectedMemberId]);

  useEffect(() => {
    setSelectedArticleIds([]);
  }, [selectedMemberId]);

  const [processQuality, setProcessQuality] = useState(85);
  const [processFormat, setProcessFormat] = useState<'' | ProcessOutputFormat>('');
  const [processing, setProcessing] = useState(false);

  const [replaceFromId, setReplaceFromId] = useState(rootId);
  const [replaceToId, setReplaceToId] = useState(rootId);
  const [selectedArticleIds, setSelectedArticleIds] = useState<number[]>([]);
  const [replacing, setReplacing] = useState(false);

  const [renameOpen, setRenameOpen] = useState(false);
  const [renameTargetId, setRenameTargetId] = useState<number | null>(null);
  const [renameValue, setRenameValue] = useState('');
  const [renaming, setRenaming] = useState(false);

  const [previewZoomUrl, setPreviewZoomUrl] = useState<string | null>(null);

  const [groupSelect, setGroupSelect] = useState<string>(() =>
    familyData.root.groupId != null ? String(familyData.root.groupId) : '0',
  );
  const [savingGroup, setSavingGroup] = useState(false);

  const refreshFamily = useCallback(async () => {
    const res = await getAttachmentFamily(rootId);
    if (res.code === 200 && res.data) {
      onFamilyData(res.data);
      setGroupSelect(res.data.root.groupId != null ? String(res.data.root.groupId) : '0');
      setSelectedMemberId((prev) => {
        const members = collectFamilyMembers(res.data!);
        if (members.some((m) => m.id === prev)) return prev;
        return res.data!.root.id;
      });
    }
    return res;
  }, [onFamilyData, rootId]);

  useEffect(() => {
    const m = collectFamilyMembers(familyData);
    setReplaceFromId(familyData.root.id);
    setReplaceToId(m.length > 1 ? m[1]!.id : familyData.root.id);
  }, [familyData]);

  const members = useMemo(() => collectFamilyMembers(familyData), [familyData]);
  const previewAttachment = useMemo(
    () => findMember(familyData, selectedMemberId),
    [familyData, selectedMemberId],
  );

  const formatOptions = useMemo(
    () => [
      { value: '', label: '与源图相同（默认）' },
      { value: 'jpeg', label: 'JPEG' },
      { value: 'png', label: 'PNG' },
      { value: 'gif', label: 'GIF' },
      { value: 'webp', label: 'WebP' },
      { value: 'ico', label: 'ICO' },
    ],
    [],
  );

  const memberOptions = useMemo(
    () =>
      members.map((m) => ({
        value: String(m.id),
        label: (m.originalName || m.name).slice(0, 48),
      })),
    [members],
  );

  const groupOptions = useMemo(
    () => [{ value: '0', label: '无分类' }, ...groups.map((g) => ({ value: String(g.id), label: g.name }))],
    [groups],
  );

  const handleCopyUrl = (url: string) => {
    void navigator.clipboard.writeText(url);
    toast.success('链接已复制');
  };

  const runProcess = async () => {
    if (familyData.root.type !== 'image') return;
    setProcessing(true);
    try {
      const res = await processAttachment(rootId, {
        quality: processQuality,
        ...(processFormat ? { format: processFormat } : {}),
      });
      if (res.code === 200) {
        toast.success('已生成派生附件');
        await refreshFamily();
        if (res.data) setSelectedMemberId(res.data.id);
        onListRefresh();
      } else {
        toast.error(res.message || '处理失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '处理失败');
    } finally {
      setProcessing(false);
    }
  };

  const runReplace = async () => {
    if (selectedArticleIds.length === 0) {
      toast.error('请选择至少一篇文章');
      return;
    }
    if (replaceFromId === replaceToId) {
      toast.error('源与目标不能相同');
      return;
    }
    setReplacing(true);
    try {
      const res = await replaceAttachmentInArticles({
        fromAttachmentId: replaceFromId,
        toAttachmentId: replaceToId,
        articleIds: selectedArticleIds,
      });
      if (res.code === 200 && res.data) {
        toast.success(`已更新 ${res.data.updated} 篇文章`);
        await refreshFamily();
        setSelectedArticleIds([]);
        onListRefresh();
      } else {
        toast.error(res.message || '替换失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '替换失败');
    } finally {
      setReplacing(false);
    }
  };

  const openRenameFor = (m: Attachment) => {
    setRenameTargetId(m.id);
    setRenameValue(m.originalName || m.name);
    setRenameOpen(true);
  };

  const submitRename = async () => {
    if (renameTargetId == null) return;
    const n = renameValue.trim();
    if (!n) {
      toast.error('请输入文件名');
      return;
    }
    setRenaming(true);
    try {
      const res = await updateAttachment(renameTargetId, { name: n });
      if (res.code === 200) {
        toast.success('已重命名');
        setRenameOpen(false);
        await refreshFamily();
        onListRefresh();
      } else {
        toast.error(res.message || '重命名失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '重命名失败');
    } finally {
      setRenaming(false);
    }
  };

  const saveGroup = async () => {
    const gid = parseInt(groupSelect, 10);
    if (Number.isNaN(gid)) {
      toast.error('分类无效');
      return;
    }
    setSavingGroup(true);
    try {
      const res = await updateAttachment(rootId, { groupId: gid });
      if (res.code === 200) {
        toast.success('已保存分类');
        await refreshFamily();
        onListRefresh();
      } else {
        toast.error(res.message || '保存失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '保存失败');
    } finally {
      setSavingGroup(false);
    }
  };

  const handleDeleteMember = async (m: Attachment) => {
    const isRoot = !m.parentId;
    if (
      !window.confirm(
        isRoot
          ? '删除根附件将同时删除其全部派生文件，确定？'
          : '确定删除该派生附件？',
      )
    ) {
      return;
    }
    try {
      const res = await deleteAttachment(m.id);
      if (res.code === 200) {
        toast.success('删除成功');
        if (isRoot) {
          onClose();
          onListRefresh();
          return;
        }
        const r = await refreshFamily();
        if (r?.code !== 200 || !r.data) {
          onClose();
          onListRefresh();
        }
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '删除失败');
    }
  };

  const toggleArticle = (id: number) => {
    setSelectedArticleIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  const selectAllArticles = () => {
    if (!referencingArticlesForMember.length) return;
    setSelectedArticleIds(referencingArticlesForMember.map((a) => a.articleId));
  };

  const extHint =
    processFormat === 'jpeg' ? 'jpg' : processFormat || '*';

  return (
    <>
      <AdminModal title="附件预览与处理" onClose={onClose} maxWidth="6xl">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div className="space-y-4">
            {previewAttachment?.type === 'image' && (
              <button
                type="button"
                className="flex max-h-72 w-full cursor-zoom-in justify-center overflow-hidden rounded border border-border bg-muted/30 p-2"
                onClick={() => previewAttachment && setPreviewZoomUrl(previewAttachment.url)}
                title="点击放大"
              >
                <img
                  src={previewAttachment.url}
                  alt=""
                  className="max-h-64 object-contain"
                />
              </button>
            )}
            {previewAttachment && previewAttachment.type !== 'image' && (
              <p className="text-sm text-muted-foreground">非图片附件，可复制链接。</p>
            )}
            <div className="flex flex-wrap gap-2">
              {previewAttachment && (
                <button
                  type="button"
                  className="rounded border px-3 py-1.5 text-sm"
                  onClick={() => previewAttachment && handleCopyUrl(previewAttachment.url)}
                >
                  <Copy className="mr-1 inline h-4 w-4" />
                  复制 URL
                </button>
              )}
            </div>

            <div>
              <div className="mb-2 text-sm font-medium">家族成员</div>
              <ul className="max-h-48 space-y-1 overflow-y-auto rounded border border-border p-2">
                {members.map((m) => (
                  <li
                    key={m.id}
                    className={`flex items-center gap-1 rounded px-1 py-0.5 ${
                      selectedMemberId === m.id ? 'bg-primary/15' : ''
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() => setSelectedMemberId(m.id)}
                      className="min-w-0 flex-1 rounded px-2 py-1.5 text-left text-sm hover:bg-accent/50"
                    >
                      <span className="mr-2 text-xs text-muted-foreground">
                        {!m.parentId ? '根' : m.variant || '派生'}
                      </span>
                      <span className="font-medium">{m.originalName || m.name}</span>
                      <span className="ml-2 text-xs text-muted-foreground">{formatSize(m.size)}</span>
                    </button>
                    <button
                      type="button"
                      className="shrink-0 rounded p-1 hover:bg-accent"
                      title="重命名"
                      onClick={(e) => {
                        e.stopPropagation();
                        openRenameFor(m);
                      }}
                    >
                      <Pencil className="h-4 w-4" />
                    </button>
                    <button
                      type="button"
                      className="shrink-0 rounded p-1 text-destructive hover:bg-destructive/10"
                      title="删除"
                      onClick={(e) => {
                        e.stopPropagation();
                        void handleDeleteMember(m);
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </li>
                ))}
              </ul>
            </div>

            <div className="space-y-2 border-t border-border pt-3">
              <div className="text-sm font-medium">分类</div>
              <CustomSelect
                ariaLabel="附件分类"
                value={groupSelect}
                options={groupOptions}
                onChange={setGroupSelect}
                size="md"
                className="w-full"
              />
              <button
                type="button"
                disabled={savingGroup}
                onClick={() => void saveGroup()}
                className="w-full rounded border border-border bg-background px-3 py-2 text-sm hover:bg-accent disabled:opacity-50"
              >
                {savingGroup ? '保存中…' : '保存分类'}
              </button>
            </div>

            {familyData.root.type === 'image' && (
              <div className="space-y-3 border-t border-border pt-3">
                <div>
                  <div className="mb-1 flex items-center justify-between text-sm">
                    <span>压缩质量</span>
                    <span className="text-muted-foreground">{processQuality}%</span>
                  </div>
                  <input
                    type="range"
                    min={1}
                    max={100}
                    value={processQuality}
                    onChange={(e) => setProcessQuality(parseInt(e.target.value, 10) || 85)}
                    className="w-full accent-primary"
                  />
                </div>
                <label className="block text-sm">
                  输出格式
                  <div className="mt-1">
                    <CustomSelect
                      ariaLabel="输出格式"
                      value={processFormat}
                      options={formatOptions}
                      onChange={(v) => setProcessFormat((v || '') as '' | ProcessOutputFormat)}
                      size="md"
                      className="w-full"
                    />
                  </div>
                </label>
                <p className="text-xs text-muted-foreground">
                  派生文件名将形如：根文件名-compress{processQuality}.{extHint}
                </p>
                <button
                  type="button"
                  disabled={processing}
                  onClick={() => void runProcess()}
                  className="w-full rounded bg-primary px-3 py-2 text-sm text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
                >
                  {processing ? '处理中…' : '生成派生副本'}
                </button>
              </div>
            )}
          </div>

          <div className="space-y-3 border-t border-border pt-2 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0">
            <div className="text-sm font-medium">引用文章（当前成员）</div>
            {referencingArticlesForMember.length ? (
              <>
                <div className="flex gap-2">
                  <button type="button" className="text-xs text-primary" onClick={selectAllArticles}>
                    全选
                  </button>
                  <button
                    type="button"
                    className="text-xs text-muted-foreground"
                    onClick={() => setSelectedArticleIds([])}
                  >
                    清空
                  </button>
                </div>
                <ul className="max-h-48 space-y-2 overflow-y-auto rounded border border-border p-2">
                  {referencingArticlesForMember.map((ar) => (
                    <li key={ar.articleId} className="flex items-start gap-2 text-sm">
                      <input
                        type="checkbox"
                        checked={selectedArticleIds.includes(ar.articleId)}
                        onChange={() => toggleArticle(ar.articleId)}
                        className="mt-1"
                      />
                      <span className="flex-1">{ar.title}</span>
                    </li>
                  ))}
                </ul>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">暂无关联文章（正文 URL 解析）</p>
            )}

            <div className="space-y-2 border-t border-border pt-3">
              <div className="text-sm font-medium">文中替换链接</div>
              <label className="block text-xs text-muted-foreground">
                将链接从
                <div className="mt-1">
                  <CustomSelect
                    ariaLabel="将链接从"
                    value={String(replaceFromId)}
                    options={memberOptions}
                    onChange={(v) => setReplaceFromId(parseInt(v, 10))}
                    size="sm"
                    className="w-full"
                  />
                </div>
              </label>
              <label className="block text-xs text-muted-foreground">
                替换为
                <div className="mt-1">
                  <CustomSelect
                    ariaLabel="替换为"
                    value={String(replaceToId)}
                    options={memberOptions}
                    onChange={(v) => setReplaceToId(parseInt(v, 10))}
                    size="sm"
                    className="w-full"
                  />
                </div>
              </label>
              <button
                type="button"
                disabled={replacing || selectedArticleIds.length === 0}
                onClick={() => void runReplace()}
                className="w-full rounded border border-primary bg-primary/10 px-3 py-2 text-sm text-primary hover:bg-primary/15 disabled:opacity-50"
              >
                {replacing ? '替换中…' : '替换选中文章中的链接'}
              </button>
            </div>
          </div>
        </div>
      </AdminModal>

      {renameOpen && (
        <AdminModal
          title="重命名附件"
          onClose={() => setRenameOpen(false)}
          onConfirm={() => void submitRename()}
          loading={renaming}
          maxWidth="md"
        >
          <p className="text-xs text-muted-foreground">
            将移动存储文件并更新数据库中的公开 URL；正文中若包含旧链接将批量替换为新链接。
          </p>
          <input
            type="text"
            value={renameValue}
            onChange={(e) => setRenameValue(e.target.value)}
            className="mt-2 w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm"
          />
        </AdminModal>
      )}

      {previewZoomUrl && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/75 p-4"
          role="presentation"
          onClick={() => setPreviewZoomUrl(null)}
        >
          <button
            type="button"
            className="absolute right-4 top-4 rounded bg-background/90 px-3 py-1 text-sm text-foreground shadow"
            onClick={() => setPreviewZoomUrl(null)}
          >
            关闭
          </button>
          <img
            src={previewZoomUrl}
            alt=""
            className="max-h-[90vh] max-w-[95vw] object-contain"
            onClick={(e) => e.stopPropagation()}
          />
        </div>
      )}
    </>
  );
}
