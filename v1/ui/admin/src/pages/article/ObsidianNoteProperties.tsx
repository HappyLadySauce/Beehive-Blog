import { useState, useId, useCallback, useMemo, useEffect } from 'react';
import {
  AlignLeft,
  ChevronDown,
  ChevronRight,
  Hash,
  List,
  Tag as TagIcon,
  X,
} from 'lucide-react';
import { useArticleNotePropsStore } from './articleNotePropsStore';
import CustomSelect from '../../components/CustomSelect';
import { Popover, PopoverContent, PopoverTrigger } from '../../app/components/ui/popover';
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '../../app/components/ui/command';
import { resolveOrCreateCategoryByName, resolveOrCreateTagByName } from '../../lib/ensureTaxonomy';
import { cn } from '../../app/components/ui/utils';

function PropRow({
  icon: Icon,
  label,
  children,
  className,
  readOnly,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  children: React.ReactNode;
  className?: string;
  /** 预览区：无悬停高亮，避免像可编辑区域 */
  readOnly?: boolean;
}) {
  return (
    <div
      className={cn(
        'beehive-prop-row flex min-h-9 items-start gap-2 rounded-md border border-transparent px-1 py-1',
        !readOnly && 'hover:border-border/80 hover:bg-muted/40',
        className,
      )}
    >
      <Icon className="mt-1.5 size-3.5 shrink-0 text-muted-foreground" aria-hidden />
      <span className="w-[5.5rem] shrink-0 pt-1 text-xs text-muted-foreground">{label}</span>
      <div className="min-w-0 flex-1">{children}</div>
    </div>
  );
}

function ReadonlyCell({ children }: { children: React.ReactNode }) {
  return (
    <div className="break-words px-1 py-0.5 text-xs leading-snug text-foreground">{children}</div>
  );
}

