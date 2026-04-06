import request from '../utils/request';

export interface ArticleStatItem {
  id: number;
  title: string;
  viewCount: number;
  likeCount: number;
}

export interface SiteStatsResponse {
  articleCount: number;
  userCount: number;
  commentCount: number;
  todayViews: number;
  topArticles: ArticleStatItem[];
}

export const getSiteStats = () => {
  return request.get<any, { code: number; message: string; data: SiteStatsResponse }>('/api/v1/admin/stats');
};
