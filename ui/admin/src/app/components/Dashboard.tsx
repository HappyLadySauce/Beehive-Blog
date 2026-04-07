import { FileText, MessageSquare, Eye, Users, LayoutDashboard } from 'lucide-react';
import StatCard from './StatCard';
import { useEffect, useState } from 'react';
import { getSiteStats, SiteStatsResponse } from '../../api/site';
import { toast } from 'sonner';

export default function Dashboard() {
  const [statsData, setStatsData] = useState<SiteStatsResponse | null>(null);

  const fetchStats = async () => {
    try {
      const res = await getSiteStats();
      if (res.code === 200) {
        setStatsData(res.data);
      } else {
        toast.error(res.message || '获取统计数据失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求统计数据失败');
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

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <LayoutDashboard className="w-5 h-5 text-gray-600" />
        <h2 className="text-lg font-medium text-gray-900">控制台概览</h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      <div className="bg-white border border-gray-200 rounded">
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900">热门文章 Top 10</h3>
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
