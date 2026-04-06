import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Editor } from '@bytemd/react';
import gfm from '@bytemd/plugin-gfm';
import 'bytemd/dist/index.css';
import request from '../../utils/request';
import { toast } from 'sonner';
import { Save, ArrowLeft } from 'lucide-react';

const plugins = [gfm()];

export default function ArticleEdit() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('draft');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (id) {
      // Fetch article details
      const fetchArticle = async () => {
        try {
          const res = await request.get<any, any>(`/api/v1/admin/articles/${id}`);
          if (res.code === 200) {
            setTitle(res.data.title);
            setContent(res.data.content);
            setStatus(res.data.status);
          } else {
            toast.error(res.message || '获取文章失败');
          }
        } catch (error: any) {
          toast.error(error.response?.data?.message || '请求文章失败');
        }
      };
      fetchArticle();
    }
  }, [id]);

  const handleSave = async () => {
    if (!title) {
      toast.error('请输入文章标题');
      return;
    }
    
    setLoading(true);
    try {
      const payload = {
        title,
        content,
        status,
      };

      let res;
      if (id) {
        res = await request.put<any, any>(`/api/v1/admin/articles/${id}`, payload);
      } else {
        res = await request.post<any, any>('/api/v1/admin/articles', payload);
      }

      if (res.code === 200) {
        toast.success(id ? '更新成功' : '发布成功');
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
    const res = await Promise.all(
      files.map(async (file) => {
        const formData = new FormData();
        formData.append('file', file);
        try {
          const response = await request.post<any, any>('/api/v1/admin/upload-image', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
          if (response.code === 200) {
            return {
              url: response.data.url,
              alt: response.data.alt || file.name,
            };
          }
        } catch (error) {
          console.error('Upload failed', error);
        }
        return null;
      })
    );
    return res.filter(Boolean) as { url: string; alt: string }[];
  };

  return (
    <div className="h-full flex flex-col space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button 
            onClick={() => navigate('/articles')}
            className="p-1.5 text-gray-600 hover:bg-gray-100 rounded transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <h2 className="text-lg font-medium text-gray-900">{id ? '编辑文章' : '新建文章'}</h2>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="draft">草稿</option>
            <option value="published">发布</option>
          </select>
          <button 
            onClick={handleSave}
            disabled={loading}
            className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <Save className="w-4 h-4" />
            {loading ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded p-4">
        <input
          type="text"
          placeholder="输入文章标题..."
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full text-2xl font-medium border-none focus:ring-0 p-0 placeholder-gray-300"
        />
      </div>

      <div className="flex-1 bg-white border border-gray-200 rounded overflow-hidden editor-container">
        <Editor
          value={content}
          plugins={plugins}
          onChange={(v) => setContent(v)}
          uploadImages={uploadImages}
        />
      </div>
      
      <style>{`
        .editor-container .bytemd {
          height: calc(100vh - 250px);
          border: none;
        }
      `}</style>
    </div>
  );
}
