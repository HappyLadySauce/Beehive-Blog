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
    <div className="space-y-[clamp(1rem,0.8vw,1.5rem)]">
      <div className="flex items-center gap-3">
        <LayoutDashboard className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
        <h2 className="text-xl font-semibold tracking-tight text-foreground">控制台概览</h2>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      <div className="admin-card admin-card-glass rounded border">
        <div className="border-b border-border p-[clamp(0.75rem,0.5vw+0.55rem,1rem)]">
          <div className="flex items-center justify-between">
            <h3 className="text-[clamp(0.86rem,0.13vw+0.82rem,0.96rem)] font-medium text-foreground">热门文章 Top 10</h3>
          </div>
        </div>
        <div className="p-[clamp(0.75rem,0.5vw+0.55rem,1rem)]">
          <div className="space-y-3">
            {statsData?.topArticles?.length ? (
              statsData.topArticles.map((article) => (
                <div key={article.id} className="flex items-start gap-3 rounded border border-border p-[clamp(0.7rem,0.35vw+0.6rem,0.9rem)] hover:bg-muted/50 transition-colors">
                  <div className="flex-shrink-0 w-2 h-2 mt-1.5 rounded-full bg-primary"></div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm text-foreground truncate">{article.title}</span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span>{article.viewCount} 次浏览</span>
                      <span>• {article.likeCount} 次点赞</span>
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground text-sm">暂无热门文章</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
