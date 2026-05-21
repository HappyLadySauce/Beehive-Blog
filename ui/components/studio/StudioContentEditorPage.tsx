"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { ChangeEvent, FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { ArrowLeft, Code2, Eye, FileInput, Heading2, Loader2, PencilLine, Save, UploadCloud } from "lucide-react";
import { useRouter } from "next/navigation";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { useAuth } from "@/components/auth/AuthProvider";
import { ToastMessage } from "@/components/toast/ToastProvider";
import { attachmentContentUrl, uploadLocalAttachmentsBatch } from "@/lib/api/attachments";
import { humanizeApiError } from "@/lib/api/client";
import { createContent, getContent, setContentCategories, setContentTags, transitionContentStatusTo, updateContent } from "@/lib/api/contents";
import { listCategories } from "@/lib/api/categories";
import { listTags } from "@/lib/api/tags";
import type { AttachmentResponse, CategoryItem, ContentAIAccess, ContentDetailResponse, ContentStatus, ContentType, ContentVisibility, TagItem } from "@/lib/api/types";
import styles from "./Studio.module.css";
import type { MarkdownScrollTarget } from "./StudioMarkdownCodeMirror";
import { StudioSelect } from "./StudioSelect";

const MarkdownCodeMirror = dynamic(
  () => import("./StudioMarkdownCodeMirror").then((module) => module.StudioMarkdownCodeMirror),
  {
    loading: () => <div className={styles.contentEditorLoading}>正在加载 Markdown 编辑器...</div>,
    ssr: false
  }
);

type EditorMode = "create" | "edit";
type EditorViewMode = "live" | "source" | "preview";
type Message = { tone: "success" | "error"; text: string } | null;

type ContentEditorForm = {
  type: string;
  title: string;
  slug: string;
  excerpt: string;
  body: string;
  coverAttachmentID: string;
  status: string;
  visibility: string;
  aiAccess: string;
  tagIDs: number[];
  categoryIDs: number[];
};

type StudioContentEditorPageProps = {
  contentId?: number;
  mode: EditorMode;
};

type OutlineItem = {
  id: string;
  level: number;
  line: number;
  text: string;
};

const emptyForm: ContentEditorForm = {
  aiAccess: "allowed",
  body: "",
  categoryIDs: [],
  coverAttachmentID: "",
  excerpt: "",
  slug: "",
  status: "draft",
  tagIDs: [],
  title: "",
  type: "article",
  visibility: "public"
};

const contentTypeOptions = [
  { value: "article", label: "文章" },
  { value: "note", label: "笔记" },
  { value: "project", label: "项目" },
  { value: "experience", label: "经历" },
  { value: "reflection", label: "复盘" },
  { value: "portfolio", label: "作品集" }
] as const;

const createStatusOptions = [
  { value: "draft", label: "草稿" },
  { value: "review", label: "审核" }
] as const;

const editStatusOptions = [
  ...createStatusOptions,
  { value: "published", label: "已发布" },
  { value: "archived", label: "已归档" }
] as const;

const visibilityOptions = [
  { value: "public", label: "公开" },
  { value: "member", label: "成员" },
  { value: "private", label: "私有" }
] as const;

const aiAccessOptions = [
  { value: "allowed", label: "AI 可访问" },
  { value: "denied", label: "AI 禁止" }
] as const;

function formFromContent(content: ContentDetailResponse): ContentEditorForm {
  return {
    aiAccess: content.ai_access,
    body: content.body ?? "",
    coverAttachmentID: content.cover_attachment_id ? String(content.cover_attachment_id) : "",
    excerpt: content.excerpt ?? "",
    slug: content.slug,
    status: content.status,
    tagIDs: content.tags?.map((tag) => tag.id) ?? [],
    categoryIDs: content.categories?.map((cat) => cat.id) ?? [],
    title: content.title,
    type: content.type,
    visibility: content.visibility
  };
}

function isMarkdownFile(file: File) {
  const name = file.name.toLowerCase();
  return name.endsWith(".md") || name.endsWith(".markdown") || file.type === "text/markdown";
}

