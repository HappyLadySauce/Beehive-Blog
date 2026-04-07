import { useState, useEffect, useRef } from 'react';
import { getAttachments, deleteAttachment, uploadAttachment, Attachment } from '../../api/attachment';
import { toast } from 'sonner';
import { Image as ImageIcon, Trash2, Copy, Upload, Download } from 'lucide-react';

export default function Attachments() {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchAttachments = async () => {
    setLoading(true);
    try {
      const res = await getAttachments();
      if (res.code === 200) {
        setAttachments(res.data.items || []);
      } else {
        toast.error(res.message || '获取附件失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求附件失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAttachments();
  }, []);

  const handleDelete = async (id: number) => {
    if (!window.confirm('确定要删除该附件吗？')) return;
    try {
      const res = await deleteAttachment(id);
      if (res.code === 200) {
        toast.success('删除成功');
        fetchAttachments();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '删除请求失败');
    }
  };

  const handleCopyUrl = (url: string) => {
    navigator.clipboard.writeText(url);
    toast.success('链接已复制到剪贴板');
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;

    setUploading(true);
    try {
      for (let i = 0; i < files.length; i++) {
        const res = await uploadAttachment(files[i]);
        if (res.code === 200) {
          toast.success(`文件 ${files[i].name} 上传成功`);
        } else {
          toast.error(res.message || `文件 ${files[i].name} 上传失败`);
        }
      }
      fetchAttachments();
    } catch (error: any) {
      toast.error(error.response?.data?.message || '上传请求失败');
    } finally {
      setUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ImageIcon className="w-5 h-5 text-muted-foreground" />
          <h2 className="text-lg font-medium text-foreground">附件管理</h2>
        </div>
        <div>
          <input 
            type="file" 
            multiple 
            className="hidden" 
            ref={fileInputRef} 
            onChange={handleFileChange} 
          />
          <button 
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <Upload className="w-4 h-4" />
            {uploading ? '上传中...' : '上传附件'}
          </button>
        </div>
      </div>

      {loading ? (
        <div className="py-12 text-center text-muted-foreground">加载中...</div>
      ) : attachments.length === 0 ? (
        <div className="py-12 text-center text-muted-foreground bg-card border border-border rounded">暂无附件</div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {attachments.map((attachment) => (
            <div key={attachment.id} className="bg-card border border-border rounded overflow-hidden group relative">
              <div className="aspect-square bg-muted flex items-center justify-center overflow-hidden">
                {attachment.type === 'image' ? (
                  <img src={attachment.thumbUrl || attachment.url} alt={attachment.name} className="w-full h-full object-cover" />
                ) : (
                  <div className="text-muted-foreground flex flex-col items-center">
                    <Download className="w-8 h-8 mb-2" />
                    <span className="text-xs uppercase">{attachment.type}</span>
                  </div>
                )}
              </div>
              <div className="p-2 border-t border-border">
                <div className="text-xs font-medium text-foreground truncate" title={attachment.originalName}>
                  {attachment.originalName}
                </div>
                <div className="text-[10px] text-muted-foreground mt-0.5">
                  {(attachment.size / 1024).toFixed(1)} KB
                </div>
              </div>
              
              {/* Hover Actions */}
              <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                <button 
                  onClick={() => handleCopyUrl(attachment.url)}
                  className="p-2 bg-card text-foreground rounded-full hover:text-primary transition-colors"
                  title="复制链接"
                >
                  <Copy className="w-4 h-4" />
                </button>
                <button 
                  onClick={() => handleDelete(attachment.id)}
                  className="p-2 bg-card text-foreground rounded-full hover:text-red-600 transition-colors"
                  title="删除"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
