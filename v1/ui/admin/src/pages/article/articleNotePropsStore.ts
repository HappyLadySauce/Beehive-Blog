import { create } from 'zustand';
import type { CategoryBrief, TagListItem } from '../../api/taxonomy';

export type StatusOption = { value: string; label: string };

/** 供注入到 ByteMD 编辑/预览栏的笔记属性（与 edit.tsx state 同步） */
export interface ArticleNotePropsState {
  showNoteProperties: boolean;
  title: string;
  setTitle: (v: string) => void;
  slug: string;
  setSlug: (v: string) => void;
  summary: string;
  setSummary: (v: string) => void;
  status: string;
  handleStatusSelect: (v: string) => void;
  statusOptions: StatusOption[];
  categoryId: number | null;
  setCategoryId: (v: number | null) => void;
  categories: CategoryBrief[];
  setCategories: (list: CategoryBrief[]) => void;
  tags: TagListItem[];
  setTags: (list: TagListItem[]) => void;
  selectedTagIds: number[];
  setSelectedTagIds: (ids: number[] | ((prev: number[]) => number[])) => void;
  toggleTag: (id: number) => void;
}

const noop = () => {};

export const useArticleNotePropsStore = create<ArticleNotePropsState>(() => ({
  showNoteProperties: true,
  title: '',
  setTitle: noop,
  slug: '',
  setSlug: noop,
  summary: '',
  setSummary: noop,
  status: 'draft',
  handleStatusSelect: noop,
  statusOptions: [],
  categoryId: null,
  setCategoryId: noop,
  categories: [],
  setCategories: noop,
  tags: [],
  setTags: noop,
  selectedTagIds: [],
  setSelectedTagIds: noop,
  toggleTag: noop,
}));
