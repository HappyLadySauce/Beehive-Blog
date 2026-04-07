import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Editor } from '@bytemd/react';
import gfm from '@bytemd/plugin-gfm';
import 'bytemd/dist/index.css';
import { getArticle, createArticle, updateArticle } from '../../api/article';
import { getCategories, getTags, CategoryBrief, TagListItem } from '../../api/taxonomy';
import request from '../../utils/request';
import { toast } from 'sonner';
import { ArrowLeft, Save } from 'lucide-react';

const plugins = [gfm()];

export default function ArticleEdit() {
  const { id } = useParams();
  const navigate = useNavigate();
  const articleId = id ? parseInt(id, 10) : undefined;

  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [summary, setSummary] = useState('');
  const [status, setStatus] = useState('draft');
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<CategoryBrief[]>([]);
  const [tags, setTags] = useState<TagListItem[]>([]);

  useEffect(() => {
    const loadFilters = async () => {
      try {
        const [catRes, tagRes] = await Promise.all([
          getCategories({ pageSize: 200 }),
          getTags({ pageSize: 200 }),
        ]);
        if (catRes.code === 200) setCategories(catRes.data.list || []);
        if (tagRes.code === 200) setTags(tagRes.data.list || []);
      } catch {
        // 筛选器加载失败不阻断主流程
      }
    };
    loadFilters();
  }, []);

  useEffect(() => {
    if (!articleId) return;
    const fetchArticle = async () => {
      try {
        const res = await getArticle(articleId);
        if (res.code === 200) {
          const data = res.data;
          setTitle(data.title);
          setContent(data.content);
          setSummary(data.summary || '');
          setStatus(data.status);
          setCategoryId(data.category?.id ?? null);
          setSelectedTagIds(data.tags?.map(t => t.id) || []);
        } else {
          toast.error(res.message || '获取文章失败');
        }
      } catch (error: any) {
        toast.error(error.response?.data?.message || '请求文章失败');
      }
    };
    fetchArticle();
  }, [articleId]);

  const handleSave = async () => {
    if (!title.trim()) {
      toast.error('请输入文章标题');
      return;
    }
    if (!content.trim()) {
      toast.error('请输入文章内容');
      return;
    }

    setLoading(true);
    try {
      const payload = {
        title: title.trim(),
        content,
        summary: summary.trim() || undefined,
        status,
        categoryId: categoryId ?? undefined,
        tagIds: selectedTagIds.length > 0 ? selectedTagIds : undefined,
      };

      let res;
      if (articleId) {
        res = await updateArticle(articleId, payload);
      } else {
        res = await createArticle(payload);
      }

      if (res.code === 200) {
        toast.success(articleId ? '更新成功' : '发布成功');
        navigate('/articles');
      } else {
        toast.error(res.message || '保存失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '保存请求失败');
    } finally {
      setLoading(false);
    }
  };

  const uploadImages = async (files: File[]) => {
    const results = await Promise.all(
      files.map(async (file) => {
        const formData = new FormData();
        formData.append('file', file);
        try {
          const response = await request.post<any, any>('/api/v1/admin/upload-image', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
          if (response.code === 200) {
            return { url: response.data.url, alt: response.data.alt || file.name };
          }
        } catch (error) {
          console.error('Upload failed', error);
        }
        return null;
      })
    );
    return results.filter(Boolean) as { url: string; alt: string }[];
  };

  const toggleTag = (tagId: number) => {
    setSelectedTagIds(prev =>
      prev.includes(tagId) ? prev.filter(id => id !== tagId) : [...prev, tagId]
    );
  };

  return (
    <div className="h-full flex flex-col space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/articles')}
            className="p-1.5 text-muted-foreground hover:bg-accent rounded transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <h2 className="text-lg font-medium text-foreground">{articleId ? '编辑文章' : '新建文章'}</h2>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="px-3 py-1.5 text-sm border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            <option value="draft">草稿</option>
            <option value="published">发布</option>
            <option value="private">私密</option>
            <option value="archived">归档</option>
          </select>
          <button
            onClick={handleSave}
            disabled={loading}
            className="px-4 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <Save className="w-4 h-4" />
            {loading ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="bg-card border border-border rounded p-4">
        <input
          type="text"
          placeholder="输入文章标题..."
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full text-2xl font-medium border-none focus:ring-0 p-0 bg-transparent text-foreground placeholder:text-muted-foreground outline-none"
        />
      </div>

      <div className="flex gap-4">
        <div className="flex-1 bg-card border border-border rounded overflow-hidden editor-container">
          <Editor
            value={content}
            plugins={plugins}
            onChange={(v) => setContent(v)}
            uploadImages={uploadImages}
          />
        </div>

        <div className="w-64 flex-shrink-0 space-y-4">
          <div className="bg-card border border-border rounded p-4 space-y-3">
            <h3 className="text-sm font-medium text-foreground">分类</h3>
            <select
              value={categoryId ?? ''}
              onChange={(e) => setCategoryId(e.target.value ? parseInt(e.target.value, 10) : null)}
              className="w-full px-3 py-2 text-sm border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent"
            >
              <option value="">无分类</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>{cat.name}</option>
              ))}
            </select>
          </div>

          <div className="bg-card border border-border rounded p-4 space-y-3">
            <h3 className="text-sm font-medium text-foreground">标签</h3>
            <div className="flex flex-wrap gap-2 max-h-48 overflow-y-auto">
              {tags.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => toggleTag(tag.id)}
                  className={`px-2 py-1 text-xs rounded border transition-colors ${
                    selectedTagIds.includes(tag.id)
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'bg-card text-foreground border-border hover:border-primary/50'
                  }`}
                >
                  {tag.name}
                </button>
              ))}
              {tags.length === 0 && (
                <span className="text-xs text-muted-foreground">暂无标签</span>
              )}
            </div>
          </div>

          <div className="bg-card border border-border rounded p-4 space-y-3">
            <h3 className="text-sm font-medium text-foreground">摘要</h3>
            <textarea
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="可选，留空则自动截取正文..."
              rows={4}
              className="w-full px-3 py-2 text-sm border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent resize-none"
            />
          </div>
        </div>
      </div>

      <style>{`
        .editor-container .bytemd {
          height: calc(100vh - 280px);
          border: none;
        }
      `}</style>
    </div>
  );
}