export default function ObsidianNoteProperties({ readOnly = false }: { readOnly?: boolean }) {
  const idBase = useId();
  const [open, setOpen] = useState(false);
  const [catSearch, setCatSearch] = useState('');
  const [tagInput, setTagInput] = useState('');
  const [expanded, setExpanded] = useState(true);

  useEffect(() => {
    if (!open) {
      setCatSearch('');
    }
  }, [open]);

  const title = useArticleNotePropsStore((s) => s.title);
  const setTitle = useArticleNotePropsStore((s) => s.setTitle);
  const slug = useArticleNotePropsStore((s) => s.slug);
  const setSlug = useArticleNotePropsStore((s) => s.setSlug);
  const summary = useArticleNotePropsStore((s) => s.summary);
  const setSummary = useArticleNotePropsStore((s) => s.setSummary);
  const status = useArticleNotePropsStore((s) => s.status);
  const handleStatusSelect = useArticleNotePropsStore((s) => s.handleStatusSelect);
  const statusOptions = useArticleNotePropsStore((s) => s.statusOptions);
  const categoryId = useArticleNotePropsStore((s) => s.categoryId);
  const setCategoryId = useArticleNotePropsStore((s) => s.setCategoryId);
  const categories = useArticleNotePropsStore((s) => s.categories);
  const setCategories = useArticleNotePropsStore((s) => s.setCategories);
  const tags = useArticleNotePropsStore((s) => s.tags);
  const setTags = useArticleNotePropsStore((s) => s.setTags);
  const selectedTagIds = useArticleNotePropsStore((s) => s.selectedTagIds);
  const setSelectedTagIds = useArticleNotePropsStore((s) => s.setSelectedTagIds);

  const currentCategoryName = useMemo(() => {
    if (categoryId == null) return '';
    return categories.find((c) => c.id === categoryId)?.name ?? '';
  }, [categoryId, categories]);

  const selectedTags = useMemo(
    () => selectedTagIds.map((id) => tags.find((t) => t.id === id)).filter(Boolean),
    [selectedTagIds, tags],
  );

  const statusLabel = useMemo(
    () => statusOptions.find((o) => o.value === status)?.label ?? status,
    [status, statusOptions],
  );

  const onPickCategory = useCallback(
    async (id: number | null) => {
      setCategoryId(id);
      setOpen(false);
    },
    [setCategoryId],
  );

  const onCreateCategory = useCallback(
    async (name: string) => {
      const { id, categories: next } = await resolveOrCreateCategoryByName(name, categories);
      setCategories(next);
      if (id != null) {
        setCategoryId(id);
      }
      setOpen(false);
    },
    [categories, setCategories, setCategoryId],
  );

  const removeTag = useCallback(
    (id: number) => {
      setSelectedTagIds((prev) => prev.filter((x) => x !== id));
    },
    [setSelectedTagIds],
  );

  const addTagFromInput = useCallback(async () => {
    const raw = tagInput.trim();
    if (!raw) return;
    const parts = raw.split(/[,，]/).map((s) => s.trim()).filter(Boolean);
    let nextTags = tags;
    let nextIds = [...selectedTagIds];
    for (const name of parts) {
      const existing = nextTags.find((t) => t.name === name);
      if (existing) {
        if (!nextIds.includes(existing.id)) {
          nextIds.push(existing.id);
        }
        continue;
      }
      const { id, tags: refreshed } = await resolveOrCreateTagByName(name, nextTags);
      nextTags = refreshed;
      setTags(refreshed);
      if (id != null && !nextIds.includes(id)) {
        nextIds.push(id);
      }
    }
    setSelectedTagIds(nextIds);
    setTagInput('');
  }, [tagInput, tags, selectedTagIds, setTags, setSelectedTagIds]);

  const onTagKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      void addTagFromInput();
    }
  };

  if (readOnly) {
    return (
      <div className="beehive-note-properties beehive-note-properties--readonly text-sm">
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="mb-1 flex w-full items-center gap-1 rounded px-0.5 py-1 text-left text-xs font-medium text-muted-foreground hover:text-foreground"
        >
          {expanded ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" />}
          笔记属性
        </button>
        {expanded && (
          <div className="space-y-0.5 border-b border-border/60 pb-2">
            <PropRow icon={AlignLeft} label="title" readOnly>
              <ReadonlyCell>{title.trim() ? title : '—'}</ReadonlyCell>
            </PropRow>
            <PropRow icon={Hash} label="slug" readOnly>
              <ReadonlyCell>{slug.trim() ? slug : '—'}</ReadonlyCell>
            </PropRow>
            <PropRow icon={List} label="categories" readOnly>
              <ReadonlyCell>{currentCategoryName || '无分类'}</ReadonlyCell>
            </PropRow>
            <PropRow icon={AlignLeft} label="status" readOnly>
              <ReadonlyCell>{statusLabel}</ReadonlyCell>
            </PropRow>
            <PropRow icon={TagIcon} label="tags" readOnly>
              <div className="flex flex-wrap gap-1 px-1 py-0.5">
                {selectedTags.length === 0 ? (
                  <span className="text-xs text-muted-foreground">—</span>
                ) : (
                  selectedTags.map((t) =>
                    t ? (
                      <span
                        key={t.id}
                        className="inline-flex rounded-md border border-border bg-muted/50 px-1.5 py-0.5 text-xs text-foreground"
                      >
                        {t.name}
                      </span>
                    ) : null,
                  )
                )}
              </div>
            </PropRow>
            <PropRow icon={AlignLeft} label="summary" readOnly>
              <ReadonlyCell>
                {summary.trim() ? (
                  <span className="whitespace-pre-wrap">{summary}</span>
                ) : (
                  '—'
                )}
              </ReadonlyCell>
            </PropRow>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="beehive-note-properties text-sm">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="mb-1 flex w-full items-center gap-1 rounded px-0.5 py-1 text-left text-xs font-medium text-muted-foreground hover:text-foreground"
      >
        {expanded ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" />}
        笔记属性
      </button>
      {expanded && (
        <div className="space-y-0.5 border-b border-border/60 pb-2">
          <PropRow icon={AlignLeft} label="title">
            <input
              id={`${idBase}-title`}
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded border border-transparent bg-transparent px-1 py-0.5 text-foreground outline-none hover:border-border focus:border-ring focus:ring-1 focus:ring-ring"
            />
          </PropRow>
          <PropRow icon={Hash} label="slug">
            <input
              id={`${idBase}-slug`}
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="url-slug"
              className="w-full rounded border border-transparent bg-transparent px-1 py-0.5 text-foreground outline-none hover:border-border focus:border-ring focus:ring-1 focus:ring-ring"
            />
          </PropRow>
          <PropRow icon={List} label="categories">
            <Popover open={open} onOpenChange={setOpen}>
              <PopoverTrigger asChild>
                <button
                  type="button"
                  className="flex w-full min-w-0 items-center justify-between rounded border border-transparent px-1 py-0.5 text-left text-xs hover:border-border"
                >
                  <span className={cn('truncate', !currentCategoryName && 'text-muted-foreground')}>
                    {currentCategoryName || '无分类'}
                  </span>
                  <ChevronDown className="size-3.5 shrink-0 opacity-50" />
                </button>
              </PopoverTrigger>
              <PopoverContent className="w-[var(--radix-popover-trigger-width)] min-w-[220px] p-0" align="start">
                <Command shouldFilter={false}>
                  <CommandInput
                    placeholder="搜索或创建分类..."
                    onValueChange={setCatSearch}
                  />
                  <CommandList>
                    <CommandGroup heading="分类">
                      <CommandItem value="__none__ 无分类" onSelect={() => void onPickCategory(null)}>
                        无分类
                      </CommandItem>
                      {categories
                        .filter((c) => {
                          const q = catSearch.trim().toLowerCase();
                          if (!q) return true;
                          return (
                            c.name.toLowerCase().includes(q) ||
                            c.slug.toLowerCase().includes(q)
                          );
                        })
                        .map((c) => (
                          <CommandItem
                            key={c.id}
                            value={`${c.name} ${c.slug}`}
                            onSelect={() => void onPickCategory(c.id)}
                          >
                            {c.name}
                          </CommandItem>
                        ))}
                      {(() => {
                        const q = catSearch.trim();
                        const canCreate =
                          q.length > 0 &&
                          !categories.some((c) => c.name === q) &&
                          !categories.some((c) => c.name.toLowerCase() === q.toLowerCase());
                        if (!canCreate) {
                          return null;
                        }
                        return (
                          <CommandItem
                            value={`__create__ ${q}`}
                            onSelect={() => void onCreateCategory(q)}
                          >
                            创建「{q}」
                          </CommandItem>
                        );
                      })()}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </PropRow>
          <PropRow icon={AlignLeft} label="status">
            <div className="max-w-full">
              <CustomSelect
                value={status}
                onChange={handleStatusSelect}
                options={statusOptions}
                className="w-full min-w-0"
                size="sm"
                ariaLabel="文章状态"
                triggerClassName="!h-auto !min-h-7 !border-transparent !bg-transparent !px-1 !py-0.5 !text-xs !shadow-none hover:!bg-muted/40"
              />
            </div>
          </PropRow>
          <PropRow icon={TagIcon} label="tags">
            <div className="flex flex-wrap items-center gap-1">
              {selectedTags.map((t) =>
                t ? (
                  <span
                    key={t.id}
                    className="inline-flex items-center gap-0.5 rounded-md border border-border bg-muted/50 px-1.5 py-0.5 text-xs"
                  >
                    {t.name}
                    <button
                      type="button"
                      className="rounded p-0.5 hover:bg-background"
                      onClick={() => removeTag(t.id)}
                      aria-label={`移除 ${t.name}`}
                    >
                      <X className="size-3" />
                    </button>
                  </span>
                ) : null,
              )}
              <input
                id={`${idBase}-tags`}
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={onTagKeyDown}
                placeholder="输入标签，回车添加"
                className="min-w-[6rem] flex-1 rounded border border-transparent bg-transparent px-1 py-0.5 text-xs outline-none placeholder:text-muted-foreground hover:border-border focus:border-ring"
              />
            </div>
          </PropRow>
          <PropRow icon={AlignLeft} label="summary">
            <textarea
              id={`${idBase}-summary`}
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="摘要（可选）"
              rows={2}
              className="w-full resize-none rounded border border-transparent bg-transparent px-1 py-0.5 text-foreground outline-none hover:border-border focus:border-ring focus:ring-1 focus:ring-ring"
            />
          </PropRow>
        </div>
      )}
    </div>
  );
}
