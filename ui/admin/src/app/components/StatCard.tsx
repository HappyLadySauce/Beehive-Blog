import { LucideIcon } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  trend?: {
    value: string;
    positive: boolean;
  };
  color: 'blue' | 'green' | 'purple' | 'orange';
}

export default function StatCard({ title, value, icon: Icon, trend, color }: StatCardProps) {
  const colorClasses = {
    blue: 'bg-primary/15 text-primary',
    green: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
    purple: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
    orange: 'bg-orange-500/10 text-orange-600 dark:text-orange-400',
  };

  return (
    <div className="bg-card border border-border rounded p-5 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-3">
        <div className={`p-2.5 rounded ${colorClasses[color]}`}>
          <Icon className="w-5 h-5" />
        </div>
      </div>
      <div>
        <p className="text-xs text-muted-foreground mb-1">{title}</p>
        <h3 className="text-2xl font-semibold text-foreground">{value}</h3>
        {trend && (
          <div className="mt-1.5 flex items-center gap-1">
            <span className={`text-xs ${trend.positive ? 'text-green-600' : 'text-red-600'}`}>
              {trend.positive ? '↑' : '↓'} {trend.value}
            </span>
            <span className="text-xs text-muted-foreground">较上月</span>
          </div>
        )}
      </div>
    </div>
  );
}
