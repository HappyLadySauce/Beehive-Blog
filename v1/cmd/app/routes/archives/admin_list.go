package archives

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/articlequery"
	"k8s.io/klog/v2"
)

func parseAdminStatusFilter(raw string) ([]models.ArticleStatus, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	valid := map[string]models.ArticleStatus{
		"draft":     models.ArticleStatusDraft,
		"published": models.ArticleStatusPublished,
		"archived":  models.ArticleStatusArchived,
		"private":   models.ArticleStatusPrivate,
		"scheduled": models.ArticleStatusScheduled,
	}
	var out []models.ArticleStatus
	seen := make(map[models.ArticleStatus]struct{})
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		st, ok := valid[p]
		if !ok {
			return nil, errors.New("invalid status filter")
		}
		if _, dup := seen[st]; dup {
			continue
		}
		seen[st] = struct{}{}
		out = append(out, st)
	}
	return out, nil
}

// AdminListArticles 管理员文章分页列表（含草稿等）。
func (a *ArticleAdmin) AdminListArticles(ctx context.Context, req *v1.AdminArticleListRequest) (*v1.AdminArticleListResponse, int, error) {
	if req == nil {
		req = &v1.AdminArticleListRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	statuses, err := parseAdminStatusFilter(req.Status)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid status filter")
	}

	db := a.svc.DB.WithContext(ctx)
	q := articlequery.AdminArticleQuery(db, req.Keyword, req.Category, req.Author, req.Tag, statuses)
	rows, total, err := articlequery.ListAdminPage(ctx, db, q, page, pageSize, req.Sort)
	if err != nil {
		klog.ErrorS(err, "AdminListArticles query failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	list := make([]v1.AdminArticleListItem, 0, len(rows))
	for i := range rows {
		item := articlequery.MapListItem(&rows[i])
		list = append(list, v1.AdminArticleListItem{
			ArticleListItem: item,
			Status:          string(rows[i].Status),
		})
	}
	return &v1.AdminArticleListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}