function readFileText(file: File) {
  if (typeof file.text === "function") return file.text();
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ""));
    reader.onerror = () => reject(reader.error);
    reader.readAsText(file);
  });
}

function attachmentDisplayName(attachment: AttachmentResponse) {
  return attachment.original_name || attachment.filename || `attachment-${attachment.id}`;
}

function escapeMarkdownLabel(value: string) {
  return value.replace(/([\\[\]])/g, "\\$1");
}

function markdownReferenceForAttachment(attachment: AttachmentResponse) {
  const label = escapeMarkdownLabel(attachmentDisplayName(attachment));
  const url = attachmentContentUrl(attachment.id);
  return attachment.mime_type.startsWith("image/") ? `![${label}](${url})` : `[${label}](${url})`;
}

function insertMarkdownAt(body: string, from: number, to: number, markdown: string) {
  const start = Math.max(0, Math.min(from, body.length));
  const end = Math.max(start, Math.min(to, body.length));
  const before = body.slice(0, start);
  const after = body.slice(end);
  const prefix = before.length > 0 && !before.endsWith("\n") ? "\n\n" : "";
  const suffix = after.length > 0 && !after.startsWith("\n") ? "\n\n" : "";
  const insertion = `${prefix}${markdown}${suffix}`;
  return {
    body: `${before}${insertion}${after}`,
    cursor: before.length + insertion.length
  };
}

