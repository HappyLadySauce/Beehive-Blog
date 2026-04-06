import { FileText, MessageSquare, Eye, ThumbsUp, TrendingUp, Clock, LayoutDashboard, FolderOpen, Image, Settings, Users } from 'lucide-react';
import StatCard from './StatCard';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useEffect, useState } from 'react';
import { getSiteStats, SiteStatsResponse } from '../../api/site';
import { toast } from 'sonner';

export default function Dashboard() {
  const [statsData, setStatsData] = useState<SiteStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchStats = async () => {
    setLoading(true);
    try {
      const res = await getSiteStats();
      if (res.code === 200) {
        setStatsData(res.data);
      } else {
        toast.error(res.message || '获取统计数据失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求统计数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const stats = [
    { title: '文章', value: statsData?.articleCount.toString() || '0', icon: FileText, color: 'blue' as const },
    { title: '用户', value: statsData?.userCount.toString() || '0', icon: Users, color: 'green' as const },
    { title: '今日访问', value: statsData?.todayViews.toString() || '0', icon: Eye, color: 'purple' as const },
    { title: '评论', value: statsData?.commentCount.toString() || '0', icon: MessageSquare, color: 'orange' as const },
  ];

  const viewsData = [
    { date: '3/1', views: 1200 },
    { date: '3/5', views: 1800 },
    { date: '3/10', views: 1600 },
    { date: '3/15', views: 2400 },
    { date: '3/20', views: 2200 },
    { date: '3/25', views: 2800 },
    { date: '3/30', views: 3200 },
    { date: '4/5', views: 3600 },
  ];

  const recentArticles = [
    { id: 1, title: 'React 18 新特性详解', status: 'published', views: 1234, date: '2026-04-05' },
    { id: 2, title: 'TypeScript 最佳实践', status: 'published', views: 987, date: '2026-04-04' },
    { id: 3, title: 'Tailwind CSS 高级技巧', status: 'draft', views: 0, date: '2026-04-03' },
    { id: 4, title: 'Node.js 性能优化指南', status: 'published', views: 2341, date: '2026-04-02' },
    { id: 5, title: 'Web 安全最佳实践', status: 'scheduled', views: 0, date: '2026-04-10' },
  ];

  const statusColors = {
    published: 'bg-green-100 text-green-800',
    draft: 'bg-gray-100 text-gray-800',
    scheduled: 'bg-blue-100 text-blue-800',
  };

  const statusLabels = {
    published: '已发布',
    draft: '草稿',
    scheduled: '定时发布',
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <LayoutDashboard className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">控制台概览</h2>
        </div>
        <div className="flex items-center gap-2">
          <button 
            onClick={fetchStats}
            disabled={loading}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors disabled:opacity-50"
          >
            {loading ? '刷新中...' : '刷新'}
          </button>
          <button className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
            新建
          </button>
        </div>
      </div>

      {/* 统计卡片 - Box 布局 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* 快捷操作区域 - Box 布局 */}
      <div className="bg-white border border-gray-200 rounded">
        <div className="p-4 border-b border-gray-200">
          <h3 className="text-sm font-medium text-gray-900">快捷访问</h3>
        </div>
        <div className="p-4">
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-green-100 text-green-600 rounded flex items-center justify-center group-hover:bg-green-200">
                <FileText className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">文章生成</div>
            </button>
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-blue-100 text-blue-600 rounded flex items-center justify-center group-hover:bg-blue-200">
                <Eye className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">代码生成</div>
            </button>
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-purple-100 text-purple-600 rounded flex items-center justify-center group-hover:bg-purple-200">
                <Settings className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">主题管理</div>
            </button>
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-orange-100 text-orange-600 rounded flex items-center justify-center group-hover:bg-orange-200">
                <MessageSquare className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">评论管理</div>
            </button>
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-pink-100 text-pink-600 rounded flex items-center justify-center group-hover:bg-pink-200">
                <ThumbsUp className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">统计管理</div>
            </button>
            <button className="p-4 border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-all text-center group">
              <div className="w-10 h-10 mx-auto mb-2 bg-teal-100 text-teal-600 rounded flex items-center justify-center group-hover:bg-teal-200">
                <Image className="w-5 h-5" />
              </div>
              <div className="text-xs text-gray-700">附件管理</div>
            </button>
          </div>
        </div>
      </div>

      {/* 热门文章 - Box 布局 */}
      <div className="bg-white border border-gray-200 rounded">
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900">热门文章 Top 10</h3>
            <button className="text-xs text-blue-600 hover:text-blue-700">查看更多</button>
          </div>
        </div>
        <div className="p-4">
          <div className="space-y-3">
            {statsData?.topArticles?.length ? (
              statsData.topArticles.map((article) => (
                <div key={article.id} className="flex items-start gap-3 p-3 border border-gray-100 rounded hover:bg-gray-50 transition-colors">
                  <div className="flex-shrink-0 w-2 h-2 mt-1.5 rounded-full bg-blue-500"></div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm text-gray-900 truncate">{article.title}</span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-500">
                      <span>{article.viewCount} 次浏览</span>
                      <span>• {article.likeCount} 次点赞</span>
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-gray-500 text-sm">暂无热门文章</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
