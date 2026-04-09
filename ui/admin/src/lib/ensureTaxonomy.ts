import { getCategories, getTags, createCategory, createTag } from '../api/taxonomy';
import type { CategoryBrief, TagListItem } from '../api/taxonomy';
import { toast } from 'sonner';

/** 保存前拉取最新分类、标签列表（用于 FM 名称与下拉数据一致）。 */
export async function fetchLatestTaxonomy(): Promise<{
  categories: CategoryBrief[];
  tags: TagListItem[];
}> {
  const [catRes, tagRes] = await Promise.all([
    getCategories({ pageSize: 200 }),
    getTags({ pageSize: 200 }),
  ]);
  return {
    categories: catRes.code === 200 ? catRes.data.list || [] : [],
    tags: tagRes.code === 200 ? tagRes.data.list || [] : [],
  };
}

/**
 * 将当前选择的分类/标签 id 与最新列表对齐（去掉已删除项，避免无效 id 提交）。
 */
export function alignTaxonomyIds(
  categoryId: number | null,
  tagIds: number[],
  categories: CategoryBrief[],
  tags: TagListItem[],
): { categoryId: number | null; tagIds: number[] } {
  let c = categoryId;
  if (c != null && !categories.some((x) => x.id === c)) {
    c = null;
  }
  const t = tagIds.filter((id) => tags.some((x) => x.id === id));
  return { categoryId: c, tagIds: t };
}

/**
 * 按名称创建分类（若已存在则返回已有 id）。
 */
export async function resolveOrCreateCategoryByName(
  name: string,
  categories: CategoryBrief[],
): Promise<{ id: number | null; categories: CategoryBrief[] }> {
  const trimmed = name.trim();
  if (!trimmed) {
    return { id: null, categories };
  }
  const found = categories.find((c) => c.name === trimmed);
  if (found) {
    return { id: found.id, categories };
  }
  const res = await createCategory({ name: trimmed });
  if (res.code !== 200 || !res.data) {
    toast.error(res.message || '创建分类失败');
    return { id: null, categories };
  }
  const catRes = await getCategories({ pageSize: 200 });
  const next = catRes.code === 200 ? catRes.data.list || [] : [...categories, res.data];
  return { id: res.data.id, categories: next };
}

/**
 * 按名称创建标签（若已存在则返回已有 id）。
 */
export async function resolveOrCreateTagByName(
  name: string,
  tags: TagListItem[],
): Promise<{ id: number | null; tags: TagListItem[] }> {
  const trimmed = name.trim();
  if (!trimmed) {
    return { id: null, tags };
  }
  const found = tags.find((t) => t.name === trimmed);
  if (found) {
    return { id: found.id, tags };
  }
  const res = await createTag({ name: trimmed });
  if (res.code !== 200 || !res.data) {
    toast.error(res.message || '创建标签失败');
    return { id: null, tags };
  }
  const tagRes = await getTags({ pageSize: 200 });
  const next = tagRes.code === 200 ? tagRes.data.list || [] : [...tags, res.data];
  return { id: res.data.id, tags: next };
}
