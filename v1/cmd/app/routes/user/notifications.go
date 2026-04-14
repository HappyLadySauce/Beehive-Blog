package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"k8s.io/klog/v2"
)

const (
	defaultNotificationPageSize = 20
	maxNotificationPageSize     = 100
)

// ListNotifications returns paginated notifications for the current user.
func (s *UserService) ListNotifications(ctx context.Context, userID int64, page, pageSize int, isRead *bool) (*v1.NotificationListResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultNotificationPageSize
	}
	if pageSize > maxNotificationPageSize {
		pageSize = maxNotificationPageSize
	}
	offset := (page - 1) * pageSize

	base := s.svc.DB.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)
	if isRead != nil {
		base = base.Where("is_read = ?", *isRead)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		klog.ErrorS(err, "Failed to count notifications", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	q := s.svc.DB.WithContext(ctx).Where("user_id = ?", userID)
	if isRead != nil {
		q = q.Where("is_read = ?", *isRead)
	}
	var rows []models.Notification
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "Failed to list notifications", "userID", userID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.NotificationItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, v1.NotificationItem{
			ID:         r.ID,
			Type:       string(r.Type),
			Title:      r.Title,
			Content:    r.Content,
			IsRead:     r.IsRead,
			SourceID:   r.SourceID,
			SourceType: r.SourceType,
			CreatedAt:  r.CreatedAt,
			ReadAt:     r.ReadAt,
		})
	}
	return &v1.NotificationListResponse{Items: items, Total: total}, http.StatusOK, nil
}