function extractOutline(body: string): OutlineItem[] {
  return body.split("\n").flatMap((line, index) => {
    const match = /^(#{1,4})\s+(.+?)\s*$/.exec(line);
    if (!match) return [];
    return [{
      id: `heading-${index + 1}-${match[2].toLowerCase().replace(/[^\p{L}\p{N}]+/gu, "-").replace(/^-|-$/g, "") || "section"}`,
      level: match[1].length,
      line: index + 1,
      text: match[2].replace(/[*_`[\]()]/g, "")
    }];
  });
}

export function StudioContentEditorPage({ contentId, mode }: StudioContentEditorPageProps) {
  const router = useRouter();
  const { claims } = useAuth();
  const [form, setForm] = useState<ContentEditorForm>(emptyForm);
  const [tags, setTags] = useState<TagItem[]>([]);
  const [categories, setCategories] = useState<CategoryItem[]>([]);
  const [originalStatus, setOriginalStatus] = useState("");
  const [loading, setLoading] = useState(mode === "edit");
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [editorNotice, setEditorNotice] = useState("拖拽文件或图片到编辑区可自动上传引用");
  const [selection, setSelection] = useState({ from: 0, to: 0 });
  const [viewMode, setViewMode] = useState<EditorViewMode>("live");
  const [scrollTarget, setScrollTarget] = useState<MarkdownScrollTarget | null>(null);
  const scrollIDRef = useRef(0);

  const statusOptions = mode === "create" ? createStatusOptions : editStatusOptions;
  const title = mode === "create" ? "新建内容" : form.title || "编辑内容";
  const bodyStats = useMemo(() => `${form.body.length} 字符 / ${Math.max(1, form.body.split(/\s+/).filter(Boolean).length)} 词`, [form.body]);
  const outline = useMemo(() => extractOutline(form.body), [form.body]);

  useEffect(() => {
    let active = true;
    const load = mode === "edit" && contentId
      ? Promise.all([listTags({ page: 1, page_size: 100 }), listCategories({ page: 1, page_size: 100 }), getContent(contentId)])
      : Promise.all([listTags({ page: 1, page_size: 100 }), listCategories({ page: 1, page_size: 100 })]);

    load
      .then((result) => {
        if (!active) return;
        const [tagPayload, catPayload, content] = result as [Awaited<ReturnType<typeof listTags>>, Awaited<ReturnType<typeof listCategories>>, ContentDetailResponse | undefined];
        setTags(tagPayload.items);
        setCategories(catPayload.items);
        if (content) {
          setForm(formFromContent(content));
          setOriginalStatus(content.status);
          setSelection({ from: (content.body ?? "").length, to: (content.body ?? "").length });
        }
      })
      .catch((error: unknown) => {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [contentId, mode]);

  function setField<K extends keyof ContentEditorForm>(key: K, value: ContentEditorForm[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function toggleTag(id: number, checked: boolean) {
    setForm((current) => ({
      ...current,
      tagIDs: checked ? Array.from(new Set([...current.tagIDs, id])) : current.tagIDs.filter((tagID) => tagID !== id)
    }));
  }

  function toggleCategory(id: number, checked: boolean) {
    setForm((current) => ({
      ...current,
      categoryIDs: checked ? Array.from(new Set([...current.categoryIDs, id])) : current.categoryIDs.filter((catID) => catID !== id)
    }));
  }

  async function handleEditorFiles(files: File[]) {
    if (files.length === 0 || uploading) return;
    setMessage(null);
    setUploading(true);
    try {
      const markdownFiles = files.filter(isMarkdownFile);
      const assetFiles = files.filter((file) => !isMarkdownFile(file));
      let nextBody = form.body;
      let nextSelection = selection;

      if (markdownFiles.length > 0) {
        const imported = await Promise.all(markdownFiles.map(async (file) => ({ file, text: await readFileText(file) })));
        nextBody = imported.map((item) => item.text.trimEnd()).join("\n\n");
        nextSelection = { from: nextBody.length, to: nextBody.length };
        setForm((current) => ({ ...current, body: nextBody }));
        setSelection(nextSelection);
        setEditorNotice(imported.length === 1 ? `已载入 ${imported[0].file.name}` : `已载入 ${imported.length} 个 Markdown 文件`);
      }

      if (assetFiles.length > 0) {
        if (!claims?.uid) {
          setMessage({ tone: "error", text: "当前会话缺少用户 ID，请重新登录后再上传。" });
          return;
        }
        const formData = new FormData();
        for (const file of assetFiles) {
          formData.append("files", file);
        }
        formData.set("owner_user_id", String(claims.uid));
        formData.set("purpose", "content");
        formData.set("access_scope", "public");

        const result = await uploadLocalAttachmentsBatch(formData);
        const uploadedItems = result.items.filter((item) => !item.error);
        const failedItems = result.items.filter((item) => item.error);
        if (uploadedItems.length > 0) {
          const markdown = uploadedItems.map((item) => markdownReferenceForAttachment(item.attachment)).join("\n");
          const insertAt = nextBody === form.body ? selection : { from: nextBody.length, to: nextBody.length };
          const inserted = insertMarkdownAt(nextBody, insertAt.from, insertAt.to, markdown);
          nextBody = inserted.body;
          nextSelection = { from: inserted.cursor, to: inserted.cursor };
          setForm((current) => ({ ...current, body: nextBody }));
          setSelection(nextSelection);
          setEditorNotice(`已上传并插入 ${uploadedItems.length} 个引用`);
        }
        if (failedItems.length > 0) {
          setMessage({ tone: "error", text: failedItems.map((item) => `${item.filename}: ${item.error}`).join("; ") });
        }
      }
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setUploading(false);
    }
  }

  function onMarkdownImport(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    event.currentTarget.value = "";
    void handleEditorFiles(files);
  }

  function onAssetImport(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    event.currentTarget.value = "";
    void handleEditorFiles(files);
  }

  function scrollToOutlineItem(item: OutlineItem) {
    scrollIDRef.current += 1;
    setScrollTarget({ id: scrollIDRef.current, line: item.line });
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const titleValue = form.title.trim();
    const slugValue = form.slug.trim();
    if (!form.type || !titleValue || !slugValue) {
      setMessage({ tone: "error", text: "类型、标题和 Slug 不能为空。" });
      return;
    }

    const coverAttachmentID = form.coverAttachmentID ? Number(form.coverAttachmentID) : null;
    if (coverAttachmentID !== null && (!Number.isInteger(coverAttachmentID) || coverAttachmentID <= 0)) {
      setMessage({ tone: "error", text: "封面附件 ID 必须是正整数。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      if (mode === "create") {
        const created = await createContent({
          ai_access: form.aiAccess,
          body: form.body || null,
          cover_attachment_id: coverAttachmentID,
          excerpt: form.excerpt || null,
          slug: slugValue,
          status: form.status === "review" ? "review" : "draft",
          title: titleValue,
          type: form.type,
          visibility: form.visibility
        });
        if (form.tagIDs.length > 0) {
          await setContentTags(created.id, { tag_ids: form.tagIDs });
        }
        if (form.categoryIDs.length > 0) {
          await setContentCategories(created.id, { category_ids: form.categoryIDs });
        }
        setMessage({ tone: "success", text: "内容已创建。" });
        router.replace(`/studio/content/${created.id}/edit`);
        return;
      }

      if (!contentId) {
        setMessage({ tone: "error", text: "缺少内容 ID。" });
        return;
      }
      const updated = await updateContent(contentId, {
        ai_access: form.aiAccess,
        body: form.body || null,
        cover_attachment_id: coverAttachmentID,
        excerpt: form.excerpt || null,
        slug: slugValue,
        title: titleValue,
        type: form.type,
        visibility: form.visibility
      });
      await setContentTags(contentId, { tag_ids: form.tagIDs });
      await setContentCategories(contentId, { category_ids: form.categoryIDs });
      if (form.status !== originalStatus) {
        await transitionContentStatusTo(contentId, originalStatus, form.status as ContentStatus);
      }
      setOriginalStatus(form.status);
      setForm(formFromContent({ ...updated, status: form.status as ContentStatus, tags: tags.filter((tag) => form.tagIDs.includes(tag.id)), categories: categories.filter((cat) => form.categoryIDs.includes(cat.id)) }));
      setMessage({ tone: "success", text: "内容已保存。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className={styles.contentEditorPage}>
        <div className={styles.contentEditorLoading}>
          <Loader2 aria-hidden className="spin" size={22} />
          正在加载内容...
        </div>
      </div>
    );
  }

  return (
    <form className={styles.contentEditorPage} onSubmit={onSubmit}>
      <ToastMessage message={message} />
      <header className={styles.contentEditorHeader} aria-label="编辑器顶部栏">
        <div className={styles.contentEditorTitleGroup}>
          <Link className="secondary-button icon-button" href="/studio/content" aria-label="返回内容列表" prefetch={false}>
            <ArrowLeft aria-hidden size={18} />
          </Link>
          <div>
            <h1>{title}</h1>
          </div>
        </div>
        <div className={styles.contentEditorHeaderActions}>
          <div className={styles.editorModeTabs} role="group" aria-label="编辑模式">
            <button className={viewMode === "live" ? styles.editorModeTabActive : styles.editorModeTab} type="button" onClick={() => setViewMode("live")}>
              <PencilLine aria-hidden size={15} />
              编辑预览
            </button>
            <button className={viewMode === "source" ? styles.editorModeTabActive : styles.editorModeTab} type="button" onClick={() => setViewMode("source")}>
              <Code2 aria-hidden size={15} />
              源码
            </button>
            <button className={viewMode === "preview" ? styles.editorModeTabActive : styles.editorModeTab} type="button" onClick={() => setViewMode("preview")}>
              <Eye aria-hidden size={15} />
              预览
            </button>
          </div>
          <label className="secondary-button">
            <FileInput aria-hidden size={16} />
            导入
            <input
              aria-label="导入 Markdown 文件"
              className={styles.fileInputHidden}
              type="file"
              accept=".md,.markdown,text/markdown"
              multiple
              disabled={viewMode === "preview"}
              onChange={onMarkdownImport}
            />
          </label>
          <label className={`secondary-button icon-button ${viewMode === "preview" ? styles.disabledToolButton : ""}`} aria-label="上传文件">
            <UploadCloud aria-hidden size={16} />
            <input aria-label="上传文件" className={styles.fileInputHidden} type="file" multiple disabled={viewMode === "preview"} onChange={onAssetImport} />
          </label>
          <button className="primary-button" disabled={saving || uploading} type="submit">
            {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            保存
          </button>
        </div>
      </header>

      <main className={styles.contentEditorGrid}>
        <aside className={styles.contentEditorNavigator} aria-label="文档导航">
          <div className={styles.contentEditorNavMeta}>
            <strong>{title}</strong>
            <span>{form.slug || "未设置 Slug"}</span>
            <small>{bodyStats}</small>
            <small>{editorNotice}</small>
          </div>
          <div className={styles.markdownNavHeader}>
            <Heading2 aria-hidden size={16} />
            目录
          </div>
          {outline.length > 0 ? (
            <div className={styles.contentEditorOutline}>
              {outline.map((item) => (
                <button
                  className={styles.contentEditorOutlineItem}
                  data-level={item.level}
                  key={`${item.id}-${item.line}`}
                  type="button"
                  onClick={() => scrollToOutlineItem(item)}
                >
                  {item.text}
                </button>
              ))}
            </div>
          ) : (
            <p className={styles.contentEditorOutlineEmpty}>暂无目录</p>
          )}
        </aside>

        <section className={styles.contentEditorMain} aria-label="Markdown 编辑器">
          {viewMode === "preview" ? (
            <article className={styles.contentEditorPreview} aria-label="Markdown 预览">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{form.body || "暂无内容"}</ReactMarkdown>
            </article>
          ) : (
            <MarkdownCodeMirror
              mode={viewMode}
              scrollTarget={scrollTarget}
              uploading={uploading}
              value={form.body}
              onChange={(value) => setField("body", value)}
              onFiles={handleEditorFiles}
              onSelectionChange={setSelection}
            />
          )}
        </section>

        <aside className={styles.contentEditorInspector} aria-label="内容属性">
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="内容状态" options={statusOptions} value={form.status} onChange={(value) => setField("status", value)} />
          </label>
          <label className={styles.field}>
            <span>标题</span>
            <input aria-label="标题" value={form.title} onChange={(event) => setField("title", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Slug</span>
            <input aria-label="Slug" value={form.slug} onChange={(event) => setField("slug", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>类型</span>
            <StudioSelect ariaLabel="内容类型" options={contentTypeOptions} value={form.type} onChange={(value) => setField("type", value as ContentType)} />
          </label>
          <label className={styles.field}>
            <span>可见性</span>
            <StudioSelect ariaLabel="内容可见性" options={visibilityOptions} value={form.visibility} onChange={(value) => setField("visibility", value as ContentVisibility)} />
          </label>
          <label className={styles.field}>
            <span>AI 访问</span>
            <StudioSelect ariaLabel="AI 访问" options={aiAccessOptions} value={form.aiAccess} onChange={(value) => setField("aiAccess", value as ContentAIAccess)} />
          </label>
          <label className={styles.field}>
            <span>封面附件 ID</span>
            <input inputMode="numeric" value={form.coverAttachmentID} onChange={(event) => setField("coverAttachmentID", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>摘要</span>
            <textarea className={styles.textarea} rows={5} value={form.excerpt} onChange={(event) => setField("excerpt", event.target.value)} />
          </label>
          <fieldset className={styles.editorTagFieldset}>
            <legend>标签</legend>
            {tags.length > 0 ? (
              tags.map((tag) => (
                <label key={tag.id}>
                  <span>{tag.name}</span>
                  <input checked={form.tagIDs.includes(tag.id)} type="checkbox" onChange={(event) => toggleTag(tag.id, event.target.checked)} />
                </label>
              ))
            ) : (
              <span>暂无可用标签</span>
            )}
          </fieldset>
          <fieldset className={styles.editorTagFieldset}>
            <legend>分类</legend>
            {categories.length > 0 ? (
              categories.map((cat) => (
                <label key={cat.id}>
                  <span>{cat.name}</span>
                  <input checked={form.categoryIDs.includes(cat.id)} type="checkbox" onChange={(event) => toggleCategory(cat.id, event.target.checked)} />
                </label>
              ))
            ) : (
              <span>暂无可用分类</span>
            )}
          </fieldset>
        </aside>
      </main>
    </form>
  );
}
